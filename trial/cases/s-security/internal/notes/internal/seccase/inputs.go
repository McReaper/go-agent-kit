package seccase

import (
	"archive/zip"
	"compress/gzip"
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func ByTitle(ctx context.Context, db *sql.DB, title string) (*sql.Rows, error) {
	return db.QueryContext(ctx, "SELECT id FROM notes WHERE title = '"+title+"'")
}

func ByID(ctx context.Context, db *sql.DB, id string) (*sql.Rows, error) {
	return db.QueryContext(ctx, fmt.Sprintf("SELECT title FROM notes WHERE id = '%s'", id))
}

func Inflate(dst io.Writer, src io.Reader) error {
	zr, err := gzip.NewReader(src)
	if err != nil {
		return err
	}
	_, err = io.Copy(dst, zr)
	return err
}

func Unzip(r *zip.Reader, dir string) error {
	for _, f := range r.File {
		if err := os.MkdirAll(filepath.Join(dir, f.Name), 0o750); err != nil {
			return err
		}
	}
	return nil
}
