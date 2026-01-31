# Texere Justfile
# Repository: https://github.com/texere-ot
# Author: Texere Team
# License: MIT

# Default recipe (run when `just` is executed without arguments)
default:
    @just --list

# Build the project
build:
    go build -v ./...

# Run tests
test:
    go test -v ./...

# Run tests with coverage
test-cover:
    go test -cover ./... | grep -v "no Go files"

# Run tests for OT package
test-ot:
    go test -v ./pkg/concordia/...

# Run tests for Rope package
test-rope:
    go test -v ./pkg/rope/...

# Run tests for Document package
test-document:
    go test -v ./pkg/document/...

# Run tests with race detection
test-race:
    go test -race ./...

# Run benchmarks
bench:
    go test -bench=. -benchmem ./...

# Run OT benchmarks
bench-ot:
    go test -bench=. -benchmem ./pkg/concordia/...

# Run Rope benchmarks
bench-rope:
    go test -bench=. -benchmem ./pkg/rope/...

# Format code
fmt:
    go fmt ./...
    @echo "✅ Code formatted"

# Lint code
lint:
    @if command -v golangci-lint >/dev/null 2>&1; then \
        golangci-lint run ./...; \
    else \
        echo "⚠️  golangci-lint not installed. Run: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
    fi

# Run lint and format
check: fmt lint

# Tidy dependencies
tidy:
    go mod tidy
    @echo "✅ Dependencies tidied"

# Verify dependencies
verify:
    go mod verify

# Download dependencies
deps:
    go mod download
    @echo "✅ Dependencies downloaded"

# Clean build artifacts
clean:
    rm -f coverage.out *.prof *.test
    rm -f rope.test.exe
    find . -name "*.out" -delete
    @echo "✅ Cleaned build artifacts"

# Clean all (including cached data)
clean-all: clean
    go clean -cache -testcache
    @echo "✅ Cleaned everything"

# Generate coverage report
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "✅ Coverage report generated: coverage.html"

# Show coverage in terminal
coverage-show:
    go test -cover ./... | grep -v "no Go files"

# Install development tools
install-tools:
    @echo "Installing development tools..."
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    @echo "✅ Tools installed"

# Run go vet
vet:
    go vet ./...

# Update dependencies
update:
    go get -u ./...
    go mod tidy

# Show project info
info:
    @echo "Texere - OT and Rope Text Editing Library"
    @echo ""
    @echo "Go version:"
    @go version
    @echo ""
    @echo "Module: github.com/texere-ot"
    @echo ""
    @echo "Packages:"
    @echo "  - pkg/ot (OT)"
    @echo "  - pkg/rope (Rope)"
    @echo "  - pkg/concordia (Document)"

# Watch tests (requires entr)
watch-test:
    @find . -name "*.go" | entr -c just test

# Build example (if examples exist)
build-examples:
    @if [ -d "examples" ]; then \
        find examples -name "*.go" -exec go build {} \; ; \
        echo "✅ Examples built"; \
    else \
        echo "⚠️  No examples directory"; \
    fi

# Run example
run-example name:
    @if [ -f "examples/{{name}}.go" ]; then \
        go run examples/{{name}}.go; \
    else \
        echo "⚠️  Example not found: examples/{{name}}.go"; \
    fi

# Create a new release
release version:
    @echo "Creating release {{version}}..."
    git tag -a {{version}} -m "Release {{version}}"
    git push origin {{version}}
    @echo "✅ Release {{version}} created"

# Show git statistics
stats:
    @echo "Code statistics:"
    @echo ""
    @echo "Lines of code:"
    @find . -name "*.go" -not -path "./.git/*" | xargs wc -l | tail -1
    @echo ""
    @echo "Number of Go files:"
    @find . -name "*.go" -not -path "./.git/*" | wc -l
    @echo ""
    @echo "Git stats:"
    @git shortlog -sn --all

# Run a comprehensive check
ci: tidy vet test
    @echo "✅ CI checks passed"

# Quick development workflow
dev: fmt test
    @echo "✅ Dev workflow complete"
