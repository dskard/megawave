package microwave

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	m := New()
	if m == nil {
		t.Fatal("New() returned nil")
	}
	if m.Display() != "00:00" {
		t.Errorf("expected initial display 00:00, got %s", m.Display())
	}
	if m.IsCooking() {
		t.Error("expected IsCooking() to be false initially")
	}
}

func TestDisplay(t *testing.T) {
	m := New()

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

func TestMaxFourDigits(t *testing.T) {
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

func TestInvalidDigit(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))

	m.PressDigit(-1)
	m.PressDigit(10)

	logs := buf.String()
	if !strings.Contains(logs, "invalid digit ignored") {
		t.Error("expected 'invalid digit ignored' warning in logs")
	}

	// Display should still be 00:00
	if m.Display() != "00:00" {
		t.Errorf("expected 00:00 after invalid digits, got %s", m.Display())
	}
}

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

func TestTotalSeconds(t *testing.T) {
	m := New()

	// 00:05 = 5 seconds
	m.PressDigit(5)
	m.mu.Lock()
	s := m.totalSeconds()
	m.mu.Unlock()
	if s != 5 {
		t.Errorf("expected 5 seconds, got %d", s)
	}

	// Reset and test 01:30 = 90 seconds
	m = New()
	m.PressDigit(1)
	m.PressDigit(3)
	m.PressDigit(0)
	m.mu.Lock()
	s = m.totalSeconds()
	m.mu.Unlock()
	if s != 90 {
		t.Errorf("expected 90 seconds for 01:30, got %d", s)
	}

	// Reset and test 99:99 = 6039 seconds (max)
	m = New()
	m.PressDigit(9)
	m.PressDigit(9)
	m.PressDigit(9)
	m.PressDigit(9)
	m.mu.Lock()
	s = m.totalSeconds()
	m.mu.Unlock()
	if s != 6039 {
		t.Errorf("expected 6039 seconds for 99:99, got %d", s)
	}
}

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

func TestWithLogger(t *testing.T) {
	var buf bytes.Buffer
	logger := slog.New(slog.NewJSONHandler(&buf, nil))

	m := New(WithLogger(logger))
	m.PressDigit(1)

	if buf.Len() == 0 {
		t.Error("expected logger to capture output")
	}
}

func TestDigitPressCount(t *testing.T) {
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
