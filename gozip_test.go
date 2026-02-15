package gozip

import (
	"archive/zip"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/sanderhahn/gozip/pathsafe"
)

func TestZip(t *testing.T) {

	zippath := "test.zip"

	t.Cleanup(func() {
		os.Remove("hello.txt")
		os.Remove(zippath)
		os.RemoveAll("files")
		os.RemoveAll("extract")
	})

	os.WriteFile(zippath, []byte("<possibly an exefile>"), 0644)

	os.MkdirAll("files/emptydir", 0755)
	os.WriteFile("hello.txt", []byte("Hello World"), 0777)
	os.WriteFile("files/hello.tpl", []byte("<h1>Hello World</h1>"), 0644)

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

	if err := Zip(zippath, []string{"files", "hello.txt"}); err != nil {
		t.Fatal(err)
	}

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
}

func TestZipPathTraversal(t *testing.T) {
	// Create a zip file with a path traversal entry (../readme.md)
	zippath := "poc.zip"
	defer os.Remove(zippath)

	f, err := os.Create(zippath)
	if err != nil {
		t.Fatal(err)
	}

	w := zip.NewWriter(f)
	// Create an entry with a path traversal attack
	pw, err := w.Create("../poc.txt")
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	_, err = pw.Write([]byte("malicious content"))
	if err != nil {
		f.Close()
		t.Fatal(err)
	}
	w.Close()
	f.Close()

	// Attempt to unzip - should return an error about illegal path
	err = Unzip(zippath, "extract_secure")
	if err == nil {
		os.RemoveAll("extract_secure")
		t.Fatal("expected error for path traversal, got nil")
	}
	if !strings.Contains(err.Error(), "illegal path") {
		t.Errorf("expected 'illegal path' error, got: %v", err)
	}

	os.RemoveAll("extract_secure")
}

func TestSafeJoinSymlinkEscape(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	dest := t.TempDir()
	outside := t.TempDir()

	linkPath := filepath.Join(dest, "linkout")
	if err := os.Symlink(outside, linkPath); err != nil {
		t.Fatalf("create symlink: %v", err)
	}

	_, err := pathsafe.SafeJoin(dest, filepath.Join("linkout", "evil.txt"))
	if err == nil {
		t.Fatal("expected error for symlink escape, got nil")
	}
}

func TestSafeJoinDestIsSymlink(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("symlink tests require elevated privileges on Windows")
	}

	base := t.TempDir()
	realDest := filepath.Join(base, "real")
	if err := os.MkdirAll(realDest, 0755); err != nil {
		t.Fatalf("mkdir real dest: %v", err)
	}

	linkDest := filepath.Join(base, "destlink")
	if err := os.Symlink(realDest, linkDest); err != nil {
		t.Fatalf("create dest symlink: %v", err)
	}

	got, err := pathsafe.SafeJoin(linkDest, "file.txt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	resolved, err := filepath.EvalSymlinks(realDest)
	if err != nil {
		t.Fatalf("eval symlinks: %v", err)
	}
	want := filepath.Join(resolved, "file.txt")
	if got != want {
		t.Fatalf("safeJoin symlink dest = %q, want %q", got, want)
	}
}

func TestFilePermissionsPreserved(t *testing.T) {
	zippath := "perms_test.zip"

	t.Cleanup(func() {
		os.Remove(zippath)
		os.RemoveAll("permfiles")
		os.RemoveAll("extract_perms")
	})

	cases := []struct {
		name string
		perm os.FileMode
	}{
		{"permfiles/readonly.txt", 0444},
		{"permfiles/readwrite.txt", 0644},
		{"permfiles/executable.sh", 0755},
		{"permfiles/owneronly.txt", 0600},
		{"permfiles/allex.sh", 0777},
	}

	if err := os.MkdirAll("permfiles", 0755); err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		if err := os.WriteFile(tc.name, []byte("content of "+tc.name), tc.perm); err != nil {
			t.Fatal(err)
		}
		// Explicitly chmod to bypass umask
		if err := os.Chmod(tc.name, tc.perm); err != nil {
			t.Fatal(err)
		}
	}

	if err := Zip(zippath, []string{"permfiles"}); err != nil {
		t.Fatal(err)
	}

	os.RemoveAll("permfiles")

	if err := Unzip(zippath, "extract_perms"); err != nil {
		t.Fatal(err)
	}

	for _, tc := range cases {
		extracted := path.Join("extract_perms", tc.name)
		info, err := os.Stat(extracted)
		if err != nil {
			t.Errorf("missing extracted file %s: %v", tc.name, err)
			continue
		}
		got := info.Mode().Perm()
		if got != tc.perm {
			t.Errorf("%s: permissions = %o, want %o", tc.name, got, tc.perm)
		}
	}
}
