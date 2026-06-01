# Release process

This project uses Conventional Commits and Semantic Versioning.

## Versioning

- Patch releases fix bugs and false positives without changing intended behavior.
- Minor releases add rules, profiles, output fields, or CLI features.
- Major releases are reserved for breaking CLI or stable JSON/SARIF contract changes after v1.

## Pre-release checklist

Run:

```sh
go test ./...
go vet ./...
go run ./cmd/codex-action-guard audit --all --fail-on high
```

Also verify:

- `README.md` still matches CLI behavior.
- `docs/rules.md` matches `internal/guard/rules.go`.
- Generated profiles audit cleanly.
- `CHANGELOG.md` has release notes.
- The version tag follows `vMAJOR.MINOR.PATCH`.

## Tagging

```sh
go test ./...
go vet ./...
go run ./cmd/codex-action-guard audit --all --fail-on high
git tag v0.1.0
git push origin v0.1.0
```

Pushing a `vMAJOR.MINOR.PATCH` tag runs the Release workflow. It builds `codex-action-guard` for:

- Linux `amd64` and `arm64`
- macOS `amd64` and `arm64`
- Windows `amd64` and `arm64`

The workflow uploads compressed archives and `SHA256SUMS` to the GitHub Release for the tag. If the release already exists, the workflow uploads artifacts with `--clobber`.

## v0.1.0 release flow

1. Run the local gates:

   ```sh
   go test ./...
   go vet ./...
   go run ./cmd/codex-action-guard audit --all --fail-on high
   ```

2. Verify the six release targets build locally:

   ```bash
   for target in \
     "linux amd64" "linux arm64" \
     "darwin amd64" "darwin arm64" \
     "windows amd64" "windows arm64"
   do
     read -r goos goarch <<< "$target"
     ext=""
     if [ "$goos" = "windows" ]; then
       ext=".exe"
     fi
     GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
       go build -trimpath -o "/tmp/codex-action-guard-$goos-$goarch$ext" ./cmd/codex-action-guard
   done
   ```

3. Tag and push the release:

   ```sh
   git tag v0.1.0
   git push origin v0.1.0
   ```

4. Verify the GitHub Release has Linux, macOS, and Windows archives for `amd64` and `arm64`, plus `SHA256SUMS`:

   ```sh
   gh release view v0.1.0
   gh release download v0.1.0 --pattern SHA256SUMS --dir /tmp/codex-action-guard-release
   ```

5. Update the floating v0 action tag to the same commit:

   ```sh
   git tag -f v0 v0.1.0
   git push -f origin v0
   ```

The Release workflow listens only for semver-like `v*.*.*` tags, so pushing the floating `v0` tag does not create a separate GitHub Release.

## GitHub Action wrapper

The v0 action wrapper uses `go run` from the checked-out action source. A later release can switch the wrapper to download a prebuilt binary by platform and verify checksums before execution.

## Local artifact smoke test

Before tagging, verify the same target set locally:

```bash
for target in \
  "linux amd64" "linux arm64" \
  "darwin amd64" "darwin arm64" \
  "windows amd64" "windows arm64"
do
  read -r goos goarch <<< "$target"
  GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 go build -o "/tmp/codex-action-guard-$goos-$goarch" ./cmd/codex-action-guard
done
```
