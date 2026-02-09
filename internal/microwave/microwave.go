package microwave

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"sync"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

// Microwave represents the microwave state
type Microwave struct {
	digits     [4]int // Stored as 4 digits: [M1, M2, S1, S2]
	digitCount int    // Number of digits entered (max 4 affect display)
	isCooking  bool
	mu         sync.Mutex

	logger          *slog.Logger
	tracer          trace.Tracer
	meter           metric.Meter
	buttonPresses   metric.Int64Counter
	cookingSessions metric.Int64Counter
}

// Option is a functional option for configuring Microwave
type Option func(*Microwave)

// New creates a new Microwave with the given options
func New(opts ...Option) *Microwave {
	m := &Microwave{
		digits:     [4]int{0, 0, 0, 0},
		digitCount: 0,
		isCooking:  false,
		logger:     slog.New(slog.NewTextHandler(io.Discard, nil)),
		tracer:     otel.Tracer("megawave"),
		meter:      otel.Meter("megawave"),
	}

	for _, opt := range opts {
		opt(m)
	}

	// Initialize metrics
	var err error
	m.buttonPresses, err = m.meter.Int64Counter("microwave.button_presses",
		metric.WithDescription("Total button presses"),
	)
	if err != nil {
		m.logger.Warn("failed to create button_presses counter", "error", err)
	}

	m.cookingSessions, err = m.meter.Int64Counter("microwave.cooking_sessions",
		metric.WithDescription("Total cooking sessions started"),
	)
	if err != nil {
		m.logger.Warn("failed to create cooking_sessions counter", "error", err)
	}

	return m
}

// WithLogger sets the logger for the microwave
func WithLogger(l *slog.Logger) Option {
	return func(m *Microwave) {
		m.logger = l
	}
}

// WithTracer sets the OpenTelemetry tracer
func WithTracer(t trace.Tracer) Option {
	return func(m *Microwave) {
		m.tracer = t
	}
}

// WithMeter sets the OpenTelemetry meter
func WithMeter(meter metric.Meter) Option {
	return func(m *Microwave) {
		m.meter = meter
	}
}

// displayString returns the display without locking (caller must hold lock)
func (m *Microwave) displayString() string {
	return fmt.Sprintf("%d%d:%d%d", m.digits[0], m.digits[1], m.digits[2], m.digits[3])
}

// Display returns the current display value as MM:SS
func (m *Microwave) Display() string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.displayString()
}

// IsCooking returns whether the microwave is currently cooking
func (m *Microwave) IsCooking() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.isCooking
}

// PressDigit handles a digit button press (0-9)
// PressDigit does not accept negative integers or integers above 9.
// PressDigit ignores digit button presses while the microwave is cooking.
func (m *Microwave) PressDigit(d int) {
	if d < 0 || d > 9 {
		m.logger.Warn("invalid digit ignored", "digit", d)
		return
	}

	cooking := m.IsCooking()

	// Always log and record metrics, even while cooking
	m.logger.Info("digit pressed", "digit", d, "cooking", cooking)
	if m.buttonPresses != nil {
		m.buttonPresses.Add(context.Background(), 1,
			metric.WithAttributes(
				attribute.String("type", "digit"),
				attribute.Bool("while_cooking", cooking),
			),
		)
	}

	// Don't allow pressing digits while cooking
	if cooking {
		m.logger.Warn("digit ignored while cooking", "digit", d)
		return
	}

	m.mu.Lock()
	if m.digitCount >= 4 {
		m.mu.Unlock()
		m.logger.Warn("max digits reached, display not updated", "digit", d)
		return
	}

	// Shift digits left and add new digit
	m.digits[0] = m.digits[1]
	m.digits[1] = m.digits[2]
	m.digits[2] = m.digits[3]
	m.digits[3] = d
	m.digitCount++

	display := m.displayString()
	digitCount := m.digitCount
	m.mu.Unlock()

	m.logger.Debug("display updated", "display", display, "digitCount", digitCount)
	fmt.Print(display + "\r\n")
}

// PressStart handles the START button press.
// Note: The assignment states the microwave "cannot be stopped." We interpret this
// as meaning there is no STOP button on the microwave interface. However, we still
// respect context cancellation (e.g., Ctrl-C) to allow graceful application shutdown.
// If the intent was to ignore all interrupts during cooking, use context.Background()
// instead of the passed context.
func (m *Microwave) PressStart(ctx context.Context) {
	cooking := m.IsCooking()

	// Always log and record metrics, even while cooking
	m.logger.Info("start pressed", "cooking", cooking)
	if m.buttonPresses != nil {
		m.buttonPresses.Add(ctx, 1,
			metric.WithAttributes(
				attribute.String("type", "start"),
				attribute.Bool("while_cooking", cooking),
			),
		)
	}

	if cooking {
		m.logger.WarnContext(ctx, "start ignored, already cooking")
		return
	}

	m.mu.Lock()
	seconds := m.totalSeconds()
	m.mu.Unlock()

	if seconds == 0 {
		m.logger.Warn("cannot start with zero time")
		return
	}

	// Start tracing span for cooking session
	ctx, span := m.tracer.Start(ctx, "cooking_session")
	defer span.End()

	span.SetAttributes(
		attribute.String("initial_display", m.Display()),
		attribute.Int("duration_seconds", seconds),
	)

	// Record cooking session metric
	if m.cookingSessions != nil {
		m.cookingSessions.Add(ctx, 1)
	}

	m.logger.InfoContext(ctx, "cooking started",
		"display", m.Display(),
		"seconds", seconds,
	)

	m.mu.Lock()
	m.isCooking = true
	m.mu.Unlock()

	completed := m.countdown(ctx, seconds)

	m.mu.Lock()
	m.isCooking = false
	// Reset state for next use
	// countdown may not have completed, leaving a non-zero time in the digits
	m.digits = [4]int{0, 0, 0, 0}
	m.digitCount = 0
	m.mu.Unlock()

	if completed {
		m.logger.InfoContext(ctx, "cooking complete")
	} else {
		m.logger.InfoContext(ctx, "cooking canceled")
	}
}

// totalSeconds calculates total seconds from the digit display
// Must be called with lock held
func (m *Microwave) totalSeconds() int {
	minutes := m.digits[0]*10 + m.digits[1]
	seconds := m.digits[2]*10 + m.digits[3]
	return minutes*60 + seconds
}

// countdown runs the cooking countdown. Returns true if completed, false if canceled.
// I recognize that the assignment stated that the microwave could not be stopped
// once started. This function allows for returning false for canceled for more
// efficient testing of edge cases and use with a sample driver program.
func (m *Microwave) countdown(ctx context.Context, seconds int) bool {
	for seconds > 0 {
		// Convert seconds back to display format
		mins := seconds / 60
		secs := seconds % 60

		// if mins > 99, add the extra seconds to secs
		// we shouldn't run into a situation where
		// we don't have enough room to display the
		// digits because we did input checking in
		// PressDigit to make sure the largest value
		// we accepted was 99:99. if we do find ourself
		// in a strange situation, then log a warning
		// and print **:** ?
		overflowMins := mins - 99
		if overflowMins == 1 {
			mins = 99
			secs = secs + overflowMins*60
		} else if overflowMins > 1 {
			// Trouble
			m.logger.WarnContext(ctx, "unexpected overflowMins > 1", "overflowMins", overflowMins)
			mins = 99
			secs = 99
		}

		// update the digits in the display
		// generate a new string from the display digits
		m.mu.Lock()
		m.digits[0] = mins / 10
		m.digits[1] = mins % 10
		m.digits[2] = secs / 10
		m.digits[3] = secs % 10
		display := m.displayString()
		m.mu.Unlock()

		m.logger.DebugContext(ctx, "tick", "display", display, "remaining", seconds)
		fmt.Print(display + "\r\n")

		// Wait for 1 second or context cancellation
		select {
		case <-ctx.Done():
			return false
		case <-time.After(1 * time.Second):
			seconds--
		}
	}

	// Print final 00:00
	m.mu.Lock()
	m.digits = [4]int{0, 0, 0, 0}
	display := m.displayString()
	m.mu.Unlock()
	fmt.Print(display + "\r\n")
	m.logger.DebugContext(ctx, "tick", "display", display, "remaining", seconds)
	return true
}
