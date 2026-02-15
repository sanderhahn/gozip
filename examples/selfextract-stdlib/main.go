package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/sanderhahn/gozip/pathsafe"
)

func main() {
	var listFlag bool
	var dest string
	flag.BoolVar(&listFlag, "l", false, "list embedded files")
	flag.StringVar(&dest, "d", "extracted", "destination directory for extraction")
	flag.Parse()

	self, err := os.Executable()
	if err != nil {
		log.Fatalf("cannot determine own path: %v", err)
	}

	rc, err := zip.OpenReader(self)
	if err != nil {
		fmt.Println("No embedded files found. Append files first:")
		fmt.Printf("  gozip -c %s <files...>\n", os.Args[0])
		os.Exit(1)
	}
	defer rc.Close()

	r := &rc.Reader

	if listFlag {
		for _, zf := range r.File {
			fmt.Println(zf.Name)
		}
		return
	}

	fmt.Printf("Extracting to %s ...\n", dest)
	for _, zf := range r.File {
		outPath, err := sanitizePath(dest, zf.Name)
		if err != nil {
			log.Printf("warning: dangerous entry %q skipped: %v", zf.Name, err)
			continue
		}

		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, zf.Mode().Perm()); err != nil {
				log.Fatalf("mkdir failed: %v", err)
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			log.Fatalf("mkdir failed: %v", err)
		}

		if err := extractFile(zf, outPath); err != nil {
			log.Fatalf("extract %s failed: %v", zf.Name, err)
		}
	}
	fmt.Println("Done.")
}

// sanitizePath ensures the target path stays within the destination directory,
// preventing zip slip / path traversal attacks where a malicious zip entry
// contains paths like "../../../etc/cron.d/evil".
func sanitizePath(dest, name string) (string, error) {
	target, err := pathsafe.SafeJoin(dest, name)
	if err != nil {
		return "", fmt.Errorf("illegal file path %q: %w", name, err)
	}
	return target, nil
}

func extractFile(zf *zip.File, outPath string) error {
	src, err := zf.Open()
	if err != nil {
		return err
	}

	dst, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode().Perm())
	if err != nil {
		src.Close()
		return err
	}

	_, copyErr := io.Copy(dst, src)
	closeErr := dst.Close()
	srcErr := src.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	return srcErr
}
