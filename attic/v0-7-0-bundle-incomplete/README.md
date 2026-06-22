# Atticked v0.7.0-era incomplete delta-bundle code

`delta.go.disabled` (was `src/bundle/delta.go`) implemented delta/patch
bundle updates against an early draft of the bundle pipeline. It was
**dead by induction** and is preserved here rather than deleted.

## Why it was dead (#228)

The patch-application path could never run:

- `applyPatchOperation` (called from `ApplyOperations` on `OperationPatch`
  operations) requires a delta bundle that contains patch operations.
- Producing such a bundle goes through `CompressDelta`, which was a no-op
  stub (`// TODO: Implement tar + gzip compression of delta directory` at
  the old `delta.go:540`). With no way to *create* a delta bundle, the
  patch path was unreachable.

## Verification before atticking

Beyond the patch-path argument, a whole-file reachability check confirmed
the file was entirely dead: none of its exported entry points or types
(`GenerateDelta`, `LoadDeltaBundle`, `PrepareUpdate`, `ApplyOperations`,
`RollbackUpdate`, `CreateBackup`, `DeltaBundle`, `DeltaManifest`, …) were
referenced anywhere outside `delta.go`. (The `CreateBackup` /
`RollbackUpdate` matches elsewhere are unrelated *methods* on other types
in the `bundle` and `update` packages, not calls to these package-level
funcs.) Moving the file out and running `go build ./...` compiles cleanly,
which is the definitive proof that nothing depended on it.

## Disposition

Atticked to `attic/v0-7-0-bundle-incomplete/` with a `.disabled`
extension so it is not compiled, following the v0.10.0 #177 /
`attic/v0-7-0-bundle-disabled/` pattern. If delta/patch bundle updates are
wanted in the future, this should be rebuilt fresh against the current
`src/bundle` API rather than revived as-is.
