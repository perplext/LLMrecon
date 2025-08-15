package update

import (
	"context"
	"fmt"
	"runtime"
	
	"github.com/perplext/LLMrecon/src/version"
)

// executeBinaryPatchUpdate executes a binary patch update
func (e *UpdateExecutor) executeBinaryPatchUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	// Get platform
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	// Find appropriate patch
	var patchInfo *PatchInfo
	for i, patch := range pkg.Manifest.Components.Patches.Binary {
		// Check if patch is for current version
		currentVersion, ok := e.CurrentVersions["binary"]
		if !ok {
			return fmt.Errorf("current binary version not found")
		}
		fromVersion, err := version.ParseVersion(patch.FromVersion)
		if err != nil {
			return fmt.Errorf("failed to parse patch from version: %w", err)
		}
		if currentVersion.String() != fromVersion.String() {
			continue
		}

		patchInfo = &pkg.Manifest.Components.Patches.Binary[i]
		break
	}

	if patchInfo == nil {
		return fmt.Errorf("no suitable patch found for binary")
	}

	// Get patch path in package
	patchPath := pkg.GetBinaryPatchPath(platform, patchInfo.FromVersion, patchInfo.ToVersion)

	// Get binary path in installation
	installBinaryPath := filepath.Join(e.InstallDir, "bin", fmt.Sprintf("LLMrecon-%s", platform))

	// Get backup path
	backupBinaryPath := filepath.Join(transaction.BackupDir, "bin", fmt.Sprintf("LLMrecon-%s", platform))

	// Create session directory for patched binary
	sessionDir := filepath.Join(transaction.SessionDir, "bin")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory for binary: %w", err)
	}

	// Extract patch to session directory
	sessionPatchPath := filepath.Join(sessionDir, filepath.Base(patchPath))
	if err := pkg.ExtractFile(patchPath, sessionPatchPath); err != nil {
		return fmt.Errorf("failed to extract patch: %w", err)
	}
	// Create temporary file for patched binary
	sessionBinaryPath := filepath.Join(sessionDir, fmt.Sprintf("LLMrecon-%s", platform))
	if err := copyFile(installBinaryPath, sessionBinaryPath); err != nil {
		return fmt.Errorf("failed to copy binary for patching: %w", err)
	}

	// Apply patch
	if err := applyBinaryPatch(sessionPatchPath, sessionBinaryPath); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	// Add operation to transaction
	operation := transaction.AddOperation(
		BinaryUpdateComponent,
		"",
		sessionBinaryPath,
		installBinaryPath,
		backupBinaryPath,
	)

	// Execute operation
	if err := transaction.ExecuteOperation(ctx, operation); err != nil {
		return fmt.Errorf("failed to execute binary update operation: %w", err)
	}

	// Notify component updated
	e.NotificationManager.NotifyComponentUpdated(
		transaction.ID,
		pkg.Manifest.PackageID,
		"binary",
		"",
		map[string]interface{}{
			"from_version": patchInfo.FromVersion,
			"to_version":   patchInfo.ToVersion,
		},
	)

	// Update current version
	binaryVersion, err := version.ParseVersion(patchInfo.ToVersion)
	if err != nil {
		return fmt.Errorf("failed to parse binary version: %w", err)
	}
	e.CurrentVersions["binary"] = binaryVersion

	return nil

// executeTemplatesPatchUpdate executes a templates patch update
func (e *UpdateExecutor) executeTemplatesPatchUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	// Find appropriate patch
	var patchInfo *PatchInfo
	for i, patch := range pkg.Manifest.Components.Patches.Templates {
		// Check if patch is for current version
		currentVersion, ok := e.CurrentVersions["templates"]
		if !ok {
			return fmt.Errorf("current templates version not found")
		}
		fromVersion, err := version.ParseVersion(patch.FromVersion)
		if err != nil {
			return fmt.Errorf("failed to parse patch from version: %w", err)
		}
		if currentVersion.String() != fromVersion.String() {
			continue
		}

		patchInfo = &pkg.Manifest.Components.Patches.Templates[i]
		break
	}

	if patchInfo == nil {
		return fmt.Errorf("no suitable patch found for templates")
	}

	// Get patch path in package
	patchPath := pkg.GetTemplatesPatchPath(patchInfo.FromVersion, patchInfo.ToVersion)

	// Get templates path in installation
	installTemplatesPath := filepath.Join(e.InstallDir, "templates")

	// Get backup path
	backupTemplatesPath := filepath.Join(transaction.BackupDir, "templates")
	// Create session directory for patched templates
	sessionDir := filepath.Join(transaction.SessionDir, "templates")
	if err := os.MkdirAll(sessionDir, 0700); err != nil {
		return fmt.Errorf("failed to create session directory for templates: %w", err)
	}

	// Extract patch to session directory
	sessionPatchPath := filepath.Join(sessionDir, filepath.Base(patchPath))
	if err := pkg.ExtractFile(patchPath, sessionPatchPath); err != nil {
		return fmt.Errorf("failed to extract patch: %w", err)
	}

	// Create temporary directory for patched templates
	sessionTemplatesPath := filepath.Join(transaction.SessionDir, "templates-patched")
	if err := copyDir(installTemplatesPath, sessionTemplatesPath); err != nil {
		return fmt.Errorf("failed to copy templates for patching: %w", err)
	}

	// Apply patch
	if err := applyDirectoryPatch(sessionPatchPath, sessionTemplatesPath); err != nil {
		return fmt.Errorf("failed to apply patch: %w", err)
	}

	// Add operation to transaction
	operation := transaction.AddOperation(
		TemplatesUpdateComponent,
		"",
		sessionTemplatesPath,
		installTemplatesPath,
		backupTemplatesPath,
	)
	// Execute operation
	if err := transaction.ExecuteOperation(ctx, operation); err != nil {
		return fmt.Errorf("failed to execute templates update operation: %w", err)
	}

	// Notify component updated
	e.NotificationManager.NotifyComponentUpdated(
		transaction.ID,
		pkg.Manifest.PackageID,
		"templates",
		"",
		map[string]interface{}{
			"from_version": patchInfo.FromVersion,
			"to_version":   patchInfo.ToVersion,
		},
	)

	// Update current version
	templatesVersion, err := version.ParseVersion(patchInfo.ToVersion)
	if err != nil {
		return fmt.Errorf("failed to parse templates version: %w", err)
	}
	e.CurrentVersions["templates"] = templatesVersion

	return nil

// executeModulesPatchUpdate executes a modules patch update
func (e *UpdateExecutor) executeModulesPatchUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	// Update each module
	for _, modulePatch := range pkg.Manifest.Components.Patches.Modules {
		// Check if patch is for current version
		currentVersion, ok := e.CurrentVersions[fmt.Sprintf("module.%s", modulePatch.ID)]
		if !ok {
			return fmt.Errorf("current module %s version not found", modulePatch.ID)
		}
		fromVersion, err := version.ParseVersion(modulePatch.FromVersion)
		if err != nil {
			return fmt.Errorf("failed to parse patch from version: %w", err)
		}
		if currentVersion.String() != fromVersion.String() {
			continue
		}
		// Get patch path in package
		patchPath := pkg.GetModulePatchPath(modulePatch.ID, modulePatch.FromVersion, modulePatch.ToVersion)

		// Get module path in installation
		installModulePath := filepath.Join(e.InstallDir, "modules", modulePatch.ID)

		// Get backup path
		backupModulePath := filepath.Join(transaction.BackupDir, "modules", modulePatch.ID)

		// Create session directory for patched module
		sessionDir := filepath.Join(transaction.SessionDir, "modules", modulePatch.ID)
		if err := os.MkdirAll(sessionDir, 0700); err != nil {
			return fmt.Errorf("failed to create session directory for module: %w", err)
		}

		// Extract patch to session directory
		sessionPatchPath := filepath.Join(sessionDir, filepath.Base(patchPath))
		if err := pkg.ExtractFile(patchPath, sessionPatchPath); err != nil {
			return fmt.Errorf("failed to extract patch: %w", err)
		}

		// Create temporary directory for patched module
		sessionModulePath := filepath.Join(transaction.SessionDir, "modules-patched", modulePatch.ID)
		if err := copyDir(installModulePath, sessionModulePath); err != nil {
			return fmt.Errorf("failed to copy module for patching: %w", err)
		}

		// Apply patch
		if err := applyDirectoryPatch(sessionPatchPath, sessionModulePath); err != nil {
			return fmt.Errorf("failed to apply patch: %w", err)
		}

		// Add operation to transaction
		operation := transaction.AddOperation(
			ModuleUpdateComponent,
			modulePatch.ID,
			sessionModulePath,
			installModulePath,
			backupModulePath,
		)

		// Execute operation
		if err := transaction.ExecuteOperation(ctx, operation); err != nil {
			return fmt.Errorf("failed to execute module update operation: %w", err)
		}

		// Notify component updated
		e.NotificationManager.NotifyComponentUpdated(
			transaction.ID,
			pkg.Manifest.PackageID,
			"module",
			modulePatch.ID,
			map[string]interface{}{
				"from_version": modulePatch.FromVersion,
				"to_version":   modulePatch.ToVersion,
			},
		)

		// Update current version
		moduleVersion, err := version.ParseVersion(modulePatch.ToVersion)
		if err != nil {
			return fmt.Errorf("failed to parse module %s version: %w", modulePatch.ID, err)
		}
		e.CurrentVersions[fmt.Sprintf("module.%s", modulePatch.ID)] = moduleVersion
	}

