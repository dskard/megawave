# Contributing to Megawave

Thank you for your interest in contributing to Megawave! This document provides guidelines and information for contributors.

## Getting Started

1. Fork the repository
2. Clone your fork:
   ```bash
   git clone https://github.com/YOUR_USERNAME/megawave.git
   cd megawave
   ```
3. Install dependencies:
   ```bash
   just deps
   ```
4. Install development tools (goimports, golangci-lint):
   ```bash
   just tools
   ```
5. Create a branch for your changes:
   ```bash
   git checkout -b feature/your-feature-name
   ```

## Project Structure

```
.
├── .claude/skills/      # Claude Code skills for development
├── cmd/megawave/        # Application entry point
├── internal/
│   ├── microwave/       # Core microwave logic
│   └── telemetry/       # Logging and OpenTelemetry setup
├── docs/                # Documentation
└── Justfile             # Task definitions
```

## Development Workflow

### Running Tests

```bash
# Run tests (includes race detector)
just test

# Run tests with coverage
just coverage
```

### Code Quality

```bash
# Format code
just fmt

# Run linter
just lint

# Run all checks (format, lint, test)
just check
```

Always run `just check` before submitting a pull request.

### Testing Conventions

Tests are organized as follows:
- **Unit tests**: Grouped under `// FunctionName Test Cases` comments
- **Integration tests**: At the bottom under `// Integration Test Cases`
- **Concurrency tests**: Prefixed with `TestIntegration...Concurrent`

Each test includes documentation:
- Header comment describing what it verifies
- "Test logic:" line explaining how the test works
- Inline comments for key steps

Example:
```go
// TestCountdownCompletesSuccessfully verifies that countdown returns true when completed.
// Test logic: Calls countdown with 1 second, waits for completion, checks return value
// is true and display shows 00:00.
func TestCountdownCompletesSuccessfully(t *testing.T) {
    m := New()
    // Call countdown with 1 second duration
    result := m.countdown(context.Background(), 1)
    // Verify countdown completed successfully
    if !result {
        t.Error("countdown() should return true when completed")
    }
}
```

## Submitting Changes

1. Ensure all tests pass: `just check`
2. Commit your changes with a clear commit message
3. Push to your fork
4. Open a pull request against the `main` branch

## Reporting Issues

When reporting issues, please include:
- A clear description of the problem
- Steps to reproduce the issue
- Expected vs actual behavior
- Go version and operating system
