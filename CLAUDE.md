# Megawave

A microwave oven simulator with observability (logging, tracing, metrics).

## Build & Run

```bash
just deps             # Download dependencies
just build            # Build binary to bin/megawave
just run [args]       # Build and run in development mode
just run-prod [args]  # Run with Grafana observability
just clean            # Remove build artifacts

# Examples with arguments
just run -log-level=debug
just run-prod -log-level=debug
```

## Testing

```bash
just test       # Run all tests (includes race detector)
just coverage   # Generate coverage report (coverage.html)
```

Tests use the standard Go testing package. Test files are co-located with source files as `*_test.go`.

### Test Organization

- **Unit tests**: Grouped under `// FunctionName Test Cases` comments
- **Integration tests**: Placed at the bottom under `// Integration Test Cases` comment
- **Concurrency tests**: Prefixed with `TestIntegration...Concurrent`

### Test Documentation Style

Every test should have:
1. A header comment describing what it verifies
2. A "Test logic:" line explaining how the test works
3. Inline comments for key steps

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

### Writing Tests

Use `/megawave-write-test-for <function-name>` to generate tests for a function. This skill:
1. Locates and reads the function implementation
2. Analyzes test requirements (branches, edge cases, concurrency)
3. Creates a test plan and asks for confirmation
4. Writes documented unit, integration, and concurrency tests

## Code Quality

```bash
just fmt        # Format code (go fmt + goimports)
just lint       # Run golangci-lint
just check      # Run fmt, lint, and test
```

## Configuration

Flags override environment variables:

| Flag | Env Var | Default | Description |
|------|---------|---------|-------------|
| `-env` | `MEGAWAVE_ENV` | `development` | Environment (production/development/test) |
| `-log-level` | `MEGAWAVE_LOG_LEVEL` | `info` | Log level (debug/info/warn/error) |
| `-log-file` | `MEGAWAVE_LOG_FILE` | `megawave.log` | Log file path (development only) |
| `-otlp-endpoint` | `MEGAWAVE_OTLP_ENDPOINT` | none | OTLP collector (host:port) |

## Project Structure

- `cmd/megawave/` - Interactive terminal application
- `internal/microwave/` - Core microwave logic (display, digits, countdown)
- `internal/telemetry/` - Logging and OpenTelemetry setup

## Architecture

### Microwave Package

The `internal/microwave` package implements the core logic:

- **Digit entry**: 4-digit display (MM:SS), shifts left on each digit press
- **Cooking**: Countdown timer, prints display each second
- **State**: Cannot stop/pause once cooking starts

Uses functional options pattern for dependency injection:
```go
m := microwave.New(
    microwave.WithLogger(logger),
    microwave.WithTracer(tracer),
    microwave.WithMeter(meter),
)
```

### Telemetry Package

The `internal/telemetry` package handles:

- **Config parsing**: `ParseConfig()` reads flags and env vars
- **Logger creation**: `NewLogger(cfg)` returns environment-specific slog handler
- **OTel init**: `InitOTel(ctx, cfg)` sets up tracing, logging, and metrics exporters

Behavior by environment:
- **Production**: JSON logs to stdout, OTel traces/metrics to OTLP endpoint
- **Development**: Text logs to file (default: megawave.log)
- **Test**: Logs discarded

## Concurrency

### Mutex Strategy

The Microwave struct uses `sync.Mutex` to protect mutable state:

**Protected state** (requires lock):
- `digits [4]int`
- `digitCount int`
- `isCooking bool`

**Immutable after construction** (no lock needed):
- `logger`, `tracer`, `meter` - set once in `New()`, never modified after

**Rules:**
1. Lock before reading/writing protected state
2. Release lock before I/O operations (logging, printing, metrics)
3. Copy values to local variables before unlocking when needed for I/O
4. Keep critical sections short

**Helper functions requiring lock held by caller:**
- `displayString()` - reads digits
- `totalSeconds()` - reads digits

### Reviewing Mutex Usage

Use `/megawave-mutex-bot <file-or-function>` to check mutex patterns. This skill:
1. Identifies all lock/unlock points
2. Verifies state access happens under lock
3. Verifies I/O happens outside locks
4. Flags suspicious patterns (red/yellow)

## Observability

```bash
just grafana-up   # Start Grafana + Loki + Tempo + Prometheus
just grafana-down # Stop Grafana stack
```

Grafana at http://localhost:3000, OTLP at localhost:4318

## Conventions

- Follow standard Go conventions (gofmt, effective Go)
- Use `internal/` for code that should not be imported externally
- Place tests alongside source files as `*_test.go`
- Use table-driven tests where appropriate
- Use functional options for constructor configuration
