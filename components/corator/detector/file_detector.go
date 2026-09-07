package detector

import (
	"io"
	"mime/multipart"
	"net/http"
	"strings"
)

// FileDetector adalah implementasi untuk mendeteksi file dari multipart/form-data.
type FileDetector struct{}

// NewFileDetector membuat instance baru dari FileDetector.
func NewFileDetector() *FileDetector {
	return &FileDetector{}
}

// Detect memeriksa request untuk file upload.
func (d *FileDetector) Detect(req *http.Request) ([]DetectionResult, error) {
	// Cek apakah requestnya adalah multipart/form-data
	contentType := req.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "multipart/form-data") {
		return nil, nil // Bukan file upload, tidak ada yang dideteksi.
	}

	var results []DetectionResult

	// Batasi ukuran form di memori (misal: 32MB), sisanya akan ditulis ke disk sementara.
	if err := req.ParseMultipartForm(32 << 20); err != nil {
		// Jika error karena body terlalu besar, kita bisa abaikan, WAF mungkin akan menanganinya.
		// Jika error lain, kita kembalikan errornya.
		if err != http.ErrNotMultipart {
			return nil, err
		}
		return nil, nil
	}

	// Jika tidak ada form, tidak ada yang perlu diperiksa
	if req.MultipartForm == nil || req.MultipartForm.File == nil {
		return nil, nil
	}

	// Iterasi semua file yang di-upload
	for fieldName, fileHeaders := range req.MultipartForm.File {
		for _, fileHeader := range fileHeaders {
			result, err := d.processFile(fieldName, fileHeader)
			if err != nil {
				// Mungkin logging di sini lebih baik daripada menghentikan proses
				continue
			}
			results = append(results, result)
		}
	}

	return results, nil
}

// processFile memproses satu file header menjadi DetectionResult.
func (d *FileDetector) processFile(fieldName string, fileHeader *multipart.FileHeader) (DetectionResult, error) {
	file, err := fileHeader.Open()
	if err != nil {
		return DetectionResult{}, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return DetectionResult{}, err
	}

	return DetectionResult{
		Data:        data,
		FileName:    fileHeader.Filename,
		SourceField: "multipart-field:" + fieldName,
		MimeType:    fileHeader.Header.Get("Content-Type"),
	}, nil
}
