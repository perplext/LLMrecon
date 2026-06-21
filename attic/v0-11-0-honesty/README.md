# v0.11.0 honesty cleanup — atticed command scaffolding

These files were moved out of the build tree during the v0.11.0 stabilization
release. They are preserved here (as `.disabled`, so the Go toolchain ignores
them) rather than deleted, so the intent can be revived if the corresponding
features are ever finished. Full history remains in git.

| File | Issue | Why atticed |
|------|-------|-------------|
| `package.go.disabled` | #227 | 262-line orphan. Built a `packageCmd` with create/verify/apply subcommands, but `rootCmd.AddCommand(packageCmd)` was commented out (`// TODO: Uncomment when rootCmd is properly defined`), so it was never reachable from the CLI. The underlying `UpdatePackage` create/verify feature is itself a stub (`CreatePackage` writes no payloads). |
| `owasp.go.disabled` | #226 | Carried `//go:build ignore` — excluded from every build. OWASP command scaffolding that was never wired. |
| `reporting.go.disabled` | #226 | `//go:build ignore` — reporting command scaffolding, never wired. |
| `template_security.go.disabled` | #226 | `//go:build ignore` — template-security command scaffolding, never wired. |
| `sign.go.disabled` | #226 | `//go:build ignore` — signing command scaffolding, never wired (digital-signature verification remains unimplemented; see `src/update/verification.go`). |

## Reviving one of these

1. `git mv attic/v0-11-0-honesty/<name>.go.disabled src/cmd/<name>.go`
2. Remove the `//go:build ignore` line (for the #226 files) and make it compile
   against the current tree.
3. Wire it into the root command and add tests on its public surface.
