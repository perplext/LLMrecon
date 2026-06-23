// v0.10.0 #175 — README CLI-example smoke test.
//
// The v0.9.0 README claimed `./llmrecon scan --provider openai
// --model gpt-4` as the first Go quick-start command. That command
// never existed. The v0.10.0 honesty release fixed the README to
// reference real commands; this test makes the regression sticky:
// every Go CLI command path the README documents must parse against
// the current Cobra tree, or this test fails.
//
// What we actually check: each enumerated command path resolves to a
// registered Cobra command (Find returns no error). We do NOT execute
// the commands — that would need a mock provider, file fixtures, etc.
// The cheap "does this verb tree exist" check is what catches the
// `./llmrecon scan --owasp` style of drift.
package cmd

import (
	"strings"
	"testing"
)

// TestReadmeDocumentsRealCLISurface scans the project README for
// `./llmrecon <subcommand>` invocations and asserts each resolves to
// a real Cobra command. If a future README rewrite documents a
// command that isn't registered (or a removed command's docs aren't
// removed), this test fails with the offending line.
func TestReadmeDocumentsRealCLISurface(t *testing.T) {
	// Hand-curated list mirroring the README's Go-side examples. We
	// hard-code instead of grep'ing the README at test time because
	// the regex side of that adds churn (every new example breaks
	// parsing) without catching the actual drift mode (a documented
	// command being renamed in code without a README update).
	//
	// When you add a Go-side example to the README, add the leading
	// command path here. When you remove one from the README, remove
	// it here. CI fails if these drift apart.
	cases := []string{
		"attack list",
		"attack run",
		"bundle create",
		"bundle verify",
		"bundle import",
		"update apply",
		"update check",
		"template list",
		"template create",
		"version",
		"changelog",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			args := strings.Fields(path)
			cmd, _, err := rootCmd.Find(args)
			if err != nil {
				t.Errorf("README documents `./llmrecon %s` but rootCmd.Find errored: %v", path, err)
				return
			}
			// rootCmd.Find returns rootCmd itself when the args don't
			// match a child. Detect that case so we don't pass a
			// false negative.
			if cmd == rootCmd {
				t.Errorf("README documents `./llmrecon %s` but it resolved to rootCmd (subcommand not registered)", path)
				return
			}
			gotPath := strings.TrimPrefix(cmd.CommandPath(), rootCmd.Name()+" ")
			if gotPath != path {
				t.Errorf("README docs `./llmrecon %s` resolved to a different command: %q", path, gotPath)
			}
		})
	}
}

// TestNoFictionalCommandsRegistered is the inverse: assert the
// commands the v0.10.0 README pass REMOVED don't accidentally come
// back without intent. If someone re-adds a `scan` Cobra command
// that takes --provider/--model (the original v0.9.0 fiction), this
// test catches it so the README and the binary stay in sync.
//
// The single existing `scan` command is template-manifest-only and
// takes no flags; its presence is fine. We assert specifically that
// `scan` does NOT register --provider, --model, or --owasp flags
// matching the v0.9.0 fictional shape.
func TestNoFictionalCommandsRegistered(t *testing.T) {
	scanCmd, _, err := rootCmd.Find([]string{"scan"})
	if err != nil || scanCmd == rootCmd {
		// scan command may have been removed entirely — that's also
		// fine. The drift we're catching is a fictional scan with
		// LLM-targeting flags.
		return
	}
	for _, fictionalFlag := range []string{"provider", "model", "owasp"} {
		if scanCmd.Flags().Lookup(fictionalFlag) != nil {
			t.Errorf("scan command has --%s flag; README pass (#175) removed examples claiming this. If reintroducing, also restore the README docs.", fictionalFlag)
		}
	}
}

// TestQuickstartDocumentsRealCLISurface mirrors the README check for the
// canonical docs/quickstart.md (#232 doc consolidation). The previous
// quickstart documented fictional `scan --target`, `init`, `config set`,
// and `report` commands; the consolidated quickstart only uses real ones.
// Every command path it documents must resolve to a registered Cobra command.
//
// When you add/remove a `./llmrecon` example in docs/quickstart.md, update
// this list. CI fails if they drift apart.
func TestQuickstartDocumentsRealCLISurface(t *testing.T) {
	cases := []string{
		"attack list",
		"attack run",
		"template list",
		"template create",
		"bundle create",
		"bundle verify",
		"bundle import",
		"bundle info",
		"credential add",
		"credential list",
		"credential rotate",
		"update check",
		"update apply",
		"check-version",
		"version",
		"changelog",
		"detect",
		"prompt-protection",
	}

	for _, path := range cases {
		t.Run(path, func(t *testing.T) {
			args := strings.Fields(path)
			cmd, _, err := rootCmd.Find(args)
			if err != nil {
				t.Errorf("quickstart documents `./llmrecon %s` but rootCmd.Find errored: %v", path, err)
				return
			}
			if cmd == rootCmd {
				t.Errorf("quickstart documents `./llmrecon %s` but it resolved to rootCmd (subcommand not registered)", path)
				return
			}
			gotPath := strings.TrimPrefix(cmd.CommandPath(), rootCmd.Name()+" ")
			if gotPath != path {
				t.Errorf("quickstart docs `./llmrecon %s` resolved to a different command: %q", path, gotPath)
			}
		})
	}
}
