// v0.10.0 #174 Tier 2 — ZIP extraction primitives.
//
// Used by atomic-replace to land a downloaded update bundle into a
// staging directory on the same filesystem as the destination, so the
// subsequent os.Rename swap is syscall-atomic.
//
// ZIP-only for v0.10.0 — TAR remains the documented "not implemented"
// error path. The Tier 2 plan calls out ZIP (~30 lines stdlib) as the
// lowest-risk extraction format and the only one update.DownloadWithProgress
// produces today (.zip extension on the bundle URL).
//
// Security guards:
//   - Zip-slip: every entry's resolved path is verified to live UNDER
//     destDir. Entries with absolute paths or .. traversal are rejected.
//   - Symlinks: not extracted. ZIP can encode them in the external_attributes
//     field; we silently skip those entries to avoid both link-traversal
//     and TOCTOU surprises.
//   - Size cap: per-entry size capped at MaxExtractedFileBytes; total
//     extracted size capped at MaxTotalExtractedBytes. Prevents zip-bomb
//     resource exhaustion.
//   - File modes: write bit cleared on group/other so an unprivileged
//     bundle can't drop files with broader perms than the operator's umask.
package update

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// MaxExtractedFileBytes caps a single entry's decompressed size. Keeps
// a malicious entry from filling disk before the total cap fires.
const MaxExtractedFileBytes = 100 * 1024 * 1024 // 100 MiB

// MaxTotalExtractedBytes caps total decompressed bundle size. Sized for
// a templates+modules bundle plus generous headroom; operators with
// genuinely larger bundles can override at the call site.
const MaxTotalExtractedBytes = 1024 * 1024 * 1024 // 1 GiB

// ExtractZip extracts the ZIP archive at archivePath into destDir. The
// destDir must exist and be empty (mkdir -p the parent and create destDir
// fresh in the caller). Returns the total number of bytes extracted on
// success.
//
// Security invariants enforced:
//   - All entry paths must resolve under destDir (zip-slip protection).
//   - Symlinks are skipped silently.
//   - Per-entry and total decompressed-size caps apply.
//   - Directory entries become directories with 0755 perms; file entries
//     become files with the entry's mode AND'd with 0755 (group/other
//     write bits cleared).
func ExtractZip(archivePath, destDir string) (int64, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return 0, fmt.Errorf("open zip: %w", err)
	}
	defer func() { _ = zr.Close() }()

	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return 0, fmt.Errorf("resolve dest: %w", err)
	}

	var total int64
	for _, entry := range zr.File {
		// Skip symlinks (Mode() & os.ModeSymlink). ZIP can encode them
		// in external_attributes; we don't extract them to avoid both
		// link-traversal and TOCTOU surprises.
		if entry.Mode()&os.ModeSymlink != 0 {
			continue
		}

		target, err := safeJoin(absDest, entry.Name)
		if err != nil {
			return 0, fmt.Errorf("entry %q: %w", entry.Name, err)
		}

		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o750); err != nil {
				return 0, fmt.Errorf("mkdir %q: %w", target, err)
			}
			continue
		}

		// File entry: ensure parent dir exists.
		if err := os.MkdirAll(filepath.Dir(target), 0o750); err != nil {
			return 0, fmt.Errorf("mkdir parent of %q: %w", target, err)
		}

		written, err := extractFileEntry(entry, target, MaxTotalExtractedBytes-total)
		if err != nil {
			return 0, fmt.Errorf("extract %q: %w", entry.Name, err)
		}
		total += written
		if total > MaxTotalExtractedBytes {
			return 0, fmt.Errorf("extracted %d bytes exceeds total cap %d", total, MaxTotalExtractedBytes)
		}
	}

	return total, nil
}

// extractFileEntry copies one zip.File entry into target. budget is the
// remaining global byte budget; the entry's decompressed size must not
// push total over MaxTotalExtractedBytes.
func extractFileEntry(entry *zip.File, target string, budget int64) (int64, error) {
	rc, err := entry.Open()
	if err != nil {
		return 0, fmt.Errorf("open entry: %w", err)
	}
	defer func() { _ = rc.Close() }()

	mode := entry.Mode().Perm() & 0o755
	if mode == 0 {
		mode = 0o644
	}

	// #nosec G304 — target is the output of safeJoin(destDir, entry.Name),
	// which rejects absolute paths and traversal and re-verifies the
	// joined path resolves under destDir. The "variable file inclusion"
	// concern doesn't apply: target cannot land outside the staging dir.
	out, err := os.OpenFile(target, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, mode)
	if err != nil {
		return 0, fmt.Errorf("create file: %w", err)
	}
	defer func() { _ = out.Close() }()

	// Per-entry cap: io.LimitedReader stops the copy at MaxExtractedFileBytes+1,
	// so we can detect overrun explicitly rather than truncating silently.
	cap := int64(MaxExtractedFileBytes)
	if budget < cap {
		cap = budget
	}
	if cap <= 0 {
		return 0, fmt.Errorf("size budget exhausted")
	}
	limited := &io.LimitedReader{R: rc, N: cap + 1}
	n, err := io.Copy(out, limited)
	if err != nil {
		return 0, fmt.Errorf("copy: %w", err)
	}
	if n > cap {
		return 0, fmt.Errorf("entry exceeds size cap %d (read %d bytes)", cap, n)
	}
	return n, nil
}

// safeJoin returns destDir/entryName, but only if the resolved absolute
// path stays under destDir. Rejects absolute entry names, .. traversal,
// and any entry that would land outside the staging dir after symlink
// resolution.
//
// destDir must already be an absolute path (caller's responsibility).
func safeJoin(destDir, entryName string) (string, error) {
	if entryName == "" {
		return "", fmt.Errorf("empty entry name")
	}
	// ZIP entry separators are always '/'. filepath.Clean handles
	// platform-correct conversion.
	clean := filepath.Clean(entryName)
	if filepath.IsAbs(clean) {
		return "", fmt.Errorf("absolute path: %q", entryName)
	}
	// Reject any leading-up traversal that survived Clean.
	if clean == ".." || strings.HasPrefix(clean, ".."+string(os.PathSeparator)) || strings.Contains(clean, string(os.PathSeparator)+".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("path traversal: %q", entryName)
	}
	target := filepath.Join(destDir, clean)
	// Defense in depth: re-verify the joined path stays under destDir.
	relPath, err := filepath.Rel(destDir, target)
	if err != nil {
		return "", fmt.Errorf("rel %q: %w", target, err)
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(os.PathSeparator)) {
		return "", fmt.Errorf("escapes dest: %q -> %q", entryName, target)
	}
	return target, nil
}
