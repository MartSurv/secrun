# secrun

Secure environment variable runner. Keep secrets out of your project directories.

## Problem

AI coding tools (Claude Code, Cursor, etc.) mount your project directory into sandboxed containers. Any `.env` file in your project is readable by the AI. **secrun** stores secrets outside your project in encrypted vaults and injects them at runtime.

## Install

**Homebrew:**

```sh
brew install MartSurv/tap/secrun
```

**Shell script:**

```sh
curl -sSL https://raw.githubusercontent.com/MartSurv/secrun/main/install.sh | sh
```

**Go:**

```sh
go install github.com/MartSurv/secrun@latest
```

## Quick Start

```sh
# Initialize a project vault
cd ~/my-project
secrun init

# Import secrets from .env.example (prompts for each value)
secrun import

# Or set secrets individually
secrun set STRIPE_KEY
# Enter value: sk_live_xxx (hidden input)

# Run your app with secrets injected
secrun run -- yarn dev
```

Secrets are never written to your project directory. They only exist in encrypted storage and in your app's process memory at runtime.

## Commands

| Command | Description |
|---|---|
| `secrun init [project]` | Create a new project vault |
| `secrun set [project] <KEY> [VALUE]` | Store a secret (prompts if VALUE omitted) |
| `secrun get [project] <KEY>` | Print a secret value |
| `secrun list [project]` | List secret names |
| `secrun delete [project] <KEY>` | Remove a secret |
| `secrun import [project]` | Import from .env.example (interactive) |
| `secrun import [project] --from .env` | Import from a file |
| `secrun export [project]` | Print all as KEY=VALUE |
| `secrun run [project] -- <cmd>` | Run command with secrets injected (`--` is optional) |
| `secrun projects` | List all projects |

The `[project]` argument is optional — secrun infers it from the current directory name.

## Storage

Secrets are stored in AES-256-GCM encrypted vault files at `~/.config/secrun/vaults/`. Works on macOS and Linux.

On macOS, `secrun init` offers to save the master password to Keychain — so subsequent commands never prompt for it.

## Session Caching

On first `secrun run`, you enter your master password (or it's loaded from Keychain on macOS). A background daemon caches decrypted secrets in memory for 4 hours (configurable with `--ttl`). Subsequent runs don't prompt.

The daemon uses an authenticated protocol — other processes (including sandboxed AI tools) cannot retrieve cached secrets.

## Security

- Secrets stored with AES-256-GCM encryption with authenticated associated data (Argon2id key derivation, OWASP-compliant parameters)
- Vault files use atomic writes with file locking to prevent corruption
- Symlink protection on all file operations
- Session daemon requires auth token on every request
- Daemon prevents memory from being swapped to disk (mlockall)
- Core dumps disabled in daemon process
- Daemon auth token passed via stdin (not visible in `ps` or `/proc`)
- Master password optionally saved to macOS Keychain (never stored as plaintext on disk)
- Secrets never appear in process arguments or shell history (use `secrun set KEY` without value)

## License

MIT
