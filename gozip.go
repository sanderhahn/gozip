package gozip

import (
	"archive/zip"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	pathrs "github.com/cyphar/filepath-securejoin/pathrs-lite"
	"golang.org/x/sys/unix"
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

	if err := os.MkdirAll(destAbs, 0755); err != nil {
		return err
	}

	rootDir, err := os.OpenFile(destAbs, unix.O_PATH|unix.O_DIRECTORY|unix.O_CLOEXEC, 0)
	if err != nil {
		return err
	}
	defer rootDir.Close()

	for _, f := range r.File {
		entry := filepath.ToSlash(f.Name)
		entry = strings.TrimLeft(entry, "/")

		// Check for path traversal attempts
		if strings.Contains(entry, "..") {
			return fmt.Errorf("illegal path %q: contains path traversal", f.Name)
		}

		if f.FileInfo().IsDir() {
			dirHandle, err := pathrs.MkdirAllHandle(rootDir, entry, f.FileInfo().Mode().Perm())
			if err != nil {
				return fmt.Errorf("illegal path %q: %w", f.Name, err)
			}
			dirHandle.Close()
		} else {
			parentDir := filepath.Dir(entry)
			if parentDir != "." {
				parentHandle, err := pathrs.MkdirAllHandle(rootDir, parentDir, 0755)
				if err != nil {
					return fmt.Errorf("illegal path %q: %w", f.Name, err)
				}
				parentHandle.Close()
			}
			if err := extractFile(f, rootDir, entry); err != nil {
				return err
			}
		}
	}
	return nil
}

func extractFile(f *zip.File, rootDir *os.File, entry string) error {
	perms := f.FileInfo().Mode().Perm()

	// Safely resolve the parent directory within the root using OpenatInRoot,
	// then create the file relative to the directory handle via /proc/self/fd.
	// Note: This relies on procfs being mounted, which is required by the
	// pathrs-lite package (Linux-only).
	parentDir := filepath.Dir(entry)
	baseName := filepath.Base(entry)

	var parentPathHandle *os.File
	var err error
	if parentDir == "." {
		// File is directly in root; use the root handle itself
		parentPathHandle = rootDir
	} else {
		parentPathHandle, err = pathrs.OpenatInRoot(rootDir, parentDir)
		if err != nil {
			return err
		}
		defer parentPathHandle.Close()
	}

	// Reopen the O_PATH handle as a usable directory handle
	dirHandle, err := pathrs.Reopen(parentPathHandle, unix.O_DIRECTORY|unix.O_RDONLY)
	if err != nil {
		return err
	}
	defer dirHandle.Close()

	filePath := fmt.Sprintf("/proc/self/fd/%d/%s", dirHandle.Fd(), baseName)

	out, err := os.OpenFile(filePath, os.O_CREATE|os.O_RDWR, perms)
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
	return os.Chtimes(filePath, mtime, mtime)
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
