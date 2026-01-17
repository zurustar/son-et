# Build Workflow

## Overview
This document describes the comprehensive workflow for running and debugging FILLY programs during development. For high-level development workflow, see `development-workflow.md`.

## CRITICAL: Program Execution Rules for AI Agent

**RULE 1: ALWAYS use controlBashProcess for long-running programs**
- NEVER use executeBash for `go run cmd/son-et/main.go` commands
- These programs run indefinitely and will block execution
- Use `controlBashProcess` with action="start" instead

**RULE 2: Check for running processes BEFORE starting new ones**
- Use `listProcesses` to check for existing processes
- Stop any running processes with `controlBashProcess` action="stop"
- Only then start a new process

**RULE 3: Always stop processes when done**
- After verification or testing, stop the process
- Use `controlBashProcess` with action="stop" and the processId
- Never leave processes running indefinitely

**RULE 4: NEVER use separate sleep commands**
- NEVER execute sleep in a separate executeBash call
- ALWAYS combine sleep with other commands in a single one-liner
- Use shell command chaining: `command & sleep 5; kill $!`
- NEVER ask user to execute sleep separately

**Example Correct Workflow:**
```
1. listProcesses() - Check for existing processes
2. If processes exist: controlBashProcess(action="stop", processId=X)
3. controlBashProcess(action="start", command="sh -c 'go run cmd/son-et/main.go samples/kuma2 > log.txt 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null'")
4. Wait for process to complete (it will auto-terminate after 5 seconds)
5. Read log.txt to check output
```

**NEVER do this:**
```
executeBash("go run cmd/son-et/main.go samples/kuma2")  # WRONG - will hang forever
executeBash("sleep 5")  # WRONG - separate sleep command
```

**ALWAYS do this:**
```
executeBash("go run cmd/son-et/main.go samples/kuma2 > log.txt 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat log.txt")  # CORRECT - one-liner
```

## Execution Modes

The son-et interpreter supports three execution modes:

1. **Direct Mode** - Execute TFY projects directly from a directory (for development)
2. **Headless Mode** - Execute without GUI for testing and debugging (NEW)
3. **Embedded Mode** - Create standalone executables with embedded projects (for distribution)

## Headless Mode (Testing & Debugging)

### Overview

Headless mode allows you to run FILLY scripts without opening a GUI window. This is ideal for:
- Automated testing in CI/CD environments
- Debugging timing and logic issues
- Testing within Kiro or other development tools
- Verifying script behavior without visual inspection

### Quick Start

Run a FILLY project in headless mode with auto-termination:

```bash
go run cmd/son-et/main.go --headless --timeout=5s samples/my_game
```

**IMPORTANT:** Flags must come BEFORE the directory argument:
```bash
# CORRECT
go run cmd/son-et/main.go --headless --timeout=5s samples/kuma2

# WRONG (flags will be ignored)
go run cmd/son-et/main.go samples/kuma2 --headless --timeout=5s
```

### Command Line Options

**--headless**
- Run without GUI window
- All rendering operations are logged but not displayed
- Audio is initialized but muted (volume set to 0)
- MIDI timing still works correctly for MIDI_TIME mode
- Script logic (timing, audio, state) executes normally

**--timeout=DURATION**
- Auto-terminate after specified duration
- Formats: `5s` (5 seconds), `500ms` (500 milliseconds), `2m` (2 minutes)
- Program exits with status code 0 after timeout
- Prevents orphaned processes

**Environment Variable:**
```bash
HEADLESS=1 go run cmd/son-et/main.go samples/kuma2
```

### Audio in Headless Mode

In headless mode:
- Audio system is fully initialized (required for MIDI_TIME synchronization)
- All audio playback (MIDI and WAV) is muted (volume = 0)
- MIDI timing events still fire correctly for mes(MIDI_TIME) blocks
- No sound output, but timing behavior is identical to normal mode

This ensures that scripts using MIDI_TIME mode work correctly in headless testing.

### Timestamped Logging

All important logs include timestamps in `[HH:MM:SS.mmm]` format:

```
[19:43:20.557] runHeadless: Initializing headless execution
[19:43:20.558] RegisterSequence: 0 (16 ops)
[19:43:20.577] VM: Wait(2 steps) -> 24 ticks
[19:43:20.992] VM: Executing [3] Call (Tick 26) [Seq 0]
```

**Benefits:**
- Verify mes(TIME) timing accuracy
- Confirm Wait() durations are correct
- Debug timing-related issues
- Measure performance of operations

### Example Usage

**Basic headless test:**
```bash
go run cmd/son-et/main.go --headless --timeout=5s samples/kuma2
```

**Capture logs to file:**
```bash
go run cmd/son-et/main.go --headless --timeout=5s samples/kuma2 > test.log 2>&1
cat test.log
```

**Verify timing accuracy:**
```bash
go run cmd/son-et/main.go --headless --timeout=10s samples/kuma2 2>&1 | grep "^\["
```

### What Gets Logged in Headless Mode

**Rendering Operations (skipped but logged):**
- `OpenWin (headless): pic=0, pos=(0,0), size=(0x0)`
- `PutCast (headless): args=[...]`
- `MoveCast (headless): args=[...]`

**VM Execution (with timestamps):**
- `[19:43:20.577] VM: Executing [0] Call (Tick 1) [Seq 0]`
- `[19:43:20.577] VM: Wait(2 steps) -> 24 ticks`
- `[19:43:20.992] VM: Executing [3] Call (Tick 26) [Seq 0]`

**Asset Loading:**
- `LoadPic: title.bmp`
- `Loaded and decoded TITLE.BMP (200x200, ID=0)`

**Audio:**
- `PlayMIDI: [kuma.mid]`
- `PlayWAVE: title2.wav`

### Timing Verification Example

From the logs, you can verify timing accuracy:

```
[19:43:20.577] VM: Wait(2 steps) -> 24 ticks
[19:43:20.992] VM: Executing [3] Call (Tick 26) [Seq 0]
```

- Wait(2 steps) = 24 ticks = 400ms (at 60 FPS)
- Actual wait: 20.992 - 20.577 = 415ms ✅ (within tolerance)

This confirms that mes(TIME) is accurately synchronized with real time.

## Direct Mode (Development)

### Quick Start

Run a FILLY project directly without building:

```bash
go run cmd/son-et/main.go samples/my_game
```

The interpreter will:
- Locate TFY files in the directory
- Convert them to OpCode at runtime
- Load assets from the project directory
- Execute the project immediately

### Directory Structure

Your project directory should contain:

```
my_game/
├── script.tfy          # Main FILLY script
├── image1.bmp          # Images
├── image2.bmp
├── music.mid           # MIDI files
├── sound.wav           # Sound effects
└── default.sf2         # SoundFont (optional)
```

### Running with Debug Output

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/my_game 2>&1 | tee debug.log
```

### Running with Timestamped Logging

For detailed debugging with timestamps (RECOMMENDED - prevents orphaned processes):

```bash
# Run in background with timeout and log capture
go run cmd/son-et/main.go samples/my_game > debug.log 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat debug.log
```

Or with debug level:

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/my_game > debug.log 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat debug.log
```

## Embedded Mode (Distribution)

### Building Standalone Executables

Create a standalone executable with your project embedded:

1. **Create build configuration** (e.g., `build_kuma2.go`):

```go
// +build embed_kuma2

package main

import "embed"

//go:embed samples/kuma2/*
var embeddedFS embed.FS

var embeddedProject = "kuma2"
```

2. **Build with tag**:

```bash
go build -tags embed_kuma2 -o kuma2 ./cmd/son-et
```

3. **Run standalone**:

```bash
./kuma2
```

The executable contains all assets and runs without external files.

## Troubleshooting and Debugging

### Runtime Errors

**Parsing Errors:**
- Check TFY syntax
- Verify file paths in #include directives
- Review error messages with line numbers

**Asset Loading Errors:**
- Verify assets exist in project directory
- Check case-insensitive filename matching
- Ensure file extensions are correct

**Execution Errors:**
- Enable debug logging: `DEBUG_LEVEL=2`
- Check for timing mode issues (TIME vs MIDI_TIME)
- Verify MIDI SoundFont file exists

### Common Issues

**Issue**: Project not found
- **Check**: Directory path is correct
- **Check**: TFY files exist in directory
- **Solution**: Use absolute path or verify relative path

**Issue**: Assets not loading
- **Check**: Assets in same directory as TFY files
- **Check**: Filenames match (case-insensitive)
- **Solution**: Verify file extensions and paths

**Issue**: MIDI not playing
- **Check**: SoundFont file (`.sf2`) exists
- **Check**: MIDI file in project directory
- **Solution**: Copy GeneralUser-GS.sf2 to project directory

**Issue**: Execution hangs
- **Check**: Timing mode (TIME vs MIDI_TIME)
- **Check**: mes() block structure
- **Solution**: Enable debug logging to identify hang point

## User Verification Commands

When requesting user verification, always provide commands that:
- Execute from repository root directory
- Include timestamped logging
- Use only macOS default commands
- Capture both stdout and stderr

### Standard Verification Command Template

For sample in `samples/[SAMPLE_NAME]/`:

```bash
# Run with timeout and log capture (RECOMMENDED - prevents orphaned processes)
go run cmd/son-et/main.go samples/[SAMPLE_NAME] > execution.log 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat execution.log
```

Or with debug level:

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/[SAMPLE_NAME] > execution.log 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat execution.log
```

### Example Usage

For `samples/kuma2`:

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/kuma2 > execution.log 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat execution.log
```

### Log Analysis

After user provides the log, analyze:
1. **Startup phase**: Verify interpreter initialization
2. **Parsing phase**: Check for TFY syntax errors
3. **Asset loading**: Confirm images, MIDI, and audio files load correctly
4. **Runtime phase**: Look for errors, warnings, or unexpected behavior
5. **Execution flow**: Verify the program follows expected logic

## Testing During Development

### CRITICAL: Test Process Management for AI Agent

**RULE 4: ALWAYS use executeBash with timeout for unit tests**
- Unit tests should complete quickly and not open GUI windows
- Always use `-timeout=30s` flag to prevent hangs
- Example: `go test -timeout=30s ./pkg/compiler/... ./pkg/engine/...`

**RULE 5: Track process IDs when running GUI tests**
- When tests open GUI windows (Ebiten), they may not terminate automatically
- Use `controlBashProcess` to start tests that may open windows
- Store the returned processId for cleanup
- Always call `listProcesses` after tests to verify cleanup

**RULE 6: Finding orphaned test processes**
- Go test processes run from cache: `/Users/.../Library/Caches/go-build/.../main`
- DO NOT search for specific names like "son-et" or ".test"
- Use broad search first: `ps aux | grep "go-build" | grep -v grep`
- Then narrow down based on actual process names found

**Example Test Execution Workflow:**
```
1. For unit tests (no GUI):
   executeBash("go test -timeout=30s ./pkg/compiler/...")
   
2. For GUI tests:
   processId = controlBashProcess(action="start", command="go test ./pkg/engine/...")
   Wait for completion or timeout
   controlBashProcess(action="stop", processId=processId)
   
3. After any test run:
   listProcesses() - Verify no orphaned processes
   If orphaned processes exist:
     - Use: ps aux | grep "go-build" | grep -v grep
     - Kill with: kill -9 <PID>
```

### Running Sample Projects

The `samples/` directory contains example FILLY projects:

```bash
# Run a sample directly
go run cmd/son-et/main.go samples/kuma2

# Run with debug output
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/kuma2
```

### Testing New Features

When implementing new features:

1. **Write the feature** in `pkg/compiler/` or `pkg/engine/`
2. **Add tests** (unit tests and property-based tests)
3. **Test with sample scripts** to verify end-to-end functionality
4. **Update documentation** if needed

### Quick Test Cycle

```bash
# 1. Make changes to compiler or engine
# 2. Run tests with timeout
go test -timeout=30s ./pkg/compiler/... ./pkg/engine/...

# 3. Check for orphaned processes
ps aux | grep "go-build" | grep -v grep

# 4. Test with a sample script
go run cmd/son-et/main.go samples/kuma2

# 5. If successful, commit changes
```

## Asset Requirements

### Required Assets

**For MIDI Playback:**
- A SoundFont file (`.sf2`) must be available
- Default location: `./default.sf2` or `./GeneralUser-GS.sf2`
- Place in project directory or repository root

**For Image Display:**
- BMP files referenced in `LoadPic()` calls
- Must be in the project directory
- Case-insensitive matching is supported

**For Audio Playback:**
- WAV files referenced in `PlayWAVE()` calls
- Must be in the project directory

### Asset Organization

Recommended directory structure:

```
my_game/
├── script.tfy          # Main FILLY script
├── image1.bmp          # Images
├── image2.bmp
├── music.mid           # MIDI files
├── sound.wav           # Sound effects
└── default.sf2         # SoundFont (optional)
```

## Command Reference

### Direct Mode

```bash
# Basic execution
go run cmd/son-et/main.go <directory>

# With debug output
DEBUG_LEVEL=2 go run cmd/son-et/main.go <directory>

# With timestamped logging
DEBUG_LEVEL=2 go run cmd/son-et/main.go <directory> 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee debug.log
```

### Embedded Mode

```bash
# Build standalone executable
go build -tags embed_<project> -o <output> ./cmd/son-et

# Run standalone
./<output>
```

### Help

```bash
# Display usage information
go run cmd/son-et/main.go --help
go run cmd/son-et/main.go
```
