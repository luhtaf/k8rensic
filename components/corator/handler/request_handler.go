package handler

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/corazawaf/coraza/v3"
	"github.com/google/uuid"
	"github.com/luhtaf/corator/detector"
	"github.com/luhtaf/corator/logger"
	"github.com/luhtaf/corator/uploader"
)

// RequestHandler adalah middleware utama yang mengatur alur request.
type RequestHandler struct {
	WAF       coraza.WAF
	Detectors []detector.Detector
	Uploader  uploader.Uploader
	Loggers   []logger.Logger
	Backend   *url.URL
}

// NewRequestHandler membuat instance baru dari RequestHandler.
func NewRequestHandler(waf coraza.WAF, detectors []detector.Detector, up uploader.Uploader, logs []logger.Logger, backendURL *url.URL) *RequestHandler {
	return &RequestHandler{
		WAF:       waf,
		Detectors: detectors,
		Uploader:  up,
		Loggers:   logs,
		Backend:   backendURL,
	}
}

// ServeHTTP adalah metode yang membuat RequestHandler menjadi http.Handler.
func (rh *RequestHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// 1. Generate Request ID unik
	requestID := uuid.New().String()

	// 2. Clone body request agar bisa dibaca berkali-kali
	bodyBytes, _ := io.ReadAll(req.Body)
	req.Body.Close() // Tutup body asli

	// Buat reader baru untuk aplikasi kita (detector, WAF)
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	// 3. Jalankan semua detektor
	var allResults []detector.DetectionResult
	for _, d := range rh.Detectors {
		results, err := d.Detect(req)
		if err != nil {
			log.Printf("[%s] Error saat deteksi: %v", requestID, err)
			continue
		}
		if len(results) > 0 {
			allResults = append(allResults, results...)
		}
	}

	// 4. Proses hasil deteksi secara asinkron
	if len(allResults) > 0 {
		rh.processDetections(req, requestID, allResults)
	}

	// 5. Jalankan Coraza WAF
	tx := rh.WAF.NewTransaction()
	defer func() {
		tx.ProcessLogging()
		tx.Close()
	}()

	// Ambil IP dan Port client
	clientIP, clientPortStr, _ := strings.Cut(req.RemoteAddr, ":")
	clientPort, _ := strconv.Atoi(clientPortStr)

	// KOREKSI: Gunakan 0 untuk port yang tidak diketahui, bukan ""
	tx.ProcessConnection(clientIP, clientPort, "", 0)
	tx.ProcessURI(req.Method, req.URL.String(), req.Proto)

	// Menambahkan semua header request ke transaksi
	for k, vv := range req.Header {
		for _, v := range vv {
			tx.AddRequestHeader(k, v)
		}
	}

	// Proses header request
	tx.ProcessRequestHeaders()

	// Kita hanya proses body jika Coraza belum terinterupsi oleh header.
	if !tx.IsInterrupted() {
		tx.ProcessRequestBody()
	}

	// Cek apakah ada interupsi setelah semua proses
	if tx.IsInterrupted() {
		log.Printf("[%s] Request diblokir oleh WAF", requestID)
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("Request diblokir oleh WAF"))
		return
	}

	// 6. Teruskan request ke backend
	proxy := httputil.NewSingleHostReverseProxy(rh.Backend)
	// Berikan body yang masih fresh ke proxy
	req.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	proxy.ServeHTTP(w, req)
}

// processDetections menjalankan upload dan logging dalam goroutine.
func (rh *RequestHandler) processDetections(req *http.Request, requestID string, results []detector.DetectionResult) {
	for _, result := range results {
		go func(res detector.DetectionResult) {
			uniqueFilename := fmt.Sprintf("%s_%s", requestID, res.FileName)

			// Upload file
			uploadPath, err := rh.Uploader.Upload(context.Background(), bytes.NewReader(res.Data), uniqueFilename)
			if err != nil {
				log.Printf("[%s] Gagal upload file %s: %v", requestID, res.FileName, err)
				return
			}

			// Buat event log
			event := logger.LogEvent{
				Timestamp:   time.Now(),
				RequestID:   requestID,
				Domain:      req.Host,
				Path:        req.URL.Path,
				Method:      req.Method,
				RemoteAddr:  req.RemoteAddr,
				FileName:    res.FileName,
				FileSize:    int64(len(res.Data)),
				MimeType:    res.MimeType,
				UploadPath:  uploadPath,
				SourceField: res.SourceField,
			}

			// Kirim ke semua logger aktif
			for _, l := range rh.Loggers {
				l.Log(event)
			}
			log.Printf("[%s] File terdeteksi dan diunggah: %s dari field %s", requestID, uploadPath, res.SourceField)

		}(result)
	}
}
