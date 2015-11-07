package gozip

import (
	"io/ioutil"
	"os"
	"path"
	"testing"
	"time"
)

func TestZip(t *testing.T) {

	zippath := "test.zip"

	ioutil.WriteFile(zippath, []byte("<possibly an exefile>"), 0644)

	os.MkdirAll("files/emptydir", 0755)
	ioutil.WriteFile("hello.txt", []byte("Hello World"), 0777)
	ioutil.WriteFile("files/hello.tpl", []byte("<h1>Hello World</h1>"), 0644)

	testfileheader := "hello.txt"
	info, err := os.Stat(testfileheader)
	if err != nil {
		t.Error(err)
	}
	actualperms := info.Mode().Perm()

	filetime := time.Date(2009, time.November, 10, 23, 0, 0, 0, time.UTC)
	err = os.Chtimes(testfileheader, filetime, filetime)
	if err != nil {
		t.Error(err)
	}

	Zip(zippath, []string{"files", "hello.txt"})

	if !IsZip(zippath) {
		t.Error("zip test failed")
	}

	os.RemoveAll("files")
	os.Remove("hello.txt")

	list, err := UnzipList(zippath)
	if err != nil || len(list) != 4 {
		t.Error("unzip list failed")
	}

	if err := Unzip(zippath, "extract"); err != nil {
		t.Error("unzip failed")
	}

	if _, err := os.Stat("extract/files/hello.tpl"); os.IsNotExist(err) {
		t.Error("unzip didn't work")
	}
	if _, err := os.Stat("extract/files/emptydir"); os.IsNotExist(err) {
		t.Error("unzip didn't create empty dir")
	}

	info, err = os.Stat(path.Join("extract", testfileheader))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	if info.Mode().Perm() != actualperms {
		t.Error("unzip didn't set file perms")
	}
	if !info.ModTime().Equal(filetime) {
		t.Error("unzip didn't set file modtime")
	}

	if !t.Failed() {
		os.Remove("hello.txt")
		os.Remove("test.zip")
		os.RemoveAll("files")
		os.RemoveAll("extract")
	}

}
