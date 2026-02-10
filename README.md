# gobashly

A Go clone of [bashly](https://github.com/DannyBen/bashly) – the CLI generator from YAML.

## What is gobashly?

`go-bashly` reads a YAML configuration file and generates a complete bash command-line interface. It supports:

- Command routing and dispatch
- Argument and flag parsing
- Help/usage rendering
- Validation with friendly error messages
- Library file merging
- Feature toggles (inspect args, view markers, deps array, env var names, sourcing)
- Script formatting (internal/external formatter, tab indentation)

## Installation

### From Source

```bash
git clone https://github.com/dimitar-trifonov/go-bashly.git
cd go-bashly
go build -o go-bashly .
sudo mv go-bashly /usr/local/bin/
```

### Install with Go (local checkout)

```bash
git clone https://github.com/dimitar-trifonov/go-bashly.git
cd go-bashly
go install .
```

This installs `go-bashly` in `$GOBIN` (if set) or `$GOPATH/bin` (usually `~/go/bin`). Make sure that directory is in your `$PATH`.

### Install with Go (@latest)

If you have access to the repository on GitHub, you can install directly:

```bash
go install github.com/dimitar-trifonov/go-bashly@latest
```

## Quick Start

1. Create a project directory:

```bash
mkdir my-cli && cd my-cli
```

2. Create a configuration file `src/bashly.yml`:

```yaml
name: mycli
help: My awesome CLI tool
version: 0.1.0

args:
- name: source
  required: true
  help: Source file to process

flags:
- long: --verbose
  short: -v
  help: Enable verbose output
```

3. Generate the script:

```bash
go-bashly generate
```

4. Run your new CLI:

```bash
./mycli --help
./mycli file.txt --verbose
```

## Commands

### `go-bashly version`

Show version information.

### `go-bashly inspect`

Inspect the command tree and configuration.

```bash
go-bashly inspect [--format tree|json] [--workdir <dir>]
```

- `--format tree`: Human-friendly tree view (default)
- `--format json`: JSON output
- `--workdir`: Working directory (default: current directory)

### `go-bashly generate`

Generate the bash script and missing command partials.

```bash
go-bashly generate [--workdir <dir>] [--force] [--dry-run]
```

- `--workdir`: Working directory (default: current directory)
- `--force`: Overwrite existing files
- `--dry-run`: Show what would be generated without writing files

## Configuration

`go-bashly` looks for configuration in this order:

1. `BASHLY_CONFIG_PATH` environment variable
2. `src/bashly.yml` (default)

### Settings

You can customize behavior with a `settings.yml` file or environment variables:

```yaml
# settings.yml
source_dir: src
target_dir: .
commands_dir: ~
lib_dir: lib
enable_inspect_args: development
enable_view_markers: development
formatter: internal
tab_indent: false
```

Environment variables take precedence and use the `BASHLY_` prefix:

```bash
export BASHLY_SOURCE_DIR=src
export BASHLY_ENABLE_INSPECT_ARGS=production
export BASHLY_FORMATTER="shfmt --case-indent --indent 2"
```

## Feature Toggles

Control optional script features via settings:

| Setting | Values | Default |
|---------|--------|---------|
| `enable_inspect_args` | `always`/`never`/`development`/`production` | `development` |
| `enable_view_markers` | `always`/`never`/`development`/`production` | `development` |
| `enable_deps_array` | `always`/`never`/`development`/`production` | `always` |
| `enable_env_var_names_array` | `always`/`never`/`development`/`production` | `always` |
| `enable_sourcing` | `always`/`never`/`development`/`production` | `development` |

## Library Files

Place shared bash functions in `src/lib/*.sh` (or configure via `lib_dir`). They will be merged into the generated script.

## Formatting

Choose how the generated script is formatted:

```yaml
formatter: internal     # Built-in formatter (removes excess blank lines)
formatter: none         # No formatting
formatter: "shfmt --case-indent --indent 2"  # External formatter
tab_indent: true       # Convert leading 2 spaces to tabs
```

## Examples

See the [ruby-bashly examples](../ruby-bashly/examples/) for inspiration. Most examples work with `go-bashly`:

```bash
# Copy an example to test
cp -a ../ruby-bashly/examples/minimal ./my-example
cd my-example
go-bashly generate
./download --help
```

## Differences from Ruby bashly

- **No ERB support**: `go-bashly` does not evaluate ERB in YAML files.
- **Go binary**: Distributed as a single static binary.
- **Settings resolution**: Full environment variable and per-environment override support.
- **Ralph-governed development**: Built using meaning-first ELST bundles and autonomous slices.

## Development

```bash
go build -o go-bashly .
./go-bashly version
```

Run tests:

```bash
go test ./...
```

## License

MIT
