package detector

import "net/http"

// DetectionResult menampung informasi tentang file yang ditemukan.
type DetectionResult struct {
	Data        []byte // Konten file setelah di-decode
	FileName    string // Nama file asli atau hasil generate
	SourceField string // Field tempat file ditemukan (e.g., "form-field:user_avatar")
	MimeType    string // Tipe MIME dari data
}

// Detector adalah interface umum untuk semua implementasi detektor.
type Detector interface {
	Detect(req *http.Request) ([]DetectionResult, error)
}
