# Build Workflow

## Overview
This document describes the proper workflow for building and testing FILLY programs during development.

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
   go get github.com/zurustar/filly2exe/pkg/engine
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

## Troubleshooting

### Images Not Loading
- Verify assets are in the same directory as the generated Go file (for `//go:embed`).
- Check that `//go:embed` directive in generated code lists the assets.
- Ensure file extensions match (case-insensitive).

### Build Errors
- Make sure you're building from within the directory containing the generated `.go` file and assets.
- Verify all asset files are present before building.

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
