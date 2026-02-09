package microwave

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"sync"
	"testing"
	"time"

	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

// Constructor Test Cases

// TestNew verifies that New() creates a properly initialized Microwave with default values.
// Test logic: Creates a new Microwave and checks all fields have expected default values:
// digits=[0,0,0,0], digitCount=0, isCooking=false, and all dependencies are non-nil.
func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}

	// Check digits default to [0, 0, 0, 0]
	if m.digits != [4]int{0, 0, 0, 0} {
		t.Errorf("expected digits [0,0,0,0], got %v", m.digits)
	}

	// Check digitCount defaults to 0
	if m.digitCount != 0 {
		t.Errorf("expected digitCount 0, got %d", m.digitCount)
	}

	// Check isCooking defaults to false
	if m.isCooking {
		t.Error("expected isCooking to be false")
	}

	// Check logger is initialized
	if m.logger == nil {
		t.Error("expected logger to be non-nil")
	}

	// Check tracer is initialized
	if m.tracer == nil {
		t.Error("expected tracer to be non-nil")
	}

	// Check meter is initialized
	if m.meter == nil {
		t.Error("expected meter to be non-nil")
	}

	// Check metrics are initialized
	if m.buttonPresses == nil {
		t.Error("expected buttonPresses to be non-nil")
	}
	if m.cookingSessions == nil {
		t.Error("expected cookingSessions to be non-nil")
	}
}

// TestNewWithLogger verifies that WithLogger option sets the logger correctly.
// Test logic: Creates a custom logger and passes it via WithLogger option,
// then verifies the Microwave's logger field points to the supplied logger.
func TestNewWithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))
	if m == nil {
		t.Fatal("expected New with logger to return non-nil")
	}

	// Check logger is set to user supplied logger
	if m.logger != logger {
		t.Error("expected logger to be set to user supplied logger")
	}
}

// TestNewWithTracer verifies that WithTracer option sets the tracer correctly.
// Test logic: Creates a tracer and passes it via WithTracer option,
// then verifies the Microwave's tracer field points to the supplied tracer.
func TestNewWithTracer(t *testing.T) {
	tracer := otel.Tracer("test")
	m := New(WithTracer(tracer))
	if m == nil {
		t.Fatal("expected New with tracer to return non-nil")
	}

	// Check tracer is set to user supplied tracer
	if m.tracer != tracer {
		t.Error("expected tracer to be set to user supplied tracer")
	}
}

// TestNewWithMeter verifies that WithMeter option sets the meter correctly.
// Test logic: Creates a meter and passes it via WithMeter option,
// then verifies the Microwave's meter field points to the supplied meter.
func TestNewWithMeter(t *testing.T) {
	meter := otel.Meter("test")
	m := New(WithMeter(meter))
	if m == nil {
		t.Fatal("expected New with meter to return non-nil")
	}

	// Check meter is set to user supplied meter
	if m.meter != meter {
		t.Error("expected meter to be set to user supplied meter")
	}
}

// displayString Test Cases

// TestDisplayString verifies that displayString formats digits as MM:SS.
// Test logic: Sets digits to [1,2,3,4] and verifies displayString returns "12:34".
func TestDisplayString(t *testing.T) {
	m := New()
	m.digits = [4]int{1, 2, 3, 4}

	got := m.displayString()
	expected := "12:34"

	if got != expected {
		t.Errorf("displayString() = %q, expected %q", got, expected)
	}
}

// Display Test Cases

// TestDisplay verifies that Display returns correct MM:SS format for various digit sequences.
// Test logic: Uses table-driven tests to press different digit sequences and verify
// the display shows the expected time format after each sequence.
func TestDisplay(t *testing.T) {
	var m *Microwave

	tests := []struct {
		digits   []int
		expected string
	}{
		{[]int{}, "00:00"},
		{[]int{1}, "00:01"},
		{[]int{1, 3}, "00:13"},
		{[]int{1, 3, 5}, "01:35"},
		{[]int{1, 2, 3, 4}, "12:34"},
	}

	for _, tt := range tests {
		m = New() // Reset
		for _, d := range tt.digits {
			m.PressDigit(d)
		}
		if got := m.Display(); got != tt.expected {
			t.Errorf("after pressing %v, expected %s, got %s", tt.digits, tt.expected, got)
		}
	}
}

// IsCooking Tests Cases

// TestIsCooking verifies that IsCooking returns the correct cooking state.
// Test logic: Uses table-driven tests to set isCooking to true/false and verify
// IsCooking() returns the expected value in each case.
func TestIsCooking(t *testing.T) {
	tests := []struct {
		name     string
		setState bool
	}{
		{"returns false by default", false},
		{"returns true when cooking", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New()
			m.mu.Lock()
			m.isCooking = tt.setState
			m.mu.Unlock()

			if got := m.IsCooking(); got != tt.setState {
				t.Errorf("IsCooking() = %t, want %t", got, tt.setState)
			}
		})
	}
}

// PressDigit Test Cases

// TestPressDigitInvalidDigit verifies that invalid digits (< 0 or > 9) are ignored.
// Test logic: Presses invalid digits -1 and 10, verifies display remains 00:00
// and "invalid digit ignored" warning appears in logs.
func TestPressDigitInvalidDigit(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	m.PressDigit(-1)

	if got := m.Display(); got != "00:00" {
		t.Errorf("Display() = %s, want 00:00", got)
	}

	m.PressDigit(10)

	if got := m.Display(); got != "00:00" {
		t.Errorf("Display() = %s, want 00:00", got)
	}

	logs := buf.String()
	if !strings.Contains(logs, "invalid digit ignored") {
		t.Error("expected 'invalid digit ignored' warning in logs")
	}
}

// TestPressDigitIgnoresPressesWhileCooking verifies that digit presses are ignored while cooking.
// Test logic: Presses digit 4, sets isCooking=true, presses digit 9, then verifies
// display still shows 00:04 and "digit ignored while cooking" warning appears in logs.
func TestPressDigitIgnoresPressesWhileCooking(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	m.PressDigit(4)

	if got := m.Display(); got != "00:04" {
		t.Errorf("Display() = %s, want 00:04", got)
	}

	m.mu.Lock()
	m.isCooking = true
	m.mu.Unlock()

	m.PressDigit(9)

	if got := m.Display(); got != "00:04" {
		t.Errorf("Display() = %s, want 00:04", got)
	}

	logs := buf.String()
	if !strings.Contains(logs, "digit ignored while cooking") {
		t.Error("expected 'digit ignored while cooking' warning in logs")
	}
	if !strings.Contains(logs, `"digit":9`) {
		t.Error("expected digit 9 in logs")
	}

}

// TestPressDigitMaxFourDigits verifies that only 4 digits are accepted.
// Test logic: Presses 6 digits (1-6), verifies display updates for first 4 then stays
// at 12:34 for digits 5 and 6, and "max digits reached" warning appears in logs.
func TestPressDigitMaxFourDigits(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	m := New(WithLogger(logger))

	// Enter 4 digits - display should update
	m.PressDigit(1)
	if m.Display() != "00:01" {
		t.Errorf("expected 00:01, got %s", m.Display())
	}
	m.PressDigit(2)
	if m.Display() != "00:12" {
		t.Errorf("expected 00:12, got %s", m.Display())
	}
	m.PressDigit(3)
	if m.Display() != "01:23" {
		t.Errorf("expected 01:23, got %s", m.Display())
	}
	m.PressDigit(4)
	if m.Display() != "12:34" {
		t.Errorf("expected 12:34, got %s", m.Display())
	}

	// 5th digit - display should NOT update
	m.PressDigit(5)
	if m.Display() != "12:34" {
		t.Errorf("expected 12:34 after 5th digit, got %s", m.Display())
	}

	// 6th digit - display should still NOT update
	m.PressDigit(6)
	if m.Display() != "12:34" {
		t.Errorf("expected 12:34 after 6th digit, got %s", m.Display())
	}

	// Verify warning logged for digits 5 and 6
	logs := buf.String()
	if !strings.Contains(logs, "max digits reached") {
		t.Error("expected 'max digits reached' warning in logs")
	}

	// Verify all 6 digit presses were logged
	count := strings.Count(logs, "digit pressed")
	if count != 6 {
		t.Errorf("expected 6 'digit pressed' logs, got %d", count)
	}
}

// TestPressDigitShiftsLeft verifies that digits shift left as new digits are entered.
// Test logic: Presses digits 1, 2, 3, 4 sequentially and verifies display shows
// 00:01, 00:12, 01:23, 12:34 after each press (left-shift behavior).
func TestPressDigitShiftsLeft(t *testing.T) {
	m := New()

	m.PressDigit(1)
	if m.Display() != "00:01" {
		t.Errorf("expected 00:01, got %s", m.Display())
	}

	m.PressDigit(2)
	if m.Display() != "00:12" {
		t.Errorf("expected 00:12, got %s", m.Display())
	}

	m.PressDigit(3)
	if m.Display() != "01:23" {
		t.Errorf("expected 01:23, got %s", m.Display())
	}

	m.PressDigit(4)
	if m.Display() != "12:34" {
		t.Errorf("expected 12:34, got %s", m.Display())
	}
}

// PressStart Test Cases

// TestPressStartIgnoresPressesWhileCooking verifies that PressStart is ignored while cooking.
// Test logic: Sets up time, sets isCooking=true, calls PressStart, then verifies
// "start ignored, already cooking" warning appears and cooking state is unchanged.
func TestPressStartIgnoresPressesWhileCooking(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	// Make sure there is time on the clock
	// so we avoid warning about starting with zero time.
	m.PressDigit(4)
	if got := m.Display(); got != "00:04" {
		t.Errorf("Display() = %s, want 00:04", got)
	}

	// Set cooking to true
	m.mu.Lock()
	m.isCooking = true
	m.mu.Unlock()

	// press start
	m.PressStart(context.Background())

	logs := buf.String()
	if !strings.Contains(logs, "start ignored, already cooking") {
		t.Error("expected 'start ignored, already cooking' warning in logs")
	}

	if m.IsCooking() != true {
		t.Error("should still be cooking after pressing start while cooking")
	}
}

// TestPressStartWithZeroTime verifies that PressStart is ignored when no time is set.
// Test logic: Calls PressStart without entering any digits, verifies "cannot start
// with zero time" warning appears and cooking does not start.
func TestPressStartWithZeroTime(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	// Don't enter any digits, try to start
	m.PressStart(context.Background())

	logs := buf.String()
	if !strings.Contains(logs, "cannot start with zero time") {
		t.Error("expected 'cannot start with zero time' warning")
	}

	if m.IsCooking() {
		t.Error("should not be cooking with zero time")
	}
}

// totalSeconds test cases

// TestTotalSeconds verifies that totalSeconds correctly converts digits to seconds.
// Test logic: Sets various digit combinations and verifies totalSeconds returns
// the correct number of seconds: 00:00=0, 00:05=5, 01:30=90, 99:99=6039.
func TestTotalSeconds(t *testing.T) {
	m := New()

	// 00:00 = 0 seconds
	m.mu.Lock()
	m.digits = [4]int{0, 0, 0, 0}
	s := m.totalSeconds()
	m.mu.Unlock()
	if s != 0 {
		t.Errorf("expected 0 seconds, got %d", s)
	}

	// 00:05 = 5 seconds
	m = New()
	m.mu.Lock()
	m.digits = [4]int{0, 0, 0, 5}
	s = m.totalSeconds()
	m.mu.Unlock()
	if s != 5 {
		t.Errorf("expected 5 seconds, got %d", s)
	}

	// Reset and test 01:30 = 90 seconds
	m = New()
	m.mu.Lock()
	m.digits = [4]int{0, 1, 3, 0}
	s = m.totalSeconds()
	m.mu.Unlock()
	if s != 90 {
		t.Errorf("expected 90 seconds for 01:30, got %d", s)
	}

	// Reset and test 99:99 = 6039 seconds (max)
	m = New()
	m.mu.Lock()
	m.digits = [4]int{9, 9, 9, 9}
	s = m.totalSeconds()
	m.mu.Unlock()
	if s != 6039 {
		t.Errorf("expected 6039 seconds for 99:99, got %d", s)
	}
}

// countdown Test Cases

// TestCountdownCompletesSuccessfully verifies that countdown returns true when it completes normally.
// Test logic: Calls countdown with 1 second, waits for completion, then checks the return value
// is true and the display shows 00:00.
func TestCountdownCompletesSuccessfully(t *testing.T) {
	m := New()

	// Call countdown with 1 second
	result := m.countdown(context.Background(), 1)

	// Should return true when completed normally
	if !result {
		t.Error("countdown() should return true when completed")
	}

	// Display should be 00:00 after countdown
	if got := m.Display(); got != "00:00" {
		t.Errorf("Display() = %s, want 00:00", got)
	}
}

// TestCountdownWithZeroSeconds verifies that countdown handles zero seconds correctly.
// Test logic: Calls countdown with 0 seconds, verifies it returns true immediately
// (the loop doesn't execute but final display is still printed) and display shows 00:00.
func TestCountdownWithZeroSeconds(t *testing.T) {
	m := New()

	// Call countdown with 0 seconds - should return immediately
	result := m.countdown(context.Background(), 0)

	// Should return true (loop doesn't execute, but final display is printed)
	if !result {
		t.Error("countdown(0) should return true")
	}

	// Display should be 00:00
	if got := m.Display(); got != "00:00" {
		t.Errorf("Display() = %s, want 00:00", got)
	}
}

// TestCountdownCancellation verifies that countdown returns false when context is already canceled.
// Test logic: Creates a context, cancels it immediately before calling countdown, then verifies
// countdown returns false due to the pre-canceled context.
func TestCountdownCancellation(t *testing.T) {
	m := New()

	// Create a context that we'll cancel
	ctx, cancel := context.WithCancel(context.Background())

	// Cancel immediately
	cancel()

	// Call countdown - should return false due to cancellation
	result := m.countdown(ctx, 10)

	if result {
		t.Error("countdown() should return false when context is canceled")
	}
}

// TestCountdownUpdatesDisplay verifies that countdown properly updates the digits array.
// Test logic: Sets initial digits to a non-zero value (12:34), runs countdown for 1 second,
// then verifies the display is reset to 00:00 after countdown completes.
func TestCountdownUpdatesDisplay(t *testing.T) {
	m := New()

	// Set initial digits to something non-zero
	m.mu.Lock()
	m.digits = [4]int{1, 2, 3, 4}
	m.mu.Unlock()

	// Run countdown for 1 second
	m.countdown(context.Background(), 1)

	// After countdown, display should be 00:00
	if got := m.Display(); got != "00:00" {
		t.Errorf("Display() = %s, want 00:00", got)
	}
}

// TestCountdownOverflowMinsEqualsOne verifies the overflowMins == 1 branch in countdown.
// Test logic: Starts countdown with 6000 seconds (100 minutes), which triggers the overflow
// handling where mins=100 becomes mins=99 and the extra 60 seconds are added to secs.
// Verifies display shows 99:xx after first tick.
func TestCountdownOverflowMinsEqualsOne(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	m := New(WithLogger(logger))

	// Create a context that cancels after we check the display
	ctx, cancel := context.WithCancel(context.Background())

	// 100 minutes = 6000 seconds, which triggers overflowMins == 1
	// 100 mins -> mins=100, overflowMins=1, so mins=99, secs=secs+60
	// Start with 6000 seconds (100:00)
	done := make(chan bool)
	go func() {
		m.countdown(ctx, 6000)
		done <- true
	}()

	// Wait briefly for first tick
	time.Sleep(100 * time.Millisecond)

	// Check display shows 99:xx (overflow handled)
	display := m.Display()
	if display[0:2] != "99" {
		t.Errorf("Display() = %s, expected to start with 99", display)
	}

	cancel()
	<-done
}

// TestCountdownOverflowMinsGreaterThanOne verifies the overflowMins > 1 branch in countdown.
// Test logic: Starts countdown with 12000 seconds (200 minutes), which triggers the extreme
// overflow case where display is capped at 99:99 and a warning is logged.
// Verifies display shows 99:99 and the warning message appears in logs.
func TestCountdownOverflowMinsGreaterThanOne(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	m := New(WithLogger(logger))

	ctx, cancel := context.WithCancel(context.Background())

	// 200 minutes = 12000 seconds, which triggers overflowMins > 1
	done := make(chan bool)
	go func() {
		m.countdown(ctx, 12000)
		done <- true
	}()

	// Wait briefly for first tick, then cancel and wait for goroutine
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	// Check display shows 99:99 (max overflow)
	if got := m.Display(); got != "99:99" {
		t.Errorf("Display() = %s, want 99:99", got)
	}

	// Check warning was logged
	logs := buf.String()
	if !strings.Contains(logs, "unexpected overflowMins > 1") {
		t.Error("expected 'unexpected overflowMins > 1' warning in logs")
	}
}

// Logging Test Cases

// TestLogging verifies that PressDigit logs the digit pressed message.
// Test logic: Creates a logger that writes to a buffer, presses digit 5,
// then verifies "digit pressed" message and digit value appear in logs.
func TestLogging(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))
	m.PressDigit(5)

	logs := buf.String()
	if !strings.Contains(logs, "digit pressed") {
		t.Error("expected 'digit pressed' in logs")
	}
	if !strings.Contains(logs, `"digit":5`) {
		t.Error("expected digit value 5 in logs")
	}
}

// TestLoggingDigitPressCount verifies that each digit press is logged separately.
// Test logic: Presses digit 9 twice, then counts occurrences of "digit":9 in logs
// to verify each press was logged individually.
func TestLoggingDigitPressCount(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	// Press digit 9 twice
	m.PressDigit(9)
	m.PressDigit(9)

	logs := buf.String()

	// Verify digit 9 was pressed exactly 2 times
	count := strings.Count(logs, `"digit":9`)
	if count != 2 {
		t.Errorf("expected 2 presses of digit 9, got %d", count)
	}
}

// Integration and Concurrency Test Cases
// Run with the race detector to check for data races:
//
//	go test -race -run <test-name> ./internal/microwave

// TestIntegrationTracerCreatesSpan verifies that cooking creates an OpenTelemetry span.
// Test logic: Sets up in-memory span exporter, starts cooking for 1 second, then
// verifies a "cooking_session" span was created and exported.
func TestIntegrationTracerCreatesSpan(t *testing.T) {
	// Create an in-memory span exporter
	exporter := tracetest.NewInMemoryExporter()
	tp := trace.NewTracerProvider(trace.WithSyncer(exporter))
	tracer := tp.Tracer("test")

	m := New(WithTracer(tracer))

	// Enter 1 second and start cooking
	m.PressDigit(1)
	m.PressStart(context.Background())

	// Verify a span was created
	spans := exporter.GetSpans()
	if len(spans) == 0 {
		t.Error("expected at least one span to be created")
	}

	// Verify the cooking_session span exists
	found := false
	for _, span := range spans {
		if span.Name == "cooking_session" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'cooking_session' span")
	}
}

// TestIntegrationMeterRecordsMetrics verifies that button presses are recorded as metrics.
// Test logic: Sets up manual metric reader, presses digits 1 and 2, collects metrics,
// then verifies "microwave.button_presses" metric exists in the collected data.
func TestIntegrationMeterRecordsMetrics(t *testing.T) {
	// Create a manual reader to collect metrics
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	meter := mp.Meter("test")

	m := New(WithMeter(meter))

	// Press some digits
	m.PressDigit(1)
	m.PressDigit(2)

	// Collect metrics
	var rm metricdata.ResourceMetrics
	if err := reader.Collect(context.Background(), &rm); err != nil {
		t.Fatalf("failed to collect metrics: %v", err)
	}

	// Verify button_presses metric exists
	found := false
	for _, sm := range rm.ScopeMetrics {
		for _, metric := range sm.Metrics {
			if metric.Name == "microwave.button_presses" {
				found = true
				break
			}
		}
	}
	if !found {
		t.Error("expected 'microwave.button_presses' metric")
	}
}

// TestIntegrationDisplayConcurrent verifies that Display() is safe for concurrent access.
// Test logic: Spawns 100 writer goroutines calling PressDigit and 100 reader goroutines
// calling Display() simultaneously. Verifies display format remains valid (MM:SS).
func TestIntegrationDisplayConcurrent(t *testing.T) {
	m := New()

	// Run concurrent readers and writers
	var wg sync.WaitGroup
	for i := range 100 {
		wg.Add(2)

		// Writer goroutine
		go func(digit int) {
			defer wg.Done()
			m.PressDigit(digit % 10)
		}(i)

		// Reader goroutine
		go func() {
			defer wg.Done()
			display := m.Display()
			// Verify format is always valid MM:SS
			if len(display) != 5 || display[2] != ':' {
				t.Errorf("corrupted display: %q", display)
			}
		}()
	}

	wg.Wait()
}

// TestIntegrationIsCookingConcurrent verifies that IsCooking() is safe for concurrent access.
// Test logic: Spawns 100 writer goroutines toggling isCooking and 100 reader goroutines
// calling IsCooking() simultaneously. Must pass with race detector enabled.
func TestIntegrationIsCookingConcurrent(t *testing.T) {
	m := New()

	// Run concurrent readers and writers
	var wg sync.WaitGroup
	for range 100 {
		wg.Add(2)

		// Writer goroutine
		go func() {
			defer wg.Done()
			m.mu.Lock()
			m.isCooking = !m.isCooking
			m.mu.Unlock()
		}()

		// Reader goroutine
		go func() {
			defer wg.Done()
			_ = m.IsCooking()
		}()
	}

	wg.Wait()
}

// TestIntegrationPressDigitWhileCooking verifies that digit presses are ignored during cooking.
// Test logic: Starts cooking in a goroutine, waits for cooking to begin, presses digit 5,
// then cancels and verifies "digit ignored while cooking" warning appears in logs.
func TestIntegrationPressDigitWhileCooking(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	// Enter 2 seconds
	m.PressDigit(2)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start cooking in a goroutine
	done := make(chan bool)
	go func() {
		m.PressStart(ctx)
		done <- true
	}()

	// Wait for cooking to start
	timeout := time.After(5 * time.Second)
	for !m.IsCooking() {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for cooking to start")
		default:
			// continue waiting
		}
	}

	// Try to press digit while cooking
	m.PressDigit(5)

	// Cancel cooking and wait for goroutine to finish
	cancel()
	<-done

	logs := buf.String()
	if !strings.Contains(logs, "digit ignored while cooking") {
		t.Error("expected 'digit ignored while cooking' warning")
	}
}

// TestIntegrationPressStartWhileCooking verifies that start presses are ignored during cooking.
// Test logic: Starts cooking in a goroutine, waits for cooking to begin, calls PressStart,
// then cancels and verifies "start ignored, already cooking" warning appears in logs.
func TestIntegrationPressStartWhileCooking(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	// Enter 2 seconds
	m.PressDigit(2)

	// Create a cancellable context
	ctx, cancel := context.WithCancel(context.Background())

	// Start cooking in a goroutine
	done := make(chan bool)
	go func() {
		m.PressStart(ctx)
		done <- true
	}()

	// Wait for cooking to start
	timeout := time.After(5 * time.Second)
	for !m.IsCooking() {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for cooking to start")
		default:
			// continue waiting
		}
	}

	// Try to start again while cooking
	m.PressStart(ctx)

	// Cancel cooking and wait for goroutine to finish
	cancel()
	<-done

	logs := buf.String()
	if !strings.Contains(logs, "start ignored, already cooking") {
		t.Error("expected 'start ignored, already cooking' warning")
	}
}

// TestIntegrationPressStartStartsCooking verifies the full cooking cycle.
// Test logic: Sets time to 2 seconds, starts cooking in goroutine, verifies IsCooking is true,
// waits for completion, then verifies logs, state reset, and display is ready for reuse.
func TestIntegrationPressStartStartsCooking(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	// Make sure there is time on the clock
	m.mu.Lock()
	m.digits = [4]int{0, 0, 0, 2}
	m.mu.Unlock()

	// Start cooking in a goroutine
	done := make(chan bool)
	go func() {
		m.PressStart(context.Background())
		done <- true
	}()

	// Wait for cooking to start
	timeout := time.After(5 * time.Second)
	for !m.IsCooking() {
		select {
		case <-timeout:
			t.Fatal("timed out waiting for cooking to start")
		default:
			// continue waiting
		}
	}

	// Verify cooking flag is true while cooking
	if !m.IsCooking() {
		t.Error("expected IsCooking() to be true while cooking")
	}

	// Wait for cooking to complete
	<-done

	// Look for Log messages
	logs := buf.String()

	// Look for the cooking started message
	if !strings.Contains(logs, "cooking started") {
		t.Error("expected 'cooking started' warning in logs")
	}

	// Look for the cooking complete message
	if !strings.Contains(logs, "cooking complete") {
		t.Error("expected 'cooking complete' warning in logs")
	}

	// Check that we are not still cooking
	if m.IsCooking() {
		t.Error("should not be cooking after cooking completes")
	}

	// Check that the display is back to zero
	if got := m.Display(); got != "00:00" {
		t.Errorf("Display() = %s, want 00:00", got)
	}

	// Check that the digitCount is at zero
	m.mu.Lock()
	got := m.digitCount
	m.mu.Unlock()
	if got != 0 {
		t.Errorf("digitCount = %d, want 0", got)
	}

	// Verify the display is ready for use again
	m.PressDigit(5)
	if got := m.Display(); got != "00:05" {
		t.Errorf("Display() = %s, want 00:05 (digitCount should have been reset)", got)
	}
}

// TestIntegrationCountdownConcurrent verifies that countdown() is safe for concurrent access.
// Test logic: Starts countdown in a goroutine while spawning 50 reader goroutines that call
// Display() and 50 that call IsCooking() concurrently. Must pass with race detector enabled.
func TestIntegrationCountdownConcurrent(t *testing.T) {
	m := New()

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup

	// Start countdown in a goroutine
	wg.Add(1)
	go func() {
		defer wg.Done()
		m.countdown(ctx, 5)
	}()

	// Spawn multiple readers that check Display and IsCooking
	for range 50 {
		wg.Add(2)

		go func() {
			defer wg.Done()
			_ = m.Display()
		}()

		go func() {
			defer wg.Done()
			_ = m.IsCooking()
		}()
	}

	// Let it run briefly then cancel
	time.Sleep(100 * time.Millisecond)
	cancel()
	wg.Wait()
}

// TestIntegrationCountdownCancellationMidway verifies countdown can be canceled mid-execution.
// Test logic: Starts countdown with 10 seconds in a goroutine, waits 500ms for it to begin,
// then cancels the context. Verifies countdown returns false due to mid-execution cancellation.
func TestIntegrationCountdownCancellationMidway(t *testing.T) {
	m := New()

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)
	var result bool
	go func() {
		result = m.countdown(ctx, 10)
		done <- true
	}()

	// Wait for countdown to start
	time.Sleep(500 * time.Millisecond)

	// Cancel midway
	cancel()

	// Wait for countdown to finish with timeout
	timeout := time.After(5 * time.Second)
	select {
	case <-done:
		// Expected
	case <-timeout:
		t.Fatal("timed out waiting for countdown to finish")
	}

	// Should return false due to cancellation
	if result {
		t.Error("countdown() should return false when canceled midway")
	}
}
