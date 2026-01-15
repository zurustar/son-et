# Task 14.1: Analyze Variable Usage in mes() Blocks

## Summary

Implemented variable analysis for mes() blocks during code generation. The transpiler now scans all mes() blocks in a function and identifies which variables are referenced inside them. These variables are marked as "needs VM registration" and stored for use in subsequent code generation phases.

## Implementation Details

### Changes Made

1. **Added `vmVars` field to Generator struct** (`pkg/compiler/codegen/codegen.go`)
   - New field: `vmVars map[string]bool`
   - Stores variables that need VM registration (used in mes() blocks)

2. **Implemented `scanMesBlocksForVMVars()` function**
   - Scans all mes() blocks in a function body
   - Collects all variables referenced inside mes() blocks
   - Returns a map of variable names that need VM registration
   - Handles nested structures (if/for/while/switch/step blocks inside mes)

3. **Updated `Generate()` function**
   - Calls `scanMesBlocksForVMVars()` for the main function
   - Stores results in `g.vmVars` for use in later code generation phases

4. **Enhanced `collectVariablesInBlock()` function**
   - Added support for IndexExpression to handle array access
   - Improved variable collection for nested expressions

### How It Works

The implementation follows a two-phase approach:

**Phase 1: Analysis (Task 14.1 - Current)**
```
1. Parse the FILLY script into AST
2. During code generation, scan all mes() blocks
3. For each mes() block, collect all variable references
4. Mark these variables in g.vmVars map
5. Store for use in Phase 2
```

**Phase 2: Code Generation (Task 14.2 - Next)**
```
1. For variables in g.vmVars, generate: varname = engine.Assign("varname", value).(type)
2. For other variables, generate normal: varname = value
3. This ensures mes() blocks can access parent scope variables
```

### Example

**Input FILLY Script:**
```filly
main() {
    winW = WinInfo(0)
    winH = WinInfo(1)
    p39 = LoadPic("P39.BMP")
    
    mes(MIDI_TIME) {
        OpenWin(p39, winW-320, winH-240, 640, 480, 0, 0, 0)
    }
}
```

**After Task 14.1:**
- `g.vmVars` contains: `{"winw": true, "winh": true, "p39": true}`
- These variables are identified as needing VM registration

**After Task 14.2 (Next):**
- Will generate: `winw = engine.Assign("winW", engine.WinInfo(0)).(int)`
- Will generate: `winh = engine.Assign("winH", engine.WinInfo(1)).(int)`
- Will generate: `p39 = engine.Assign("p39", engine.LoadPic("P39.BMP")).(int)`

## Test Coverage

Created comprehensive tests in `pkg/compiler/codegen/codegen_vm_vars_test.go`:

1. **TestScanMesBlocksForVMVars** - Tests variable collection from mes() blocks
   - Simple variable references
   - Multiple variables
   - Variables in nested expressions (a+b, c*2)
   - Multiple mes() blocks
   - No mes() blocks (edge case)
   - Variables in step blocks inside mes
   - Variables in control flow (if/for/while) inside mes

2. **TestVMVarsInGeneratedCode** - Tests integration with code generation
   - Verifies vmVars is populated during Generate()
   - Checks that expected variables are marked
   - Validates generated code structure

All tests pass successfully.

## Requirements Validated

- ✅ **Requirement 1.1**: Transpiler parses FILLY scripts and generates Go code
- ✅ **Requirement 1.2**: Generated code compiles without errors
- ✅ **Requirement 1.4**: Case-insensitive identifier handling (variables stored lowercase)

## Next Steps

Task 14.2 will use the `g.vmVars` map to generate `engine.Assign()` calls for variables that need VM registration, enabling proper variable scope between main function and mes() blocks.

## Technical Notes

### Variable Collection Algorithm

The `scanMesBlocksForVMVars()` function uses a recursive walker pattern:

1. **Walks the AST** looking for MesBlockStatement nodes
2. **For each mes() block**, calls `collectVariablesInBlock()` to get all variable references
3. **Aggregates** variables from all mes() blocks into a single map
4. **Avoids recursion** into mes() block bodies after collection (prevents double-counting)

### Case Sensitivity

FILLY is case-insensitive, so all variable names are converted to lowercase:
- `winW`, `winw`, `WINW` all map to `"winw"`
- This ensures consistent variable lookup in the VM

### Scope Handling

The current implementation identifies variables used in mes() blocks but doesn't yet distinguish between:
- Variables defined in parent scope (need Assign)
- Variables defined locally in mes() block (don't need Assign)

This distinction will be handled in Task 14.3 when we enhance `collectVariablesInBlock()` to detect all variable references including nested expressions.

## Files Modified

- `pkg/compiler/codegen/codegen.go` - Added vmVars field and scanMesBlocksForVMVars function
- `pkg/compiler/codegen/codegen_vm_vars_test.go` - New test file with comprehensive tests

## Build Status

✅ All existing tests pass
✅ New tests pass
✅ No regressions introduced
