// Self-extracting binary example
//
// This program demonstrates how to use the gozip library to create a
// self-extracting Go binary. When the binary runs, it extracts the files
// that were appended to it as a zip archive.
//
// Build and use:
//
//	go build -o selfextract .
//	gozip -c selfextract payload/           # append payload files to the binary
//	./selfextract                           # extracts payload to ./extracted/
//	./selfextract -l                        # lists embedded files
//	./selfextract -d /tmp/myfiles           # extracts to custom directory
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/sanderhahn/gozip"
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

	if !gozip.IsZip(self) {
		fmt.Println("No embedded files found. Append files first:")
		fmt.Printf("  gozip -c %s <files...>\n", os.Args[0])
		os.Exit(1)
	}

	if listFlag {
		files, err := gozip.UnzipList(self)
		if err != nil {
			log.Fatalf("listing failed: %v", err)
		}
		for _, f := range files {
			fmt.Println(f)
		}
		return
	}

	fmt.Printf("Extracting to %s ...\n", dest)
	if err := gozip.Unzip(self, dest); err != nil {
		log.Fatalf("extraction failed: %v", err)
	}
	fmt.Println("Done.")
}
