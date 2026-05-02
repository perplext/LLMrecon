# JBFuzz seed corpus

Starter seeds for `src/attacks/adaptive/jbfuzz.go`. Per JBFuzz §4.1
(arXiv 2503.08990), labs train heavily against the public DAN-family
templates, so seeds here use **psychological-theme** scaffolds:
assumed-responsibility, character-roleplay, scenario-immersion,
authority-deferral, and similar. Each seed is a single-prompt JSON
blob; the engine mutates the `prompt` field via synonym substitution,
refusal-suppression prefix, and rephrase ops to discover model-specific
bypasses.

Schema (one file per seed):

    {
      "id":       "stable-id",
      "name":     "human-readable",
      "category": "psychological-theme",
      "prompt":   "...",
      "indicators": ["...", "..."]
    }

`indicators` are strings that, if present in a model response, suggest
the seed succeeded *before* mutation; the engine uses them as the
absence-of-refusal signal during fitness scoring (refusal-heuristic
default). Embedding-based fitness is opt-in via the engine config and
does not require this field.

Add new seeds by dropping a JSON file in this directory. The loader
glob's `*.json` (this README is ignored).
