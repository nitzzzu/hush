```
██╗  ██╗██╗   ██╗███████╗██╗  ██╗
██║  ██║██║   ██║██╔════╝██║  ██║
███████║██║   ██║███████╗███████║
██╔══██║██║   ██║╚════██║██╔══██║
██║  ██║╚██████╔╝███████║██║  ██║
╚═╝  ╚═╝ ╚═════╝ ╚══════╝╚═╝  ╚═╝
```

> Your secrets deserve silence



https://github.com/user-attachments/assets/dabc42ac-29f1-4853-b09c-e970ce831c79



**hush** is a CLI secrets manager for AI-assisted workflows. It encrypts secrets at rest with AES-256-GCM, stores the encryption key in your OS keychain, and injects secrets into subprocesses at runtime, so AI agents can use your credentials without seeing their values.

A Go reimplementation of [psst](https://github.com/Michaelliv/psst).

---

## Why hush?

When you run an AI agent, you often need to give it access to API keys, database URLs, or other credentials. Pasting them into a prompt or `.env` file leaks your secrets into model context, logs, or version control.

hush keeps secrets out of that path:

```bash
# The agent runs with DATABASE_URL injected, but never sees the value
hush run -- python manage.py migrate
```

Secrets live encrypted on your machine. Your OS keychain holds the key. Nothing leaves your device.

---

## Installation

```bash
# Clone and build
git clone https://github.com/nitzzzu/hush
cd hush
go build -o hush ./cmd/hush/

# Move to your PATH
mv hush /usr/local/bin/hush
```

Requires Go 1.21+.

---

## Quick Start

```bash
# Initialize a vault in the current project
hush init

# Store a secret
hush set DATABASE_URL "postgresql://user:pass@localhost/mydb"

# Run your app with secrets injected
hush run -- npm start
```

---

## Commands

### Vault

```bash
hush init                     # Initialize local vault (.hush/vault.db)
hush init --global            # Initialize global vault (~/.hush/vault.db)
hush init --env staging       # Initialize named environment vault
```

### Secrets

```bash
# Set
hush set API_KEY "sk-abc123"
hush set API_KEY                          # interactive prompt (hidden input)
cat key.pem | hush set TLS_CERT -f -     # read from stdin

# Get
hush get API_KEY                          # masked: ********
hush get API_KEY --reveal                 # prints plaintext
hush get API_KEY --json                   # full metadata as JSON

# List
hush list                                 # all secrets
hush list --tag prod                      # filter by tag
hush list --json

# Delete
hush rm OLD_SECRET
```

### Tags

```bash
hush set API_KEY "sk-abc123" --tag prod --tag api
hush tag DATABASE_URL production
hush untag DATABASE_URL staging
```

### Subprocess Injection

```bash
# Inject all secrets
hush run -- npm start
hush run -- python manage.py runserver

# Inject only secrets with a specific tag
hush --tag prod run -- ./deploy.sh
hush --tag database run -- alembic upgrade head
```

### Version History

```bash
hush history API_KEY           # list all versions
hush rollback API_KEY 1        # restore to version 1
```

Up to 10 versions are retained per secret.

### Import / Export

```bash
# Import
hush import -f .env.local
cat .env | hush import
hush import -f .env --overwrite    # overwrite existing secrets

# Export
hush export                        # dotenv format to stdout
hush export -f secrets.env
hush export --format shell         # export KEY="value"
```

### Leak Scanning

```bash
hush scan                   # scan all git-tracked files
hush scan --staged          # scan only staged files (pre-commit hook)
hush scan -d ./src          # scan a specific directory
hush scan --json
```

Exits with code `1` if secret values appear in scanned files. Use it as a pre-commit guard.

### Environments

```bash
hush --env staging set API_KEY "sk-staging-xyz"
hush --env prod set API_KEY "sk-prod-abc"
hush --env staging run -- ./integration-test.sh
hush env list
```

---

## Global Flags

| Flag | Description |
|---|---|
| `--global, -g` | Use global vault (`~/.hush/`) instead of local |
| `--env <name>` | Target a named environment vault |
| `--tag <tag>` | Filter secrets by tag (repeatable, OR logic) |
| `--json` | JSON output |
| `--quiet, -q` | Suppress output, use exit codes only |

---

## Environment Variables

| Variable | Description |
|---|---|
| `HUSH_GLOBAL=1` | Force global vault |
| `HUSH_ENV=<name>` | Select environment |
| `HUSH_PASSWORD=<pwd>` | Password-derived key (CI/headless mode) |

---

## CI / Headless Usage

In environments without a keychain (Docker, GitHub Actions, etc.), use `HUSH_PASSWORD`:

```bash
HUSH_PASSWORD=my-ci-password hush run -- ./deploy.sh
```

The key is derived from the password and vault path using SHA-256. Store `HUSH_PASSWORD` as a CI secret, not the individual credentials.

---

## Pre-commit Hook

Prevent accidental secret commits:

```bash
# .git/hooks/pre-commit
#!/bin/sh
hush scan --staged --quiet || { echo "hush: secret leak detected"; exit 1; }
```

---

## Security

- **Encryption:** AES-256-GCM (authenticated encryption, nonce per ciphertext)
- **Key storage:** OS keychain: macOS Keychain, Linux libsecret/KWallet, Windows Credential Manager
- **Local-first:** No cloud sync, no telemetry, no network calls
- **Subprocess isolation:** Secrets visible only inside the child process environment
- **Output masking:** Default display hides values; `--reveal` required for plaintext
- **Version limit:** Max 10 history entries per secret (oldest auto-pruned)

---

## Vault Layout

```
.hush/
├── vault.db              # local vault (SQLite, AES-256-GCM encrypted values)
└── envs/
    ├── staging/
    │   └── vault.db
    └── prod/
        └── vault.db

~/.hush/
└── vault.db              # global vault
```

---

## License

MIT
