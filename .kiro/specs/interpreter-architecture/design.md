# Design Document: Interpreter Architecture

## Overview

The son-et interpreter system executes FILLY language scripts through a virtual machine (VM) that interprets OpCode sequences. The system supports two execution modes:

1. **Direct Mode**: Interprets TFY scripts from a directory at runtime for rapid development iteration
2. **Embedded Mode**: Executes pre-compiled OpCode embedded in the son-et executable for distribution

This specification focuses on the CLI interface, execution modes, and build process. Common design elements (OpCode structure, VM architecture, concurrency model, etc.) are defined in [COMMON_DESIGN.md](../COMMON_DESIGN.md). Runtime engine functionality is defined in [core-engine/design.md](../core-engine/design.md).

## Common Design Elements

This specification depends on the following common design elements defined in [COMMON_DESIGN.md](../COMMON_DESIGN.md):
- OpCode Structure
- Variable Scope Model
- Concurrency Model
- Rendering Pipeline
- Timing & Synchronization
- Virtual Display Architecture
- Data Models
- Error Handling
- Critical Design Constraints

## Architecture

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        son-et CLI                            │
│  ┌──────────────┐              ┌──────────────┐            │
│  │ Direct Mode  │              │ Embedded Mode│            │
│  │              │              │              │            │
│  │ Read TFY     │              │ Load Embedded│            │
│  │ from disk    │              │ OpCode       │            │
│  └──────┬───────┘              └──────┬───────┘            │
│         │                             │                     │
│         └─────────────┬───────────────┘                     │
│                       │                                     │
│              ┌────────▼────────┐                           │
│              │   Interpreter   │                           │
│              │                 │                           │
│              │ TFY → OpCode    │                           │
│              └────────┬────────┘                           │
│                       │                                     │
│              ┌────────▼────────┐                           │
│              │   Virtual Machine│                          │
│              │                 │                           │
│              │ • Sequencer     │                           │
│              │ • ExecuteOp     │                           │
│              │ • ResolveArg    │                           │
│              │ • Scope Chain   │                           │
│              └────────┬────────┘                           │
│                       │                                     │
│              ┌────────▼────────┐                           │
│              │  Engine (pkg/engine)                        │
│              │                 │                           │
│              │ • Graphics      │                           │
│              │ • Audio/MIDI    │                           │
│              │ • Window Mgmt   │                           │
│              └─────────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

### Component Responsibilities

**CLI (cmd/son-et/main.go)**:
- Parse command-line arguments
- Determine execution mode (direct vs embedded)
- Initialize the interpreter with appropriate input source
- Handle errors and display usage information

**Interpreter (pkg/compiler/interpreter)**:
- Parse TFY scripts using existing lexer and parser
- Convert AST to OpCode sequences
- Handle #include directives recursively
- Scan for asset references (LoadPic, PlayMIDI, PlayWAVE)
- Manage variable declarations and scope information

**Virtual Machine (pkg/engine)**:
- Execute OpCode sequences through Sequencer
- Manage variable scope chain
- Handle timing and synchronization (TIME vs MIDI_TIME modes)
- Provide ExecuteOp for command execution
- Provide ResolveArg for variable and expression resolution

**Engine (pkg/engine)**:
- Provide FILLY language functions (graphics, audio, etc.)
- Manage game state and resources
- Handle Ebiten integration for rendering and input

## Components and Interfaces

### Interpreter Component

**Location**: `pkg/compiler/interpreter/interpreter.go`

**Responsibilities**:
- Convert TFY scripts to OpCode sequences
- Manage global and function-level variable declarations
- Track asset references for embedding

**Key Types**:

```go
type Interpreter struct {
    assets        []string              // Discovered asset files
    globals       map[string]bool       // Global variable names
    userFuncs     map[string]*Function  // User-defined functions
}

type Function struct {
    Name       string
    Parameters []Parameter
    Body       []OpCode
    Locals     map[string]string // Variable name -> type
}

type Parameter struct {
    Name    string
    Type    string
    Default any
}
```

**Key Methods**:

```go
// Interpret converts a TFY script to OpCode sequences
func (i *Interpreter) Interpret(program *ast.Program) (*Script, error)

// interpretFunction converts a function AST to OpCode
func (i *Interpreter) interpretFunction(fn *ast.FunctionStatement) (*Function, error)

// interpretStatement converts a statement AST to OpCode
func (i *Interpreter) interpretStatement(stmt ast.Statement) ([]OpCode, error)

// interpretExpression converts an expression AST to OpCode
func (i *Interpreter) interpretExpression(expr ast.Expression) (OpCode, error)

// scanAssets discovers asset references in the AST
func (i *Interpreter) scanAssets(program *ast.Program) []string
```

### Script Structure

**Location**: `pkg/compiler/interpreter/script.go`

**Responsibilities**:
- Represent a compiled TFY script
- Store OpCode sequences for all functions
- Track global variables and assets

**Key Types**:

```go
type Script struct {
    Globals   map[string]string    // Variable name -> type
    Functions map[string]*Function // Function name -> Function
    Main      *Function            // Main function OpCode
    Assets    []string             // Asset file names
}
```

### VM Execution

**Location**: `pkg/engine/engine.go` (existing)

**Enhancements Needed**:
- Support function calls through OpCode
- Manage scope chain for variable resolution
- Handle user-defined function registration

**Key Types** (existing):

```go
type OpCode struct {
    Cmd  string
    Args []any
}

type Sequencer struct {
    commands     []OpCode
    pc           int
    waitTicks    int
    active       bool
    ticksPerStep int
    vars         map[string]any
    parent       *Sequencer
    mode         int
    onComplete   func()
}
```

**New OpCode Commands**:

```go
// Function call
OpCode{Cmd: "Call", Args: []any{"functionName", arg1, arg2, ...}}

// Function definition (for user functions)
OpCode{Cmd: "DefineFunc", Args: []any{"functionName", []OpCode{...}}}

// Return from function
OpCode{Cmd: "Return", Args: []any{returnValue}}
```

## CLI Interface

### Command-Line Arguments

**Direct Mode:**
```bash
son-et <directory>
```
- Executes the project in the specified directory
- Locates main function in TFY files
- Converts TFY to OpCode at runtime
- Loads assets from project directory

**Help:**
```bash
son-et --help
son-et
```
- Displays usage information

**Embedded Mode:**
- No command-line arguments needed
- Executable runs embedded project automatically
- Assets loaded from embedded data

### Build Process

**Embedded Mode Build:**

Use Go build tags to conditionally compile embedded projects:

```go
// +build embed_kuma2

package main

import "embed"

//go:embed samples/kuma2/*
var embeddedFS embed.FS

var embeddedProject = "kuma2"
```

**Build Command:**
```bash
go build -tags embed_kuma2 -o kuma2 ./cmd/son-et
```

**Runtime Detection:**
```go
func main() {
    if embeddedProject != "" {
        // Embedded mode: execute embedded project
        executeEmbedded(embeddedFS, embeddedProject)
    } else {
        // Direct mode: execute from command-line argument
        executeDirect(os.Args[1])
    }
}
```

## Asset Management

### Direct Mode

- Load assets from project directory at runtime
- Use os.ReadFile for asset loading
- Perform case-insensitive file matching

### Embedded Mode

- Embed assets using //go:embed at build time
- Use embed.FS for asset loading
- Assets are part of the executable

### Unified Interface

```go
type AssetLoader interface {
    ReadFile(name string) ([]byte, error)
    ReadDir(name string) ([]fs.DirEntry, error)
}

// DirectAssetLoader loads from filesystem
type DirectAssetLoader struct {
    baseDir string
}

// EmbeddedAssetLoader loads from embed.FS
type EmbeddedAssetLoader struct {
    fs embed.FS
}
```

## Performance Considerations

**OpCode Generation:**
- Generate OpCode once (at startup for direct mode, at build time for embedded mode)
- Cache generated OpCode in memory
- No runtime parsing overhead after initial generation

**Variable Resolution:**
- Use hash maps for O(1) variable lookup
- Limit scope chain depth to avoid deep recursion
- Cache frequently accessed variables

**Asset Loading:**
- Load assets lazily (on first use)
- Cache loaded assets in memory
- Use case-insensitive matching for cross-platform compatibility

## Data Models

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for common data models (Picture, Cast, Window, VM Sequence).

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system—essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Direct Execution Equivalence

*For any* TFY script, executing it in direct mode SHALL produce the same behavior as executing the equivalent embedded OpCode.

**Validates: Requirements 2.1, 2.3**

### Property 2: Asset Discovery Completeness

*For any* TFY script, the interpreter SHALL discover all asset references (LoadPic, PlayMIDI, PlayWAVE) and include them in the asset list.

**Validates: Requirements C4.1, C4.2**

### Property 3: CLI Argument Handling

*For any* valid directory path, running `son-et <directory>` SHALL execute the project in that directory.

**Validates: Requirements 3.1**

### Property 4: Embedded Mode Execution

*For any* embedded project, running the executable without arguments SHALL execute the embedded project.

**Validates: Requirements 2.5**

## Error Handling

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for common error handling strategies.

### Interpreter-Specific Errors

**Parsing Errors:**
- When TFY script contains invalid syntax
- Action: Report error with file, line, and column numbers, halt execution

**Asset Discovery Errors:**
- When asset file referenced in script cannot be found
- Action: Report warning with filename and line number, continue execution

**Directory Not Found:**
- When specified directory does not exist
- Action: Report error, display usage information, exit

## Testing Strategy

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for common testing strategies.

### Integration Testing

**End-to-End Scenarios:**
- Test direct mode execution with sample scripts
- Test embedded mode execution
- Verify asset loading in both modes
- Test error reporting with invalid scripts

**Platform Testing:**
- Verify builds on macOS
- Test with various project structures

## Implementation Notes

See [COMMON_DESIGN.md](../COMMON_DESIGN.md) for critical design constraints and common mistakes to avoid.

### Development Workflow

**Testing Direct Mode:**
```bash
go run cmd/son-et/main.go samples/kuma2
```

**Building Embedded Mode:**
```bash
go build -tags embed_kuma2 -o kuma2 ./cmd/son-et
./kuma2
```

**Debugging:**
- Use DEBUG_LEVEL=2 for detailed logging
- Check asset discovery with verbose output
- Verify OpCode generation correctness


## Critical Implementation Details

### step() Block Conversion

**IMPORTANT:** `step(n)` is NOT a loop construct. It sets the time duration for Wait operations.

**Correct Conversion:**
```go
case *ast.StepBlockStatement:
    // step(n) block: step(65) { ... }
    // This sets the step resolution and then executes the body once
    
    // First, emit SetStep operation
    setStepOp := OpCode{
        Cmd:  OpSetStep,
        Args: []any{int(s.Count)},
    }
    
    // Then, interpret the body statements
    bodyOps := []OpCode{}
    for _, stmt := range s.Body.Statements {
        ops, err := i.interpretStatement(stmt)
        if err != nil {
            return nil, err
        }
        bodyOps = append(bodyOps, ops...)
    }
    
    // Return SetStep followed by body operations
    result := []OpCode{setStepOp}
    result = append(result, bodyOps...)
    return result, nil
```

**WRONG Approach (DO NOT USE):**
```go
// WRONG: Treating step(n) as a loop
return []OpCode{{
    Cmd:  OpStep,
    Args: []any{s.Count, bodyOps},  // This creates a loop, which is incorrect
}}, nil
```

**Why This Matters:**
- `step(65)` means "set each Wait(1) to 65 time units"
- The block body executes once, not 65 times
- Misinterpreting this causes scripts to repeat incorrectly

### main() Function Execution

**IMPORTANT:** The `main()` function should NOT be wrapped in `RegisterSequence`.

**Correct Approach (in cmd/son-et/main.go):**
```go
engine.InitDirect(assetLoader, imageDecoder, func() {
    // Convert interpreter OpCode to engine OpCode format
    engineOps := convertToEngineOpCodes(script.Main.Body)
    
    // Execute the main sequence directly (not wrapped in RegisterSequence)
    // This allows mes() blocks to register their own sequences
    for _, op := range engineOps {
        engine.ExecuteOpDirect(op)
    }
})
```

**WRONG Approach (DO NOT USE):**
```go
// WRONG: Wrapping main() in RegisterSequence causes nested sequence deadlock
engine.RegisterSequence(engine.Time, engineOps)
```

**Why This Matters:**
- `mes()` blocks internally call `RegisterSequence`
- If `main()` is already in a sequence, you get nested sequences
- Nested sequences cause deadlock in TIME mode (outer waits for inner, but inner can't start)

## Implementation Guidelines

Based on real-world debugging experience:

1. **Verify Language Semantics:** Always validate FILLY behavior with actual samples before implementing
2. **Test with Real Scripts:** Use actual FILLY samples (like kuma2) to validate design assumptions
3. **Avoid Premature Abstraction:** Emit OpCodes directly rather than creating unnecessary wrapper constructs
4. **Document Execution Context:** Clearly specify whether functions expect to be called within sequences or directly
5. **Cross-Reference Common Design:** See [COMMON_DESIGN.md](../COMMON_DESIGN.md#lessons-learned-from-implementation) for high-level lessons learned
