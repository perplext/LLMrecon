# Quick Start Guide

Get up and running with LLMrecon. This is the single canonical quickstart —
it consolidates what used to be spread across several overlapping guides.

> Every `llmrecon` command shown here is verified against the real CLI by a
> smoke test (`src/cmd/readme_smoke_test.go`); documented commands that don't
> exist fail CI.

## 1. Build

LLMrecon is built from source with Go (1.21+).

```bash
# Clone, then build the CLI
go build -o llmrecon ./src/main.go

# Verify
./llmrecon version
```

For the Python attack-harness components (ML optimization, Ollama test
harness), see the Python sections of `CLAUDE.md` and install:

```bash
pip install numpy pandas requests rich
```

## 2. Explore the attack modules

The attack-module ecosystem is the core of the tool.

```bash
# Enumerate every registered attack module
./llmrecon attack list

# Machine-readable form
./llmrecon attack list --json
```

## 3. Run your first attack (mock provider)

`--provider=mock` runs a module's full state machine against a deterministic
local mock — no API key, no network, no cost. Ideal for trying things out.

```bash
./llmrecon attack run \
  --module=jbfuzz \
  --provider=mock \
  --payload="<the harmful instruction to test>" \
  --metadata=allow_experimental=true
```

Useful flags (`./llmrecon attack run --help` for the full list):

- `--provider` — `mock` (default), `openai`, or `anthropic`.
- `--api-key` — provider API key; takes precedence over the provider's
  `*_API_KEY` env var.
- `--metadata key=value` — repeatable; sets safety gates (e.g.
  `allow_experimental=true`, `i_understand_risks=true`).
- `--success-indicators` — comma-separated substrings that mark a success.
- `--emit-jsonl <path>` — append the result as one JSON line for the Python
  ingest pipeline (`python -m ml.data.ingest`).

## 4. Run against a real provider

Real providers read the API key from `--api-key` (preferred) or the
`OPENAI_API_KEY` / `ANTHROPIC_API_KEY` env var.

```bash
# Flag (takes precedence)
./llmrecon attack run --module=jbfuzz --provider=openai --api-key="sk-..." \
  --payload="..." --metadata=allow_experimental=true

# Or via env var
export OPENAI_API_KEY="sk-..."
./llmrecon attack run --module=jbfuzz --provider=openai --payload="..." \
  --metadata=allow_experimental=true
```

## 5. Templates

```bash
# List vulnerability templates
./llmrecon template list

# Create a new template
./llmrecon template create
```

## 6. Bundles (air-gapped distribution)

```bash
# Create a signed bundle
./llmrecon bundle create

# Verify an extracted bundle's signature / checksums / manifest
./llmrecon bundle verify <path>

# Import an extracted bundle directory
./llmrecon bundle import <path>

# Inspect a bundle
./llmrecon bundle info <path>
```

## 7. Credentials

```bash
./llmrecon credential add
./llmrecon credential list
./llmrecon credential rotate <id>
```

## 8. Updates

```bash
./llmrecon update check     # see what's available
./llmrecon update apply     # apply available updates
./llmrecon check-version    # check the binary version against the latest
```

## Common commands

```bash
./llmrecon version          # build/version info
./llmrecon changelog        # version history
./llmrecon detect           # detect vulnerabilities in LLM responses
./llmrecon prompt-protection # manage prompt-injection protection
./llmrecon help             # full command tree
./llmrecon attack run --help # per-command help
```

## Next steps

- 📖 [User Guide](user-guide.md) — detailed usage
- 🧪 [Template Guide](template-guide.md) — writing custom templates
- 🤝 [Contributing](../CONTRIBUTING.md)

## Need help?

- 🐛 [Report issues](https://github.com/perplext/LLMrecon/issues)
- 📚 Browse the rest of `docs/`
