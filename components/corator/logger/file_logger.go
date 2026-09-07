package logger

import (
	"os"

	"github.com/luhtaf/corator/config" // Nama package diganti sesuai modul Anda
	"github.com/rs/zerolog"
)

// FileLogger adalah implementasi logger yang menulis ke file lokal.
type FileLogger struct {
	logger zerolog.Logger
}

// NewFileLogger membuat instance baru dari FileLogger.
func NewFileLogger(cfg config.FileLoggerConfig) (*FileLogger, error) {
	file, err := os.OpenFile(
		cfg.Path,
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	if err != nil {
		return nil, err
	}

	// Menggunakan timestamp Unix dan pesan default dari zerolog
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(file).With().Timestamp().Logger()

	return &FileLogger{logger: logger}, nil
}

// Log mencatat event ke file.
func (l *FileLogger) Log(event LogEvent) {
	l.logger.Info().
		Str("request_id", event.RequestID).
		Str("domain", event.Domain).
		Str("path", event.Path).
		Str("method", event.Method).
		Str("remote_addr", event.RemoteAddr).
		Str("file_name", event.FileName).
		Int64("file_size", event.FileSize).
		Str("mime_type", event.MimeType).
		Str("upload_path", event.UploadPath).
		Str("source_field", event.SourceField).
		Msg("file intercepted")
}
