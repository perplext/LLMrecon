// v0.10.0 #174 Tier 2 — atomic-replace primitives for `update apply`.
//
// The honesty invariant: kill -9 mid-apply leaves the operator with EITHER
// the original installation OR the new one — never a half-extracted state.
//
// Strategy:
//
//   1. Stage: extract bundle into a sibling tmp directory ON THE SAME
//      FILESYSTEM as dest. This guarantees os.Rename is syscall-atomic
//      (cross-FS rename returns EXDEV).
//   2. Validate: caller-supplied check on staged contents. Manifest
//      schema, file presence, etc.
//   3. Backup: os.Rename(dest, dest+".bak."+ts). Atomic at the syscall
//      level — kernel either commits the rename or it doesn't, no
//      intermediate "partially renamed" state visible to other processes.
//   4. Apply: os.Rename(staged, dest). Atomic.
//   5. Cleanup: optional — caller decides retention.
//
// Failure paths:
//
//   - Kill -9 during steps 1-2: tmp dir on disk; original dest untouched.
//     Operator runs `update apply` again or removes tmp manually.
//   - Kill -9 between steps 3 and 4: dest is GONE, .bak exists. The
//     RecoverFromInterruptedApply helper finds and restores; operators
//     also have manual instructions in the error message.
//   - Kill -9 after step 4: success path; tmp may contain artifacts.
//
// EXDEV mitigation: tmp staging is created via os.MkdirTemp under the
// PARENT of dest, never an arbitrary global tmpdir. Renames between
// siblings on the same FS never EXDEV.
package update

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// StagedApplyOptions controls AtomicReplaceFromZip behavior.
type StagedApplyOptions struct {
	// ArchivePath is the path to the downloaded ZIP bundle.
	ArchivePath string

	// DestDir is the install directory the bundle replaces. May not
	// exist yet on a first install — the function handles both cases.
	DestDir string

	// Validate runs against the staged directory before the dest swap.
	// nil means no validation. Return non-nil error to abort the apply
	// without touching dest. The staged dir is removed on validation
	// error so partial extracts don't leak.
	Validate func(stagedDir string) error

	// KeepBackup keeps the dest.bak.<ts> directory after a successful
	// apply. Default false: backup is removed once the new dest is in
	// place. Set true if the operator passed --backup explicitly.
	KeepBackup bool
}

// AtomicReplaceResult reports what happened on disk so the caller can
// emit an accurate summary message and the test harness can assert.
type AtomicReplaceResult struct {
	// StagedBytes is the total decompressed size landed in staging.
	StagedBytes int64

	// BackupPath is the dest.bak.<ts> path. Empty when there was no
	// pre-existing dest (first install) or when KeepBackup=false and
	// the cleanup happened.
	BackupPath string

	// FirstInstall reports whether dest didn't exist before the apply.
	FirstInstall bool
}

// AtomicReplaceFromZip extracts the ZIP at opts.ArchivePath, validates
// the staged content, then atomically swaps it for opts.DestDir.
//
// Returns a non-nil error on any failure; in the failure case the
// on-disk state matches one of:
//
//   - dest unchanged, staged tmp dir cleaned up (steps 1-2 failed)
//   - dest unchanged, staged tmp dir cleaned up (validation failed)
//   - dest absent, .bak present (kill -9 between steps 3-4: caller
//     surfaces RecoveryHint to operator)
//
// Never returns nil while leaving a half-extracted dest.
func AtomicReplaceFromZip(opts StagedApplyOptions) (*AtomicReplaceResult, error) {
	if opts.ArchivePath == "" {
		return nil, fmt.Errorf("AtomicReplaceFromZip: ArchivePath required")
	}
	if opts.DestDir == "" {
		return nil, fmt.Errorf("AtomicReplaceFromZip: DestDir required")
	}
	absDest, err := filepath.Abs(opts.DestDir)
	if err != nil {
		return nil, fmt.Errorf("resolve dest: %w", err)
	}

	parent := filepath.Dir(absDest)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return nil, fmt.Errorf("mkdir parent %q: %w", parent, err)
	}

	// Stage tmp dir as a SIBLING of dest. Same filesystem → os.Rename
	// is atomic. If we used os.TempDir() this could land on /tmp which
	// is often a separate FS, and the rename would EXDEV.
	stagedDir, err := os.MkdirTemp(parent, ".llmrecon-staged-")
	if err != nil {
		return nil, fmt.Errorf("mkdir staged: %w", err)
	}
	stageOK := false
	defer func() {
		// Clean up staged dir only if we never reached the apply rename.
		// Once renamed in, the staged path is gone (not a directory we
		// own anymore) and RemoveAll on the now-stale path no-ops.
		if !stageOK {
			_ = os.RemoveAll(stagedDir)
		}
	}()

	bytes, err := ExtractZip(opts.ArchivePath, stagedDir)
	if err != nil {
		return nil, fmt.Errorf("extract: %w", err)
	}

	if opts.Validate != nil {
		if err := opts.Validate(stagedDir); err != nil {
			return nil, fmt.Errorf("validate: %w", err)
		}
	}

	result := &AtomicReplaceResult{StagedBytes: bytes}

	// Step 3: backup existing dest (if any).
	_, statErr := os.Stat(absDest)
	switch {
	case statErr == nil:
		// Dest exists — back it up.
		backupPath := absDest + ".bak." + time.Now().UTC().Format("20060102T150405Z")
		// Avoid clobbering an older .bak from a previous failed run.
		for i := 0; ; i++ {
			if _, err := os.Stat(backupPath); os.IsNotExist(err) {
				break
			}
			backupPath = fmt.Sprintf("%s.%d", absDest+".bak."+time.Now().UTC().Format("20060102T150405Z"), i+1)
		}
		if err := os.Rename(absDest, backupPath); err != nil {
			return nil, fmt.Errorf("backup rename %q -> %q: %w", absDest, backupPath, err)
		}
		result.BackupPath = backupPath
	case os.IsNotExist(statErr):
		result.FirstInstall = true
	default:
		return nil, fmt.Errorf("stat dest %q: %w", absDest, statErr)
	}

	// Step 4: apply. If this fails, dest is currently absent (we
	// renamed it to .bak above). The error message names the .bak so
	// the operator can recover with one mv.
	if err := os.Rename(stagedDir, absDest); err != nil {
		hint := ""
		if result.BackupPath != "" {
			hint = fmt.Sprintf(" — original install preserved at %q; restore with: mv %q %q", result.BackupPath, result.BackupPath, absDest)
		}
		return result, fmt.Errorf("apply rename %q -> %q: %w%s", stagedDir, absDest, err, hint)
	}
	stageOK = true // staged dir is now dest; defer cleanup is a no-op.

	// Step 5: cleanup of backup if not requested.
	if !opts.KeepBackup && result.BackupPath != "" {
		_ = os.RemoveAll(result.BackupPath)
		result.BackupPath = ""
	}

	return result, nil
}

// RecoverFromInterruptedApply scans destParent for any
// "{destBase}.bak.*" siblings of destBase. If destBase is absent and a
// .bak exists, it renames the most recent .bak back to destBase.
//
// Used by `update apply --recover` (future flag) and by tests that
// kill -9 between backup and apply rename.
//
// Returns the path it restored, or empty + nil error if no recovery
// was needed.
func RecoverFromInterruptedApply(destDir string) (string, error) {
	absDest, err := filepath.Abs(destDir)
	if err != nil {
		return "", fmt.Errorf("resolve dest: %w", err)
	}
	if _, err := os.Stat(absDest); err == nil {
		return "", nil // dest exists, no recovery needed
	} else if !os.IsNotExist(err) {
		return "", fmt.Errorf("stat dest: %w", err)
	}

	parent := filepath.Dir(absDest)
	base := filepath.Base(absDest)
	prefix := base + ".bak."

	entries, err := os.ReadDir(parent)
	if err != nil {
		return "", fmt.Errorf("read parent: %w", err)
	}

	var newest string
	var newestTS string
	for _, e := range entries {
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		ts := strings.TrimPrefix(name, prefix)
		if ts > newestTS {
			newest = filepath.Join(parent, name)
			newestTS = ts
		}
	}

	if newest == "" {
		return "", nil
	}

	if err := os.Rename(newest, absDest); err != nil {
		return "", fmt.Errorf("rename %q -> %q: %w", newest, absDest, err)
	}
	return absDest, nil
}
