package gozip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	securejoin "github.com/cyphar/filepath-securejoin"
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

// Unzip extracts all files from the zip archive at zippath into the destination directory.
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
			if err := extractFile(f, fullname); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractFile(f *zip.File, fullname string) error {
	perms := f.FileInfo().Mode().Perm()
	out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
	if err != nil {
		return err
	}
	defer out.Close()

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
