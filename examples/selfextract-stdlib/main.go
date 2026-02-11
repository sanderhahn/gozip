package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
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

	f, err := os.Open(self)
	if err != nil {
		log.Fatalf("cannot open executable: %v", err)
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		log.Fatalf("cannot stat executable: %v", err)
	}

	r, err := zip.NewReader(f, info.Size())
	if err != nil {
		fmt.Println("No embedded files found. Append files first:")
		fmt.Printf("  gozip -c %s <files...>\n", os.Args[0])
		os.Exit(1)
	}

	if listFlag {
		for _, zf := range r.File {
			fmt.Println(zf.Name)
		}
		return
	}

	fmt.Printf("Extracting to %s ...\n", dest)
	for _, zf := range r.File {
		outPath := filepath.Join(dest, zf.Name)

		if zf.FileInfo().IsDir() {
			if err := os.MkdirAll(outPath, zf.Mode()); err != nil {
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

func extractFile(zf *zip.File, outPath string) error {
	src, err := zf.Open()
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, zf.Mode())
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	return err
}
