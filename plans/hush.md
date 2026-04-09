# hush 🤫 — Implementation Plan

> Your secrets deserve silence.

A Go reimplementation of [psst](https://github.com/Michaelliv/psst), an AI-native secrets manager.

---

## Overview

`hush` is a CLI tool that lets AI agents *use* secrets without *seeing* them. Secrets are
encrypted at rest (AES-256-GCM), keyed via the OS keychain, and injected into subprocesses
at runtime — the agent only receives the exit code and (redacted) output.

---

## Architecture

```
cmd/hush/          ← main entry-point
internal/
  vault/           ← Vault struct: SQLite + AES-256-GCM CRUD
  crypto/          ← encrypt / decrypt helpers
  keychain/        ← OS keychain abstraction
  commands/        ← one file per CLI command
  scanner/         ← secret-leak scanner
```

---

## Tech Stack

| Concern | Library |
|---------|---------|
| CLI framework | `github.com/spf13/cobra` |
| SQLite (no CGo) | `modernc.org/sqlite` |
| OS keychain | `github.com/99designs/keyring` |
| Env parsing | stdlib + simple parser |
| Testing | stdlib `testing` + `github.com/stretchr/testify` |

---

## Implementation Checklist

- [x] Create `/plans/hush.md`
- [ ] Initialize Go module (`go mod init github.com/nitzzzu/hush`)
- [ ] Implement `internal/crypto` — AES-256-GCM encrypt/decrypt helpers
- [ ] Implement `internal/keychain` — OS keychain get/set/delete (with `HUSH_PASSWORD` fallback)
- [ ] Implement `internal/vault` — Vault struct backed by SQLite
  - [ ] `InitVault(path)` — create DB + schema (secrets table, history table)
  - [ ] `SetSecret(name, value, tags)` — encrypt + upsert, archive previous version
  - [ ] `GetSecret(name)` — decrypt and return
  - [ ] `ListSecrets(tags...)` — list names (optionally filtered by tags)
  - [ ] `DeleteSecret(name)` — remove secret
  - [ ] `GetHistory(name)` — return version history
  - [ ] `Rollback(name, version)` — restore previous version
  - [ ] `AddTag / RemoveTag` — tag management
  - [ ] `AllSecrets(tags...)` — return all decrypted name→value map for injection
  - [ ] Vault path resolution (local `.hush/`, global `~/.hush/`, envs `envs/<name>/`)
- [ ] Implement `cmd/hush` CLI scaffold with global flags
  - [ ] `--global / -g`, `--env`, `--tag`, `--json`, `--quiet / -q`
  - [ ] `HUSH_GLOBAL`, `HUSH_ENV` env var support
- [ ] Implement `init` command
- [ ] Implement `set` command (interactive prompt, `--stdin`, `--tag`)
- [ ] Implement `get` command
- [ ] Implement `list` command (with `--tag` filter, `--json`)
- [ ] Implement `list envs` sub-command
- [ ] Implement `rm` command
- [ ] Implement `tag` / `untag` commands
- [ ] Implement `history` command
- [ ] Implement `rollback` command
- [ ] Implement `import` command (from file, `--stdin`, `--from-env`)
- [ ] Implement `export` command (stdout / `--env-file`)
- [ ] Implement `run` command — inject ALL secrets, mask output
- [ ] Implement `exec` command — `hush SECRET [SECRET...] -- cmd` pattern, mask output
- [ ] Implement `scan` command — scan tracked/staged/path files for leaked secret values
- [ ] Write tests for `internal/crypto`
- [ ] Write tests for `internal/vault`
- [ ] Write tests for `commands/exec`
- [ ] Write tests for `commands/import`
- [ ] Write tests for `commands/export`
- [ ] Write README.md
- [ ] Add GitHub Actions CI workflow

---

## Security Model

- Secrets encrypted at rest with AES-256-GCM
- Per-vault random 256-bit encryption key
- Encryption key stored in OS keychain (macOS Keychain, libsecret, Windows Credential Manager)
- `HUSH_PASSWORD` env var as fallback for headless/CI environments
- Secrets automatically redacted as `[REDACTED]` in command output
- `--no-mask` flag for debugging
- Secrets never printed to stdout except via explicit `hush get`

---

## Command Reference

```
hush init                              Create local vault (.hush/)
hush init --global                     Create global vault (~/.hush/)
hush init --env <name>                 Create vault for specific environment

hush set <NAME>                        Set secret (interactive prompt)
hush set <NAME> --stdin                Set secret from stdin
hush set <NAME> --tag <t>              Set secret with tags (repeatable)
hush get <NAME>                        Get secret value
hush list                              List secret names
hush list --tag <t>                    List secrets filtered by tag
hush list envs                         List available environments
hush rm <NAME>                         Remove secret
hush tag <NAME> <t1> [t2...]           Add tags to secret
hush untag <NAME> <t1>...              Remove tags from secret
hush history <NAME>                    Show version history
hush rollback <NAME> --to N            Restore secret to version N

hush import <file>                     Import from .env file
hush import --stdin                    Import from stdin
hush import --from-env                 Import from environment variables
hush export                            Export to stdout (.env format)
hush export --env-file <f>             Export to file

hush run <command>                     Run command with ALL secrets injected
hush <NAME> [NAME...] -- <cmd>         Inject specific secrets and run
hush --tag <t> -- <cmd>                Inject secrets with tag and run

hush scan                              Scan tracked files for leaked secrets
hush scan --staged                     Scan only git staged files
hush scan --path <dir>                 Scan specific directory
```
