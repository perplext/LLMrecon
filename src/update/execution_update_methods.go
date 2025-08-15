package update

import (
	"context"
	"fmt"
	"runtime"
	
	"github.com/perplext/LLMrecon/src/version"
)

// executeFullUpdate executes a full update from the package
func (e *UpdateExecutor) executeFullUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	e.Logger.Info("UpdateExecutor", "Executing full update", transaction.ID, map[string]interface{}{
		"package_id": pkg.Manifest.PackageID,
		"version":    pkg.Manifest.Components.Binary.Version,
	})

	// Update binary if included
	if pkg.Manifest.Components.Binary.Version != "" {
		if err := e.executeBinaryUpdate(ctx, pkg, transaction); err != nil {
			return fmt.Errorf("failed to update binary: %w", err)
		}
	}

	// Update templates if included
	if pkg.Manifest.Components.Templates.Version != "" {
		if err := e.executeTemplatesUpdate(ctx, pkg, transaction); err != nil {
			return fmt.Errorf("failed to update templates: %w", err)
		}
	}
	// Update modules if included
	if len(pkg.Manifest.Components.Modules) > 0 {
		if err := e.executeModulesUpdate(ctx, pkg, transaction); err != nil {
			return fmt.Errorf("failed to update modules: %w", err)
		}
	}

	return nil

// executeDifferentialUpdate executes a differential update from the package
func (e *UpdateExecutor) executeDifferentialUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	e.Logger.Info("UpdateExecutor", "Executing differential update", transaction.ID, map[string]interface{}{
		"package_id": pkg.Manifest.PackageID,
		"version":    pkg.Manifest.Components.Binary.Version,
	})

	// Update binary if included
	if len(pkg.Manifest.Components.Patches.Binary) > 0 {
		if err := e.executeBinaryPatchUpdate(ctx, pkg, transaction); err != nil {
			return fmt.Errorf("failed to update binary: %w", err)
		}
	}

	// Update templates if included
	if len(pkg.Manifest.Components.Patches.Templates) > 0 {
		if err := e.executeTemplatesPatchUpdate(ctx, pkg, transaction); err != nil {
			return fmt.Errorf("failed to update templates: %w", err)
		}
	}

	// Update modules if included
	if len(pkg.Manifest.Components.Patches.Modules) > 0 {
		if err := e.executeModulesPatchUpdate(ctx, pkg, transaction); err != nil {
			return fmt.Errorf("failed to update modules: %w", err)
		}
	}

	return nil

// executeBinaryUpdate executes a binary update
func (e *UpdateExecutor) executeBinaryUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	// Get platform
	platform := fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)

	// Get binary path in package
	binaryPath := pkg.GetBinaryPath(platform)

	// Get binary path in installation
	installBinaryPath := filepath.Join(e.InstallDir, "bin", filepath.Base(binaryPath))

	// Get backup path
	backupBinaryPath := filepath.Join(transaction.BackupDir, "bin", filepath.Base(binaryPath))

	// Add operation to transaction
	operation := transaction.AddOperation(
		BinaryUpdateComponent,
		"",
		binaryPath,
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
			"version": pkg.Manifest.Components.Binary.Version,
		},
	)

	// Update current version
	binaryVersion, err := version.ParseVersion(pkg.Manifest.Components.Binary.Version)
	if err != nil {
		return fmt.Errorf("failed to parse binary version: %w", err)
	}
	e.CurrentVersions["binary"] = binaryVersion

	return nil

// executeTemplatesUpdate executes a templates update
func (e *UpdateExecutor) executeTemplatesUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	// Get templates path in package
	templatesPath := pkg.GetTemplatesPath()

	// Get templates path in installation
	installTemplatesPath := filepath.Join(e.InstallDir, "templates")

	// Get backup path
	backupTemplatesPath := filepath.Join(transaction.BackupDir, "templates")
	// Add operation to transaction
	operation := transaction.AddOperation(
		TemplatesUpdateComponent,
		"",
		templatesPath,
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
			"version": pkg.Manifest.Components.Templates.Version,
		},
	)

	// Update current version
	templatesVersion, err := version.ParseVersion(pkg.Manifest.Components.Templates.Version)
	if err != nil {
		return fmt.Errorf("failed to parse templates version: %w", err)
	}
	e.CurrentVersions["templates"] = templatesVersion

	return nil

// executeModulesUpdate executes a modules update
func (e *UpdateExecutor) executeModulesUpdate(ctx context.Context, pkg *UpdatePackage, transaction *UpdateTransaction) error {
	// Update each module
	for _, moduleInfo := range pkg.Manifest.Components.Modules {
		// Get module path in package
		modulePath := pkg.GetModulePath(moduleInfo.ID)

		// Get module path in installation
		installModulePath := filepath.Join(e.InstallDir, "modules", moduleInfo.ID)

		// Get backup path
		backupModulePath := filepath.Join(transaction.BackupDir, "modules", moduleInfo.ID)

		// Add operation to transaction
		operation := transaction.AddOperation(
			ModuleUpdateComponent,
			moduleInfo.ID,
			modulePath,
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
			moduleInfo.ID,
			map[string]interface{}{
				"version": moduleInfo.Version,
			},
		)

		// Update current version
		moduleVersion, err := version.ParseVersion(moduleInfo.Version)
		if err != nil {
			return fmt.Errorf("failed to parse module %s version: %w", moduleInfo.ID, err)
		}
		e.CurrentVersions[fmt.Sprintf("module.%s", moduleInfo.ID)] = moduleVersion
	}

