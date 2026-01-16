# Common Requirements

This document defines requirements that are shared across multiple son-et specifications.

## Glossary

See [GLOSSARY.md](GLOSSARY.md) for common terms.

## Requirements

### Requirement C1: Complete OpCode-Based Execution

**User Story:** As a developer, I want all TFY code to execute as OpCode sequences, so that variable scope is consistently managed through the VM.

#### Acceptance Criteria

1. WHEN the system processes a TFY script, THE System SHALL convert all statements to OpCode sequences
2. WHEN the system encounters code outside mes() blocks, THE System SHALL execute it as OpCode
3. WHEN the system encounters code inside mes() blocks, THE System SHALL execute it as OpCode
4. WHEN the system processes function definitions, THE System SHALL execute function bodies as OpCode sequences
5. THE System SHALL provide unified variable scope management through the VM for all code

### Requirement C2: Variable Scope Unification

**User Story:** As a developer, I want consistent variable scope management, so that variables behave predictably throughout my script.

#### Acceptance Criteria

1. WHEN a variable is declared in the main function, THE VM SHALL make it accessible to all mes() blocks
2. WHEN a variable is declared in a user function, THE VM SHALL scope it to that function
3. WHEN a variable is declared globally, THE VM SHALL make it accessible to all functions and mes() blocks
4. WHEN a mes() block references a variable, THE VM SHALL resolve it through the scope chain
5. WHEN a variable is assigned in a mes() block, THE VM SHALL update the variable in the correct scope

### Requirement C3: VM-Based Execution Engine

**User Story:** As a developer, I want a robust VM execution engine, so that my scripts execute reliably and predictably.

#### Acceptance Criteria

1. THE System SHALL execute OpCode structures through a virtual machine
2. THE System SHALL use a Sequencer component for managing OpCode execution and timing
3. THE System SHALL provide an ExecuteOp function for executing individual OpCode commands
4. THE System SHALL provide a ResolveArg function for resolving variable references and nested expressions
5. THE System SHALL support all FILLY language OpCode commands

### Requirement C4: Asset Management

**User Story:** As a developer, I want automatic asset discovery and loading, so that I don't need to manually specify asset files.

#### Acceptance Criteria

1. WHEN the system scans a project directory, THE System SHALL identify all LoadPic, PlayMIDI, and PlayWAVE calls in TFY scripts
2. WHEN the system identifies asset references, THE System SHALL locate the files in the project directory
3. THE System SHALL perform case-insensitive file matching for asset references (Windows 3.1 compatibility)
4. WHEN an asset file is not found, THE System SHALL report a clear error message
5. THE System SHALL support loading assets from both filesystem (direct mode) and embedded data (embedded mode)

### Requirement C5: Backward Compatibility

**User Story:** As a developer, I want existing TFY scripts to work without modification, so that I don't need to rewrite my applications.

#### Acceptance Criteria

1. WHEN the interpreter processes an existing TFY script, THE System SHALL execute it correctly without script modifications
2. WHEN the interpreter encounters FILLY language constructs, THE System SHALL support all existing syntax
3. WHEN the interpreter processes mes() blocks, THE System SHALL maintain timing behavior compatibility
4. WHEN the interpreter processes user-defined functions, THE System SHALL maintain calling convention compatibility
5. WHEN the interpreter processes variable declarations, THE System SHALL maintain scope behavior compatibility

### Requirement C6: Error Reporting

**User Story:** As a developer, I want clear error messages with source locations, so that I can quickly identify and fix issues.

#### Acceptance Criteria

1. WHEN a parsing error occurs, THE System SHALL report the error with TFY script line and column numbers
2. WHEN a runtime error occurs, THE System SHALL report the error with the OpCode command that failed
3. WHEN an asset loading error occurs, THE System SHALL report which asset file could not be loaded
4. WHEN a variable resolution error occurs, THE System SHALL report which variable was not found
5. WHEN a function call error occurs, THE System SHALL report which function was called incorrectly
