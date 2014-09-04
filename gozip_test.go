package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestZip(t *testing.T) {

	path := "test.zip"

	ioutil.WriteFile(path, []byte(""), 0644)

	os.MkdirAll("views", 0755)
	ioutil.WriteFile("views/hello.tpl", []byte("<h1>Hello World</h1>"), 0644)
	ioutil.WriteFile("hello.txt", []byte("Hello World"), 0644)

	Zip(path, []string{"views", "hello.txt"})

	if !IsZip(path) {
		t.Error("zip test failed")
	}

	os.RemoveAll("views")
	os.Remove("hello.txt")

	if err := Unzip(path); err != nil {
		t.Error("unzip failed")
	}

	if _, err := os.Stat("views/hello.tpl"); os.IsNotExist(err) {
		t.Error("unzip didn't work")
	}
	if _, err := os.Stat("views/hello.tpl"); os.IsNotExist(err) {
		t.Error("hello.txt")
	}

	os.RemoveAll("views")
	os.Remove("hello.txt")

	if !t.Failed() {
		os.Remove("test.zip")
	}

}
