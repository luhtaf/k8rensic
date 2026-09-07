package detector

import (
	"encoding/base64"
	"fmt"
	"mime"
	"net/http"
	"regexp"
	"strings"
)

// Base64Detector adalah implementasi untuk mendeteksi file dari string Base64.
type Base64Detector struct {
	// Regex untuk mencari kandidat string base64.
	// Minimal 20 karakter, hanya berisi karakter base64, dan mungkin punya padding '='.
	base64Regex *regexp.Regexp
}

// NewBase64Detector membuat instance baru dari Base64Detector.
func NewBase64Detector() *Base64Detector {
	return &Base64Detector{
		base64Regex: regexp.MustCompile(`^(?:[A-Za-z0-9+/]{4}){5,}(?:[A-Za-z0-9+/]{2}==|[A-Za-z0-9+/]{3}=)?$`),
	}
}

// Detect memeriksa form values dan query params untuk string Base64.
func (d *Base64Detector) Detect(req *http.Request) ([]DetectionResult, error) {
	var results []DetectionResult

	// Parse form agar kita bisa mengakses query dan body values
	if err := req.ParseForm(); err != nil {
		return nil, fmt.Errorf("gagal parse form: %w", err)
	}

	// Iterasi melalui semua nilai di form (query + body)
	for fieldName, values := range req.Form {
		for _, value := range values {
			// Cek apakah value ini adalah kandidat base64
			if !d.base64Regex.MatchString(value) {
				continue
			}

			// Coba decode string base64
			decodedData, err := base64.StdEncoding.DecodeString(value)
			if err != nil {
				continue // Bukan base64 yang valid, abaikan.
			}

			// Cek tipe konten dari data yang sudah di-decode
			mimeType := http.DetectContentType(decodedData)

			// Abaikan jika hanya teks biasa
			if strings.HasPrefix(mimeType, "text/plain") {
				continue
			}

			// Buat nama file yang unik dan dapat dilacak
			exts, _ := mime.ExtensionsByType(mimeType)
			ext := ".bin" // default extension
			if len(exts) > 0 {
				ext = exts[0]
			}
			fileName := fmt.Sprintf("base64_%s%s", fieldName, ext)

			results = append(results, DetectionResult{
				Data:        decodedData,
				FileName:    fileName,
				SourceField: "form-field:" + fieldName,
				MimeType:    mimeType,
			})
		}
	}

	return results, nil
}
