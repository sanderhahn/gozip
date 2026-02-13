// Package pathsafe provides path sanitization utilities for safe file extraction
// from archives (e.g., zip files).
//
// Security Model:
//
// This package aims to prevent common path traversal and zip-slip attacks by
// validating and sanitizing file paths before extraction. It checks for:
//   - Absolute paths
//   - Path traversal attempts (../)
//   - Windows volume and UNC paths
//   - Existing symlinks in the target path
//
// However, users should be aware of inherent limitations:
//
//  1. TOCTOU Races: The validation functions use a check-then-use pattern which
//     is vulnerable to time-of-check to time-of-use (TOCTOU) race conditions.
//     An attacker with filesystem write access could create malicious symlinks
//     between validation and file creation.
//
//  2. Environment Trust: This implementation assumes a reasonably trusted
//     extraction environment where the destination directory and its parents
//     are not controlled by an attacker. For untrusted environments, additional
//     hardening measures (such as creating the destination with secure permissions
//     up-front, or using platform-specific atomic/no-follow file operations) may
//     be necessary.
//
// This package is suitable for typical use cases where archives come from trusted
// or semi-trusted sources, and the extraction environment is under the user's
// control. For high-security scenarios involving untrusted archives in shared
// or hostile environments, consider using additional security layers beyond
// path validation alone.
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
// Security Note: This function provides defense-in-depth against zip-slip and
// path traversal attacks, but it cannot eliminate all race conditions. Specifically:
//
//  1. TOCTOU (time-of-check to time-of-use) races: An attacker with filesystem
//     access could replace a validated path component with a symlink between the
//     SafeJoin check and subsequent file creation. This is an inherent limitation
//     of check-then-use patterns.
//
//  2. Destination directory symlinks: If the destination directory (or a parent)
//     can be replaced with a symlink by an attacker, extraction can be redirected
//     outside the intended location.
//
// For maximum security in hostile environments:
//  - Ensure the destination directory and its parents are not writable by
//    untrusted users.
//  - Consider creating the destination directory up-front with secure permissions.
//  - On platforms that support it, use atomic file operations with O_NOFOLLOW
//    or similar flags to prevent symlink following.
//
// This implementation is suitable for typical use cases where the extraction
// environment is trusted (e.g., user extracts their own archives). For
// untrusted environments or archives from untrusted sources, additional
// hardening may be required.
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
// and confirms the final path resolves within dest.
//
// Security Limitation: This function returns nil (no error) as soon as it
// encounters a non-existent path component. This means it cannot detect symlinks
// that are created later at that location before file extraction occurs, enabling
// a TOCTOU (time-of-check to time-of-use) race condition.
//
// This is an inherent limitation of check-then-use validation patterns. An attacker
// with filesystem write access could:
//  1. Wait for SafeJoin to validate a path with non-existent components
//  2. Create a malicious symlink at that location before file creation
//  3. Redirect extraction outside the intended destination
//
// Complete mitigation would require atomic path walking (e.g., openat-style APIs
// on Unix) or platform-specific no-follow flags during file creation.
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
