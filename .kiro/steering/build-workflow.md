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

**Example Correct Workflow:**
```
1. listProcesses() - Check for existing processes
2. If processes exist: controlBashProcess(action="stop", processId=X)
3. controlBashProcess(action="start", command="go run cmd/son-et/main.go samples/kuma2")
4. Wait 2-3 seconds
5. getProcessOutput(processId=Y) - Check output
6. controlBashProcess(action="stop", processId=Y) - Stop when done
```

**NEVER do this:**
```
executeBash("go run cmd/son-et/main.go samples/kuma2")  # WRONG - will hang forever
```

## Execution Modes

The son-et interpreter supports two execution modes:

1. **Direct Mode** - Execute TFY projects directly from a directory (for development)
2. **Embedded Mode** - Create standalone executables with embedded projects (for distribution)

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

For detailed debugging with timestamps:

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/my_game 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee debug.log
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
# Run with timestamped logging
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/[SAMPLE_NAME] 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee execution.log
```

### Example Usage

For `samples/kuma2`:

```bash
DEBUG_LEVEL=2 go run cmd/son-et/main.go samples/kuma2 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee execution.log
```

### Log Analysis

After user provides the log, analyze:
1. **Startup phase**: Verify interpreter initialization
2. **Parsing phase**: Check for TFY syntax errors
3. **Asset loading**: Confirm images, MIDI, and audio files load correctly
4. **Runtime phase**: Look for errors, warnings, or unexpected behavior
5. **Execution flow**: Verify the program follows expected logic

## Testing During Development

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
# 2. Run tests
go test -timeout=30s ./pkg/compiler/... ./pkg/engine/...

# 3. Test with a sample script
go run cmd/son-et/main.go samples/kuma2

# 4. If successful, commit changes
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
