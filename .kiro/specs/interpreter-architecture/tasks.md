# Implementation Plan: Interpreter Architecture

## Overview

This implementation plan covers the son-et interpreter system, which provides two execution modes:
1. **Direct Mode**: Execute TFY projects directly from a directory for rapid development iteration
2. **Embedded Mode**: Create standalone executables with embedded projects for distribution

The interpreter builds on the existing compiler infrastructure (lexer, parser, AST) and runtime engine (VM, graphics, audio) defined in the core-engine spec.

**Implementation Status Summary:**
- ⚠️ Interpreter component - Pending
- ⚠️ CLI interface - Pending
- ⚠️ Asset management (direct vs embedded) - Pending
- ⚠️ Build system for embedded mode - Pending
- ⚠️ Documentation updates - Pending

## Tasks

- [x] 0. Asset Management Infrastructure (Moved from core-engine)
  - [x] 0.1 Add dependency injection for external dependencies
    - Create AssetLoader interface to abstract asset sources
    - Create ImageDecoder interface to abstract BMP decoding
    - Allow mock implementations for testing
    - Add constructor parameters for dependency injection
    - _Requirements: C4.1, C4.2_
    - _Note: Originally implemented in core-engine/tasks.md Task 0.3_
    - _Note: Moved here as asset loading strategy is interpreter-architecture concern_

  - [x] 0.2 Implement EmbedFSAssetLoader for embedded mode
    - Implement AssetLoader for embed.FS
    - Use embed.FS.ReadFile for file loading
    - Support case-insensitive filename matching
    - Handle embedded directory structure
    - _Requirements: 2.3, C4.3_
    - _Note: Already implemented in pkg/engine/asset_loader.go_

- [x] 1. Create Interpreter Component
  - [x] 1.1 Create interpreter package structure
    - Create `pkg/compiler/interpreter/` directory
    - Define `Interpreter` struct with assets, globals, userFuncs fields
    - Define `Script` struct to hold compiled OpCode sequences
    - Define `Function` struct for user-defined functions
    - _Requirements: 1.1, 1.2, 1.3_

  - [x] 1.2 Implement AST to OpCode conversion
    - Implement `Interpret(program *ast.Program) (*Script, error)` method
    - Convert function definitions to OpCode sequences
    - Convert statements to OpCode sequences
    - Convert expressions to OpCode sequences
    - Handle #include directives recursively
    - _Requirements: 1.3, C1.1, C1.2_

  - [x] 1.3 Implement asset discovery
    - Implement `scanAssets(program *ast.Program) []string` method
    - Walk AST to find LoadPic, PlayMIDI, PlayWAVE calls
    - Extract string literal filenames
    - Return unique list of asset filenames
    - _Requirements: 1.4, C4.1, C4.2_

  - [x] 1.4 Implement variable scope tracking
    - Track global variable declarations
    - Track function-level local variables
    - Generate scope information for VM execution
    - Support case-insensitive variable names
    - _Requirements: C2.1, C2.2, C2.3_

  - [x] 1.5 Write unit tests for interpreter
    - Test simple script conversion to OpCode
    - Test function definition conversion
    - Test asset discovery
    - Test variable scope tracking
    - Test #include handling
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [x] 2. Implement CLI Interface
  - [x] 2.1 Update cmd/son-et/main.go for direct mode
    - Parse command-line arguments
    - Detect direct mode: `son-et <directory>`
    - Display help: `son-et --help` or `son-et`
    - Validate directory exists
    - _Requirements: 3.1, 3.2, 3.3_

  - [x] 2.2 Implement direct mode execution
    - Locate TFY files in specified directory
    - Read and parse TFY files
    - Convert to OpCode using interpreter
    - Initialize engine with DirectAssetLoader
    - Execute OpCode through VM
    - _Requirements: 1.1, 1.2, 1.3, 1.4, 1.5_

  - [x] 2.3 Implement error reporting for direct mode
    - Report parsing errors with file, line, column
    - Report asset loading errors with filename
    - Report runtime errors with TFY line numbers
    - Display helpful error messages
    - _Requirements: 1.5, C6.1, C6.2_

  - [x] 2.4 Write integration tests for CLI
    - Test direct mode with sample projects
    - Test help display
    - Test error handling (directory not found, invalid TFY)
    - _Requirements: 3.1, 3.2, 3.3_

- [x] 3. Implement Asset Management for Direct Mode
  - [x] 3.1 Implement FilesystemAssetLoader for direct mode
    - Implement AssetLoader for filesystem access (direct mode)
    - Use os.ReadFile for file loading
    - Support case-insensitive filename matching
    - Handle relative paths from project directory
    - _Requirements: 1.4, C4.3_

  - [x] 3.2 Write unit tests for FilesystemAssetLoader
    - Test filesystem loading with test files
    - Test case-insensitive matching
    - Test relative path resolution
    - Test error handling (file not found)
    - _Requirements: C4.1, C4.2, C4.3_

- [x] 4. Implement Embedded Mode Build System
  - [x] 4.1 Create build tag structure
    - Design build tag naming convention (e.g., embed_kuma2)
    - Create template for embedded project files
    - Document build tag usage
    - _Requirements: 3.4, 2.1, 2.4_

  - [x] 4.2 Implement embedded mode detection
    - Add build-time variable for embedded project name
    - Detect embedded mode in main.go
    - Route to embedded execution path
    - _Requirements: 2.5, 3.5_

  - [x] 4.3 Implement embedded mode execution
    - Load embedded OpCode at startup
    - Initialize engine with EmbeddedAssetLoader
    - Execute embedded OpCode through VM
    - Handle errors gracefully
    - _Requirements: 2.2, 2.3, 2.5_

  - [x] 4.4 Create example embedded build
    - Create build file for sample project (e.g., kuma2)
    - Add //go:embed directives for assets
    - Test build process
    - Verify executable runs standalone
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [x] 4.5 Write integration tests for embedded mode
    - Test embedded executable creation
    - Test embedded execution
    - Verify assets are embedded correctly
    - Test error handling
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

- [ ] 5. Checkpoint - Verify Core Functionality
  - Run all unit tests and integration tests
  - Test direct mode with multiple sample projects
  - Test embedded mode build and execution
  - Verify asset loading in both modes
  - Ask user if questions arise

- [ ] 6. Write Property-Based Tests
  - [~] 6.1 Write property test for direct execution equivalence
    - **Property 1: Direct Execution Equivalence**
    - For any TFY script, direct mode and embedded mode produce same behavior
    - **Validates: Requirements 2.1, 2.3**

  - [~] 6.2 Write property test for asset discovery completeness
    - **Property 2: Asset Discovery Completeness**
    - For any TFY script, all asset references are discovered
    - **Validates: Requirements C4.1, C4.2**

  - [~] 6.3 Write property test for CLI argument handling
    - **Property 3: CLI Argument Handling**
    - For any valid directory path, son-et executes the project
    - **Validates: Requirements 3.1**

  - [~] 6.4 Write property test for embedded mode execution
    - **Property 4: Embedded Mode Execution**
    - For any embedded project, executable runs without arguments
    - **Validates: Requirements 2.5**

- [ ] 7. Update Documentation
  - [~] 7.1 Update README.md
    - Add interpreter usage section
    - Document direct mode: `son-et <directory>`
    - Document embedded mode build process
    - Add examples for both modes
    - Remove transpiler-based workflow references
    - _Requirements: 4.1, 4.4_

  - [~] 7.2 Update build-workflow.md
    - Replace transpiler workflow with interpreter workflow
    - Document direct mode development workflow
    - Document embedded mode build workflow
    - Update troubleshooting section
    - Add debugging tips for interpreter mode
    - _Requirements: 4.2, 4.4_

  - [~] 7.3 Update development-workflow.md
    - Update feature implementation workflow for interpreter
    - Update build and verification phase
    - Update debugging procedures
    - Remove transpiler-specific instructions
    - Add interpreter-specific best practices
    - _Requirements: 4.3, 4.4_

  - [~] 7.4 Create interpreter usage guide
    - Document command-line options
    - Provide examples for common use cases
    - Document error messages and solutions
    - Add FAQ section
    - _Requirements: 4.4, 4.5_

- [ ] 8. Integration Testing and Validation
  - [~] 8.1 Test with all existing sample projects
    - Run direct mode on each sample project
    - Verify correct execution and output
    - Check for any regressions
    - _Requirements: All_

  - [~] 8.2 Build embedded executables for samples
    - Create embedded builds for key samples
    - Test standalone execution
    - Verify asset embedding
    - Check executable size and performance
    - _Requirements: 2.1, 2.2, 2.3, 2.4, 2.5_

  - [~] 8.3 Performance benchmarking
    - Benchmark direct mode startup time
    - Benchmark OpCode generation time
    - Compare with previous transpiler performance
    - Profile and optimize hot paths
    - _Requirements: All_

  - [~] 8.4 Cross-platform testing
    - Test on macOS (primary platform)
    - Verify asset loading on different filesystems
    - Test case-insensitive matching
    - _Requirements: All_

- [ ] 9. Final Checkpoint - Complete System Validation
  - Ensure all tests pass
  - Verify all requirements are implemented
  - Review code coverage (aim for >80% on critical paths)
  - Ask user if questions arise
  - Confirm spec is complete and ready for production use

## Notes

- The interpreter builds on existing compiler infrastructure (lexer, parser, AST)
- The interpreter uses existing runtime engine (VM, graphics, audio)
- Direct mode is for development; embedded mode is for distribution
- Asset management is unified through AssetLoader interface
- Documentation updates are critical for user adoption
- Property tests validate correctness across all execution modes

## Dependencies

This spec depends on:
- **core-engine spec**: Provides compiler infrastructure and runtime engine
- **COMMON_DESIGN.md**: Defines shared design elements (OpCode, VM, etc.)
- **COMMON_REQUIREMENTS.md**: Defines shared requirements (C1-C6)

## Success Criteria

Implementation is complete when:
1. ✅ Direct mode executes TFY projects from directories
2. ✅ Embedded mode creates standalone executables
3. ✅ Asset loading works in both modes
4. ✅ All property tests pass
5. ✅ Documentation is updated and accurate
6. ✅ All sample projects run correctly in both modes
