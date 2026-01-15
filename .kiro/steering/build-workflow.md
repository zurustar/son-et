# Build Workflow

## Overview
This document describes the comprehensive workflow for building, testing, and debugging FILLY programs during development. For high-level development workflow, see `development-workflow.md`.

## Build Process

### General Workflow

To build a FILLY project, it is recommended to work in a directory containing your script and assets.

**Workflow:**

1. **Prepare your workspace**
   Ensure your `.tfy` script and all asset files (images, MIDI) are in the same directory.

   ```bash
   mkdir build_work
   cp samples/my_game/* build_work/
   ```

2. **Generate Go code**
   Run the transpiler (`son-et`) against your script.

   ```bash
   go run cmd/son-et/main.go build_work/game.tfy > build_work/main.go
   ```

3. **Build the executable**
   Move into the directory and build.

   ```bash
   cd build_work
   go mod init mygame
   go get github.com/zurustar/son-et/pkg/engine
   go build -o game main.go
   ```

4. **Run the executable**
   ```bash
   ./game
   ```

## Example: Building "My Game"

```bash
# 1. Prepare directory
mkdir -p my_build
cp my_assets/* my_build/

# 2. Generate
go run cmd/son-et/main.go my_build/script.tfy > my_build/game.go

# 3. Build
cd my_build
# (Initialize module if needed, or rely on parent go.mod if inside repo)
go build -o ../output/my_game game.go
cd ..

# 4. Run
./output/my_game
```

## Troubleshooting and Debugging

### Build Errors
- Make sure you're building from within the directory containing the generated `.go` file and assets.
- Verify all asset files are present before building.

### Images Not Loading
- Verify assets are in the same directory as the generated Go file (for `//go:embed`).
- Check that `//go:embed` directive in generated code lists the assets.
- Ensure file extensions match (case-insensitive).

### Debugging Build Issues

**Check Generated Go Code:**
```bash
cat samples/xxx/game.go
```

**Look for Syntax Errors:**
```bash
go build samples/xxx/game.go 2>&1 | head -20
```

**Check Asset Embedding:**
```bash
grep "//go:embed" samples/xxx/game.go
```

**Enable Debug Output:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee debug.log
```

**Check for Race Conditions:**
```bash
go build -race -o game main.go
./game
```

**Verify Asset Embedding:**
```bash
# Check generated code for go:embed directives
grep "//go:embed" build_work/main.go
```

### Runtime Issues

**Enable Debug Logging:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | tee debug.log
```
./game
```

**Review Timing Mode:**
- Check if `mes(MIDI_TIME)` or `mes(TIME)` is used
- Verify correct blocking/non-blocking behavior

### Common Issues

**Issue**: Images not loading
- **Check**: Assets in same directory as `game.go`
- **Check**: `//go:embed` directives in generated code
- **Check**: Case-insensitive filename matching

**Issue**: MIDI not playing
- **Check**: SoundFont file (`.sf2`) exists
- **Check**: MIDI file embedded correctly
- **Check**: `PlayMIDI()` called after `mes(MIDI_TIME)` block

**Issue**: Deadlock or freeze
- **Check**: Mutex double-locking (see `dev_guidelines.md`)
- **Check**: MIDI_TIME mode is non-blocking
- **Check**: TIME mode is blocking correctly

### Quick Build Commands

**One-liner for Quick Testing:**
```bash
go run cmd/son-et/main.go samples/xxx/script.tfy > samples/xxx/game.go && \
cd samples/xxx && \
go build -o game game.go && \
echo "Build successful. Run ./game to test."
```

## User Verification Commands

When requesting user verification, always provide commands that:
- Execute from repository root directory
- Include timestamped logging
- Use only macOS default commands
- Capture both stdout and stderr

### Standard Verification Command Template

For sample in `samples/[SAMPLE_NAME]/`:

```bash
# Build and run with timestamped logging
go run cmd/son-et/main.go samples/[SAMPLE_NAME]/[SCRIPT_NAME].tfy > samples/[SAMPLE_NAME]/game.go && \
cd samples/[SAMPLE_NAME] && \
go build -o game game.go && \
echo "$(date '+%Y-%m-%d %H:%M:%S') Build completed successfully" && \
DEBUG_LEVEL=2 ./game 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee ../../game_execution.log && \
cd ../.. && \
echo "$(date '+%Y-%m-%d %H:%M:%S') Execution completed. Log saved to game_execution.log"
```

### Example Usage

For `samples/kuma2/KUMA2.TFY`:

```bash
go run cmd/son-et/main.go samples/kuma2/KUMA2.TFY > samples/kuma2/game.go && \
cd samples/kuma2 && \
go build -o game game.go && \
echo "$(date '+%Y-%m-%d %H:%M:%S') Build completed successfully" && \
DEBUG_LEVEL=2 ./game 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee ../../game_execution.log && \
cd ../.. && \
echo "$(date '+%Y-%m-%d %H:%M:%S') Execution completed. Log saved to game_execution.log"
```

### Log Analysis

After user provides the log, analyze:
1. **Build phase**: Check for compilation errors
2. **Startup phase**: Verify engine initialization
3. **Runtime phase**: Look for errors, warnings, or unexpected behavior
4. **Asset loading**: Confirm images, MIDI, and audio files load correctly
5. **Execution flow**: Verify the program follows expected logic

## Testing During Development

### Running Sample Projects

The `samples/` directory contains example FILLY projects that can be used for testing:

```bash
# Test with a sample
go run cmd/son-et/main.go samples/xxx/SCRIPT.TFY > samples/xxx/game.go
cd samples/xxx
go build -o game game.go
./game
```

### Debugging Build Issues

**Enable Debug Output:**
```bash
DEBUG_LEVEL=2 ./game 2>&1 | while IFS= read -r line; do 
  echo "$(date '+%H:%M:%S.%3N') $line"
done | tee debug.log
```

**Check for Race Conditions:**
```bash
go build -race -o game main.go
./game
```

**Verify Asset Embedding:**
```bash
# Check generated code for go:embed directives
grep "//go:embed" build_work/main.go
```

## Integration with Kiro Development

When implementing new features:

1. **Write the feature** in the appropriate package (`pkg/compiler/` or `pkg/engine/`)
2. **Add tests** for the feature (unit tests and property-based tests)
3. **Test with sample scripts** to verify end-to-end functionality
4. **Update documentation** if the feature affects user-facing behavior

### Quick Test Cycle

```bash
# 1. Make changes to compiler or engine
# 2. Run tests
go test ./pkg/compiler/... ./pkg/engine/...

# 3. Test with a sample script
go run cmd/son-et/main.go samples/xxx/SCRIPT.TFY > /tmp/test_game.go
cd /tmp
go run test_game.go

# 4. If successful, commit changes
```

## Asset Requirements

### Required Assets for Building

**For MIDI Playback:**
- A SoundFont file (`.sf2`) must be available
- Default location: `./default.sf2` or `./GeneralUser-GS.sf2`
- The engine will look for soundfonts in the current directory

**For Image Display:**
- BMP files referenced in `LoadPic()` calls
- Must be in the same directory as the generated Go file
- Case-insensitive matching is supported

**For Audio Playback:**
- WAV files referenced in `PlayWAVE()` calls
- Must be in the same directory as the generated Go file

### Asset Organization

Recommended directory structure for a FILLY project:

```
my_game/
├── script.tfy          # Main FILLY script
├── image1.bmp          # Images
├── image2.bmp
├── music.mid           # MIDI files
├── sound.wav           # Sound effects
└── default.sf2         # SoundFont (optional, can use global)
```

After transpilation:

```
my_game/
├── script.tfy
├── game.go             # Generated Go code
├── image1.bmp
├── image2.bmp
├── music.mid
├── sound.wav
├── default.sf2
└── game                # Built executable (contains all assets)
```
