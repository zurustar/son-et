# Testing Guidelines for Engine Package

## Test Isolation Requirements

### Critical Rule: Always Call ResetEngineForTest()

**Every test that uses global engine state MUST call `ResetEngineForTest()` at the beginning.**

```go
func TestMyFeature(t *testing.T) {
    // REQUIRED: Reset engine state before test
    ResetEngineForTest()
    
    // Your test code here...
}
```

### What ResetEngineForTest() Does

`ResetEngineForTest()` cleans up all global engine state:

- **Sequencers**: Resets `mainSequencer` and `sequencers` to nil
- **Global Variables**: Clears `globalVars` map
- **Timing State**: Resets `tickCount`, `ticksPerStep`, `targetTick` to defaults
- **Sync Mode**: Resets `midiSyncMode` to false
- **PPQ**: Resets `GlobalPPQ` to 480
- **Termination Flag**: Resets `programTerminated` to false
- **Global Engine**: Calls `Reset()` on `globalEngine` if it exists

### Why Test Isolation Matters

Without proper test isolation:
- Tests may fail when run together but pass when run alone
- State from one test can leak into another test
- Background goroutines from previous tests may interfere
- Race conditions become more likely

### Common Test Isolation Issues

#### Issue 1: Background Goroutines

**Problem**: Tests that start background goroutines (e.g., for `UpdateVM`) may leave them running after the test completes.

**Solution**: Always stop background goroutines before test cleanup:

```go
func TestWithBackgroundVM(t *testing.T) {
    ResetEngineForTest()
    
    // Start background VM
    stopVM := make(chan bool)
    go func() {
        ticker := time.NewTicker(10 * time.Millisecond)
        defer ticker.Stop()
        tickNum := 0
        for {
            select {
            case <-stopVM:
                return
            case <-ticker.C:
                UpdateVM(tickNum)
                tickNum++
            }
        }
    }()
    
    // CRITICAL: Stop background goroutine before test ends
    defer func() { stopVM <- true }()
    
    // Your test code here...
}
```

#### Issue 2: Shared Global State

**Problem**: Tests that modify global variables without resetting them can affect other tests.

**Solution**: Always use `ResetEngineForTest()` and avoid direct global variable access when possible.

```go
func TestWithGlobalState(t *testing.T) {
    ResetEngineForTest()
    
    // Use globalVars through proper API
    globalVars["myvar"] = 123
    
    // Test code...
    
    // ResetEngineForTest() will clean up globalVars in next test
}
```

#### Issue 3: Timing State Not Reset

**Problem**: Tests that modify `targetTick`, `tickCount`, or other timing variables can affect subsequent tests.

**Solution**: `ResetEngineForTest()` now resets all timing state including `targetTick`.

```go
func TestWithTiming(t *testing.T) {
    ResetEngineForTest() // Resets targetTick, tickCount, etc.
    
    // Modify timing state
    atomic.StoreInt64(&targetTick, 100)
    
    // Test code...
    
    // Next test will have clean timing state
}
```

## Testing TIME Mode Sequences

### Pattern: Start UpdateVM Before RegisterSequence

When testing `RegisterSequence` in TIME mode, always start `UpdateVM` in a background goroutine BEFORE calling `RegisterSequence`:

```go
func TestTimeModeSequence(t *testing.T) {
    ResetEngineForTest()
    
    // Start UpdateVM FIRST
    stopVM := make(chan bool)
    go func() {
        ticker := time.NewTicker(10 * time.Millisecond)
        defer ticker.Stop()
        tickNum := 0
        for {
            select {
            case <-stopVM:
                return
            case <-ticker.C:
                UpdateVM(tickNum)
                tickNum++
            }
        }
    }()
    defer func() { stopVM <- true }()
    
    // Give VM time to start
    time.Sleep(50 * time.Millisecond)
    
    // NOW call RegisterSequence
    ops := []OpCode{...}
    go func() {
        RegisterSequence(Time, ops)
    }()
    
    // Test assertions...
}
```

**Why**: `RegisterSequence` in TIME mode blocks until the sequence completes. The sequence can only complete if `UpdateVM` is running. If you call `RegisterSequence` before starting `UpdateVM`, the sequence will never complete.

### Pattern: Use Ticker for Consistent Timing

Use `time.Ticker` for consistent `UpdateVM` calls:

```go
ticker := time.NewTicker(10 * time.Millisecond)
defer ticker.Stop()
tickNum := 0
for {
    select {
    case <-stopVM:
        return
    case <-ticker.C:
        UpdateVM(tickNum)
        tickNum++
    }
}
```

**Avoid**: Manual loops with `time.Sleep()` can have inconsistent timing:

```go
// DON'T DO THIS
for i := 0; i < 20; i++ {
    UpdateVM(i)
    time.Sleep(10 * time.Millisecond)
}
```

## Testing MIDI_TIME Mode Sequences

MIDI_TIME mode is non-blocking, so the pattern is different:

```go
func TestMidiTimeModeSequence(t *testing.T) {
    ResetEngineForTest()
    
    ops := []OpCode{...}
    
    // RegisterSequence returns immediately in MIDI_TIME mode
    RegisterSequence(MidiTime, ops)
    
    // Simulate MIDI ticks
    for i := 0; i < 100; i++ {
        NotifyTick(1) // Advance by 1 MIDI tick
    }
    
    // Test assertions...
}
```

## Running Tests

### Run All Tests

```bash
go test -timeout=30s ./pkg/engine/...
```

### Run Specific Test

```bash
go test -timeout=30s -v -run TestMyFeature ./pkg/engine/
```

### Run with Race Detector

```bash
go test -timeout=30s -race ./pkg/engine/...
```

**Note**: Always use `-timeout` flag to prevent tests from hanging indefinitely.

## Known Issues and Workarounds

### Issue: TestRegisterSequenceBlocksInTimeMode fails in full suite

**Status**: FIXED in Task 32.1

**Cause**: `UpdateVM` was marking sequences as finished even when they yielded (e.g., during `Wait`).

**Fix**: Use `goto` to skip sequence completion check when yielding.

### Issue: Property tests timeout

**Status**: RESOLVED (tests complete within 30s timeout)

**Workaround**: If property tests take too long, reduce the number of test cases or increase timeout.

## Best Practices

1. **Always call `ResetEngineForTest()` at the start of each test**
2. **Stop all background goroutines before test ends**
3. **Use `defer` to ensure cleanup happens even if test fails**
4. **Start `UpdateVM` before `RegisterSequence` in TIME mode tests**
5. **Use `time.Ticker` for consistent timing in tests**
6. **Always use `-timeout` flag when running tests**
7. **Run tests with `-race` flag to detect race conditions**
8. **Avoid direct global variable access when possible**

## Example: Complete Test Template

```go
func TestCompleteExample(t *testing.T) {
    // 1. Reset engine state
    ResetEngineForTest()
    
    // 2. Initialize test engine if needed
    globalEngine = NewTestEngine()
    
    // 3. Start background VM if testing TIME mode
    stopVM := make(chan bool)
    go func() {
        ticker := time.NewTicker(10 * time.Millisecond)
        defer ticker.Stop()
        tickNum := 0
        for {
            select {
            case <-stopVM:
                return
            case <-ticker.C:
                UpdateVM(tickNum)
                tickNum++
            }
        }
    }()
    defer func() { stopVM <- true }()
    
    // 4. Give VM time to start
    time.Sleep(50 * time.Millisecond)
    
    // 5. Create and register sequence
    ops := []OpCode{
        {Cmd: interpreter.OpAssign, Args: []any{Variable("x"), 1}},
    }
    
    done := make(chan bool)
    go func() {
        RegisterSequence(Time, ops)
        done <- true
    }()
    
    // 6. Wait for completion or timeout
    select {
    case <-done:
        // Success
    case <-time.After(2 * time.Second):
        t.Fatal("Test timed out")
    }
    
    // 7. Verify results
    vmLock.Lock()
    if val, ok := globalVars["x"]; !ok || val != 1 {
        t.Errorf("Expected x=1, got %v", val)
    }
    vmLock.Unlock()
}
```

## References

- `test_helpers_internal_test.go`: Test helper functions
- `test_isolation_test.go`: Test isolation unit tests
- `mes_time_test.go`: TIME mode sequence tests
