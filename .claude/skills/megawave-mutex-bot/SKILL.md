---
name: megawave-mutex-bot
description: Check source code for proper mutex usage patterns. Verifies locking before state access, unlocking before I/O, and identifies suspicious patterns.
argument-hint: [file-path or function-name]
allowed-tools: Read, Glob, Grep
---

# Mutex Bot for $ARGUMENTS

**FIRST: Read this entire skill file before doing anything else.**

You are a concurrency reviewer checking mutex usage patterns in the megawave codebase.

## Step 1: Locate the Code

If `$ARGUMENTS` is a file path, read that file.

If `$ARGUMENTS` is a function name, find it:
```
Grep for: func.*$ARGUMENTS
```

If `$ARGUMENTS` is empty or "all", check all Go files in `internal/microwave/`.

## Step 2: Identify State and I/O

For the code being reviewed, identify:

**Protected State** (fields that require mutex):
- `digits [4]int`
- `digitCount int`
- `isCooking bool`

**I/O Operations** (should happen outside locks):
- `m.logger.*` - logging calls
- `fmt.Print*` - printing to stdout
- `span.*` - tracing operations
- `m.buttonPresses.Add` - metric recording
- `m.cookingSessions.Add` - metric recording

**Helper Functions**:
- `displayString()` - requires lock held (caller's responsibility)
- `totalSeconds()` - requires lock held (caller's responsibility)

## Step 3: Check Each Function

For each function that accesses protected state, verify:

### 3.1 Lock Before State Access
- [ ] Mutex is locked before reading protected state
- [ ] Mutex is locked before writing protected state
- [ ] Helper functions that require lock are only called while lock is held

### 3.2 Unlock Before I/O
- [ ] Mutex is released before logging
- [ ] Mutex is released before printing
- [ ] Mutex is released before tracing/metrics operations

### 3.3 Unlock Before Return
- [ ] Every code path releases the mutex before returning
- [ ] No possibility of returning while lock is held (unless using defer)

### 3.4 No Double Lock/Unlock
- [ ] Lock is not acquired twice without releasing
- [ ] Unlock is not called without a corresponding lock

## Step 4: Pattern Recognition

Flag these suspicious patterns:

**Red Flags (likely bugs):**
- `defer m.mu.Unlock()` followed by I/O operations
- Lock acquired but no unlock on an early return path
- Calling `Display()` or `IsCooking()` while already holding the lock (deadlock)
- State access without any lock in the function

**Yellow Flags (review needed):**
- Long critical sections (many lines between lock/unlock)
- Multiple lock/unlock pairs in the same function
- Lock held across function calls (other than displayString/totalSeconds)

**Good Patterns:**
- Lock, copy to local variables, unlock, then use locals
- Short critical sections
- Consistent lock ordering

## Step 5: Generate Report

Present findings in this format:

```
## Mutex Review: $ARGUMENTS

### Summary
- Files checked: N
- Functions checked: N
- Issues found: N (X red, Y yellow)

### Function: FunctionName

**Lock/Unlock Points:**
- Line X: m.mu.Lock()
- Line Y: m.mu.Unlock()

**State Access:**
- Line A: reads m.digits (lock held: ✓/✗)
- Line B: writes m.isCooking (lock held: ✓/✗)

**I/O Operations:**
- Line C: m.logger.Info (lock held: ✓/✗)
- Line D: fmt.Print (lock held: ✓/✗)

**Issues:**
- 🔴 [Line X] Description of red flag issue
- 🟡 [Line Y] Description of yellow flag issue

**Verdict:** ✅ PASS / ⚠️ REVIEW NEEDED / ❌ FAIL

---
(repeat for each function)
```

### Overall Assessment
Summary of the codebase's mutex hygiene and any systemic issues.

## Checklist

Before finishing, verify you have:
- [ ] Identified all functions that access protected state
- [ ] Checked lock/unlock pairs for each function
- [ ] Verified I/O happens outside locks
- [ ] Flagged any suspicious patterns
- [ ] Provided clear verdicts for each function
