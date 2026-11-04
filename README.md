# ghactions-pin-check

> Every `uses: actions/checkout@v4` in your workflows is a tag someone else can re-point.

A small, dependency-free linter that finds GitHub Actions references pinned to
mutable tags or branches instead of an immutable commit SHA — and fails CI when
it finds them.

## Why this matters

A workflow step like:

```yaml
- uses: actions/checkout@v4
```

runs whatever commit `v4` points at *today*. Tags are mutable: the action's
owner can move `v4`, and an attacker who compromises that account can move it to
malicious code — which then runs inside your pipeline, with access to your
`GITHUB_TOKEN` and secrets. This is exactly how several real supply-chain
incidents have played out. GitHub's own hardening guide recommends pinning
third-party actions **to a full-length commit SHA**:

```yaml
- uses: actions/checkout@a81bbbf8298c0fa03ea29cdc473d45769f953675 # v4.1.7
```

`ghactions-pin-check` enforces that rule.

## What it flags

| Reference                              | Kind       | Strict | `--allow-tags` |
|----------------------------------------|------------|:------:|:--------------:|
| `actions/checkout@<40-hex-sha>`        | `sha`      | ✅     | ✅             |
| `actions/checkout@v4` / `@v4.1.0`      | `tag`      | ❌     | ✅             |
| `actions/checkout@main`                | `branch`   | ❌     | ❌             |
| `actions/checkout` (no version)        | `unpinned` | ❌     | ❌             |
| `./.github/actions/local`              | `local`    | ✅     | ✅             |
| `docker://alpine:3.20`                 | `docker`   | ✅     | ✅             |

Local (`./…`) and `docker://` references are never flagged — there is no
upstream git tag to pin.

## Install

```bash
go install github.com/moveeeax/ghactions-pin-check/cmd/ghactions-pin-check@latest
```

or build from source:

```bash
make build   # produces ./bin/ghactions-pin-check
```

## Usage

```bash
# Scan the current repo (.github/workflows is found automatically)
ghactions-pin-check

# Scan a specific repo, directory, or single file
ghactions-pin-check -path ../other-repo
ghactions-pin-check -path .github/workflows/ci.yml

# Machine-readable output
ghactions-pin-check -json

# Allow version tags, only reject branches and bare references
ghactions-pin-check -allow-tags
```

The command exits `1` when any reference violates the policy and `0` when the
workflows are clean, so it drops straight into a CI gate.

### As a CI step

```yaml
jobs:
  pin-check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@3d3c42e5aac5ba805825da76410c181273ba90b1 # v7.0.1
      - uses: actions/setup-go@b7ad1dad31e06c5925ef5d2fc7ad053ef454303e # v7.0.0
        with:
          go-version: '1.27'
      - run: go run github.com/moveeeax/ghactions-pin-check/cmd/ghactions-pin-check@latest
```

## How it works

It scans `*.yml`/`*.yaml`, extracts every `uses:` scalar (a line-oriented scan,
so reported line numbers are exact and there is no YAML dependency), classifies
the reference after the `@`, and reports anything that is not a full 40-character
commit SHA (tags allowed only with `-allow-tags`).

## Development

```bash
make test   # go test -race ./...
make vet
make build
```

## License

MIT, see [LICENSE](LICENSE).
