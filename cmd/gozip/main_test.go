package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

var gozipBin string

func TestMain(m *testing.M) {
	moduleRoot, err := moduleRootDir()
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	binDir, err := os.MkdirTemp("", "gozip-bin-")
	if err != nil {
		os.Stderr.WriteString(err.Error() + "\n")
		os.Exit(1)
	}

	binName := "gozip"
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}
	gozipBin = filepath.Join(binDir, binName)

	cmd := exec.Command("go", "build", "-o", gozipBin, "./cmd/gozip")
	cmd.Dir = moduleRoot
	if output, err := cmd.CombinedOutput(); err != nil {
		os.Stderr.WriteString(string(output))
		os.RemoveAll(binDir)
		os.Exit(1)
	}

	code := m.Run()
	os.RemoveAll(binDir)
	os.Exit(code)
}

func moduleRootDir() (string, error) {
	wd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return filepath.Clean(filepath.Join(wd, "..", "..")), nil
}

func runGozip(t *testing.T, workDir string, args ...string) string {
	t.Helper()
	cmd := exec.Command(gozipBin, args...)
	cmd.Dir = workDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("gozip %v failed: %v\n%s", args, err, output)
	}
	return string(output)
}

func TestCmdCreateAndList(t *testing.T) {
	workDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	if err := os.MkdirAll(filepath.Join(workDir, "dir"), 0755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(workDir, "dir", "sub.txt"), []byte("sub"), 0644); err != nil {
		t.Fatalf("write sub file: %v", err)
	}

	runGozip(t, workDir, "-c", "test.zip", "file.txt", "dir")

	out := runGozip(t, workDir, "-l", "test.zip")
	if !strings.Contains(out, "file.txt") {
		t.Fatalf("list missing file.txt: %q", out)
	}
	if !strings.Contains(out, "dir") {
		t.Fatalf("list missing dir entry: %q", out)
	}
	if !strings.Contains(out, "dir/sub.txt") {
		t.Fatalf("list missing dir/sub.txt: %q", out)
	}
}

func TestCmdExtract(t *testing.T) {
	workDir := t.TempDir()

	if err := os.WriteFile(filepath.Join(workDir, "file.txt"), []byte("hello"), 0644); err != nil {
		t.Fatalf("write file: %v", err)
	}

	runGozip(t, workDir, "-c", "test.zip", "file.txt")

	if err := os.Remove(filepath.Join(workDir, "file.txt")); err != nil {
		t.Fatalf("remove source file: %v", err)
	}

	extractDir := filepath.Join(workDir, "out")
	runGozip(t, workDir, "-x", "test.zip", extractDir)

	if _, err := os.Stat(filepath.Join(extractDir, "file.txt")); err != nil {
		t.Fatalf("expected extracted file: %v", err)
	}
}
