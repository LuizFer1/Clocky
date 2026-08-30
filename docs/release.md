# Releasing Clocky

1. Merge the release branch to `master`.
2. Tag and push: `git tag v0.1.0 && git push origin v0.1.0`
3. GitHub Actions (`.github/workflows/release.yml`) runs GoReleaser and publishes a GitHub Release with archives + `checksums.txt`.
4. Users install or upgrade with:
   - `scripts/install.ps1` / `scripts/install.sh`
   - `clocky update`

Pin installs with `CLOCKY_VERSION=v0.1.0`. Forks can set `CLOCKY_GITHUB_REPO=owner/name`.
