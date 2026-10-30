# cclint justfile

# Default: show available recipes
default:
    @just --list

# Build binary and symlink to ~/go/bin
build:
    go build -o cclint . && ln -sf "$(pwd)/cclint" ~/go/bin/cclint

# Build with version override
build-version version:
    go build -ldflags "-X github.com/dotcommander/cclint/cmd.Version={{version}}" -o cclint .

# Run all tests
test:
    go test ./...

# Run tests for a specific package
test-pkg pkg:
    go test ./internal/{{pkg}}/...

# Run tests with verbose output
test-verbose:
    go test -v ./...

# Run short tests only
test-short:
    go test -short ./...

# Lint all component types (default ~/.claude)
lint *args:
    go run . {{args}}

# Lint agents only
lint-agents:
    go run . agents

# Lint commands only
lint-commands:
    go run . commands

# Lint skills only
lint-skills:
    go run . skills

# Lint plugins only
lint-plugins:
    go run . plugins

# Lint with scores
scores *args:
    go run . --scores {{args}}

# Lint with improvement recommendations
improvements *args:
    go run . --improvements {{args}}

# Lint staged files (for pre-commit)
staged:
    go run . --staged

# Lint uncommitted changes
diff:
    go run . --diff

# Create baseline from current issues
baseline-create:
    go run . --baseline-create

# Lint with baseline filtering
baseline *args:
    go run . --baseline {{args}}

# JSON output
json *args:
    go run . --format json {{args}}

# Run go vet
vet:
    go vet ./...

# Check formatting
fmt-check:
    @test -z "$(gofmt -l .)" || (gofmt -l . && exit 1)

# Format code
fmt:
    gofmt -w .

# Full CI check: fmt, vet, build, test
ci: fmt-check vet build test

# Clean build artifacts
clean:
    rm -f cclint
