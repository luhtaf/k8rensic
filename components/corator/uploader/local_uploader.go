package uploader

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/luhtaf/corator/config"
)

// LocalUploader adalah implementasi uploader untuk menyimpan file ke disk lokal.
type LocalUploader struct {
	destinationPath string
}

// NewLocalUploader membuat instance baru dari LocalUploader.
func NewLocalUploader(cfg config.LocalConfig) (*LocalUploader, error) {
	if cfg.Path == "" {
		return nil, fmt.Errorf("path penyimpanan lokal tidak boleh kosong")
	}
	return &LocalUploader{
		destinationPath: cfg.Path,
	}, nil
}

// Upload menyimpan file ke path yang telah ditentukan.
func (u *LocalUploader) Upload(ctx context.Context, fileReader io.Reader, uniqueFilename string) (string, error) {
	// Pastikan direktori tujuan ada
	if err := os.MkdirAll(u.destinationPath, 0755); err != nil {
		return "", fmt.Errorf("gagal membuat direktori tujuan: %w", err)
	}

	// Gabungkan path direktori dengan nama file
	fullPath := filepath.Join(u.destinationPath, uniqueFilename)

	// Buat file baru di tujuan
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("gagal membuat file tujuan: %w", err)
	}
	defer dst.Close()

	// Salin konten dari file sumber ke file tujuan
	if _, err := io.Copy(dst, fileReader); err != nil {
		return "", fmt.Errorf("gagal menyalin konten file: %w", err)
	}

	// Kembalikan path lengkap dari file yang berhasil disimpan
	return fullPath, nil
}
