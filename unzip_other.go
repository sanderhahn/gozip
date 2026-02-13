//go:build !linux

package gozip

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
)

// Unzip extracts all files from the zip archive at zippath into the destination directory.
// On non-Linux systems, this uses securejoin.SecureJoin for path traversal protection.
func Unzip(zippath string, destination string) error {
	r, err := zip.OpenReader(zippath)
	if err != nil {
		return err
	}
	defer r.Close()

	destAbs, err := filepath.Abs(destination)
	if err != nil {
		return err
	}

	for _, f := range r.File {
		entry := filepath.ToSlash(f.Name)
		entry = strings.TrimLeft(entry, "/")

		// Check for path traversal attempts
		if strings.Contains(entry, "..") {
			return fmt.Errorf("illegal path %q: contains path traversal", f.Name)
		}

		fullname, err := securejoin.SecureJoin(destAbs, entry)
		if err != nil {
			return fmt.Errorf("illegal path %q: %w", f.Name, err)
		}

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fullname, f.FileInfo().Mode().Perm()); err != nil {
				return err
			}
		} else {
			if err := os.MkdirAll(filepath.Dir(fullname), 0755); err != nil {
				return err
			}
			if err := extractFilePath(f, fullname); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractFilePath(f *zip.File, fullname string) error {
	perms := f.FileInfo().Mode().Perm()
	out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
	if err != nil {
		return err
	}
	defer out.Close()

	return writeExtractedFile(f, out, fullname)
}
