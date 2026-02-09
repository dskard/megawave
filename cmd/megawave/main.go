package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.opentelemetry.io/otel"
	"golang.org/x/term"

	"github.com/dskard/megawave/internal/microwave"
	"github.com/dskard/megawave/internal/telemetry"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(),
		os.Interrupt,    // Ctrl-C
		syscall.SIGTERM, // kill command
	)
	defer cancel()

	// Parse config (flags override env vars)
	cfg := telemetry.ParseConfig()

	// Initialize OTel if in production
	var otelShutdown func(context.Context) error
	if cfg.Environment == telemetry.Production {
		var err error
		otelShutdown, err = telemetry.InitOTel(ctx, cfg)
		if err != nil {
			log.Fatal(err)
		}
		// Shutdown with fresh context (not the canceled one) to allow flushing
		defer func() {
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer shutdownCancel()
			_ = otelShutdown(shutdownCtx)
		}()
	}

	// Create logger based on config (returns cleanup function for file handle)
	logger, closeLog := telemetry.NewLogger(cfg)
	defer func() { _ = closeLog() }()

	// Create microwave
	m := microwave.New(
		microwave.WithLogger(logger),
		microwave.WithTracer(otel.Tracer("megawave")),
		microwave.WithMeter(otel.Meter("megawave")),
	)

	// Print instructions
	printInstructions()

	// Run interactive loop
	if err := runInteractive(ctx, cancel, m); err != nil {
		if err != context.Canceled {
			log.Fatal(err)
		}
	}

	fmt.Println("\nGoodbye!")
}

func printInstructions() {
	fmt.Println("╔════════════════════════════════════════╗")
	fmt.Println("║         MEGAWAVE MICROWAVE             ║")
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Println("║  Controls:                             ║")
	fmt.Println("║    0-9       : Enter time digits       ║")
	fmt.Println("║    Enter     : Start cooking           ║")
	fmt.Println("║    Ctrl-C    : Exit                    ║")
	fmt.Println("╠════════════════════════════════════════╣")
	fmt.Println("║  Display format: MM:SS                 ║")
	fmt.Println("║  Example: Press 1,3,5 for 01:35        ║")
	fmt.Println("╚════════════════════════════════════════╝")
	fmt.Println()
	fmt.Println("Ready. Enter time and press Enter to start.")
	fmt.Println()
}

func runInteractive(ctx context.Context, cancel context.CancelFunc, m *microwave.Microwave) error {
	// Set terminal to raw mode to capture individual keypresses
	oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
	if err != nil {
		return fmt.Errorf("failed to set raw mode: %w", err)
	}
	defer func() { _ = term.Restore(int(os.Stdin.Fd()), oldState) }()

	// Channel to receive keypresses
	keyChan := make(chan byte, 1)
	errChan := make(chan error, 1)

	// Start goroutine to read keypresses
	go func() {
		buf := make([]byte, 1)
		for {
			n, err := os.Stdin.Read(buf)
			if err != nil {
				errChan <- err
				return
			}
			if n > 0 {
				// In raw mode, Ctrl-C doesn't generate SIGINT, so we
				// cancel the context here to allow immediate interruption
				if buf[0] == 3 { // Ctrl-C
					cancel()
				}
				keyChan <- buf[0]
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-errChan:
			return err
		case char := <-keyChan:
			switch {
			case char >= '0' && char <= '9':
				// Digit pressed
				digit := int(char - '0')
				m.PressDigit(digit)

			case char == '\r' || char == '\n':
				// Enter pressed
				m.PressStart(ctx)

			case char == 3: // Ctrl-C
				return context.Canceled
			}
		}
	}
}
