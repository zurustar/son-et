# Task 0.7 Summary: Update Existing Code to Use New Architecture

## Objective
Migrate all package-level functions to use the new EngineState architecture while maintaining backward compatibility for generated code.

## Changes Made

### 1. Global EngineState Instance
- Created `globalEngine *EngineState` variable
- Initialized in `Init()` function with proper dependencies
- Used by all package-level wrapper functions

### 2. Updated Package-Level Functions
All backward-compatible wrapper functions now delegate to `globalEngine`:
- `LoadPic()` → `globalEngine.LoadPic()`
- `CreatePic()` → `globalEngine.CreatePic()`
- `OpenWin()` → `globalEngine.OpenWin()`
- `CloseWin()` → `globalEngine.CloseWin()`
- `MoveWin()` → `globalEngine.MoveWin()`
- `DelPic()` → `globalEngine.DelPic()`
- `TextWrite()` → `globalEngine.TextWrite()`
- `MovePic()` → `globalEngine.MovePic()`
- `PutCast()` → `globalEngine.PutCast()`
- `MoveCast()` → `globalEngine.MoveCast()`
- `DelCast()` → `globalEngine.DelCast()`
- `PicWidth()` → `globalEngine.PicWidth()`
- `PicHeight()` → `globalEngine.PicHeight()`

### 3. Updated State Management Functions
Functions that modify global state now sync with `globalEngine`:
- `SetFont()` - Updates both `globalEngine` and legacy globals
- `TextColor()` - Updates both `globalEngine` and legacy globals
- `BgColor()` - Updates both `globalEngine` and legacy globals
- `BackMode()` - Updates both `globalEngine` and legacy globals
- `CapTitle()` - Uses `globalEngine.windows` map
- `CloseWinAll()` - Clears both `globalEngine` and legacy globals
- `ReversePic()` - Uses `globalEngine.pictures` map
- `RegisterUserFunc()` - Syncs with `globalEngine.userFuncs`

### 4. Updated VM/ExecuteOp
- `ExecuteOp()` now checks `globalEngine` first for state access
- Falls back to legacy globals for backward compatibility
- Updated `MovePic` case to use `globalEngine.pictures`
- Updated `MoveWin` case to use `globalEngine.windows`
- Updated user function lookup to check `globalEngine.userFuncs`

### 5. Test Infrastructure Updates
- Created `mocks_test.go` with shared mock types:
  - `MockAssetLoader`
  - `MockImageDecoder`
  - `MockDirEntry`
- Renamed `test_helpers.go` to `test_helpers_internal_test.go`
- Removed duplicate mock definitions from `engine_di_test.go`
- All test files can now share mock implementations

### 6. Backward Compatibility
- Legacy global variables maintained for compatibility
- Generated code continues to work without changes
- Dual-path approach: check `globalEngine` first, fall back to globals
- All existing sample projects compile and run successfully

## Verification

### Build Status
✅ `go build ./pkg/engine/...` - Success
✅ All compiler warnings are from Ebitengine dependencies (not our code)

### Test Status
✅ All baseline tests passing
✅ All dependency injection tests passing
✅ All test helper tests passing
✅ State isolation verified
✅ No global state leakage

### Sample Project Status
✅ Sample scenarios transpile successfully
✅ Sample scenarios compile successfully
✅ Generated code uses package-level functions (backward compatible)

## Architecture Benefits

### Before (Task 0.6)
- EngineState struct with methods ✓
- Mock dependencies for testing ✓
- Test helpers and utilities ✓
- **But**: Generated code still used global state

### After (Task 0.7)
- EngineState struct with methods ✓
- Mock dependencies for testing ✓
- Test helpers and utilities ✓
- **New**: Global EngineState instance
- **New**: All package functions use EngineState
- **New**: Backward compatibility maintained
- **New**: Sample projects work without changes

## Migration Path

### For New Code
```go
// Use EngineState directly
engine := NewEngineState(
    WithAssetLoader(loader),
    WithImageDecoder(decoder),
)
picID := engine.LoadPic("test.bmp")
```

### For Generated Code (Backward Compatible)
```go
// Package-level functions work as before
picID := LoadPic("test.bmp")
winID := OpenWin(picID, 0, 0, 640, 480, 0, 0, 0xFFFFFF)
```

### For Tests
```go
// Use test helpers with mock dependencies
engine := NewTestEngine()
picID := engine.LoadPic("test.bmp")
```

## Next Steps

Task 0.7 is complete. The architecture migration is finished:
- ✅ Task 0.1: EngineState struct created
- ✅ Task 0.2: Functions refactored to use EngineState
- ✅ Task 0.3: Dependency injection added
- ✅ Task 0.4: Rendering separated from state
- ✅ Task 0.5: Test utilities created
- ✅ Task 0.6: Baseline tests written
- ✅ Task 0.7: Existing code migrated

Ready to proceed with:
- Task 0.8: Checkpoint - Verify refactoring maintains functionality
- Task 1.1+: Property-based tests for transpiler
- Task 2+: Enhanced asset embedding
- Task 4+: Control flow statements
- And beyond...

## Files Modified
- `pkg/engine/engine.go` - Main migration changes
- `pkg/engine/engine_di_test.go` - Removed duplicate mocks
- `pkg/engine/mocks_test.go` - New shared mock types
- `pkg/engine/test_helpers_internal_test.go` - Renamed from test_helpers.go
- `.kiro/specs/core-engine/tasks.md` - Task status updated

## Commit
```
feat: Migrate existing code to use EngineState architecture

- Created global EngineState instance (globalEngine)
- Updated all package-level wrapper functions to use globalEngine
- Maintained backward compatibility with legacy global variables
- Updated SetFont, TextColor, BgColor, BackMode to sync with globalEngine
- Updated CapTitle, CloseWinAll, ReversePic to use globalEngine
- Updated ExecuteOp to use globalEngine for state access
- Updated RegisterUserFunc to sync with globalEngine
- Moved mock types to mocks_test.go for sharing across test files
- Renamed test_helpers.go to test_helpers_internal_test.go
- All tests passing, sample projects compile successfully

Task 0.7 complete - backward compatibility maintained
```
