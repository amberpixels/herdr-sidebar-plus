binary := "bin/herdr-sidebar-plus"

# Default: list available commands
[private]
default:
    @just --list

# Build the binary the plugin commands run; `herdr plugin link` does not build
[group('dev')]
build:
    go build -o {{binary}} ./cmd/herdr-sidebar-plus

# Link this checkout as a local plugin
[group('dev')]
link: build
    herdr plugin link .

# Recompute every label and icon now
[group('dev')]
refresh:
    herdr plugin action invoke amberpixels.sidebar-plus.refresh

# Remove build artifacts
[group('dev')]
clean:
    rm -rf bin

# Run all tests, e.g. `just test -run Label -v`
[group('test')]
test *ARGS:
    go test {{ARGS}} ./...

# Vet - immutable
[group('lint')]
vet:
    go vet ./...

# Lint with standardgo - immutable
[group('lint')]
lint: vet
    standardgo ./...
