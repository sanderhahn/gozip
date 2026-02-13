package pathsafe

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// SafeJoin joins dest with name while preventing path traversal, absolute paths,
// and Windows volume or UNC paths, regardless of the host OS. It normalizes
// separators and rejects absolute or volume-prefixed inputs.
//
// Note: this is a pre-validation step. It reduces symlink-based escapes by
// rejecting existing symlinks in the path, but it cannot eliminate TOCTOU
// (time-of-check to time-of-use) races if an attacker can replace path
// components after validation.
func SafeJoin(dest, name string) (string, error) {
	if name == "" {
		return "", fmt.Errorf("empty path")
	}

	// Resolve the destination root to prevent symlink escapes via the root itself.
	// If it does not exist yet or cannot be resolved, fall back to the absolute path.
	if resolved, err := filepath.EvalSymlinks(dest); err == nil {
		dest = resolved
	} else {
		abs, absErr := filepath.Abs(dest)
		if absErr != nil {
			return "", absErr
		}
		dest = abs
	}

	entry := strings.ReplaceAll(name, "\\", "/")
	entry = filepath.ToSlash(entry)
	if strings.HasPrefix(entry, "//") {
		return "", fmt.Errorf("absolute path")
	}
	if strings.HasPrefix(entry, "/") {
		return "", fmt.Errorf("absolute path")
	}
	if hasWindowsVolume(entry) {
		return "", fmt.Errorf("windows volume")
	}

	clean := path.Clean(entry)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") {
		return "", fmt.Errorf("path traversal")
	}

	target := filepath.Join(dest, filepath.FromSlash(clean))
	destClean := filepath.Clean(dest)
	targetClean := filepath.Clean(target)

	rel, err := filepath.Rel(destClean, targetClean)
	if err != nil {
		return "", err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("escaped destination")
	}
	if err := ensureNoSymlinkPrefix(destClean, targetClean); err != nil {
		return "", err
	}

	return targetClean, nil
}

// ensureNoSymlinkPrefix rejects any pre-existing symlink in the target path
// and confirms the final path resolves within dest. This cannot fully prevent
// TOCTOU (time-of-check to time-of-use) races if an attacker can swap in a
// symlink after the checks but before file creation.
func ensureNoSymlinkPrefix(dest, target string) error {
	root := filepath.Clean(dest)
	cleanTarget := filepath.Clean(target)

	// Ensure the raw cleaned target is still under the destination.
	rel, err := filepath.Rel(root, cleanTarget)
	if err != nil {
		return err
	}
	if rel == ".." || strings.HasPrefix(rel, ".."+string(os.PathSeparator)) {
		return fmt.Errorf("escaped destination")
	}

	// Walk each existing path component from root to target and reject symlinks.
	current := root
	relPath := strings.TrimPrefix(rel, string(os.PathSeparator))
	if relPath == "." || relPath == "" {
		return nil
	}
	for _, part := range strings.Split(relPath, string(os.PathSeparator)) {
		if part == "" {
			continue
		}
		current = filepath.Join(current, part)
		info, err := os.Lstat(current)
		if err != nil {
			if os.IsNotExist(err) {
				// Component does not exist yet; no further symlink checks possible here.
				return nil
			}
			return err
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("symlink in path")
		}
	}

	return nil
}

func hasWindowsVolume(path string) bool {
	if len(path) < 2 {
		return false
	}
	c := path[0]
	if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') {
		return path[1] == ':'
	}
	return false
}
