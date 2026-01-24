# Requirements Document: OpCode Type Safety

## Introduction

This specification defines the requirements for improving type safety in the OpCode system by replacing the current generic `Args []any` field with strongly-typed structs for each operation type. The current implementation uses runtime type assertions throughout the VM execution, making the code error-prone and difficult to maintain. This change will provide compile-time type checking, better documentation through code, and improved maintainability.

## Glossary

- **OpCode**: A single operation instruction in the virtual machine, consisting of a command type and arguments
- **VM**: Virtual Machine - the execution engine that processes OpCode sequences
- **Code_Generator**: The component that converts AST nodes into OpCode sequences
- **Type_Assertion**: Runtime check to verify the type of an interface value
- **Compile_Time_Validation**: Type checking performed by the Go compiler before runtime
- **Parameter_Value**: A value that can be a literal (known at compile time), a variable reference (resolved at runtime), or an expression (evaluated at runtime)

## Requirements

### Requirement 1: Strongly-Typed OpCode Structures

**User Story:** As a developer, I want each OpCode type to have its own strongly-typed struct, so that I can catch type errors at compile time instead of runtime.

#### Acceptance Criteria

1. THE System SHALL define a separate struct type for each OpCmd variant
2. WHEN an OpCode is constructed, THE Code_Generator SHALL use the appropriate typed struct
3. WHEN the VM executes an OpCode, THE System SHALL eliminate runtime type assertions for argument access
4. THE System SHALL maintain the existing OpCmd enum for operation identification

### Requirement 2: Parameter Value Representation

**User Story:** As a developer, I want a unified way to represent parameters that can be literals, variables, or expressions, so that the type system accurately reflects the runtime behavior.

#### Acceptance Criteria

1. THE System SHALL define a Value type that can represent literals, variables, or nested OpCodes
2. WHEN a parameter is a literal value, THE System SHALL store it directly in the Value type
3. WHEN a parameter is a variable reference, THE System SHALL store the variable name in the Value type
4. WHEN a parameter is an expression, THE System SHALL store a nested OpCode in the Value type
5. THE System SHALL provide type-safe accessors for each Value variant

### Requirement 3: Assignment Operation Type Safety

**User Story:** As a developer, I want assignment operations to have strongly-typed parameters, so that I cannot accidentally pass incorrect argument types.

#### Acceptance Criteria

1. THE System SHALL define an AssignOp struct with fields for variable name and value
2. WHEN creating an assignment OpCode, THE Code_Generator SHALL use the AssignOp struct
3. WHEN executing an assignment, THE VM SHALL access fields directly without type assertions
4. THE AssignOp SHALL validate that the variable name is of type Variable

### Requirement 4: Binary Operation Type Safety

**User Story:** As a developer, I want binary operations to have strongly-typed parameters, so that operator and operand types are validated at compile time.

#### Acceptance Criteria

1. THE System SHALL define a BinaryOp struct with fields for operator, left operand, and right operand
2. WHEN creating a binary operation OpCode, THE Code_Generator SHALL use the BinaryOp struct
3. WHEN executing a binary operation, THE VM SHALL access operator and operands without type assertions
4. THE BinaryOp SHALL store the operator as a string type

### Requirement 5: Control Flow Operation Type Safety

**User Story:** As a developer, I want control flow operations (if, for, while) to have strongly-typed parameters, so that conditions and code blocks are properly validated.

#### Acceptance Criteria

1. THE System SHALL define an IfOp struct with fields for condition, then-block, and optional else-block
2. THE System SHALL define a ForOp struct with fields for init, condition, post, and body
3. THE System SHALL define a WhileOp struct with fields for condition and body
4. WHEN creating control flow OpCodes, THE Code_Generator SHALL use the appropriate typed structs
5. WHEN executing control flow operations, THE VM SHALL access fields directly without type assertions

### Requirement 6: Function Call Type Safety

**User Story:** As a developer, I want function calls to have strongly-typed parameters, so that function names and arguments are properly validated.

#### Acceptance Criteria

1. THE System SHALL define a CallOp struct with fields for function name and argument list
2. WHEN creating a function call OpCode, THE Code_Generator SHALL use the CallOp struct
3. WHEN executing a function call, THE VM SHALL access function name and arguments without type assertions
4. THE CallOp SHALL support variable-length argument lists

### Requirement 7: Wait and Timing Operation Type Safety

**User Story:** As a developer, I want wait and timing operations to have strongly-typed parameters, so that step counts and tick values are properly validated.

#### Acceptance Criteria

1. THE System SHALL define a WaitOp struct with a field for the wait count
2. THE System SHALL define a SetStepOp struct with a field for the step duration
3. WHEN creating wait or timing OpCodes, THE Code_Generator SHALL use the appropriate typed structs
4. WHEN executing wait operations, THE VM SHALL access the wait count without type assertions

### Requirement 8: Event Handler Registration Type Safety

**User Story:** As a developer, I want event handler registration to have strongly-typed parameters, so that event types and handler bodies are properly validated.

#### Acceptance Criteria

1. THE System SHALL define a RegisterEventHandlerOp struct with fields for event type and handler body
2. WHEN creating an event handler registration OpCode, THE Code_Generator SHALL use the RegisterEventHandlerOp struct
3. WHEN registering an event handler, THE VM SHALL access event type and body without type assertions
4. THE RegisterEventHandlerOp SHALL store the event type as a string

### Requirement 9: Array Operation Type Safety

**User Story:** As a developer, I want array operations to have strongly-typed parameters, so that array access and assignment are properly validated.

#### Acceptance Criteria

1. THE System SHALL define an ArrayAccessOp struct with fields for array reference and index
2. THE System SHALL define an ArrayAssignOp struct with fields for array reference, index, and value
3. WHEN creating array operation OpCodes, THE Code_Generator SHALL use the appropriate typed structs
4. WHEN executing array operations, THE VM SHALL access fields directly without type assertions

### Requirement 10: Backward Compatibility

**User Story:** As a developer, I want the refactoring to maintain backward compatibility with existing VM execution logic, so that all existing programs continue to work correctly.

#### Acceptance Criteria

1. WHEN the refactoring is complete, THE System SHALL execute all existing test programs without behavioral changes
2. THE System SHALL maintain the same OpCmd enum values for operation identification
3. THE System SHALL preserve the existing VM execution semantics for all operations
4. THE System SHALL maintain the same error handling behavior for invalid operations

### Requirement 11: OpCode Interface Design

**User Story:** As a developer, I want a clean interface for working with OpCodes, so that I can write generic code that works with any operation type.

#### Acceptance Criteria

1. THE System SHALL define an OpCode interface with a method to retrieve the OpCmd type
2. WHEN implementing the OpCode interface, THE System SHALL ensure all typed structs satisfy the interface
3. THE System SHALL allow OpCode slices to contain any operation type
4. THE System SHALL provide type switches or type assertions for accessing specific operation types when needed
