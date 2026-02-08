# Megawave

A microwave oven simulator written in Go with observability (logging, tracing, metrics).

## Quick Start

```bash
# Install dependencies
just deps

# Build the application
just build

# Run in development mode (logs to megawave.log)
just run

# Run interactively
./bin/megawave
```

## Prerequisites

- **Go 1.22+** - [Download](https://go.dev/dl/)
- **just** - Task runner ([Installation](https://github.com/casey/just#installation))
- **Docker** - Required for observability stack (optional)

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/dskard/megawave.git
   cd megawave
   ```

2. Install dependencies:
   ```bash
   just deps
   ```

3. Build the application:
   ```bash
   just build
   ```

   This creates the binary at `./bin/megawave`.

## Compiling

### Using just (recommended)

```bash
# Build the binary
just build

# Build and run
just run

# Clean build artifacts
just clean
```

### Using go directly

```bash
# Build
go build -o bin/megawave ./cmd/megawave

# Run
go run ./cmd/megawave

# Run tests
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

## Running

### Interactive Mode (default)

```bash
./bin/megawave
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
./bin/megawave -log-level=debug

# Custom log file location
./bin/megawave -log-file=/tmp/megawave.log

# Production mode with observability
./bin/megawave -env=production -otlp-endpoint=localhost:4318

# Using environment variables
MEGAWAVE_LOG_LEVEL=debug ./bin/megawave

# View logs in real-time while running
tail -f megawave.log
```

### Available just Commands

```bash
just                # List all available commands
just deps           # Download Go module dependencies
just build          # Build the binary
just run [args]     # Build and run (with optional args)
just test           # Run tests
just coverage       # Generate coverage report
just fmt            # Format code
just lint           # Run linter
just clean          # Remove build artifacts
just grafana-up     # Start Grafana observability stack
just grafana-down   # Stop Grafana stack
just run-prod [args] # Run with Grafana observability (with optional args)
```

Examples with arguments:

```bash
# Run with debug logging
just run -log-level=debug

# Run in production with extra flags
just run-prod -log-level=debug
```

## Observability Demo

This guide shows how to view logs, traces, and metrics from megawave using Grafana.

### Prerequisites

- Docker installed and running
- megawave built (`just build`)

### Step 1: Start the Grafana Stack

Start the all-in-one Grafana observability stack:

```bash
just grafana-up
```

This starts a container with:
- **Grafana** at http://localhost:3000
- **OTLP receiver** at http://localhost:4318

### Step 2: Run megawave in Production Mode

```bash
./bin/megawave -env=production -otlp-endpoint=localhost:4318
```

Or use the shortcut:
```bash
just run-prod
```

Interact with the microwave (press digits, start cooking).

### Step 3: View Logs in Grafana

1. Open http://localhost:3000 in your browser
2. Click **Explore** in the left sidebar
3. Select **Loki** from the data source dropdown
4. Enter the query:
   ```
   {service_name="megawave"}
   ```
5. Click **Run query**

You should see structured logs for each button press and cooking event:
- `digit pressed` - Each digit entered
- `cooking started` - When START is pressed
- `tick` - Each second of countdown
- `cooking complete` - When timer reaches 00:00

### Step 4: View Traces in Grafana

1. In Grafana, click **Explore**
2. Select **Tempo** from the data source dropdown
3. Choose **Search** tab
4. Set Service Name to `megawave`
5. Click **Run query**

Click on any trace to see:
- **cooking_session** span - The entire cooking duration
- Span attributes: `initial_display`, `duration_seconds`
- Timeline showing when each event occurred

### Step 5: View Metrics in Grafana

1. In Grafana, click **Explore**
2. Select **Prometheus** from the data source dropdown
3. Enter queries:
   ```promql
   # Total button presses
   microwave_button_presses_total

   # Cooking sessions started
   microwave_cooking_sessions_total
   ```
4. Click **Run query**

### Step 6: Correlate Logs with Traces

Logs emitted during cooking sessions include trace IDs for correlation:
1. In Loki, find a log from during cooking (e.g., "cooking started", "tick")
2. In the log details, find the `TraceId` attribute
3. Search for that trace ID in Tempo to see the full cooking session span

### Cleanup

Stop the Grafana stack when done:

```bash
just grafana-down
```

## Development

```bash
# Run tests
just test

# Run tests with coverage
just coverage

# Format code
just fmt

# Run linter
just lint

# Run all checks
just check
```

## Project Structure

```
.
├── cmd/megawave/        # Application entry point
├── internal/
│   ├── microwave/       # Core microwave logic
│   └── telemetry/       # Logging and OpenTelemetry setup
├── docs/                # Documentation
└── Justfile             # Task definitions
```
