package logger

import "time"

// LogEvent adalah struktur standar untuk setiap entri log.
// Menggunakan format JSON yang ramah untuk Elastic/SIEM.
type LogEvent struct {
	Timestamp   time.Time `json:"@timestamp"`
	RequestID   string    `json:"request_id"`
	Domain      string    `json:"domain"`
	Path        string    `json:"path"`
	Method      string    `json:"method"`
	RemoteAddr  string    `json:"remote_addr"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	MimeType    string    `json:"mime_type"`
	UploadPath  string    `json:"upload_path"`
	SourceField string    `json:"source_field"`
}

// Logger adalah interface umum untuk semua implementasi logger.
type Logger interface {
	Log(event LogEvent)
}
