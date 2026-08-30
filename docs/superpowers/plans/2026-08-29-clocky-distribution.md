# Clocky Distribution, Install & Update Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add version embedding, `clocky update` (checksum-verified GitHub Releases upgrade), install scripts, and a GoReleaser + GitHub Actions release pipeline for Windows/macOS/Linux.

**Architecture:** `internal/version` holds ldflag-injected metadata. `internal/update` fetches the latest GitHub Release, selects the OS/arch asset, verifies SHA256, and replaces the running binary (Windows-safe rename via `.new`/`.old`). CLI gains `version` and `update`. GoReleaser produces archives + `checksums.txt`; `scripts/install.{sh,ps1}` perform first-time install into the paths from the distribution spec.

**Tech Stack:** Go 1.22+ stdlib (`net/http`, `archive/zip`, `archive/tar`, `compress/gzip`, `crypto/sha256`), GoReleaser, GitHub Actions. No new runtime module deps required for MVP (optional: none).

**Spec:** `docs/superpowers/specs/2026-08-29-clocky-distribution-design.md`

**Defaults:** GitHub repo API slug `luisf/Clocky` (override with `CLOCKY_GITHUB_REPO`). Module path remains `github.com/luisf/clocky`.

---

## File map

| Path | Responsibility |
|------|----------------|
| `internal/version/version.go` | `Version`, `Commit`, `Date` vars + `String()` |
| `internal/update/semver.go` | Normalize/compare versions |
| `internal/update/asset.go` | Archive name for GOOS/GOARCH; parse checksums |
| `internal/update/github.go` | Fetch latest release JSON + download bytes |
| `internal/update/replace_unix.go` | Atomic replace on Unix |
| `internal/update/replace_windows.go` | Windows `.new` / `.old` replace |
| `internal/update/update.go` | Orchestrate check/download/verify/replace |
| `internal/update/*_test.go` | Unit + httptest tests |
| `internal/cli/version_cmd.go` | `clocky version` |
| `internal/cli/update_cmd.go` | `clocky update` flags |
| `internal/cli/cli.go` | Route + help; cleanup `.old` on start |
| `cmd/clocky/main.go` | Honor `ExitCodeError` |
| `.goreleaser.yaml` | Multi-OS builds, archives, checksums |
| `.github/workflows/release.yml` | Tag-triggered release |
| `scripts/install.sh` | Unix installer |
| `scripts/install.ps1` | Windows installer |
| `README.md` | Install one-liners + update docs |

---

### Task 1: Version package

**Files:**
- Create: `internal/version/version.go`
- Create: `internal/version/version_test.go`

- [ ] **Step 1: Write failing test**

```go
package version

import "testing"

func TestStringDefaultDev(t *testing.T) {
	// Reset in case other tests mutate (use local assignment pattern in String tests)
	if got := format("dev", "", ""); got != "dev" {
		t.Fatalf("got %q", got)
	}
	if got := format("1.2.0", "abc1234", "2026-08-29"); got != "1.2.0 (abc1234 2026-08-29)" {
		t.Fatalf("got %q", got)
	}
}
```

- [ ] **Step 2: Implement**

```go
package version

// Set via -ldflags at release build time.
var (
	Version = "dev"
	Commit  = ""
	Date    = ""
)

func String() string {
	return format(Version, Commit, Date)
}

func format(v, commit, date string) string {
	if commit == "" && date == "" {
		return v
	}
	extra := commit
	if date != "" {
		if extra != "" {
			extra += " "
		}
		extra += date
	}
	return v + " (" + extra + ")"
}
```

- [ ] **Step 3: `go test ./internal/version/ -v` — PASS**

- [ ] **Step 4: Commit**

```bash
git add internal/version/
git commit -m "feat: add version package for ldflag injection"
```

---

### Task 2: Semver compare + asset naming + checksum parse

**Files:**
- Create: `internal/update/semver.go`
- Create: `internal/update/asset.go`
- Create: `internal/update/semver_test.go`
- Create: `internal/update/asset_test.go`

- [ ] **Step 1: Failing tests**

```go
package update

import "testing"

func TestNormalizeVersion(t *testing.T) {
	if got := NormalizeVersion("v1.2.0"); got != "1.2.0" {
		t.Fatalf("got %q", got)
	}
	if got := NormalizeVersion("1.2.0"); got != "1.2.0" {
		t.Fatalf("got %q", got)
	}
}

func TestCompareSemver(t *testing.T) {
	cases := []struct{ a, b string; want int }{
		{"1.2.0", "1.2.0", 0},
		{"v1.2.0", "1.1.9", 1},
		{"1.2.0", "v1.3.0", -1},
		{"1.2.0", "1.2.0-rc.1", 1}, // treat pre-release as less if implemented simply: prefer numeric-only MVP
	}
	for _, tc := range cases {
		// MVP: numeric major.minor.patch only; non-numeric suffix → compare as plain strings after normalize
		got := CompareSemver(tc.a, tc.b)
		if got != tc.want {
			t.Fatalf("CompareSemver(%q,%q)=%d want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestAssetName(t *testing.T) {
	got, err := AssetName("1.2.0", "windows", "amd64")
	if err != nil || got != "clocky_1.2.0_windows_amd64.zip" {
		t.Fatalf("got %q err %v", got, err)
	}
	got, err = AssetName("1.2.0", "linux", "arm64")
	if err != nil || got != "clocky_1.2.0_linux_arm64.tar.gz" {
		t.Fatalf("got %q err %v", got, err)
	}
	got, err = AssetName("1.2.0", "darwin", "amd64")
	if err != nil || got != "clocky_1.2.0_darwin_amd64.tar.gz" {
		t.Fatalf("got %q err %v", got, err)
	}
	if _, err := AssetName("1.0.0", "windows", "386"); err == nil {
		t.Fatal("expected error for 386")
	}
}

func TestParseChecksums(t *testing.T) {
	in := "abc123  clocky_1.2.0_windows_amd64.zip\ndef456  clocky_1.2.0_linux_amd64.tar.gz\n"
	m, err := ParseChecksums(in)
	if err != nil {
		t.Fatal(err)
	}
	if m["clocky_1.2.0_windows_amd64.zip"] != "abc123" {
		t.Fatalf("%v", m)
	}
}
```

**MVP CompareSemver:** split on `.`, parse up to 3 ints; missing parts = 0; ignore leading `v`. If either side is `dev`, treat current as older than any release (`CompareSemver("dev", "1.0.0") == -1`). Skip complex pre-release rules: if a segment is non-numeric, fall back to string compare of normalized forms.

- [ ] **Step 2: Implement until tests pass**

```go
func NormalizeVersion(s string) string { /* trim space, trim prefix v */ }
func CompareSemver(a, b string) int { /* -1 a<b, 0 equal, 1 a>b */ }
func AssetName(version, goos, goarch string) (string, error)
func ParseChecksums(text string) (map[string]string, error) // filename → hex digest
```

Windows/darwin/linux × amd64/arm64 only. Windows → `.zip`; others → `.tar.gz`.

- [ ] **Step 3: Commit**

```bash
git add internal/update/semver.go internal/update/asset.go internal/update/semver_test.go internal/update/asset_test.go
git commit -m "feat: add update semver, asset names, and checksum parsing"
```

---

### Task 3: GitHub client + download verify (httptest)

**Files:**
- Create: `internal/update/github.go`
- Create: `internal/update/github_test.go`
- Create: `internal/update/verify.go`

- [ ] **Step 1: Types and API**

```go
package update

type Release struct {
	TagName string `json:"tag_name"`
	Assets  []Asset `json:"assets"`
}

type Asset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

type Getter interface {
	Get(url string) (body []byte, err error)
}

type HTTPGetter struct {
	Client *http.Client // default timeout 60s
	Token  string       // optional; from GH_TOKEN or GITHUB_TOKEN
}

func (h HTTPGetter) Get(url string) ([]byte, error)

func RepoSlug() string // os.Getenv("CLOCKY_GITHUB_REPO") or "luisf/Clocky"

func LatestRelease(g Getter, repo string) (*Release, error)
// GET https://api.github.com/repos/{repo}/releases/latest
// Accept: application/vnd.github+json
// Authorization: Bearer {token} if set

func FindAssetURL(rel *Release, name string) (string, error)
func FindChecksumURL(rel *Release) (string, error) // asset named checksums.txt

func SHA256Hex(data []byte) string
func VerifySHA256(data []byte, wantHex string) error
```

- [ ] **Step 2: httptest tests** covering latest JSON parse, missing asset error, checksum verify success/fail.

- [ ] **Step 3: Commit**

```bash
git commit -m "feat: add GitHub release fetch and SHA256 verify"
```

---

### Task 4: Extract archive + replace binary

**Files:**
- Create: `internal/update/extract.go`
- Create: `internal/update/extract_test.go`
- Create: `internal/update/replace_unix.go` (`//go:build !windows`)
- Create: `internal/update/replace_windows.go` (`//go:build windows`)
- Create: `internal/update/replace_test.go` (temp dir replace round-trip; build-tag aware)

**Extract API:**

```go
// ExtractBinary reads zip or tar.gz and writes the clocky binary to destPath (full path).
func ExtractBinary(archive []byte, archiveName, destPath string) error
```

- Zip: find `clocky.exe` or `clocky` (basename match)
- Tar.gz: same
- Set file mode `0755` on Unix after write

**Replace API:**

```go
// ReplaceExecutable replaces the running binary at exePath with the file at newPath.
func ReplaceExecutable(exePath, newPath string) error
```

**Windows behavior:**

1. Destination final name = `exePath`
2. Write already done to `newPath` (e.g. `clocky.exe.new`)
3. `old := exePath + ".old"` — remove existing `.old` if present
4. `os.Rename(exePath, old)` 
5. `os.Rename(newPath, exePath)`
6. Best-effort `os.Remove(old)` (may fail if locked; leave for cleanup)

**Unix behavior:**

1. `os.Rename(newPath, exePath)` (same directory)

**Cleanup helper** (called from CLI startup):

```go
func CleanupStaleOld(exePath string) {
    _ = os.Remove(exePath + ".old")
}
```

- [ ] **Step 1: TDD** — create a tiny zip in test containing `clocky` bytes; extract; replace in temp dir.

- [ ] **Step 2: Commit**

```bash
git commit -m "feat: extract release archives and replace executable safely"
```

---

### Task 5: Update orchestrator

**Files:**
- Create: `internal/update/update.go`
- Create: `internal/update/update_test.go`

```go
type Options struct {
	CurrentVersion string
	GOOS, GOARCH   string
	ExePath        string // os.Executable()
	Getter         Getter
	Repo           string
	Yes            bool
	CheckOnly      bool
	Stdin          io.Reader
	Stdout         io.Writer
	Prompt         func(question string) (bool, error) // default: read Y/n from Stdin
}

type Result struct {
	Current string
	Latest  string
	Updated bool
	Pending bool // true when CheckOnly && update available
}

func Run(opts Options) (Result, error)
```

**Flow:**

1. Defaults: Getter HTTP, Repo `RepoSlug()`, Prompt reads line
2. `LatestRelease` → normalize tags
3. If `CompareSemver(current, latest) >= 0` → print already up to date; return
4. If `CheckOnly` → print current/latest; return `Pending: true` (CLI maps to exit 3)
5. If `!Yes` → prompt; if no, return without error
6. Resolve asset name for version **without** leading `v` in filename: use `NormalizeVersion(latest)` inside `AssetName`
7. Download checksums.txt + archive; `VerifySHA256`
8. Write archive extract to `exePath + ".new"` (or temp in same dir)
9. `ReplaceExecutable(exePath, newPath)`
10. Print success

**Test with fake Getter** returning canned release JSON + archive bytes + checksums (reuse extract test zip). Use temp `ExePath` copy of a dummy file.

- [ ] **Step 1: TDD then implement**

- [ ] **Step 2: Commit**

```bash
git commit -m "feat: orchestrate clocky self-update from GitHub Releases"
```

---

### Task 6: CLI — version, update, exit codes, cleanup

**Files:**
- Create: `internal/cli/exit.go`
- Create: `internal/cli/version_cmd.go`
- Create: `internal/cli/update_cmd.go`
- Modify: `internal/cli/cli.go`
- Modify: `cmd/clocky/main.go`

- [ ] **Step 1: Exit code helper**

```go
package cli

type ExitCodeError struct {
	Code int
	Msg  string
}

func (e *ExitCodeError) Error() string {
	if e.Msg == "" {
		return ""
	}
	return e.Msg
}
```

```go
// cmd/clocky/main.go
func main() {
	cleanupOldBinary()
	if err := cli.Run(os.Args[1:]); err != nil {
		var ec *cli.ExitCodeError
		if errors.As(err, &ec) {
			if ec.Msg != "" {
				fmt.Fprintln(os.Stderr, ec.Msg)
			}
			os.Exit(ec.Code)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func cleanupOldBinary() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	update.CleanupStaleOld(exe)
}
```

- [ ] **Step 2: Commands**

`version_cmd.go`:

```go
func runVersion(args []string) error {
	fmt.Println(version.String())
	return nil
}
```

`update_cmd.go`:

```go
func runUpdate(args []string) error {
	fs := flag.NewFlagSet("update", flag.ContinueOnError)
	yes := fs.Bool("yes", false, "update without prompting")
	check := fs.Bool("check", false, "only check for updates")
	if err := fs.Parse(args); err != nil {
		return err
	}
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	res, err := update.Run(update.Options{
		CurrentVersion: version.Version,
		GOOS:           runtime.GOOS,
		GOARCH:         runtime.GOARCH,
		ExePath:        exe,
		Yes:            *yes,
		CheckOnly:      *check,
		Stdin:          os.Stdin,
		Stdout:         os.Stdout,
	})
	if err != nil {
		return err
	}
	if *check && res.Pending {
		return &ExitCodeError{Code: 3, Msg: ""} // message already printed to Stdout
	}
	return nil
}
```

Wire in `cli.go` switch: `version`, `update`. Expand help:

```text
  clocky version
  clocky update [--yes] [--check]
```

- [ ] **Step 3: Manual smoke**

```powershell
go build -ldflags "-X github.com/luisf/clocky/internal/version.Version=0.0.1" -o clocky.exe ./cmd/clocky
.\clocky.exe version
.\clocky.exe update --check
# expect API error or real check depending on repo existence — should not panic
```

- [ ] **Step 4: Commit**

```bash
git commit -m "feat: add clocky version and update commands"
```

---

### Task 7: GoReleaser + GitHub Actions

**Files:**
- Create: `.goreleaser.yaml`
- Create: `.github/workflows/release.yml`

- [ ] **Step 1: `.goreleaser.yaml`**

```yaml
version: 2

project_name: clocky

before:
  hooks:
    - go mod tidy

builds:
  - id: clocky
    main: ./cmd/clocky
    binary: clocky
    env:
      - CGO_ENABLED=0
    goos: [windows, darwin, linux]
    goarch: [amd64, arm64]
    ldflags:
      - -s -w
      - -X github.com/luisf/clocky/internal/version.Version={{.Version}}
      - -X github.com/luisf/clocky/internal/version.Commit={{.ShortCommit}}
      - -X github.com/luisf/clocky/internal/version.Date={{.Date}}

archives:
  - id: default
    formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    files:
      - none*

checksum:
  name_template: checksums.txt
  algorithm: sha256

changelog:
  sort: asc
  filters:
    exclude:
      - "^docs:"
      - "^test:"

release:
  draft: false
  prerelease: auto
```

Note: `files: [none*]` avoids bundling README into archive so archive contains primarily the binary (GoReleaser may still need a workaround — if `none*` unsupported, use empty `files: []` or document that extra files are OK as long as binary basename is `clocky`/`clocky.exe`). Prefer:

```yaml
archives:
  - formats: [tar.gz]
    format_overrides:
      - goos: windows
        formats: [zip]
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
```

and ensure `ExtractBinary` picks the binary by basename regardless of extra files.

- [ ] **Step 2: `.github/workflows/release.yml`**

```yaml
name: release

on:
  push:
    tags:
      - "v*"

permissions:
  contents: write

jobs:
  goreleaser:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"
      - uses: goreleaser/goreleaser-action@v6
        with:
          distribution: goreleaser
          version: "~> v2"
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

- [ ] **Step 3: Local snapshot smoke (optional if goreleaser installed)**

```bash
goreleaser release --snapshot --clean
```

- [ ] **Step 4: Commit**

```bash
git add .goreleaser.yaml .github/workflows/release.yml
git commit -m "ci: add GoReleaser GitHub release workflow"
```

---

### Task 8: Install scripts

**Files:**
- Create: `scripts/install.sh`
- Create: `scripts/install.ps1`

**Shared behavior:** detect arch; `REPO` default `luisf/Clocky`; version from `CLOCKY_VERSION` or latest API; download archive + checksums; verify; install to default dir or `CLOCKY_INSTALL_DIR`; fix PATH.

#### `scripts/install.sh` (bash)

Skeleton responsibilities:

```bash
#!/usr/bin/env bash
set -euo pipefail
REPO="${CLOCKY_GITHUB_REPO:-luisf/Clocky}"
INSTALL_DIR="${CLOCKY_INSTALL_DIR:-${HOME}/.local/bin}"
# uname -s/-m → darwin/linux + amd64/arm64
# curl -fsSL API latest or specific tag assets
# sha256sum / shasum -a 256 verify
# tar -xzf or unzip
# install -m 755 clocky "$INSTALL_DIR/clocky"
# echo PATH hint if needed
"$INSTALL_DIR/clocky" version
```

#### `scripts/install.ps1`

```powershell
$ErrorActionPreference = 'Stop'
$Repo = if ($env:CLOCKY_GITHUB_REPO) { $env:CLOCKY_GITHUB_REPO } else { 'luisf/Clocky' }
$InstallDir = if ($env:CLOCKY_INSTALL_DIR) { $env:CLOCKY_INSTALL_DIR } else { Join-Path $env:LOCALAPPDATA 'Clocky' }
# Detect amd64/arm64 from $env:PROCESSOR_ARCHITECTURE
# Invoke-RestMethod releases/latest
# Download zip + checksums.txt; Get-FileHash -Algorithm SHA256
# Expand-Archive; copy clocky.exe
# Add User PATH if missing via [Environment]::SetEnvironmentVariable
& "$InstallDir\clocky.exe" version
```

Mark `install.sh` executable in git: `git update-index --chmod=+x scripts/install.sh` (or document `chmod +x`).

- [ ] **Step 1: Write both scripts completely (no placeholders)**
- [ ] **Step 2: Commit**

```bash
git add scripts/install.sh scripts/install.ps1
git commit -m "feat: add cross-platform install scripts for GitHub Releases"
```

---

### Task 9: README + help polish

**Files:**
- Modify: `README.md`
- Modify: `internal/cli/cli.go` help text (if not fully updated in Task 6)

Replace install section with:

```markdown
## Install

### One-line install

**Windows (PowerShell):**
\`\`\`powershell
irm https://raw.githubusercontent.com/luisf/Clocky/master/scripts/install.ps1 | iex
\`\`\`

**macOS / Linux:**
\`\`\`bash
curl -fsSL https://raw.githubusercontent.com/luisf/Clocky/master/scripts/install.sh | bash
\`\`\`

### Update

\`\`\`bash
clocky update
clocky update --yes
clocky update --check
\`\`\`

### From source
… keep existing go build instructions …
```

Fix clone URL to `https://github.com/luisf/Clocky.git`.

- [ ] **Step 1: Edit README**
- [ ] **Step 2: `go test ./...`**
- [ ] **Step 3: Commit**

```bash
git commit -m "docs: document install scripts and clocky update"
```

---

### Task 10: End-to-end verification checklist

- [ ] **Step 1: Unit tests**

```powershell
go test ./...
```

Expected: all PASS.

- [ ] **Step 2: Local version ldflags**

```powershell
go build -ldflags "-X github.com/luisf/clocky/internal/version.Version=0.0.0-dev" -o clocky.exe ./cmd/clocky
.\clocky.exe version
```

Expected: prints `0.0.0-dev`.

- [ ] **Step 3: Update check against missing/private repo**

Should fail with a clear HTTP/API error, not panic; existing binary untouched.

- [ ] **Step 4: Document release steps in commit message or short `docs/release.md` (optional one-pager)**

```markdown
# Releasing Clocky
1. Merge to master
2. git tag v0.1.0 && git push origin v0.1.0
3. GitHub Actions publishes Release
4. Users: install script or clocky update
```

If adding `docs/release.md`:

```bash
git add docs/release.md
git commit -m "docs: add release tagging checklist"
```

---

## Spec coverage

| Spec item | Task |
|-----------|------|
| Version ldflags + `clocky version` | 1, 6, 7 |
| Semver / asset / checksums | 2 |
| GitHub latest + download | 3 |
| Extract + Windows/Unix replace | 4 |
| `clocky update` / `--yes` / `--check` exit 3 | 5, 6 |
| `CLOCKY_GITHUB_REPO` / tokens | 3, 5, 8 |
| GoReleaser + Actions | 7 |
| install.sh / install.ps1 + PATH dirs | 8 |
| README | 9 |
| Fail closed on checksum | 3, 5 |

## Self-review notes

- Repo slug default `luisf/Clocky` must match the real GitHub repo name/casing when published; override via env until then.
- GoReleaser archive may include LICENSE/README; `ExtractBinary` must select by binary basename.
- Windows replace leaves `.old` if delete fails; startup cleanup handles it.
- No package-manager manifests in this plan (explicitly out of scope).
