package phlaredb

import (
	"log"
	"os"
	"testing"

	"github.com/segmentio/parquet-go"
)

func TestParquetPageSize(t *testing.T) {
	f, err := os.OpenFile("../../testdata/01H1X4NJ7Y8JZWNNP26VKY3XM4/profiles.parquet", os.O_RDONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	newFile, err := os.OpenFile("../../testdata/01H1X4NJ7Y8JZWNNP26VKY3XM4/profiles.parquet.new", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer newFile.Close()
	stats, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}

	pqf, err := parquet.OpenFile(f, stats.Size())
	if err != nil {
		t.Fatal(err)
	}

	writer := parquet.NewWriter(newFile, parquet.PageBufferSize(5*1024*1024))

	for _, rowGroup := range pqf.RowGroups() {
		_, err := writer.WriteRowGroup(rowGroup)
		if err != nil {
			t.Fatal(err)
		}
	}

	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
}
