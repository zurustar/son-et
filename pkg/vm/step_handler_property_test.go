package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// genOpCodeCmd generates a random OpCode command type (excluding SetStep).
func genOpCodeCmdWithoutSetStep() gopter.Gen {
	cmds := []opcode.Cmd{
		opcode.Assign,
		opcode.ArrayAssign,
		opcode.Call,
		opcode.BinaryOp,
		opcode.UnaryOp,
		opcode.ArrayAccess,
		opcode.If,
		opcode.For,
		opcode.While,
		opcode.Switch,
		opcode.Break,
		opcode.Continue,
		opcode.RegisterEventHandler,
		opcode.Wait,
		opcode.DefineFunction,
	}
	return gen.IntRange(0, len(cmds)-1).Map(func(i int) opcode.Cmd {
		return cmds[i]
	})
}

// genSimpleOpCode generates a simple OpCode without nested blocks.
func genSimpleOpCode(includeSetStep bool) gopter.Gen {
	if includeSetStep {
		// 50% chance to generate SetStep
		return gen.Bool().FlatMap(func(isSetStep interface{}) gopter.Gen {
			if isSetStep.(bool) {
				return gen.IntRange(1, 100).Map(func(stepValue int) opcode.OpCode {
					return opcode.OpCode{
						Cmd:  opcode.SetStep,
						Args: []any{stepValue},
					}
				})
			}
			return genOpCodeCmdWithoutSetStep().Map(func(cmd opcode.Cmd) opcode.OpCode {
				return opcode.OpCode{
					Cmd:  cmd,
					Args: []any{},
				}
			})
		}, nil)
	}
	return genOpCodeCmdWithoutSetStep().Map(func(cmd opcode.Cmd) opcode.OpCode {
		return opcode.OpCode{
			Cmd:  cmd,
			Args: []any{},
		}
	})
}

// genOpCodeSequence generates a sequence of OpCodes.
// If includeSetStep is true, the sequence may contain SetStep opcodes.
func genOpCodeSequence(includeSetStep bool, maxDepth int) gopter.Gen {
	if maxDepth <= 0 {
		// Base case: generate simple opcodes only
		return gen.SliceOfN(5, genSimpleOpCode(includeSetStep))
	}

	return gen.IntRange(0, 5).FlatMap(func(length interface{}) gopter.Gen {
		n := length.(int)
		if n == 0 {
			return gen.Const([]opcode.OpCode{})
		}

		return gen.SliceOfN(n, genOpCodeWithNesting(includeSetStep, maxDepth-1))
	}, nil)
}

// genOpCodeWithNesting generates an OpCode that may contain nested blocks.
func genOpCodeWithNesting(includeSetStep bool, maxDepth int) gopter.Gen {
	return gen.IntRange(0, 10).FlatMap(func(choice interface{}) gopter.Gen {
		c := choice.(int)

		// 30% chance to generate SetStep if allowed
		if includeSetStep && c < 3 {
			return gen.IntRange(1, 100).Map(func(stepValue int) opcode.OpCode {
				return opcode.OpCode{
					Cmd:  opcode.SetStep,
					Args: []any{stepValue},
				}
			})
		}

		// 20% chance to generate If with nested blocks
		if c < 5 && maxDepth > 0 {
			return genOpCodeSequence(includeSetStep, maxDepth-1).FlatMap(func(thenBlock interface{}) gopter.Gen {
				return genOpCodeSequence(includeSetStep, maxDepth-1).Map(func(elseBlock []opcode.OpCode) opcode.OpCode {
					return opcode.OpCode{
						Cmd:  opcode.If,
						Args: []any{true, thenBlock.([]opcode.OpCode), elseBlock},
					}
				})
			}, nil)
		}

		// 10% chance to generate While with nested block
		if c < 6 && maxDepth > 0 {
			return genOpCodeSequence(includeSetStep, maxDepth-1).Map(func(body []opcode.OpCode) opcode.OpCode {
				return opcode.OpCode{
					Cmd:  opcode.While,
					Args: []any{true, body},
				}
			})
		}

		// 10% chance to generate For with nested blocks
		if c < 7 && maxDepth > 0 {
			return genOpCodeSequence(includeSetStep, maxDepth-1).FlatMap(func(initBlock interface{}) gopter.Gen {
				return genOpCodeSequence(includeSetStep, maxDepth-1).FlatMap(func(postBlock interface{}) gopter.Gen {
					return genOpCodeSequence(includeSetStep, maxDepth-1).Map(func(bodyBlock []opcode.OpCode) opcode.OpCode {
						return opcode.OpCode{
							Cmd:  opcode.For,
							Args: []any{initBlock.([]opcode.OpCode), true, postBlock.([]opcode.OpCode), bodyBlock},
						}
					})
				}, nil)
			}, nil)
		}

		// Otherwise, generate a simple opcode
		return genSimpleOpCode(includeSetStep)
	}, nil)
}

// containsSetStepInSequence checks if a sequence contains OpSetStep (for test verification).
func containsSetStepInSequence(opcodes []opcode.OpCode) bool {
	for _, op := range opcodes {
		if op.Cmd == opcode.SetStep {
			return true
		}
		// Check nested blocks
		for _, arg := range op.Args {
			if childOpcodes, ok := arg.([]opcode.OpCode); ok {
				if containsSetStepInSequence(childOpcodes) {
					return true
				}
			}
		}
	}
	return false
}

// Feature: fix-step-handler-removal, Property 1: OpSetStep検出とHasStepBlock設定の対応
// **Validates: Requirements 1.1, 1.2, 1.3**
//
// 任意のOpCodeシーケンスについて、そのシーケンスにOpSetStepが含まれている場合にのみ、
// containsOpSetStep関数がtrueを返すことを検証します。
func TestProperty1_OpSetStepDetectionMatchesHasStepBlock(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property 1a: OpSetStepを含むシーケンスでは、containsOpSetStepがtrueを返す
	// Validates: Requirement 1.2
	properties.Property("containsOpSetStep returns true when SetStep is present", prop.ForAll(
		func(opcodes []opcode.OpCode) bool {
			// Verify that the sequence actually contains SetStep
			hasSetStep := containsSetStepInSequence(opcodes)
			if !hasSetStep {
				// Skip this test case if no SetStep was generated
				return true
			}

			// The function under test should return true
			result := containsOpSetStep(opcodes)
			return result == true
		},
		genOpCodeSequence(true, 3), // Generate sequences that may contain SetStep
	))

	// Property 1b: OpSetStepを含まないシーケンスでは、containsOpSetStepがfalseを返す
	// Validates: Requirement 1.3
	properties.Property("containsOpSetStep returns false when SetStep is absent", prop.ForAll(
		func(opcodes []opcode.OpCode) bool {
			// The function under test should return false for sequences without SetStep
			result := containsOpSetStep(opcodes)
			return result == false
		},
		genOpCodeSequence(false, 3), // Generate sequences without SetStep
	))

	// Property 1c: containsOpSetStepの結果は、実際のSetStepの存在と一致する
	// Validates: Requirements 1.1, 1.2, 1.3
	properties.Property("containsOpSetStep result matches actual SetStep presence", prop.ForAll(
		func(opcodes []opcode.OpCode) bool {
			// Check if the sequence actually contains SetStep
			actualHasSetStep := containsSetStepInSequence(opcodes)

			// The function under test
			result := containsOpSetStep(opcodes)

			// The result should match the actual presence
			return result == actualHasSetStep
		},
		genOpCodeSequence(true, 3), // Generate mixed sequences
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_NestedSetStepDetection tests that SetStep is detected in deeply nested structures.
// **Validates: Requirements 1.1**
func TestProperty1_NestedSetStepDetection(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: SetStep in nested If blocks is detected
	properties.Property("SetStep in nested If blocks is detected", prop.ForAll(
		func(depth int) bool {
			// Create a deeply nested If structure with SetStep at the innermost level
			innermost := []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{10}},
			}

			current := innermost
			for i := 0; i < depth; i++ {
				current = []opcode.OpCode{
					{Cmd: opcode.If, Args: []any{true, current, []opcode.OpCode{}}},
				}
			}

			return containsOpSetStep(current) == true
		},
		gen.IntRange(0, 5), // Nesting depth from 0 to 5
	))

	// Property: SetStep in nested While blocks is detected
	properties.Property("SetStep in nested While blocks is detected", prop.ForAll(
		func(depth int) bool {
			// Create a deeply nested While structure with SetStep at the innermost level
			innermost := []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{10}},
			}

			current := innermost
			for i := 0; i < depth; i++ {
				current = []opcode.OpCode{
					{Cmd: opcode.While, Args: []any{true, current}},
				}
			}

			return containsOpSetStep(current) == true
		},
		gen.IntRange(0, 5), // Nesting depth from 0 to 5
	))

	// Property: SetStep in nested For blocks is detected
	properties.Property("SetStep in nested For blocks is detected", prop.ForAll(
		func(depth int) bool {
			// Create a deeply nested For structure with SetStep at the innermost level
			innermost := []opcode.OpCode{
				{Cmd: opcode.SetStep, Args: []any{10}},
			}

			current := innermost
			for i := 0; i < depth; i++ {
				current = []opcode.OpCode{
					{Cmd: opcode.For, Args: []any{[]opcode.OpCode{}, true, []opcode.OpCode{}, current}},
				}
			}

			return containsOpSetStep(current) == true
		},
		gen.IntRange(0, 5), // Nesting depth from 0 to 5
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// TestProperty1_EmptySequence tests that empty sequences return false.
// **Validates: Requirements 1.3**
func TestProperty1_EmptySequence(t *testing.T) {
	// Empty sequence should return false
	result := containsOpSetStep([]opcode.OpCode{})
	if result != false {
		t.Errorf("containsOpSetStep should return false for empty sequence, got %v", result)
	}

	// Nil sequence should return false
	result = containsOpSetStep(nil)
	if result != false {
		t.Errorf("containsOpSetStep should return false for nil sequence, got %v", result)
	}
}
