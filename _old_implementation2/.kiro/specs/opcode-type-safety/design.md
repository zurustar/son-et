# Design Document: OpCode Type Safety

## Overview

This design replaces the current `OpCode` struct with a strongly-typed system where each operation type has its own struct. The current implementation uses `Args []any`, requiring runtime type assertions throughout the VM. The new design provides compile-time type safety, better documentation, and eliminates an entire class of runtime errors.

The key insight is that while parameters can be literals, variables, or expressions (nested OpCodes), we can model this with a `Value` type that encapsulates these variants. Each operation type then has a dedicated struct with appropriately typed fields.

## Architecture

### Current Architecture

```
OpCode {
    Cmd: OpCmd (enum)
    Args: []any  // Generic slice requiring type assertions
}
```

Problems:
- No compile-time validation of argument count or types
- Type assertions scattered throughout VM code
- Easy to make mistakes with argument order
- Difficult to understand what each operation expects

### New Architecture

```
OpCode interface {
    GetCmd() OpCmd
}

// Each operation type implements OpCode
AssignOp struct { Variable, Value }
BinaryOp struct { Operator, Left, Right }
IfOp struct { Condition, ThenBlock, ElseBlock }
// ... etc
```

Benefits:
- Compile-time type checking
- Self-documenting code
- Direct field access (no type assertions)
- Impossible to construct invalid operations

## Components and Interfaces

### 1. Value Type

The `Value` type represents a parameter that can be a literal, variable reference, or expression:

```go
type Value struct {
    kind ValueKind
    data any
}

type ValueKind int

const (
    ValueLiteral ValueKind = iota  // int64, float64, string, bool
    ValueVariable                   // Variable reference
    ValueExpression                 // Nested OpCode
    ValueArray                      // []Value for array literals
)

// Constructors
func LiteralValue(v any) Value
func VariableValue(name Variable) Value
func ExpressionValue(op OpCode) Value
func ArrayValue(elements []Value) Value

// Accessors
func (v Value) AsLiteral() (any, bool)
func (v Value) AsVariable() (Variable, bool)
func (v Value) AsExpression() (OpCode, bool)
func (v Value) AsArray() ([]Value, bool)
```

### 2. OpCode Interface

```go
type OpCode interface {
    GetCmd() OpCmd
}
```

All operation structs implement this interface, allowing them to be stored in slices and processed generically.

### 3. Operation Structs

#### Assignment Operations

```go
type AssignOp struct {
    Variable Variable
    Value    Value
}

func (op AssignOp) GetCmd() OpCmd { return OpAssign }
```

#### Binary Operations

```go
type BinaryOp struct {
    Operator string  // "+", "-", "*", "/", "%", "==", "!=", "<", ">", "<=", ">=", "&&", "||"
    Left     Value
    Right    Value
}

func (op BinaryOp) GetCmd() OpCmd { return OpBinaryOp }
```

#### Unary Operations

```go
type UnaryOp struct {
    Operator string  // "-", "!"
    Operand  Value
}

func (op UnaryOp) GetCmd() OpCmd { return OpUnaryOp }
```

#### Control Flow Operations

```go
type IfOp struct {
    Condition Value
    ThenBlock []OpCode
    ElseBlock []OpCode  // Can be nil/empty
}

func (op IfOp) GetCmd() OpCmd { return OpIf }

type ForOp struct {
    Init      OpCode    // Can be nil
    Condition Value     // Can be nil (infinite loop)
    Post      OpCode    // Can be nil
    Body      []OpCode
}

func (op ForOp) GetCmd() OpCmd { return OpFor }

type WhileOp struct {
    Condition Value
    Body      []OpCode
}

func (op WhileOp) GetCmd() OpCmd { return OpWhile }

type SwitchOp struct {
    Value   Value
    Cases   []SwitchCase
    Default []OpCode  // Can be nil/empty
}

type SwitchCase struct {
    Value Value
    Body  []OpCode
}

func (op SwitchOp) GetCmd() OpCmd { return OpSwitch }
```

#### Function Operations

```go
type CallOp struct {
    Function Value    // Function name (can be variable or literal string)
    Args     []Value
}

func (op CallOp) GetCmd() OpCmd { return OpCall }
```

#### Wait and Timing Operations

```go
type WaitOp struct {
    Count Value  // Number of steps to wait
}

func (op WaitOp) GetCmd() OpCmd { return OpWait }

type SetStepOp struct {
    Duration Value  // Step duration
}

func (op SetStepOp) GetCmd() OpCmd { return OpSetStep }
```

#### Event Handler Operations

```go
type RegisterEventHandlerOp struct {
    EventType string
    Body      []OpCode
}

func (op RegisterEventHandlerOp) GetCmd() OpCmd { return OpRegisterEventHandler }
```

#### Array Operations

```go
type ArrayAccessOp struct {
    Array Value  // Array reference (variable or expression)
    Index Value  // Index (can be literal, variable, or expression)
}

func (op ArrayAccessOp) GetCmd() OpCmd { return OpArrayAccess }

type ArrayAssignOp struct {
    Array Value  // Array reference (variable or expression)
    Index Value  // Index
    Value Value  // Value to assign
}

func (op ArrayAssignOp) GetCmd() OpCmd { return OpArrayAssign }
```

#### Simple Operations (No Arguments)

```go
type BreakOp struct{}
func (op BreakOp) GetCmd() OpCmd { return OpBreak }

type ContinueOp struct{}
func (op ContinueOp) GetCmd() OpCmd { return OpContinue }
```

## Data Models

### OpCode Slice Type

Since we need to store different operation types in slices, we use the `OpCode` interface:

```go
type Program []OpCode
```

### Variable Type

The existing `Variable` type remains unchanged:

```go
type Variable string

func (v Variable) String() string {
    return string(v)
}
```

## Migration Strategy

### Phase 1: Add New Types (Non-Breaking)

1. Define `Value` type and constructors
2. Define `OpCode` interface
3. Define all operation structs
4. Keep existing `OpCode` struct temporarily

### Phase 2: Update Code Generator

1. Modify `codegen.go` to create typed operations
2. Update `generateExpression` to return `Value` instead of `any`
3. Update all `generate*` methods to use new types

### Phase 3: Update VM

1. Modify `vm.go` to use type switches instead of the old `Cmd` field
2. Update `ExecuteOp` to switch on concrete types
3. Update `evaluateValue` to work with `Value` type
4. Remove all type assertions from execution methods

### Phase 4: Update Tests and Remove Old Code

1. Update all tests to use new types
2. Remove old `OpCode` struct
3. Update serializer to handle new types

## Code Generation Changes

### Before (Current)

```go
func (g *Generator) generateAssignStatement(stmt *ast.AssignStatement) []interpreter.OpCode {
    value := g.generateExpression(stmt.Value)
    return []interpreter.OpCode{{
        Cmd:  interpreter.OpAssign,
        Args: []any{interpreter.Variable(ident.Value), value},
    }}
}
```

### After (New)

```go
func (g *Generator) generateAssignStatement(stmt *ast.AssignStatement) []interpreter.OpCode {
    value := g.generateExpression(stmt.Value)
    return []interpreter.OpCode{
        interpreter.AssignOp{
            Variable: interpreter.Variable(ident.Value),
            Value:    value,
        },
    }
}
```

## VM Execution Changes

### Before (Current)

```go
func (vm *VM) executeAssign(seq *Sequencer, op interpreter.OpCode) error {
    if len(op.Args) != 2 {
        return NewRuntimeError(...)
    }
    
    varName, ok := op.Args[0].(interpreter.Variable)
    if !ok {
        return NewRuntimeError(...)
    }
    
    value, err := vm.evaluateValue(seq, op.Args[1])
    // ...
}
```

### After (New)

```go
func (vm *VM) executeAssign(seq *Sequencer, op interpreter.AssignOp) error {
    // No type assertions needed!
    value, err := vm.evaluateValue(seq, op.Value)
    if err != nil {
        return err
    }
    
    seq.SetVariable(string(op.Variable), value)
    return nil
}
```

### Main Dispatch

```go
func (vm *VM) ExecuteOp(seq *Sequencer, op interpreter.OpCode) error {
    switch op := op.(type) {
    case interpreter.AssignOp:
        return vm.executeAssign(seq, op)
    case interpreter.BinaryOp:
        return vm.executeBinaryOp(seq, op)
    case interpreter.IfOp:
        return vm.executeIf(seq, op)
    case interpreter.CallOp:
        return vm.executeCall(seq, op)
    // ... etc
    default:
        return NewRuntimeError("Unknown", "", "Unknown OpCode type: %T", op)
    }
}
```

## Error Handling

### Compile-Time Errors (Eliminated)

The following errors become impossible:
- Wrong number of arguments
- Wrong argument types
- Wrong argument order

### Runtime Errors (Preserved)

The following errors still occur at runtime (as expected):
- Division by zero
- Array index out of bounds
- Undefined variable access
- Type mismatches in operations (e.g., adding string to int)

### Error Messages

Error messages improve because we can reference field names:

Before: `"OpAssign first argument must be Variable, got %T"`
After: `"Cannot assign to non-variable: %v"` (but this becomes a compile error)

## Testing Strategy

### Unit Tests

1. **Value Type Tests**
   - Test construction of each Value variant
   - Test accessor methods
   - Test type checking

2. **Operation Construction Tests**
   - Test creating each operation type
   - Verify fields are set correctly
   - Test GetCmd() returns correct OpCmd

3. **Code Generation Tests**
   - Test that AST nodes generate correct typed operations
   - Verify nested expressions create correct Value types
   - Test all statement types

4. **VM Execution Tests**
   - Test execution of each operation type
   - Verify behavior matches old implementation
   - Test error conditions

### Integration Tests

1. **Existing Test Suite**
   - Run all existing compiler tests
   - Run all existing VM tests
   - Verify no behavioral changes

2. **Serialization Tests**
   - Test serializing typed operations to Go code
   - Verify deserialization works correctly



## Correctness Properties

A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.

### Property 1: Code Generator Type Correctness

*For any* valid AST node of a specific type (assignment, binary operation, control flow, function call, wait/timing, event handler, or array operation), the code generator should produce an OpCode of the corresponding typed struct (AssignOp, BinaryOp, IfOp/ForOp/WhileOp, CallOp, WaitOp/SetStepOp, RegisterEventHandlerOp, or ArrayAccessOp/ArrayAssignOp respectively).

**Validates: Requirements 1.2, 3.2, 4.2, 5.4, 6.2, 7.3, 8.2, 9.3**

### Property 2: Value Round-Trip Preservation

*For any* value (literal, variable, or expression), creating a Value from it and then extracting it using the appropriate accessor should return an equivalent value.

**Validates: Requirements 2.2, 2.3, 2.4**

### Property 3: Value Type Safety

*For any* Value, exactly one of its type accessors (AsLiteral, AsVariable, AsExpression, AsArray) should return true, and the others should return false.

**Validates: Requirements 2.5**

### Property 4: OpCmd Identification

*For any* typed operation struct, calling GetCmd() should return the OpCmd value that corresponds to that operation type (e.g., AssignOp.GetCmd() returns OpAssign, BinaryOp.GetCmd() returns OpBinaryOp).

**Validates: Requirements 1.4, 10.2**

### Property 5: Behavioral Equivalence

*For any* valid program that executes successfully with the old OpCode system, executing the same program with the new typed OpCode system should produce identical results (same variable values, same function calls, same control flow).

**Validates: Requirements 10.1, 10.3**

### Property 6: Error Preservation

*For any* program that produces a runtime error with the old OpCode system, the new typed OpCode system should produce an equivalent error (same error type and similar error message).

**Validates: Requirements 10.4**

