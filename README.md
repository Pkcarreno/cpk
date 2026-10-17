# cpk

A lightweight Corepack alternative that manages the `packageManager` specification in `package.json`.

## installation

Install the binary with the Go toolchain:

```sh
go install github.com/pkcarreno/cpk@latest
```

### other methods

#### precompiled binaries

Download precompiled archives for Linux, macOS, and Windows from the [Releases](https://github.com/pkcarreno/cpk/releases/latest) page. Each release includes a SHA-256 checksum file to verify archive integrity.

#### mise

Install `cpk` with the `mise` Go backend:

```sh
mise use -g go:github.com/pkcarreno/cpk@latest
```

## usage

### commands

| Command | Arguments | Behavior |
|---|---|---|
| `cpk use` | `<package-manager[@target]>` | Resolves target version and integrity hash from the registry. Writes the specification to `package.json`. |
| `cpk up` | `[target]` | Updates the pinned package manager version in `package.json`. Preserves Yarn Classic within `1.x` when target is omitted. |
| `cpk current` | None | Reads `package.json` locally and displays the active package manager, version, and integrity hash without network requests. |
| `cpk outdated` | None | Queries the registry to check if a newer package manager version is available. Never mutates `package.json`. |
| `cpk verify` | None | Verifies that the integrity hash in `package.json` matches published registry metadata. Never mutates `package.json`. |

### options

#### global flags

| Flag | Short | Description |
|---|---|---|
| `--registry <url>` | | URL of the npm registry. Overrides `NPM_CONFIG_REGISTRY`. |
| `--quiet` | `-q` | Do not write status messages, spinners, or confirmations to `stderr`. |
| `--verbose` | `-v` | Write detailed diagnostic logs and execution timings to `stderr`. |
| `--no-color` | | Do not use ANSI color escape codes. |

#### command flags

| Flag | Command | Description |
|---|---|---|
| `--raw` | `cpk current` | Output only the unformatted specification string to `stdout`. |
| `--fail-on-outdated` | `cpk outdated` | Exit with a non-zero status code if an update is available. |

#### environment variables

| Variable | Description |
|---|---|
| `NPM_CONFIG_REGISTRY` | Default registry URL when `--registry` is not set. |
| `NO_COLOR` | Disables ANSI color escape codes when set and non-empty. |
| `CI` | Disables animated spinners and color codes when set to `true` or `1`. |

## technical overview

### surgical manifest modification

`cpk` updates `package.json` at the byte level with `sjson`. Standard JSON formatters reorder object keys and change file indentation. In contrast, `cpk` mutates only the `packageManager` value in place. It preserves existing indentation (spaces or tabs), key order, and trailing newlines. This produces a clean, single-line git diff.

### security architecture

`cpk` enforces a strict allowlist of supported package managers (`pnpm`, `yarn`, `npm`). It rejects unsupported tools locally before network requests.

During resolution, `cpk` converts registry Subresource Integrity (SRI) hashes and legacy shasums into canonical `+<algorithm>.<hex>` integrity hashes. It validates algorithm prefixes and hex digest lengths for `sha512`, `sha384`, `sha256`, `sha224`, and `sha1`.

Manifest inspection runs strictly offline by default. Remote network queries execute only during explicit update and audit requests.
