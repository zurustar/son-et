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

### 0.1. Testing Before Commit (CRITICAL)

**RULE: NEVER commit code without testing it first**

This rule has been emphasized since the beginning of development and is absolutely critical:

1. **Before ANY commit:**
   - Run the code with the relevant sample
   - Verify the behavior is correct
   - Check the logs for errors
   - Confirm with user if needed

2. **Testing workflow:**
   ```bash
   # Test your changes
   go run cmd/son-et/main.go samples/[NAME] > test.log 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat test.log
   
   # Review the log
   # Verify behavior
   
   # Only then commit
   git add .
   git commit -m "description"
   ```

3. **Why this is critical:**
   - Untested code often doesn't work
   - Committing broken code wastes time
   - It shows lack of care and attention
   - User has to repeatedly point this out

**If you commit without testing, you are doing it wrong. No exceptions.**

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
   - Interpreter changes (AST to OpCode conversion) in `pkg/compiler/interpreter/`
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

1. **Run the implementation**
   - Follow the detailed execution process in `build-workflow.md`
   - Use direct mode for rapid iteration

2. **Provide execution commands to user**
   - Generate the complete command sequence for the user
   - Use background execution with timeout to prevent orphaned processes
   - Commands must be executable from repository root
   - Use only macOS default commands
   - Standard format: `go run cmd/son-et/main.go samples/[NAME] > log.txt 2>&1 & PID=$!; sleep 5; kill $PID 2>/dev/null; cat log.txt`

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

- **Execution Success + Correct Behavior**: Move to next feature or task
- **Parsing Errors**: Fix interpreter or AST conversion issues
- **Runtime Errors**: Debug using log timestamps, fix implementation
- **Incorrect Behavior**: Review requirements, adjust implementation
- **Asset Loading Issues**: Check file paths and directory structure

**Log Analysis Process:**
1. Read the generated log file (e.g., `execution.log`)
2. Review timestamped log for error patterns
3. Identify the failure point (parsing, startup, runtime)
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

- Refer to `build-workflow.md` for detailed execution requirements and asset management

## Debugging and Execution Issues

For detailed debugging procedures, execution troubleshooting, and asset management, refer to `build-workflow.md`.

## Quick Reference Commands

### Run Tests

**CRITICAL RULE: Always set timeout when running tests**

```bash
# Run all tests with timeout (REQUIRED)
go test -timeout=30s ./pkg/compiler/... ./pkg/engine/...

# Run with race detector and timeout
go test -timeout=30s -race ./pkg/compiler/... ./pkg/engine/...

# Run specific test with timeout
go test -v -timeout=20s -run TestSpecificFunction ./pkg/engine/
```

**Why timeout is required:**
- Tests may hang due to blocking operations (e.g., TIME mode in VM)
- Timeout allows early detection of deadlocks or infinite loops
- Default timeout is too long for development feedback
- Recommended timeout: 20-30 seconds for unit tests

**Never run tests without `-timeout` flag!**

### Process Management After Tests

**CRITICAL: Always check for orphaned processes after running tests**

```bash
# Check for orphaned Go test processes
ps aux | grep "go-build" | grep -v grep

# If processes found, kill them
kill -9 <PID>

# Verify cleanup
ps aux | grep "go-build" | grep -v grep
```

**Why this is important:**
- Go tests run from cache: `/Users/.../Library/Caches/go-build/.../main`
- Tests with GUI (Ebiten) may not terminate automatically
- Orphaned processes keep windows open and consume resources
- DO NOT search for "son-et" or ".test" - use "go-build" instead

### Check Implementation Status
```bash
# Check which functions are implemented
grep -r "func.*(" pkg/engine/*.go | grep -v "^//"

# Check for TODO or FIXME
grep -r "TODO\|FIXME" pkg/
```

## Success Criteria

A feature implementation is complete when:

1. ✅ Sample scenario executes without errors
2. ✅ Interpreter converts TFY to OpCode correctly
3. ✅ Executable runs without crashes
4. ✅ User confirms correct behavior
5. ✅ (Optional) Tests pass for the feature

## Notes

- **Incremental Development**: Implement one feature at a time
- **User-Driven**: Wait for user verification before proceeding
- **Sample-Driven**: Use sample scenarios to guide implementation priorities
- **Test Later**: Focus on working implementation first, comprehensive tests later
- **Document Decisions**: Note any design decisions or deviations in code comments
