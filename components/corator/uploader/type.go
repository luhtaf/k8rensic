package uploader

import (
	"context"
	"io"
)

// Uploader adalah interface umum untuk semua implementasi uploader.
type Uploader interface {
	// Mengembalikan URL/path dari file yang diupload dan error
	Upload(ctx context.Context, fileReader io.Reader, uniqueFilename string) (string, error)
}
