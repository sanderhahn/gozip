package main

import (
	"archive/zip"
	"errors"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

func IsZip(path string) bool {
	r, err := zip.OpenReader(path)
	if err == nil {
		r.Close()
		return true
	}
	return false
}

func Zip(path string, dirs []string) (err error) {
	if IsZip(path) {
		return errors.New(path + " is already a zip file")
	}

	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return
	}
	defer f.Close()

	w := zip.NewWriter(f)
	for _, dir := range dirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			// Empty directories get lost somewhere
			if !info.IsDir() {
				p, err := w.Create(path)
				if err != nil {
					return err
				}
				content, err := ioutil.ReadFile(path)
				if err != nil {
					return err
				}
				_, err = p.Write(content)
				if err != nil {
					return err
				}
			}
			return err
		})
	}
	err = w.Close()
	return
}

func Unzip(path string) (err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		os.MkdirAll(filepath.Dir(f.Name), 0755)
		out, err := os.Create(f.Name)
		if err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		_, err = io.CopyN(out, rc, f.FileInfo().Size())
		if err != nil {
			return err
		}
		rc.Close()
		out.Close()
	}
	return
}

func UnzipList(path string) (list []string, err error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		return
	}
	for _, f := range r.File {
		list = append(list, f.Name)
	}
	return
}

func main() {
	var list, extract, create bool
	flag.BoolVar(&create, "c", false, "create zip")
	flag.BoolVar(&list, "l", false, "list zip")
	flag.BoolVar(&extract, "x", false, "extract zip")

	flag.Parse()

	args := flag.Args()
	if len(args) == 1 && list {
		path := args[0]
		list, err := UnzipList(path)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range list {
			fmt.Printf("%s\n", f)
		}
	} else if len(args) == 1 && extract {
		path := args[0]
		err := Unzip(path)
		if err != nil {
			log.Fatal(err)
		}
	} else if create && len(args) > 1 {
		err := Zip(args[0], args[1:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		flag.PrintDefaults()
	}
}
