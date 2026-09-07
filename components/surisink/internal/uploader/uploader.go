package uploader

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/luhtaf/surisink/internal/meta"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/minio/minio-go/v7/pkg/tags"
)

// Uploader uploads files to S3-compatible storage.
type Uploader struct {
	cli    *minio.Client
	bucket string
	prefix string
}

// New constructs an Uploader.
func New(endpoint, ak, sk, bucket, prefix string, useSSL bool) (*Uploader, error) {
	cli, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(ak, sk, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	return &Uploader{cli: cli, bucket: bucket, prefix: prefix}, nil
}

// EnsureBucket ensures the bucket exists.
func (u *Uploader) EnsureBucket(ctx context.Context) error {
	exists, err := u.cli.BucketExists(ctx, u.bucket)
	if err != nil {
		return err
	}
	if !exists {
		return u.cli.MakeBucket(ctx, u.bucket, minio.MakeBucketOptions{})
	}
	return nil
}

// ObjectKey builds a deterministic object key.
func (u *Uploader) ObjectKey(ts time.Time, flowID, sha256, origName string) string {
	return fmt.Sprintf("%s/%04d/%02d/%02d/%s/%s_%s",
		u.prefix, ts.Year(), ts.Month(), ts.Day(), flowID, sha256, filepath.Base(origName),
	)
}

// UploadFile uploads the file and applies tags/metadata.
func (u *Uploader) UploadFile(ctx context.Context, fm meta.FileMeta) (string, error) {
	key := u.ObjectKey(fm.TS, fm.FlowID, fm.SHA256, fm.OrigName)
	put := minio.PutObjectOptions{
		ContentType: fm.MIME,
		UserMetadata: map[string]string{
			"x-amz-meta-sha256":  fm.SHA256,
			"x-amz-meta-ts":      fm.TS.UTC().Format(time.RFC3339),
			"x-amz-meta-flow_id": fm.FlowID,
			"x-amz-meta-src":     fm.SrcIP,
			"x-amz-meta-dst":     fm.DstIP,
			"x-amz-meta-sensor":  fm.Sensor,
		},
	}
	_, err := u.cli.FPutObject(ctx, u.bucket, key, fm.Path, put)
	if err != nil {
		return "", err
	}

	kv := map[string]string{
		"sha256": fm.SHA256,
		"mime":   fm.MIME,
		"ts":     fm.TS.UTC().Format(time.RFC3339),
	}
	if fm.FlowID != "" {
		kv["flow_id"] = fm.FlowID
	}
	if fm.SrcIP != "" {
		kv["src"] = fm.SrcIP
	}
	if fm.DstIP != "" {
		kv["dst"] = fm.DstIP
	}
	if fm.Sensor != "" {
		kv["sensor"] = fm.Sensor
	}
	// minio-go v7 takes *tags.Tags here, not a plain map. The map form compiles
	// against no released version this module pins, so this call never built —
	// see docs/development.md.
	//
	// The second argument to NewTags is isObject: object tags allow 10 pairs,
	// bucket tags 50, and passing false silently accepts a set the API will
	// later reject.
	t, err := tags.NewTags(kv, true)
	if err != nil {
		return "", fmt.Errorf("building object tags: %w", err)
	}
	if err := u.cli.PutObjectTagging(ctx, u.bucket, key, t, minio.PutObjectTaggingOptions{}); err != nil {
		return "", err
	}
	return key, nil
}
