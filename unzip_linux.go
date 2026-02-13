//go:build linux

package gozip

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	pathrs "github.com/cyphar/filepath-securejoin/pathrs-lite"
	"golang.org/x/sys/unix"
)

// Unzip extracts all files from the zip archive at zippath into the destination directory.
// On Linux, this uses TOCTOU-safe pathrs APIs with file descriptor-based operations.
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
	parentDir := filepath.Dir(entry)
	baseName := filepath.Base(entry)

	var parentPathHandle *os.File
	var err error
	if parentDir == "." {
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

	return writeExtractedFile(f, out, filePath)
}
