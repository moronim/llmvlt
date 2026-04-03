# llmvlt

A CLI secret manager built for AI/ML engineers.

`llmvlt` knows what `OPENAI_API_KEY` is. It knows what `ANTHROPIC_API_KEY` looks like. It warns you when a key format is wrong, tracks which keys were active during an experiment, and injects them safely into your scripts — without ever touching your shell history.

```
$ llmvlt init --preset openai-stack
✓ Vault initialized (.llmvlt.store)
  Preset: openai-stack — OpenAI API credentials

$ llmvlt set OPENAI_API_KEY sk-proj-abc123...
✓ OPENAI_API_KEY format looks valid
✓ Secret OPENAI_API_KEY saved

$ llmvlt run -- python train.py
```

---

## Why not just use `.env` files?

`.env` files get committed. API keys get leaked. It happens to everyone.

`llmvlt` keeps your secrets in an AES-256 encrypted vault on disk — never in plaintext, never in your shell history, never accidentally pushed to GitHub. And unlike generic secret managers, it understands the specific keys that AI/ML engineers use every day.

---

## Installation

### From source (requires Go 1.25+)

```bash
go install github.com/moronim/llmvlt@latest
```

### From GitHub releases

Download the latest binary for your platform from the [releases page](https://github.com/moronim/llmvlt/releases) and add it to your PATH.

---

## Getting Started

**1. Initialize a vault in your project directory:**

```bash
llmvlt init --preset openai-stack
```

**2. Set your secrets:**

```bash
llmvlt set OPENAI_API_KEY sk-...
llmvlt set OPENAI_ORG_ID org-...
```

**3. Run your script with secrets injected:**

```bash
llmvlt run -- python train.py
```

Secrets are injected into the subprocess only — they never appear in your parent shell, in `ps aux` output, or in your shell history.

---

## Presets

Presets are the core feature that makes `llmvlt` different. Each preset knows which keys a provider uses, what they look like, and how often they should be rotated.

| Preset | Description |
|---|---|
| `openai-stack` | OpenAI API credentials |
| `anthropic-stack` | Anthropic (Claude) API credentials |
| `huggingface-stack` | Hugging Face credentials |
| `replicate-stack` | Replicate API credentials |
| `wandb-stack` | Weights & Biases credentials |
| `langchain-stack` | LangChain / LangSmith credentials |
| `together-stack` | Together AI credentials |
| `mistral-stack` | Mistral AI credentials |
| `google-ai-stack` | Google AI (Gemini) credentials |
| `cohere-stack` | Cohere API credentials |
| `full-llm-stack` | All LLM provider credentials combined |
| `mlops-stack` | ML experiment tracking & ops (W&B + LangSmith) |

List all available presets:

```bash
llmvlt presets
```

---

## Commands

### `llmvlt init`

Initialize a new vault in the current directory.

```bash
llmvlt init                          # empty vault
llmvlt init --preset openai-stack    # with a preset
```

### `llmvlt set`

Store a secret. For known provider keys, the format is validated and the operation is **blocked** if it doesn't match. Use `--force` to override.

```bash
llmvlt set OPENAI_API_KEY sk-...
llmvlt set HF_TOKEN hf_...

# If the format is wrong:
llmvlt set OPENAI_API_KEY wrong-value
# ✗ Invalid format. OPENAI_API_KEY: Should start with 'sk-' or 'sk-proj-'. Use --force to store anyway

# Override with --force:
llmvlt set OPENAI_API_KEY unusual-key --force
# ⚠ OPENAI_API_KEY: Should start with 'sk-' or 'sk-proj-' — stored despite format mismatch (--force)
```

### `llmvlt get`

Retrieve a secret value.

```bash
llmvlt get OPENAI_API_KEY
```

### `llmvlt list`

List all secret names stored in the vault. Never prints values.

```bash
llmvlt list
```

### `llmvlt run`

Inject all secrets into a subprocess. The hero command.

```bash
llmvlt run -- python train.py
llmvlt run -- jupyter notebook
llmvlt run -- pytest tests/
```

Secrets are available as environment variables inside the subprocess. They are never exposed to the parent shell.

### `llmvlt inject`

Export secrets in different formats for use in other contexts.

```bash
# Inject into current shell session
eval $(llmvlt inject)

# Generate a Jupyter cell to paste into a notebook
llmvlt inject --format jupyter

# Write a .env file (use with caution)
llmvlt inject --format dotenv --out .env
```

### `llmvlt check`

Check for keys that haven't been rotated recently.

```bash
llmvlt check

# Example output:
# ⚠  OPENAI_API_KEY — last rotated 94 days ago (recommended: 90 days)
# ✓  ANTHROPIC_API_KEY — rotated 12 days ago
# ✓  HF_TOKEN — rotated 45 days ago
```

### `llmvlt use`

Switch between providers when benchmarking the same script across multiple LLMs.

```bash
llmvlt use anthropic    # activates Anthropic keys
llmvlt use openai       # switches to OpenAI keys
```

### `llmvlt history`

Show a log of commands executed via `llmvlt run`, including which secrets were active and any experiment tags.

```bash
llmvlt run --tag "gpt4-baseline" -- python eval.py
llmvlt run --tag "claude-comparison" -- python eval.py

llmvlt history
# 2026-03-12 10:00:00  python eval.py [claude-comparison]
#            ↳ ANTHROPIC_API_KEY
# 2026-03-11 18:22:00  python eval.py [gpt4-baseline]
#            ↳ OPENAI_API_KEY

llmvlt history --last 5    # show only last 5 entries
```

---

## Building from Source

You need [Go 1.25+](https://go.dev/dl/) installed.

```bash
git clone https://github.com/moronim/llmvlt
cd llmvlt
make build
```

### Release binaries

The Makefile includes a `release` target that cross-compiles for all supported platforms:

```bash
make release
```

This produces:
- `bin/llmvlt-linux-amd64`
- `bin/llmvlt-linux-arm64`
- `bin/llmvlt-macos-arm64`
- `bin/llmvlt-windows-amd64.exe`

To publish a GitHub release:

```bash
./release.sh v0.1.0
```

---

## Security Model

- Secrets are encrypted with **AES-256** using a key derived from your master password via **Argon2id**
- The vault is a single encrypted file (`.llmvlt.store`) — never plaintext
- `llmvlt run` forks a subprocess: secrets are injected into the child process only, never the parent shell
- Secret values are never written to shell history
- File permissions on the store are set to `0600` (owner read/write only)

Inspired by [scrt](https://github.com/loderunner/scrt).

---

## Roadmap

- [ ] Cloud sync across machines
- [ ] Team vaults with access control
- [ ] Key rotation automation
- [ ] GitHub Actions integration
- [ ] More presets (Groq, DeepSeek, etc.)

---

## License

Apache 2.0