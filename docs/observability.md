# Observability

This document describes how to observe megawave using logs, traces, and metrics.

## Overview

Megawave uses OpenTelemetry to export telemetry data:

| Signal  | Library                  | Export Format | Destination |
|---------|--------------------------|---------------|-------------|
| Logs    | slog + otelslog bridge   | OTLP/HTTP     | Loki        |
| Traces  | OTel SDK                 | OTLP/HTTP     | Tempo       |
| Metrics | OTel SDK                 | OTLP/HTTP     | Prometheus  |

## Quick Start

### Prerequisites

- Docker installed and running
- megawave built (`just build`)

### Steps

```bash
# Start the Grafana observability stack
just grafana-up

# Run megawave in production mode
just run-prod

# Interact with the microwave (press digits, start cooking), then Ctrl-C to exit

# View telemetry in Grafana
open http://localhost:3000
```

## Grafana Stack

The `just grafana-up` command starts the [grafana/otel-lgtm](https://github.com/grafana/docker-otel-lgtm) Docker image, which includes:

| Component  | Purpose                | Port |
|------------|------------------------|------|
| Grafana    | Visualization UI       | 3000 |
| Loki       | Log aggregation        | -    |
| Tempo      | Distributed tracing    | -    |
| Prometheus | Metrics storage        | -    |
| OTel Collector | OTLP receiver       | 4317 (gRPC), 4318 (HTTP) |

## Configuration

| Flag | Env Var | Description |
|------|---------|-------------|
| `-env=production` | `MEGAWAVE_ENV=production` | Enable OTel export |
| `-otlp-endpoint=localhost:4318` | `MEGAWAVE_OTLP_ENDPOINT=localhost:4318` | Collector address |
| `-log-level=debug` | `MEGAWAVE_LOG_LEVEL=debug` | Include debug logs |

## Viewing Logs in Loki

1. Open Grafana at http://localhost:3000
2. Click **Explore** in the left sidebar
3. Select **Loki** from the data source dropdown
4. Enter query:
   ```
   {service_name="megawave"}
   ```
5. Click **Run query**

### Log Events

| Message | Level | When |
|---------|-------|------|
| `digit pressed` | INFO | User presses 0-9 |
| `digit ignored while cooking` | WARN | Digit pressed during countdown |
| `max digits reached` | WARN | More than 4 digits entered |
| `start pressed` | INFO | User presses Enter |
| `cooking started` | INFO | Countdown begins |
| `tick` | DEBUG | Each second of countdown |
| `cooking complete` | INFO | Countdown finished |
| `cooking canceled` | INFO | Ctrl-C during cooking |

### Useful Queries

```logql
# All logs
{service_name="megawave"}

# Only warnings and errors
{service_name="megawave"} | logfmt | level=~"WARN|ERROR"

# Cooking sessions
{service_name="megawave"} |= "cooking"

# Specific digit presses
{service_name="megawave"} | json | digit=9
```

## Viewing Traces in Tempo

1. In Grafana, click **Explore**
2. Select **Tempo** from the data source dropdown
3. Choose **Search** tab
4. Set Service Name to `megawave`
5. Click **Run query**

### Trace Structure

Each cooking session creates a trace:

```
cooking_session (span)
├── Attributes:
│   ├── initial_display: "01:30"
│   └── duration_seconds: 90
└── Duration: actual cooking time
```

### Correlating Logs and Traces

Logs emitted during a cooking session include the trace ID (via context-aware logging). In Loki:

1. Click on a log entry from during cooking (e.g., "cooking started", "tick", "cooking complete")
2. Find the `TraceId` field in the log attributes
3. Copy the trace ID and search for it in Tempo

Note: Only logs emitted with `InfoContext`/`DebugContext`/etc. include trace IDs. Logs before the cooking span starts (like "start pressed") won't have a trace ID.

## Viewing Metrics in Prometheus

1. In Grafana, click **Explore**
2. Select **Prometheus** from the data source dropdown
3. Enter PromQL queries

### Available Metrics

| Metric | Type | Description |
|--------|------|-------------|
| `microwave_button_presses_total` | Counter | Total button presses |
| `microwave_cooking_sessions_total` | Counter | Cooking sessions started |

### Useful Queries

```promql
# Total button presses
microwave_button_presses_total

# Button presses by type (digit vs start)
sum by (type) (microwave_button_presses_total)

# Presses while cooking
microwave_button_presses_total{while_cooking="true"}

# Cooking sessions over time
rate(microwave_cooking_sessions_total[5m])
```

## Creating Dashboards

Dashboards let you visualize multiple metrics, logs, and traces in one view.

### Creating Your First Dashboard

1. Open Grafana at http://localhost:3000
2. In the left sidebar, click the **Dashboards** icon (four squares)
3. Click the **New** button in the top right
4. Select **New dashboard** from the dropdown

### Adding a Panel

1. On the new dashboard, click **Add visualization**
2. Select a data source:
   - **Prometheus** for metrics
   - **Loki** for logs
   - **Tempo** for traces
3. Enter your query in the query editor (see examples below)
4. The visualization preview updates automatically
5. (Optional) Change the visualization type using the dropdown in the top right of the panel editor (Time series, Stat, Table, etc.)
6. (Optional) Set a panel title in the right sidebar under **Panel options**
7. Click **Apply** in the top right to add the panel to the dashboard

### Saving the Dashboard

1. Click the **Save** icon (floppy disk) in the top right
2. Enter a dashboard name (e.g., "Megawave")
3. Click **Save**

### Example Panels

**Button Press Rate (Time series):**
- Data source: Prometheus
- Query:
  ```promql
  rate(microwave_button_presses_total[1m])
  ```

**Total Cooking Sessions (Stat):**
- Data source: Prometheus
- Query:
  ```promql
  increase(microwave_cooking_sessions_total[1h])
  ```

**Recent Logs (Logs):**
- Data source: Loki
- Query:
  ```logql
  {service_name="megawave"} | json
  ```

### Adding More Panels

After saving, click **Add** in the top toolbar and select **Visualization** to add more panels to your dashboard.

## Troubleshooting

### No data in Grafana

1. Verify the Grafana stack is running:
   ```bash
   docker ps | grep megawave-grafana
   ```

2. Verify megawave is configured for production:
   ```bash
   ./bin/megawave -env=production -otlp-endpoint=localhost:4318
   ```

3. Check for connection errors in megawave output

### Logs missing after Ctrl-C

The OTel batch processor needs time to flush. The shutdown handler uses a 5-second timeout to ensure logs are sent before exit.

### Schema version conflicts

If you see schema URL conflicts, ensure you're using compatible OTel SDK versions. Run `go mod tidy` to update dependencies.

## Architecture

```
                              OTLP/HTTP
┌─────────────┐                :4318                ┌────────────────┐
│  megawave   │ ───────────────────────────────────>│ OTel Collector │
│             │                                     └───────┬────────┘
│  - slog     │                                             │
│  - traces   │                                             │
│  - metrics  │                         ┌───────────────────┼───────────────────┐
└─────────────┘                         │                   │                   │
                                        v                   v                   v
                                   ┌──────────┐        ┌──────────┐        ┌────────────┐
                                   │   Loki   │        │  Tempo   │        │ Prometheus │
                                   │  (logs)  │        │ (traces) │        │ (metrics)  │
                                   └────┬─────┘        └────┬─────┘        └─────┬──────┘
                                        │                   │                    │
                                        └───────────────────┼────────────────────┘
                                                            │
                                                            v
                                                       ┌──────────┐
                                                       │ Grafana  │
                                                       │  :3000   │
                                                       └──────────┘
```

## Cleanup

Stop the Grafana stack when done:

```bash
just grafana-down
```

This removes the container and all collected telemetry data.
