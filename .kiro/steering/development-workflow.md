---
inclusion: always
---

# Development Workflow

## Overview

This document describes the step-by-step workflow for implementing features in the son-et project using Kiro. This workflow is designed for incremental feature development driven by sample scenarios.

For detailed build procedures, debugging, and asset management, see `build-workflow.md`.

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

1. **Build the implementation**
   - Follow the detailed build process in `build-workflow.md`
   - Ensure all assets are properly embedded

2. **Provide execution commands to user**
   - Generate the complete command sequence for the user
   - Include timestamped logging for debugging
   - Commands must be executable from repository root
   - Use only macOS default commands

3. **User executes and reports completion**
   - User runs the provided commands
   - User reports completion and optionally describes the sample behavior

4. **Analyze execution results**
   - Read the generated log file to review execution
   - Identify any runtime problems from the log
   - Confirm successful execution or diagnose failures

5. **Request user verification**
   - Ask user to confirm the behavior is correct
   - Wait for user feedback on functionality

### 5. Feedback Loop

Based on log analysis and user completion report:

- **Build Success + Correct Behavior**: Move to next feature or task
- **Build Errors**: Fix transpiler or code generation issues
- **Runtime Errors**: Debug using log timestamps, fix implementation
- **Incorrect Behavior**: Review requirements, adjust implementation
- **Asset Loading Issues**: Check file paths and embedding directives

**Log Analysis Process:**
1. Read the generated log file (e.g., `game_execution.log`)
2. Review timestamped log for error patterns
3. Identify the failure point (build, startup, runtime)
4. Cross-reference with known issues in `build-workflow.md`
5. Implement fixes and repeat verification cycle

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

- Refer to `build-workflow.md` for detailed build requirements and asset management

## Debugging and Build Issues

For detailed debugging procedures, build troubleshooting, and asset management, refer to `build-workflow.md`.

## Quick Reference Commands

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
