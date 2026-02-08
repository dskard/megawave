# Architecture

This document describes the architecture of the megawave microwave simulator.

## Package Structure

```
cmd/megawave/          # Application entry point
internal/
  microwave/           # Core microwave logic
  telemetry/           # Logging and OpenTelemetry setup
```

## Component Overview

```
┌─────────────────────────────────────────────────────────┐
│                      main.go                            │
│  - Signal handling (Ctrl-C)                             │
│  - Terminal raw mode for keypress capture               │
│  - Wires together telemetry and microwave               │
└─────────────────┬───────────────────────────────────────┘
                  │
      ┌───────────┴───────────┐
      ▼                       ▼
┌─────────────┐       ┌───────────────┐
│  microwave  │       │   telemetry   │
│             │       │               │
│ - Digits    │       │ - Config      │
│ - Display   │       │ - Logger      │
│ - Countdown │       │ - OTel init   │
└─────────────┘       └───────────────┘
```

## Package Details

### cmd/megawave

The main package handles:

- **Configuration**: Parses flags and environment variables via `telemetry.ParseConfig()`
- **Signal handling**: Sets up context cancellation on Ctrl-C
- **Terminal mode**: Uses raw mode to capture individual keypresses without Enter
- **Event loop**: Routes keypresses to `PressDigit()` or `PressStart()`

### internal/microwave

The core microwave logic with no external dependencies (except OTel interfaces).

**State:**
- `digits [4]int` - The four display digits (MM:SS format)
- `digitCount int` - Number of digits entered (max 4)
- `isCooking bool` - Whether countdown is active

**Public API:**
- `New(opts ...Option) *Microwave` - Constructor with functional options
- `PressDigit(d int)` - Handle digit button press (0-9)
- `PressStart(ctx context.Context)` - Start cooking countdown
- `Display() string` - Get current display as "MM:SS"
- `IsCooking() bool` - Check if cooking is in progress

**Functional Options:**
- `WithLogger(*slog.Logger)` - Inject logger
- `WithTracer(trace.Tracer)` - Inject OTel tracer
- `WithMeter(metric.Meter)` - Inject OTel meter

**Concurrency:**
- Uses `sync.Mutex` to protect state
- Internal `displayString()` helper for use within locked sections
- `countdown()` respects context cancellation for graceful shutdown

### internal/telemetry

Handles all observability configuration.

**Config:**
- `ParseConfig()` - Reads from flags and env vars (flags take precedence)
- Environment: `production`, `development`, `test`
- Log level: `debug`, `info`, `warn`, `error`

**Logger Creation:**
- Production: OTel slog bridge (logs sent via OTLP)
- Development: Text logs to file
- Test: Discarded

**OTel Initialization:**
- Creates trace exporter and provider
- Creates log exporter and provider
- Sets global providers
- Returns combined shutdown function

## Data Flow

### Digit Entry

```
User presses '5'
    │
    ▼
main.go reads byte from stdin
    │
    ▼
m.PressDigit(5)
    │
    ├─► Log "digit pressed"
    ├─► Record button_presses metric
    │
    ▼
Shift digits left, append 5
    │
    ▼
Print new display
```

### Cooking

```
User presses Enter
    │
    ▼
m.PressStart(ctx)
    │
    ├─► Log "cooking started"
    ├─► Start tracing span
    ├─► Record cooking_sessions metric
    │
    ▼
countdown(ctx, seconds)
    │
    ├─► Each second: update display, print, sleep
    │
    ▼ (on ctx.Done or completion)
    │
Log "cooking complete" or "cooking canceled"
```

## Design Decisions

### Functional Options Pattern

Chosen over alternatives (config struct, builder pattern) because:
- Clean API: `New(WithLogger(l), WithTracer(t))`
- Optional dependencies with sensible defaults
- Easy to extend without breaking existing code

### Context for Cancellation

`PressStart` accepts a context to allow:
- Graceful shutdown on Ctrl-C
- Proper trace context propagation
- Future support for timeouts

### Mutex Strategy

Fine-grained locking with short critical sections:
- Lock only when accessing/modifying state
- Release before I/O operations (logging, printing)
- Internal `displayString()` for use within locked sections

### Raw Terminal Mode

Required because:
- Standard input buffers until Enter
- We need immediate response to each keypress
- Ctrl-C in raw mode is byte 3, not SIGINT
