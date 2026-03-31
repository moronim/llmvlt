# aikeys

A CLI secret manager built for AI/ML engineers.

<!-- TODO: Record and add animated GIF demo here -->
<!-- ![aikeys demo](docs/demo.gif) -->

**aikeys** knows your providers — OpenAI, Anthropic, Hugging Face, Replicate, and more. It validates key formats, injects secrets into your scripts, and tracks which experiments used which key versions.

Stop leaking API keys in notebooks. Stop juggling `.env` files across projects.

## Install

```bash
# From source
go install github.com/moronim/aikeys@latest

# Or build locally
git clone https://github.com/moronim/aikeys.git
cd aikeys && make install
```

## Quick Start

```bash
# Initialize a vault with a preset
aikeys init --preset openai-stack

# Set your keys
aikeys set OPENAI_API_KEY sk-proj-abc123...

# Run a script with secrets injected (never touches your shell)
aikeys run -- python train.py

# Tag experiment runs for reproducibility
aikeys run --tag "gpt4-baseline" -- python eval.py
```

## Core Commands

| Command | Description |
|---------|-------------|
| `aikeys init [--preset NAME]` | Create a new vault, optionally with a provider preset |
| `aikeys set KEY [VALUE]` | Store a secret (reads from stdin if no value given) |
| `aikeys get KEY` | Retrieve a secret |
| `aikeys list` | List all keys (values never shown) |
| `aikeys run -- CMD` | Run a command with secrets injected |
| `aikeys inject [--format FMT]` | Output secrets as shell exports, .env, or Jupyter cell |
| `aikeys check` | Validate formats, check for empty keys, rotation reminders |
| `aikeys use PROVIDER` | Switch active provider context for benchmarking |
| `aikeys history` | View experiment run log |
| `aikeys presets` | List all available provider presets |

## Provider Presets

Presets scaffold the exact keys each provider needs, with format validation built in.

| Preset | Secrets |
|--------|---------|
| `openai-stack` | `OPENAI_API_KEY`, `OPENAI_ORG_ID`, `OPENAI_PROJECT_ID` |
| `anthropic-stack` | `ANTHROPIC_API_KEY` |
| `huggingface-stack` | `HF_TOKEN`, `HF_HOME` |
| `replicate-stack` | `REPLICATE_API_TOKEN` |
| `wandb-stack` | `WANDB_API_KEY`, `WANDB_PROJECT`, `WANDB_ENTITY` |
| `langchain-stack` | `LANGCHAIN_API_KEY`, `LANGCHAIN_TRACING_V2`, `LANGCHAIN_ENDPOINT` |
| `together-stack` | `TOGETHER_API_KEY` |
| `mistral-stack` | `MISTRAL_API_KEY` |
| `google-ai-stack` | `GOOGLE_API_KEY` |
| `cohere-stack` | `COHERE_API_KEY` |
| `full-llm-stack` | All of the above combined |
| `mlops-stack` | W&B + LangChain |

## The `run` Command

This is the hero feature. It injects secrets into a child process only — they never appear in your parent shell, shell history, or `ps` output.

```bash
# Simple usage
aikeys run -- python train.py

# Tag for experiment tracking
aikeys run --tag "claude-3.5-eval" -- python benchmark.py

# View history
aikeys history
# 2026-03-15 14:30:00  python benchmark.py [claude-3.5-eval]
#            ↳ ANTHROPIC_API_KEY
```

## Injection Formats

```bash
# Shell (eval in bash/zsh)
eval $(aikeys inject)

# Dotenv file
aikeys inject --format dotenv -o .env

# Jupyter notebook cell
aikeys inject --format jupyter
# Output:
# import os
# os.environ['OPENAI_API_KEY'] = 'sk-...'
```

## Provider Switching

Benchmarking the same script across providers:

```bash
aikeys use anthropic
aikeys run -- python eval.py   # Only Anthropic keys injected

aikeys use openai
aikeys run -- python eval.py   # Only OpenAI keys injected

aikeys use all                 # Re-enable everything
```

## Security

- **AES-256-GCM** encryption at rest
- **Argon2id** key derivation (resistant to GPU attacks)
- Secrets injected into child process only via `run` — never in parent shell
- Vault file permissions set to `0600` (owner read/write only)
- Secrets never written to shell history

## Why not just use...

| Tool | Why aikeys is different |
|------|----------------------|
| `.env` files | Unencrypted, no validation, no rotation tracking |
| 1Password / Bitwarden | Password managers, not CLI tools. No `.env` injection or provider awareness |
| HashiCorp Vault | Enterprise infrastructure. Overkill for a solo ML engineer |
| Doppler / Infisical | Generic secret managers. No concept of AI providers or experiment tracking |
| Python `keyring` | A library, not a workflow tool. No presets, no injection |

## License

Apache 2.0
