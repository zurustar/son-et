# Implementation Plan: OpCode Type Safety

## Overview

This implementation plan converts the OpCode system from using generic `Args []any` to strongly-typed structs. The approach is incremental: first add new types alongside existing code, then migrate the code generator, then the VM, and finally remove old code. This ensures the system remains functional throughout the refactoring.

## Tasks

- [x] 1. Create new type system foundation
  - [x] 1.1 Define Value type and constructors in opcode.go
    - Create `Value` struct with `kind` and `data` fields
    - Define `ValueKind` enum (ValueLiteral, ValueVariable, ValueExpression, ValueArray)
    - Implement constructor functions: `LiteralValue()`, `VariableValue()`, `ExpressionValue()`, `ArrayValue()`
    - Implement accessor methods: `AsLiteral()`, `AsVariable()`, `AsExpression()`, `AsArray()`
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 1.2 Write property test for Value round-trip preservation
    - **Property 2: Value Round-Trip Preservation**
    - **Validates: Requirements 2.2, 2.3, 2.4**

  - [x] 1.3 Write property test for Value type safety
    - [x] 1.3.1 Write test structure and literal type safety test
      - Create test function `TestValueTypeSafetyProperty`
      - Test that literal values return true only for AsLiteral()
      - **Property 3: Value Type Safety (Literal)**
    - [x] 1.3.2 Write variable type safety test
      - Test that variable values return true only for AsVariable()
      - **Property 3: Value Type Safety (Variable)**
    - [x] 1.3.3 Write expression type safety test
      - Test that expression values return true only for AsExpression()
      - **Property 3: Value Type Safety (Expression)**
    - [x] 1.3.4 Write array type safety test
      - Test that array values return true only for AsArray()
      - **Property 3: Value Type Safety (Array)**
      - **Validates: Requirements 2.5**

  - [x] 1.4 Define OpCode interface in opcode.go
    - Create `OpCode` interface with `GetCmd() OpCmd` method
    - Keep existing `OpCode` struct temporarily (will be removed later)
    - _Requirements: 11.1_

- [x] 2. Define typed operation structs
  - [x] 2.1 Define assignment and expression operations
    - Create `AssignOp` struct with Variable and Value fields
    - Create `BinaryOp` struct with Operator, Left, Right fields
    - Create `UnaryOp` struct with Operator and Operand fields
    - Implement `GetCmd()` for each
    - _Requirements: 1.1, 3.1, 4.1_

  - [x] 2.2 Define control flow operations
    - Create `IfOp` struct with Condition, ThenBlock, ElseBlock fields
    - Create `ForOp` struct with Init, Condition, Post, Body fields
    - Create `WhileOp` struct with Condition and Body fields
    - Create `SwitchOp` and `SwitchCase` structs
    - Create `BreakOp` and `ContinueOp` structs (no fields)
    - Implement `GetCmd()` for each
    - _Requirements: 1.1, 5.1, 5.2, 5.3_

  - [x] 2.3 Define function and timing operations
    - Create `CallOp` struct with Function and Args fields
    - Create `WaitOp` struct with Count field
    - Create `SetStepOp` struct with Duration field
    - Implement `GetCmd()` for each
    - _Requirements: 1.1, 6.1, 7.1, 7.2_

  - [x] 2.4 Define event handler and array operations
    - Create `RegisterEventHandlerOp` struct with EventType and Body fields
    - Create `ArrayAccessOp` struct with Array and Index fields
    - Create `ArrayAssignOp` struct with Array, Index, Value fields
    - Implement `GetCmd()` for each
    - _Requirements: 1.1, 8.1, 9.1, 9.2_

  - [x] 2.5 Write property test for OpCmd identification
    - **Property 4: OpCmd Identification**
    - **Validates: Requirements 1.4, 10.2**

- [x] 3. Checkpoint - Verify type definitions compile
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. Update code generator to use typed operations
  - [x] 4.1 Update generateExpression to return Value
    - Modify `generateExpression()` to return `Value` instead of `any`
    - Update literal handling to use `LiteralValue()`
    - Update variable handling to use `VariableValue()`
    - Update nested OpCode handling to use `ExpressionValue()`
    - Update array literal handling to use `ArrayValue()`
    - _Requirements: 1.2_

  - [x] 4.2 Update assignment and expression generation
    - Modify `generateAssignStatement()` to create `AssignOp`
    - Modify binary operation generation to create `BinaryOp`
    - Modify unary operation generation to create `UnaryOp`
    - Update all call sites to use new types
    - _Requirements: 3.2, 4.2_

  - [x] 4.3 Update control flow generation
    - Modify `generateIfStatement()` to create `IfOp`
    - Modify `generateForStatement()` to create `ForOp`
    - Modify `generateWhileStatement()` to create `WhileOp`
    - Modify `generateSwitchStatement()` to create `SwitchOp`
    - Update break/continue to create `BreakOp`/`ContinueOp`
    - _Requirements: 5.4_

  - [x] 4.4 Update function and timing generation
    - Modify `generateExpressionStatement()` to create `CallOp` for function calls
    - Modify function call expression generation to create `CallOp`
    - Modify `generateStepStatement()` to create `WaitOp` and `SetStepOp`
    - _Requirements: 6.2, 7.3_

  - [x] 4.5 Update event handler and array generation
    - Modify `generateMesStatement()` to create `RegisterEventHandlerOp`
    - Modify array access generation to create `ArrayAccessOp`
    - Modify array assignment generation to create `ArrayAssignOp`
    - _Requirements: 8.2, 9.3_

  - [x] 4.6 Write property test for code generator type correctness
    - **Property 1: Code Generator Type Correctness**
    - **Validates: Requirements 1.2, 3.2, 4.2, 5.4, 6.2, 7.3, 8.2, 9.3**

- [x] 5. Checkpoint - Verify code generation works
  - Ensure all tests pass, ask the user if questions arise.

- [-] 6. Update VM to use typed operations
  - [x] 6.1 Update ExecuteOp to use type switch
    - Replace `switch op.Cmd` with `switch op := op.(type)`
    - Update each case to handle the typed operation struct
    - Remove old OpCmd-based dispatch
    - _Requirements: 1.3_

  - [x] 6.2 Update evaluateValue to work with Value type
    - Modify `evaluateValue()` to accept and handle `Value` type
    - Use Value accessors instead of type assertions
    - Update expression evaluation to work with typed operations
    - _Requirements: 2.5_

  - [x] 6.3 Update assignment and expression execution
    - Modify `executeAssign()` to accept `AssignOp` and use fields directly
    - Modify `evaluateBinaryOp()` to accept `BinaryOp` and use fields directly
    - Remove all type assertions from these methods
    - _Requirements: 3.3, 4.3_

  - [x] 6.4 Update control flow execution
    - Modify `executeIf()` to accept `IfOp` and use fields directly
    - Modify `executeFor()` to accept `ForOp` and use fields directly
    - Modify `executeWhile()` to accept `WhileOp` and use fields directly
    - Remove all type assertions from these methods
    - _Requirements: 5.5_

  - [ ] 6.5 Update function and timing execution
    - Modify `executeCall()` to accept `CallOp` and use fields directly
    - Modify `executeWait()` to accept `WaitOp` and use fields directly
    - Modify `executeSetStep()` to accept `SetStepOp` and use fields directly
    - Remove all type assertions from these methods
    - _Requirements: 6.3, 7.4_

  - [ ] 6.6 Update event handler and array execution
    - Modify `executeRegisterEventHandler()` to accept `RegisterEventHandlerOp` and use fields directly
    - Update array operation execution to work with typed operations
    - Remove all type assertions from these methods
    - _Requirements: 8.3, 9.4_

  - [ ] 6.7 Write property test for behavioral equivalence
    - **Property 5: Behavioral Equivalence**
    - **Validates: Requirements 10.1, 10.3**

  - [ ] 6.8 Write property test for error preservation
    - **Property 6: Error Preservation**
    - **Validates: Requirements 10.4**

- [ ] 7. Checkpoint - Verify VM execution works
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 8. Update serializer to handle typed operations
  - [ ] 8.1 Update serializeOpCode to handle typed operations
    - Add type switch to handle each operation type
    - Generate appropriate Go code for each typed struct
    - Update Value serialization to handle all variants
    - _Requirements: 10.1_

  - [ ] 8.2 Update multi-title serialization
    - Update multi-title code generation to work with typed operations
    - Ensure all title functions serialize correctly
    - _Requirements: 10.1_

- [ ] 9. Update tests to use new types
  - [ ] 9.1 Update opcode_test.go
    - Modify tests to construct typed operations
    - Update assertions to check typed fields
    - Remove tests that check old Args field
    - _Requirements: 10.1_

  - [ ] 9.2 Update codegen tests
    - Update all code generation tests to expect typed operations
    - Verify generated operations have correct types and fields
    - _Requirements: 10.1_

  - [ ] 9.3 Update serializer tests
    - Update serialization tests to work with typed operations
    - Verify serialized code compiles and runs correctly
    - _Requirements: 10.1_

- [ ] 10. Remove old OpCode struct and cleanup
  - [ ] 10.1 Remove old OpCode struct definition
    - Delete the old `OpCode` struct with `Cmd` and `Args` fields
    - Ensure OpCode interface is the only OpCode type
    - _Requirements: 1.1_

  - [ ] 10.2 Clean up any remaining references
    - Search for any remaining uses of old OpCode struct
    - Update or remove as appropriate
    - _Requirements: 10.1_

- [ ] 11. Final checkpoint - Run full test suite
  - Ensure all tests pass, ask the user if questions arise.

## Notes

- Each task references specific requirements for traceability
- Checkpoints ensure incremental validation
- Property tests validate universal correctness properties
- The migration is designed to be incremental and non-breaking until the final cleanup
