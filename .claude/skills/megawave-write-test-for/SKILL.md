---
name: megawave-write-test-for
description: Write comprehensive test cases for a function in the megawave codebase. Acts as a software engineer to create unit and integration tests.
argument-hint: [function-name]
allowed-tools: Read, Glob, Grep, Edit, Write, Bash, AskUserQuestion
---

# Write Tests for $ARGUMENTS

**FIRST: Read this entire skill file before doing anything else.**

You are a software engineer writing tests for the `$ARGUMENTS` function in the megawave codebase.

**IMPORTANT: Do NOT write any test code until the user confirms the plan.**

## Step 1: Locate and Read the Function

First, find the function `$ARGUMENTS` in the codebase:

```
Grep for: func.*$ARGUMENTS
```

Read the function implementation completely to understand:
- Input parameters and their types
- Return values and their types
- All code branches (if statements, switch/case statements)
- Any mutex locking patterns
- Dependencies on other methods or fields

## Step 2: Analyze Test Requirements

For the function, identify:

1. **Input Validation Tests**
   - Valid inputs (happy path)
   - Invalid inputs (edge cases, boundary values)
   - Zero/nil/empty values
   - Out-of-range values

2. **Return Value Tests**
   - Expected return values for various inputs
   - Error conditions and error returns

3. **Branch Coverage Tests**
   - Each `if` condition (true and false paths)
   - Each `switch`/`case` branch
   - Each `else` block
   - Early returns

4. **State Tests**
   - State changes caused by the function
   - Side effects on struct fields

5. **Concurrency Tests** (if applicable)
   - Does the function use mutex locking?
   - Does the function read or write shared state?
   - Can the function be called from multiple goroutines?

## Step 3: Create Test Plan and Ask for Confirmation

**STOP HERE AND PRESENT THE PLAN TO THE USER.**

Create a test plan in the following format and present it to the user:

```
## Test Plan for $ARGUMENTS

### Unit Tests

| # | Test Name | Purpose | Code Changes |
|---|-----------|---------|--------------|
| 1 | Test$ARGUMENTS... | What this test verifies | Brief description of test logic |
| 2 | Test$ARGUMENTS... | What this test verifies | Brief description of test logic |

### Integration Tests

| # | Test Name | Purpose | Code Changes |
|---|-----------|---------|--------------|
| 1 | TestIntegration$ARGUMENTS... | What this test verifies | Brief description of test logic |

### Concurrency Tests (if applicable)

| # | Test Name | Purpose | Code Changes |
|---|-----------|---------|--------------|
| 1 | TestIntegration$ARGUMENTSConcurrent | What this test verifies | Brief description of test logic |

### Branch Coverage Summary
- [ ] Branch 1: covered by Test...
- [ ] Branch 2: covered by Test...
```

After presenting the plan, use AskUserQuestion to ask:
- "Do you approve this test plan for $ARGUMENTS?"
- Options: "Yes, write the tests" / "No, I want to modify the plan"

**DO NOT proceed to Step 4 until the user confirms the plan.**

## Step 4: Write Unit Tests (only after user confirmation)

Place unit tests grouped under a comment header for the function:

```go
// $ARGUMENTS Test Cases

func Test$ARGUMENTS(t *testing.T) { }
func Test$ARGUMENTS<Scenario>(t *testing.T) { }
```

Unit test guidelines:
- Test the function in isolation
- Mock or directly set internal state when needed
- Use table-driven tests when testing multiple inputs
- Check both the return value AND any state changes
- Verify log messages when relevant
- **Document each test** with a header comment describing what it verifies and the test logic

Example table-driven test:
```go
// Test<FunctionName> verifies that <FunctionName> returns correct values for various inputs.
// Test logic: Creates a new Microwave, calls <FunctionName> with each input, and checks
// the return value matches the expected output.
func Test<FunctionName>(t *testing.T) {
    tests := []struct {
        name     string
        input    <type>
        expected <type>
    }{
        {"description of case", input, expected},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            m := New()

            // Call the function with test input
            got := m.<FunctionName>(tt.input)

            // Verify return value matches expected
            if got != tt.expected {
                t.Errorf("<FunctionName>() = %v, want %v", got, tt.expected)
            }
        })
    }
}
```

Example simple test (no table):
```go
// Test<FunctionName><Scenario> verifies that <FunctionName> <what it does in this scenario>.
// Test logic: Sets up <initial state>, calls <FunctionName>, then checks <expected outcome>.
func Test<FunctionName><Scenario>(t *testing.T) {
    // Create logger to capture log output
    var buf bytes.Buffer
    logger := slog.New(slog.NewJSONHandler(&buf, nil))

    m := New(WithLogger(logger))

    // Setup initial state
    m.mu.Lock()
    m.someField = someValue
    m.mu.Unlock()

    // Call the function being tested
    m.<FunctionName>(args)

    // Verify state change via public getter
    if got := m.<GetterFunction>(); got != expected {
        t.Errorf("<GetterFunction>() = %v, want %v", got, expected)
    }

    // Verify expected log message was written
    logs := buf.String()
    if !strings.Contains(logs, "expected log message") {
        t.Error("expected 'expected log message' in logs")
    }
}
```

## Step 5: Write Integration Tests (only after user confirmation)

Place integration tests at the bottom of the test file in the "Integration Test Cases" section.

Integration test guidelines:
- Test the function with real goroutines and actual cooking cycles
- Use cancellable contexts for tests involving cooking
- Add timeout protection for busy-wait loops
- Run with race detector: `go test -race`
- Test concurrent access patterns
- **Document each test** with a header comment describing what it verifies and the test logic

Example cooking-related integration test:
```go
// TestIntegration<Name> verifies <what this test checks>.
// Test logic: Starts cooking in a goroutine, waits for cooking to begin,
// performs <test action>, then cancels and verifies <expected outcome>.
func TestIntegration<Name>(t *testing.T) {
    m := New()

    // Set cooking time to 2 seconds
    m.PressDigit(2)

    // Create cancellable context for graceful shutdown
    ctx, cancel := context.WithCancel(context.Background())

    // Start cooking in a separate goroutine
    done := make(chan bool)
    go func() {
        m.PressStart(ctx)
        done <- true
    }()

    // Wait for cooking to start with timeout protection
    timeout := time.After(5 * time.Second)
    for !m.IsCooking() {
        select {
        case <-timeout:
            t.Fatal("timed out waiting for cooking to start")
        default:
            // continue waiting
        }
    }

    // Perform test action while cooking
    // ...

    // Cancel cooking and wait for goroutine to finish
    cancel()
    <-done

    // Verify expected outcome
    // ...
}
```

## Step 6: Write Concurrency Tests (if applicable, only after user confirmation)

If the function uses mutex locking or accesses shared state, write a concurrency test.

Concurrency test guidelines:
- Use multiple goroutines to read and write concurrently
- Writer goroutines modify state (call the function or set fields)
- Reader goroutines read state (call getter functions)
- Must pass with race detector: `go test -race`
- **Document each test** with a header comment describing what it verifies and the test logic

Example concurrency test:
```go
// TestIntegration<FunctionName>Concurrent verifies that <FunctionName>() is safe for concurrent access.
// Test logic: Spawns 100 writer goroutines calling <FunctionName> and 100 reader goroutines
// calling <GetterFunction> simultaneously. Must pass with race detector enabled.
func TestIntegration<FunctionName>Concurrent(t *testing.T) {
    m := New()

    var wg sync.WaitGroup
    for range 100 {
        wg.Add(2)

        // Writer goroutine - modifies state by calling <FunctionName>
        go func() {
            defer wg.Done()
            m.<FunctionName>(args)
        }()

        // Reader goroutine - reads state via <GetterFunction>
        go func() {
            defer wg.Done()
            _ = m.<GetterFunction>()
        }()
    }

    // Wait for all goroutines to complete
    wg.Wait()
}
```

## Step 7: Verify Tests

After writing tests, run them:

```bash
go test -v -run Test<FunctionName> ./internal/microwave
go test -race -run TestIntegration<FunctionName> ./internal/microwave
```

## Test File Location

Tests go in: `internal/microwave/microwave_test.go`

## Existing Test Patterns

Reference existing tests in the codebase for consistent style:
- `TestNew` - Constructor tests
- `TestDisplay` - Table-driven tests
- `TestIsCooking` - State tests
- `TestPressDigitInvalidDigit` - Input validation
- `TestIntegrationDisplayConcurrent` - Concurrency tests
- `TestIntegrationPressStartStartsCooking` - Full cycle integration tests

## Checklist

Before finishing, verify:
- [ ] All code branches have test coverage
- [ ] Unit tests are grouped under "// $ARGUMENTS Test Cases" comment
- [ ] Integration tests are in the integration section
- [ ] Each test has a header comment describing what it verifies
- [ ] Each test has a "Test logic:" line explaining how the test works
- [ ] Test code has inline comments explaining key steps
- [ ] Table-driven tests used where appropriate
- [ ] Error messages include actual and expected values
- [ ] Tests use public API (Display, IsCooking) when possible
- [ ] Mutex locking used only when accessing private fields
- [ ] Timeout protection on busy-wait loops
- [ ] Concurrency tests written for functions with mutex locking
- [ ] Tests pass with `go test -race`
