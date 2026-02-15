package pathsafe

import (
	"path/filepath"
	"testing"
)

func TestSafeJoin(t *testing.T) {
	dest := t.TempDir()
	destResolved := dest
	if resolved, err := filepath.EvalSymlinks(dest); err == nil {
		destResolved = resolved
	}

	type testCase struct {
		name    string
		entry   string
		want    string
		wantErr bool
	}

	cases := []testCase{
		{name: "file", entry: "file.txt", want: "file.txt"},
		{name: "dir file", entry: "dir/sub.txt", want: "dir/sub.txt"},
		{name: "dir file backslash", entry: `dir\sub.txt`, want: "dir/sub.txt"},
		{name: "dot segment", entry: "a/./b", want: "a/b"},
		{name: "parent cleanup", entry: "a/../b", want: "b"},
		{name: "double slash", entry: "a//b", want: "a/b"},
		{name: "empty", entry: "", wantErr: true},
		{name: "dot", entry: ".", wantErr: true},
		{name: "dotdot", entry: "..", wantErr: true},
		{name: "traversal", entry: "../evil", wantErr: true},
		{name: "deep traversal", entry: "a/../../evil", wantErr: true},
		{name: "abs unix", entry: "/abs", wantErr: true},
		{name: "abs windows", entry: `\\abs`, wantErr: true},
		{name: "drive relative", entry: "C:evil", wantErr: true},
		{name: "drive absolute", entry: `C:\\evil`, wantErr: true},
		{name: "unc", entry: "//server/share", wantErr: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SafeJoin(dest, tc.entry)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error for %q, got nil", tc.entry)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error for %q: %v", tc.entry, err)
			}
			want := filepath.Join(destResolved, filepath.FromSlash(tc.want))
			if got != want {
				t.Fatalf("SafeJoin(%q) = %q, want %q", tc.entry, got, want)
			}
		})
	}
}
