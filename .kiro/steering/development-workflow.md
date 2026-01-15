---
inclusion: always
---

# Development Workflow

## Overview

This document describes the step-by-step workflow for implementing features in the son-et project using Kiro. This workflow is designed for incremental feature development driven by sample scenarios.

## Feature Implementation Workflow

### 0. Branch Management (CRITICAL)

**RULE: Always create a new branch before starting any implementation work**

1. **Before starting any task:**
   ```bash
   git checkout -b feature/task-description
   ```

2. **Branch naming conventions:**
   - `feature/task-X.Y-description` - New feature implementation
   - `refactor/task-X.Y-description` - Code refactoring
   - `test/task-X.Y-description` - Test additions
   - `fix/issue-description` - Bug fixes
   - `docs/description` - Documentation updates

3. **Example:**
   ```bash
   # Starting task 0.1: Create Engine state struct
   git checkout -b feature/task-0.1-engine-state
   ```

4. **After completing work:**
   - Commit changes to the feature branch
   - Request user review
   - Merge to main branch only after approval

**Never commit directly to the main branch during implementation work.**

### 1. Feature Request from User

The user will provide a sample scenario (FILLY script) and request:
> "このサンプルで使っている機能を実装してください"
> (Please implement the features used in this sample)

**Important Notes:**
- Sample scenarios are located in `samples/` directory
- These samples are excluded from Git via `.gitignore`
- Samples may use features that are not yet implemented

### 2. Analysis Phase

When receiving a feature request:

**IMPORTANT: Create a feature branch first (see step 0)**

1. **Read the sample scenario file**
   - The file will be in `samples/xxx/` directory
   - Read the `.tfy` or `.fil` script file
   - Identify all FILLY functions and syntax used

2. **Map to requirements**
   - Check `.kiro/specs/core-engine/requirements.md`
   - Identify which requirements cover the needed features
   - Check `.kiro/specs/core-engine/design.md` for implementation guidance

3. **Determine implementation scope**
   - Compiler changes (lexer/parser/codegen) in `pkg/compiler/`
   - Runtime changes (engine functions) in `pkg/engine/`
   - Both if needed

### 3. Implementation Phase

1. **Implement the feature**
   - Write code in appropriate package
   - Follow design patterns from `design.md`
   - Add necessary error handling

2. **Add tests (if time permits)**
   - Unit tests for specific cases
   - Property-based tests for universal properties
   - Reference the corresponding property in `design.md`

3. **Update documentation (if needed)**
   - Update comments in code
   - Note any deviations from original design

### 4. Build and Verification Phase

**CRITICAL: Always follow this sequence**

1. **Transpile the sample**
   ```bash
   go run cmd/son-et/main.go samples/xxx/script.tfy > samples/xxx/game.go
   ```

2. **Build the executable**
   ```bash
   cd samples/xxx
   go build -o game game.go
   ```

3. **Check for build errors**
   - If compilation fails, fix the generated code or transpiler
   - If build succeeds, proceed to next step

4. **Request user verification**
   - Inform the user: "ビルドが完了しました。`./game` を実行して動作確認をお願いします"
   - Wait for user feedback

### 5. Feedback Loop

Based on user feedback:

- **Success**: Move to next feature or task
- **Runtime error**: Debug using logs, fix implementation
- **Incorrect behavior**: Review requirements, adjust implementation
- **Build error**: Fix transpiler or code generation

## Important Constraints

### Platform Constraints

- **Operating System**: macOS
- **Shell**: zsh (default on macOS)
- **Available Commands**: Only macOS standard commands
  - ✅ Use: `grep`, `awk`, `sed`, `python3`, `go`, `git`
  - ❌ Avoid: Linux-specific commands, non-standard tools

### Sample File Constraints

- Sample files are **NOT in Git repository** (`.gitignore`)
- Must read sample files directly when analyzing
- Cannot assume sample file contents without reading

### Build Constraints

- Must build from within the sample directory (for `//go:embed`)
- Assets (BMP, MIDI, WAV) must be in same directory as generated Go file
- SoundFont file (`.sf2`) must be available for MIDI playback

## Debugging Workflow

When issues occur:

### Transpiler Issues

1. **Check generated Go code**
   ```bash
   cat samples/xxx/game.go
   ```

2. **Look for syntax errors**
   ```bash
   go build samples/xxx/game.go 2>&1 | head -20
   ```

3. **Check asset embedding**
   ```bash
   grep "//go:embed" samples/xxx/game.go
   ```

### Runtime Issues

1. **Enable debug logging**
   ```bash
   DEBUG_LEVEL=2 ./game 2>&1 | tee debug.log
   ```

2. **Check for race conditions**
   ```bash
   go build -race -o game game.go
   ./game
   ```

3. **Review timing mode**
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

## Quick Reference Commands

### Transpile and Build
```bash
# One-liner for quick testing
go run cmd/son-et/main.go samples/xxx/script.tfy > samples/xxx/game.go && \
cd samples/xxx && \
go build -o game game.go && \
echo "Build successful. Run ./game to test."
```

### Run Tests
```bash
# Run all tests
go test ./pkg/compiler/... ./pkg/engine/...

# Run with race detector
go test -race ./pkg/compiler/... ./pkg/engine/...

# Run specific test
go test -v -run TestSpecificFunction ./pkg/engine/
```

### Check Implementation Status
```bash
# Check which functions are implemented
grep -r "func.*(" pkg/engine/*.go | grep -v "^//"

# Check for TODO or FIXME
grep -r "TODO\|FIXME" pkg/
```

## Success Criteria

A feature implementation is complete when:

1. ✅ Sample scenario transpiles without errors
2. ✅ Generated Go code compiles without errors
3. ✅ Executable runs without crashes
4. ✅ User confirms correct behavior
5. ✅ (Optional) Tests pass for the feature

## Notes

- **Incremental Development**: Implement one feature at a time
- **User-Driven**: Wait for user verification before proceeding
- **Sample-Driven**: Use sample scenarios to guide implementation priorities
- **Test Later**: Focus on working implementation first, comprehensive tests later
- **Document Decisions**: Note any design decisions or deviations in code comments
