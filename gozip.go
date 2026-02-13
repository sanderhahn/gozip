package gozip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// ErrAlreadyZip is returned when trying to zip into a file that is already a zip.
var ErrAlreadyZip = errors.New("file is already a zip")

// IsZip checks to see if path is already a zip file
func IsZip(path string) bool {
	r, err := zip.OpenReader(path)
	if err == nil {
		r.Close()
		return true
	}
	return false
}

// Zip takes all the files (dirs) and zips them into path.
func Zip(path string, dirs []string) error {
	if IsZip(path) {
		return fmt.Errorf("%s: %w", path, ErrAlreadyZip)
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	defer f.Close()

	startoffset, err := f.Seek(0, io.SeekEnd)
	if err != nil {
		return err
	}

	w := zip.NewWriter(f)
	w.SetOffset(startoffset)

	for _, dir := range dirs {
		err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			fh, err := zip.FileInfoHeader(info)
			if err != nil {
				return err
			}
			fh.Name = path

			p, err := w.CreateHeader(fh)
			if err != nil {
				return err
			}
			if !info.IsDir() {
				src, err := os.Open(path)
				if err != nil {
					return err
				}
				defer src.Close()
				_, err = io.Copy(p, src)
				if err != nil {
					return err
				}
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return w.Close()
}

func writeExtractedFile(f *zip.File, out *os.File, fullname string) error {
	perms := f.FileInfo().Mode().Perm()

	rc, err := f.Open()
	if err != nil {
		return err
	}
	defer rc.Close()

	_, err = io.CopyN(out, rc, f.FileInfo().Size())
	if err != nil {
		return err
	}

	// Explicitly set permissions to bypass umask
	if err := out.Chmod(perms); err != nil {
		return err
	}

	mtime := f.FileInfo().ModTime()
	return os.Chtimes(fullname, mtime, mtime)
}

// UnzipList lists all the files in the zip file at path.
func UnzipList(path string) ([]string, error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	var list []string
	for _, f := range r.File {
		list = append(list, f.Name)
	}
	return list, nil
}
