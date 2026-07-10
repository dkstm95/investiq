# Development

## Build From Source

```bash
git clone https://github.com/dkstm95/gacha.git
cd gacha
go test ./...
go build -o gacha ./cmd/gacha
./gacha doctor
ln -sf gacha gch
./gch doctor
```

## Release

```bash
git tag v0.2.6
git push origin v0.2.6
```

The tag workflow runs all checks, builds and verifies the artifacts, and only
then creates the GitHub release. Do not publish a release manually before the
workflow succeeds.

The release workflow lives in:

```text
.github/workflows/release.yml
```

The release artifact and checksum contract is documented in:

```text
docs/release-artifacts.md
```

## Agent Assets

```text
internal/agent/system-prompt.md
internal/agent/workflows/
internal/agent/templates/
```

These files are embedded into the `gacha` binary with `go:embed`. Treat
`internal/agent/` as the source of truth for the runtime prompt, workflows, and
report template.
