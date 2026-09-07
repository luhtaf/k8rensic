package uploader

import (
	"fmt"

	"github.com/luhtaf/corator/config"
)

// NewUploader adalah factory yang membuat instance uploader berdasarkan konfigurasi.
func NewUploader(cfg *config.Config) (Uploader, error) {
	switch cfg.Uploader.Type {
	case "local":
		return NewLocalUploader(cfg.Uploader.Local)
	case "s3":
		return NewS3Uploader(cfg.Uploader.S3)
	default:
		return nil, fmt.Errorf("tipe uploader tidak dikenal: %s", cfg.Uploader.Type)
	}
}
