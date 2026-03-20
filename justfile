set dotenv-load

service_name := "soon-cms"
git_commit := `git rev-parse --short HEAD`
go_cache := justfile_directory() + "/.cache/go-build"

# Run the service locally
run:
    mkdir -p "{{go_cache}}"
    HTTP_PORT="${HTTP_PORT:-3000}" GOCACHE="{{go_cache}}" go run ./cmd/soon-cms/

# Get linting tools
setup:
    go get github.com/golangci/golangci-lint/cmd/golangci-lint
    go get golang.org/x/tools/cmd/goimports

# Run tests
test: lint
    mkdir -p "{{go_cache}}" ./test
    GOCACHE="{{go_cache}}" go test \
        -v \
        -race \
        -bench=./... \
        -benchmem \
        -timeout=120s \
        -cover \
        -coverprofile=./test/coverage.txt \
        -bench=./... ./...

# Generate mocks
mocks:
    go generate ./...

# Clean, format, lint, test
full: clean fmt lint test

# Lint
lint:
    golangci-lint run --config configs/golangci.yml

# Format code
fmt:
    gofmt -w -s .
    goimports -w .
    GOCACHE="{{go_cache}}" go clean ./...

# Format and lint
pre-commit: fmt lint

# Clean build artifacts
clean:
    GOCACHE="{{go_cache}}" go clean ./...
    rm -rf bin/{{service_name}}
