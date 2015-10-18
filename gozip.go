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
	"path"
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

	startoffset, err := f.Seek(0, os.SEEK_END)
	if err != nil {
		return
	}

	w := zip.NewWriter(f)
	w.SetOffset(startoffset)

	for _, dir := range dirs {
		err = filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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

func Unzip(zippath string, destination string) (err error) {
	r, err := zip.OpenReader(zippath)
	if err != nil {
		return err
	}
	for _, f := range r.File {
		fullname := path.Join(destination, f.Name)
		if f.FileInfo().IsDir() {
			os.MkdirAll(fullname, f.FileInfo().Mode().Perm())
		} else {
			os.MkdirAll(filepath.Dir(fullname), 0755)
			perms := f.FileInfo().Mode().Perm()
			out, err := os.OpenFile(fullname, os.O_CREATE|os.O_RDWR, perms)
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

			mtime := f.FileInfo().ModTime()
			err = os.Chtimes(fullname, mtime, mtime)
			if err != nil {
				return err
			}
		}
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
	flag.BoolVar(&create, "c", false, "create zip (arguments: zipfile [files...])")
	flag.BoolVar(&list, "l", false, "list zip (arguments: zipfile)")
	flag.BoolVar(&extract, "x", false, "extract zip (arguments: zipfile [destination]")

	flag.Parse()

	args := flag.Args()
	argc := len(args)
	if list && argc == 1 {
		path := args[0]
		list, err := UnzipList(path)
		if err != nil {
			log.Fatal(err)
		}
		for _, f := range list {
			fmt.Printf("%s\n", f)
		}
	} else if extract && (argc == 1 || argc == 2) {
		path := args[0]
		dest := "."
		if argc == 2 {
			dest = args[1]
		}
		err := Unzip(path, dest)
		if err != nil {
			log.Fatal(err)
		}
	} else if create && argc > 1 {
		err := Zip(args[0], args[1:])
		if err != nil {
			log.Fatal(err)
		}
	} else {
		flag.Usage()
	}
}
