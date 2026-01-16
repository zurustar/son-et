# Requirements Document: Interpreter Architecture

## Introduction

This specification defines the son-et interpreter system's architecture and execution modes. The son-et system provides two execution modes: direct execution for rapid development iteration, and embedded executable generation for distribution.

This specification focuses on the CLI interface, execution modes, and build process. Core VM functionality and FILLY language features are defined in [COMMON_REQUIREMENTS.md](../COMMON_REQUIREMENTS.md) and [core-engine/requirements.md](../core-engine/requirements.md).

## Glossary

See [GLOSSARY.md](../GLOSSARY.md) for common terms used across all son-et specifications.

## Common Requirements

This specification depends on the following common requirements defined in [COMMON_REQUIREMENTS.md](../COMMON_REQUIREMENTS.md):
- **C1**: Complete OpCode-Based Execution
- **C2**: Variable Scope Unification
- **C3**: VM-Based Execution Engine
- **C4**: Asset Management
- **C5**: Backward Compatibility
- **C6**: Error Reporting

## Requirements

### Requirement 1: Direct Execution Mode

**User Story:** As a developer, I want to run TFY projects directly during development, so that I can iterate quickly.

#### Acceptance Criteria

1. WHEN a developer runs `son-et <directory>`, THE System SHALL execute the project immediately in direct mode
2. WHEN executing in direct mode, THE System SHALL locate the main function in TFY files within the directory
3. WHEN executing in direct mode, THE System SHALL convert TFY to OpCode at runtime
4. WHEN executing in direct mode, THE System SHALL load assets from the project directory
5. WHEN a runtime error occurs in direct mode, THE System SHALL report the error with TFY script line numbers

### Requirement 2: Embedded Executable Generation

**User Story:** As a developer, I want to create a custom son-et executable with my project embedded, so that I can distribute a standalone application.

#### Acceptance Criteria

1. WHEN a developer builds son-et with a project directory specified, THE System SHALL create an executable with the project embedded
2. WHEN the embedded executable runs, THE System SHALL execute the embedded project without requiring external TFY files
3. WHEN the embedded executable runs, THE System SHALL load assets from embedded data
4. WHEN building with embedded project, THE System SHALL convert TFY to OpCode at build time
5. WHEN the embedded executable runs without arguments, THE System SHALL execute the embedded project

### Requirement 3: Command-Line Interface

**User Story:** As a developer, I want a clear command-line interface, so that I can easily execute projects directly or build embedded executables.

#### Acceptance Criteria

1. WHEN a developer runs `son-et <directory>`, THE System SHALL execute the project in direct mode
2. WHEN a developer runs `son-et --help`, THE System SHALL display usage information
3. WHEN a developer runs `son-et` without arguments, THE System SHALL display usage information
4. WHEN building son-et with embedded project, THE System SHALL support build-time configuration for the project directory
5. WHEN the embedded son-et executable runs, THE System SHALL execute the embedded project

### Requirement 4: Documentation Updates

**User Story:** As a developer, I want updated documentation, so that I understand how to use the interpreter system.

#### Acceptance Criteria

1. THE Documentation SHALL update README.md to describe basic usage of son-et as an interpreter
2. THE Documentation SHALL update build-workflow.md to describe the new development workflow
3. THE Documentation SHALL update development-workflow.md to reflect the interpreter-based development process
4. THE Documentation SHALL provide examples of both direct execution and embedded distribution workflows
5. THE Documentation SHALL remove references to the old transpiler-based workflow
