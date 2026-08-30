# Clocky — Distribution, Install & Update Design

**Date:** 2026-08-29  
**Status:** Approved for planning  
**Scope:** Cross-platform install (Windows, macOS, Linux) and in-app `clocky update`

## Goal

Ship Clocky as a single binary per OS/arch via **GitHub Releases**, with:

1. One-line **install scripts** for first-time setup
2. A built-in **`clocky update`** command that upgrades in place safely
3. A repeatable **CI release pipeline** (GoReleaser)

Package managers (winget / Scoop / Homebrew) are **out of MVP** and listed as follow-ups.

## Approach

**GoReleaser + install scripts + `clocky update` (Approach A).**

Releases are the source of truth. Install scripts and `clocky update` both consume the same GitHub Release assets and checksums.

## Architecture

```text
Developer              GitHub Actions              User machine
─────────              ──────────────              ────────────
git tag v1.2.0  ──►  GoReleaser builds
                     uploads assets +
                     checksums.txt          ──►  install.ps1 / install.sh
                                                 (first install)

                     same assets            ──►  clocky update
                                                 (later upgrades)
```

### Release assets (per tag)

| Asset pattern | Example |
|---------------|---------|
| `clocky_<version>_windows_amd64.zip` | `clocky_1.2.0_windows_amd64.zip` |
| `clocky_<version>_windows_arm64.zip` | … |
| `clocky_<version>_darwin_amd64.tar.gz` | … |
| `clocky_<version>_darwin_arm64.tar.gz` | … |
| `clocky_<version>_linux_amd64.tar.gz` | … |
| `clocky_<version>_linux_arm64.tar.gz` | … |
| `checksums.txt` | SHA256 of every archive |

Each archive contains a single binary: `clocky.exe` (Windows) or `clocky` (Unix).

### Install locations (defaults)

| OS | Directory | Binary |
|----|-----------|--------|
| Windows | `%LOCALAPPDATA%\Clocky` | `clocky.exe` |
| macOS | `~/.local/bin` | `clocky` |
| Linux | `~/.local/bin` | `clocky` |

Install scripts also ensure the directory is on `PATH` (user-level), with clear printed instructions if a new shell is required.

Override via env: `CLOCKY_INSTALL_DIR`.

## First-time install

### Windows (PowerShell)

```powershell
irm https://raw.githubusercontent.com/<owner>/Clocky/master/scripts/install.ps1 | iex
```

### macOS / Linux

```bash
curl -fsSL https://raw.githubusercontent.com/<owner>/Clocky/master/scripts/install.sh | bash
```

### Script behavior

1. Detect OS + arch (`amd64` / `arm64`; reject unsupported)
2. Resolve latest stable GitHub Release (or pin with `CLOCKY_VERSION=v1.2.0`)
3. Download matching archive + `checksums.txt`
4. Verify SHA256
5. Extract binary into install dir
6. Ensure PATH entry (Windows user PATH / Unix shell note for `~/.local/bin`)
7. Print `clocky version` and next steps

Fail closed on checksum mismatch or HTTP errors.

## Version embedding

Build with ldflags so the binary knows its version:

```text
-X github.com/luisf/clocky/internal/version.Version={{.Version}}
-X github.com/luisf/clocky/internal/version.Commit={{.ShortCommit}}
-X github.com/luisf/clocky/internal/version.Date={{.Date}}
```

Commands:

- `clocky version` → prints embedded version (and commit/date if present)
- Dev builds without ldflags → `dev`

## `clocky update`

### UX

```text
clocky update           # check latest; if newer, prompt [Y/n], then upgrade
clocky update --yes     # non-interactive (CI / scripts)
clocky update --check   # only report; exit 0 if up to date, 3 if update available
```

English messages. Example:

```text
Current: v1.1.0
Latest:  v1.2.0
Update to v1.2.0? [Y/n]
Downloading clocky_1.2.0_windows_amd64.zip …
Checksum OK
Updated successfully to v1.2.0
```

If already latest: `Already up to date (v1.2.0)` exit 0.

### Algorithm

1. Read current version from `internal/version`
2. `GET https://api.github.com/repos/<owner>/Clocky/releases/latest`
3. Compare semver (ignore leading `v`); if current ≥ latest → done
4. Select asset for `runtime.GOOS` / `runtime.GOARCH`
5. Download archive + find its line in `checksums.txt`
6. Verify SHA256
7. Extract to a temp file next to the running executable
8. **Replace binary:**
   - **Unix:** write temp → `os.Rename` over current path (same filesystem)
   - **Windows:** cannot overwrite a running `.exe` → write `clocky.exe.new`, spawn a tiny helper or use rename-on-reboot pattern: rename current → `clocky.exe.old`, rename new → `clocky.exe`, then schedule delete of `.old` on next start (or delete `.old` at start of every `clocky` invocation)
9. Print success; new version used on next process (Windows: same shell after replace completes)

### Repo discovery

Default owner/repo baked in as constants (e.g. `github.com/luisf/clocky` → API `luisf/Clocky`). Override for forks/tests:

- Env `CLOCKY_GITHUB_REPO=owner/name`
- Optional flag later (not MVP)

### Security (MVP)

| Control | Required |
|---------|----------|
| HTTPS only | Yes |
| SHA256 vs `checksums.txt` from same release | Yes |
| Confirm before replace (unless `--yes`) | Yes |
| Code signing (Authenticode / notarization) | No (follow-up) |
| Auto-update in background | No |

### Errors

- Offline / API rate limit → clear English error; mention `GH_TOKEN` optional for higher limits
- Unsupported arch → list supported
- Checksum mismatch → abort, leave current binary untouched
- Permission denied writing install dir → suggest running with appropriate rights / fix PATH dir

## CI / GoReleaser

Files to add:

- `.goreleaser.yaml` — builds, archives, checksums, changelog
- `.github/workflows/release.yml` — on `push` tags `v*`
- `scripts/install.sh`, `scripts/install.ps1`
- `internal/version/version.go`
- `internal/update/update.go` (+ tests with httptest)

Tagging workflow for maintainers:

```bash
git tag v0.1.0
git push origin v0.1.0
```

Requires GitHub repo with Actions enabled and `contents: write` for the release workflow.

## CLI surface changes

| Command | Notes |
|---------|-------|
| `clocky version` | New |
| `clocky update` | New (`--yes`, `--check`) |
| `clocky help` | Document install URL + update |

## Testing

**Unit:**

- Semver compare (`v1.2.0` vs `1.2.0`)
- Asset name selection matrix (GOOS/GOARCH)
- Checksum line parsing
- Update decision: up-to-date / newer / newer-with-yes

**Integration (httptest / golden):**

- Mock GitHub API + asset download → update writes new binary path in temp dir

**Manual:**

- Tag a pre-release on a test repo or use `--check` against real API once public

## Out of scope (MVP)

- winget / Scoop / Chocolatey / Homebrew formulas
- Native `.msi` / `.pkg` / `.deb`
- Background auto-update daemon
- Code signing / Apple notarization
- Updating install scripts via `clocky update` (scripts stay on `master` raw URLs)

## Success criteria

1. Tagging `vX.Y.Z` produces GitHub Release with archives for win/mac/linux amd64+arm64 and `checksums.txt`
2. Install scripts place a working `clocky` on PATH on each OS
3. `clocky update` upgrades an older binary to latest with checksum verification
4. Failed checksum or download never corrupts the existing install
5. `clocky version` reports the release version from ldflags

## Follow-ups

1. Publish to winget / Scoop / Homebrew using the same release assets
2. Code signing
3. `clocky update --check` hook from `status` (optional nag)
