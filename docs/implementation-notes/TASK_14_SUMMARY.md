# Task 14: Variable Scope & VM Architecture (Phase 2)

## Summary

Successfully implemented Phase 2 of the Variable Scope & VM Architecture, enabling proper variable scope between main function and mes() blocks. The transpiler now generates `engine.Assign()` calls for variables used in mes() blocks, allowing them to be accessible via the VM's parent scope lookup mechanism.

## Implementation Status

✅ **Phase 1 Complete** (from previous work): VM supports parent scope lookup
✅ **Phase 2 Complete** (this task): Transpiler generates Assign() calls for VM variables

## Changes Made

### 1. Task 14.1: Analyze variable usage in mes() blocks

**File**: `pkg/compiler/codegen/codegen.go`

- Added `vmVars map[string]bool` field to Generator struct
- Implemented `scanMesBlocksForVMVars()` function to scan all mes() blocks and collect variable references
- Updated `Generate()` to call scanner and store results in `g.vmVars`

**Test**: `pkg/compiler/codegen/codegen_vm_vars_test.go`
- Comprehensive tests for variable collection from mes() blocks
- Tests for nested expressions, control flow, multiple mes() blocks

### 2. Task 14.2: Generate engine.Assign() calls

**File**: `pkg/compiler/codegen/codegen.go`

- Modified `genStatement()` for AssignStatement to check `g.vmVars`
- Variables in `vmVars` generate: `varname = engine.Assign("varname", value).(type)`
- Other variables generate normal: `varname = value`
- Implemented `inferType()` function for type assertions (int, string, []int)

**Test**: `pkg/compiler/codegen/codegen_assign_test.go`
- Tests for Assign() call generation
- Tests for type inference
- Tests for mixed VM and non-VM variables
- Tests for case-insensitive variable names

### 3. Task 14.3: Enhanced variable detection

**File**: `pkg/compiler/codegen/codegen.go`

- Enhanced `collectVariablesInBlock()` to detect:
  - Variables in infix expressions (e.g., `winW-320`)
  - Variables in nested function arguments
  - Variables in array subscripts
  - Variables in control flow (if/for/while/switch)
  - Exclusion of function names and constants

**Test**: `pkg/compiler/codegen/codegen_enhanced_vars_test.go`
- Tests for all enhanced detection scenarios
- Tests for excluding function names and constants

### 4. Task 14.4: Sample scenario testing

- Created test scenario with variables used in mes() blocks
- Verified transpilation generates correct Assign() calls
- Verified generated code compiles successfully
- Confirmed proper distinction between VM and non-VM variables

### 5. Task 14.5: Unit tests for variable scope

**File**: `pkg/engine/variable_scope_basic_test.go`

- Tests for `SetVMVar()` basic functionality
- Tests for `Assign()` helper function
- Tests for `ResolveArg()` variable resolution
- Tests for case-insensitive variable names

**File**: `pkg/engine/engine.go`

- Added timing mode constants: `Time = 0`, `MidiTime = 1`
- Removed conflicting global `MidiTime` variable

### 6. Task 14.6: Documentation

**This document** - Complete implementation summary

## How It Works

### Before (Broken)

```go
// Generated code (BROKEN)
func main() {
    engine.Init(assets, func() {
        var winw int
        var winh int
        
        winw = engine.WinInfo(0)  // Go local variable
        winh = engine.WinInfo(1)  // Go local variable
        
        engine.RegisterSequence(engine.MidiTime, []engine.OpCode{
            {Cmd: "OpenWin", Args: []any{
                engine.Variable("winw"),  // NOT FOUND in VM!
                // ...
            }},
        }, map[string]any{})  // Empty vars map
    })
}
```

**Problem**: `winw` and `winh` are Go local variables, not accessible in VM.

### After (Fixed)

```go
// Generated code (FIXED)
func main() {
    engine.Init(assets, func() {
        var winw int
        var winh int
        
        // Use Assign() to register in VM AND assign to Go local
        winw = engine.Assign("winW", engine.WinInfo(0)).(int)
        winh = engine.Assign("winH", engine.WinInfo(1)).(int)
        
        engine.RegisterSequence(engine.MidiTime, []engine.OpCode{
            {Cmd: "OpenWin", Args: []any{
                engine.Variable("winw"),  // FOUND in parent scope!
                // ...
            }},
        }, map[string]any{"winw": winw, "winh": winh})
    })
}
```

**Solution**: Variables are registered in VM via `Assign()` and accessible via parent scope lookup.

## Example

**Input FILLY Script:**
```filly
main() {
    winW = WinInfo(0)
    winH = WinInfo(1)
    pic = LoadPic("test.bmp")
    localVar = 42
    
    mes(MIDI_TIME) {
        OpenWin(pic, winW-320, winH-240, 640, 480, 0, 0, 0)
    }
    
    OpenWin(localVar, 0, 0, 100, 100, 0, 0, 0)
}
```

**Generated Go Code:**
```go
// Variables used in mes() blocks - use Assign()
winw = engine.Assign("winW", engine.WinInfo(0)).(int)
winh = engine.Assign("winH", engine.WinInfo(1)).(int)
pic = engine.Assign("pic", engine.LoadPic("test.bmp")).(int)

// Variable NOT used in mes() blocks - normal assignment
localvar = 42

// mes() block can access winw, winh, pic via parent scope
engine.RegisterSequence(engine.MidiTime, []engine.OpCode{
    {Cmd: "OpenWin", Args: []any{
        engine.Variable("pic"),
        engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winW"), 320}},
        engine.OpCode{Cmd: "Infix", Args: []any{"-", engine.Variable("winH"), 240}},
        640, 480, 0, 0, 0,
    }},
}, map[string]any{"pic": pic, "winh": winh, "winw": winw})
```

## Test Results

All tests pass:

✅ **Transpiler Tests** (`pkg/compiler/codegen/`)
- `TestScanMesBlocksForVMVars` - Variable collection
- `TestVMVarsInGeneratedCode` - Integration with code generation
- `TestAssignCallGeneration` - Assign() call generation
- `TestTypeInference` - Type inference for assertions
- `TestCaseInsensitiveAssign` - Case-insensitive handling
- `TestEnhancedVariableDetection` - Enhanced variable detection
- All existing codegen tests (no regressions)

✅ **Engine Tests** (`pkg/engine/`)
- `TestSetVMVarBasic` - Basic SetVMVar functionality
- `TestAssignBasic` - Basic Assign functionality
- `TestResolveArgBasic` - Variable resolution
- `TestCaseInsensitiveBasic` - Case-insensitive names

✅ **Integration Tests**
- Sample scenario transpiles correctly
- Generated code compiles without errors
- Proper distinction between VM and non-VM variables

## Requirements Validated

- ✅ **Requirement 1.1**: Transpiler parses FILLY scripts and generates Go code
- ✅ **Requirement 1.2**: Generated code compiles without errors
- ✅ **Requirement 1.4**: Case-insensitive identifier handling

## Known Limitations

1. **Array variables**: Currently, array identifiers in expressions are converted to lowercase, but array assignments might need special handling
2. **Type inference**: Limited to common cases (int, string, []int). Complex expressions may default to int
3. **Nested mes() blocks**: Not tested (rare in practice)

## Next Steps (Future Enhancements)

**Phase 3** (if needed): Full OpCode generation for all code
- Currently: mes() blocks use OpCodes, outside code uses direct Go
- Future: Everything as OpCodes for unified execution model
- Benefits: Proper scoping, debuggability, consistency
- Challenges: Performance overhead, type safety

## Files Modified

### Transpiler
- `pkg/compiler/codegen/codegen.go` - Core implementation
- `pkg/compiler/codegen/codegen_vm_vars_test.go` - Variable collection tests
- `pkg/compiler/codegen/codegen_assign_test.go` - Assign() generation tests
- `pkg/compiler/codegen/codegen_enhanced_vars_test.go` - Enhanced detection tests

### Engine
- `pkg/engine/engine.go` - Added Time/MidiTime constants
- `pkg/engine/variable_scope_basic_test.go` - Variable scope tests

### Documentation
- `docs/implementation-notes/TASK_14.1_SUMMARY.md` - Task 14.1 details
- `docs/implementation-notes/TASK_14_SUMMARY.md` - This document

## Build Status

✅ All tests pass
✅ No regressions
✅ Sample scenarios compile and run correctly
✅ Ready for merge to main branch
