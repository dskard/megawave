# Default recipe to list available commands
default:
    @just --list

# Download Go module dependencies
deps:
    go mod download

# Build the binary
build:
    go build -o bin/megawave ./cmd/megawave

# Run the application
run *args: build
    ./bin/megawave {{args}}

# Run all tests
test:
    go test -v ./...

# Run tests with coverage
coverage:
    go test -coverprofile=coverage.out ./...
    go tool cover -html=coverage.out -o coverage.html
    @echo "Coverage report generated: coverage.html"

# Format code
fmt:
    go fmt ./...
    goimports -w .

# Run linter
lint:
    golangci-lint run

# Tidy dependencies
tidy:
    go mod tidy

# Clean build artifacts
clean:
    rm -rf bin/
    rm -f coverage.out coverage.html

# Run all checks (format, lint, test)
check: fmt lint test

# Install development tools
tools:
    go install golang.org/x/tools/cmd/goimports@latest
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Start the Grafana observability stack
grafana-up:
    docker run -d --name megawave-grafana \
        -p 3000:3000 \
        -p 4317:4317 \
        -p 4318:4318 \
        grafana/otel-lgtm:latest
    @echo "Grafana running at http://localhost:3000"
    @echo "OTLP endpoint at http://localhost:4318"

# Stop the Grafana stack
grafana-down:
    docker stop megawave-grafana && docker rm megawave-grafana

# Run megawave in production mode with Grafana
run-prod *args: build
    ./bin/megawave -env=production -otlp-endpoint=localhost:4318 {{args}}
