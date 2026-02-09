# Megawave

A microwave oven simulator written in Go.

 Learn more about the design details in [docs/architecture.md](docs/architecture.md).

## Prerequisites

- **Go 1.25+** - [Download](https://go.dev/dl/)
- **just** - Task runner ([Installation](https://github.com/casey/just#installation))
- **Docker** - Required for observability stack (optional)


## Quick Start

```bash
# Install dependencies
just deps

# Build and run in development mode (logs to megawave.log)
just run
```

## Setup

1. Clone the repository:
   ```bash
   git clone https://github.com/dskard/megawave.git
   cd megawave
   ```

2. Install dependencies:
   ```bash
   just deps
   ```

## The microwave source

The core microwave logic is located in [`internal/microwave/`](internal/microwave/). This package implements digit entry, display formatting, and the cooking countdown timer.

### Running the microwave test cases

#### Using just (recommended)

```bash
# Run tests (includes race detector)
just test

# Run tests with coverage report
just coverage
```

#### Using go directly

```bash
# Run tests
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Building and running the demo program

There is a demo program named `megawave` that you can use to interact with the microwave oven.

### Using just (recommended)

These commands create the binary at `./bin/megawave`:

```bash
# Build the binary
just build

# Build and run
just run
```

### Using go directly

```bash
# Build
go build -o bin/megawave ./cmd/megawave

# Run
go run ./cmd/megawave
```

This starts an interactive session:
- Press **0-9** to enter time digits
- Press **Enter** to start cooking
- Press **Ctrl-C** to exit

### Configuration

Configuration via flags or environment variables (flags take precedence):

| Setting | Flag | Env Var | Default |
|---------|------|---------|---------|
| Environment | `-env` | `MEGAWAVE_ENV` | `development` |
| Log level | `-log-level` | `MEGAWAVE_LOG_LEVEL` | `info` |
| Log file | `-log-file` | `MEGAWAVE_LOG_FILE` | `megawave.log` |
| OTLP endpoint | `-otlp-endpoint` | `MEGAWAVE_OTLP_ENDPOINT` | none (host:port) |

### Examples

```bash
# Development mode with debug logging
just run -log-level=debug
# or
./bin/megawave -log-level=debug

# Custom log file location
just run -log-file=/tmp/megawave.log
# or
./bin/megawave -log-file=/tmp/megawave.log

# Production mode with observability
just run-prod -log-level=debug
# or
./bin/megawave -env=production -otlp-endpoint=localhost:4318

# Using environment variables
MEGAWAVE_LOG_LEVEL=debug ./bin/megawave

# View logs in real-time while running
tail -f megawave.log
```

## Observability

Megawave exports logs, traces, and metrics via OpenTelemetry. See [docs/observability.md](docs/observability.md) for detailed instructions on:

- Starting the Grafana stack (`just grafana-up`)
- Viewing logs in Loki, traces in Tempo, and metrics in Prometheus
- Creating dashboards
- Correlating logs with traces

Quick start:

```bash
just grafana-up                # Start Grafana + Loki + Tempo + Prometheus
just run-prod -log-level=debug # Run megawave with telemetry export
open http://localhost:3000     # View in Grafana
just grafana-down              # Cleanup when done
```

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for development setup, testing conventions, and how to submit changes.
