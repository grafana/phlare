package filesystem

import (
	"context"
	"os"
	"path/filepath"

	phlareobjstore "github.com/grafana/phlare/pkg/objstore"
	thanosobjstore "github.com/thanos-io/objstore"
	"github.com/thanos-io/objstore/providers/filesystem"
)

type Bucket struct {
	thanosobjstore.Bucket
	rootDir string
}

// NewBucket returns a new filesystem.Bucket.
func NewBucket(rootDir string) (*Bucket, error) {
	b, err := filesystem.NewBucket(rootDir)
	if err != nil {
		return nil, err
	}
	return &Bucket{Bucket: b, rootDir: rootDir}, nil
}

func (b *Bucket) ReaderAt(ctx context.Context, filename string) (phlareobjstore.ReaderAtCloser, error) {
	f, err := os.Open(filepath.Join(b.rootDir, filename))
	if err != nil {
		return nil, err
	}

	return &FileReaderAt{File: f}, nil
}

type FileReaderAt struct {
	*os.File
}

func (b *FileReaderAt) ReadAt(p []byte, off int64) (n int, err error) {
	return b.File.ReadAt(p, off)
}
