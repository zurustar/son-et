package vm

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/zurustar/son-et/pkg/opcode"
)

// genHandlerCountAndTarget generates a handler count between 1 and 10 for registering
// multiple handlers, and a target index within that range.
func genHandlerCountAndTarget() gopter.Gen {
	return gen.IntRange(1, 10).FlatMap(func(v interface{}) gopter.Gen {
		count := v.(int)
		return gen.IntRange(1, count).Map(func(target int) [2]int {
			return [2]int{count, target}
		})
	}, nil)
}

// Feature: required-builtin-functions, Property 1: FreezeMesはハンドラを無効化する
// 任意の登録済みEventHandlerに対して、FreezeMes(mes_no)を呼び出した後、
// そのハンドラのActiveフラグはfalseであり、イベントディスパッチ時にそのハンドラの
// 実行はスキップされる。
// **Validates: Requirements 1.1, 1.2**
func TestProperty1_FreezeMesDeactivatesHandler(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After FreezeMes(n), handler n has Active == false
	properties.Property("FreezeMes sets Active to false for any registered handler", prop.ForAll(
		func(countAndTarget [2]int, eventType EventType) bool {
			count := countAndTarget[0]
			target := countAndTarget[1]

			vm := New([]opcode.OpCode{})

			// Register 'count' handlers
			handlers := make([]*EventHandler, count)
			for i := 0; i < count; i++ {
				handlers[i] = NewEventHandler("", eventType, []opcode.OpCode{}, vm, nil)
				vm.handlerRegistry.Register(handlers[i])
			}

			// All handlers should be active initially
			for i, h := range handlers {
				if !h.Active {
					t.Logf("handler %d not active initially", i+1)
					return false
				}
			}

			// Freeze the target handler
			_, err := vm.builtins["FreezeMes"](vm, []any{int64(target)})
			if err != nil {
				return false
			}

			// The target handler must be inactive
			if handlers[target-1].Active {
				return false
			}

			// All other handlers must remain active
			for i, h := range handlers {
				if i == target-1 {
					continue
				}
				if !h.Active {
					t.Logf("non-target handler %d became inactive", i+1)
					return false
				}
			}

			return true
		},
		genHandlerCountAndTarget(),
		genEventType(),
	))

	// Property: Frozen handler is skipped by ShouldExecute check (simulates dispatcher behavior)
	// Requirement 1.2: EventDispatcher skips handlers with Active=false
	properties.Property("Frozen handler would be skipped by event dispatch", prop.ForAll(
		func(eventType EventType) bool {
			vm := New([]opcode.OpCode{})
			handler := NewEventHandler("", eventType, []opcode.OpCode{}, vm, nil)
			vm.handlerRegistry.Register(handler)

			// Freeze the handler
			_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
			if err != nil {
				return false
			}

			// Active == false means the dispatcher will skip this handler
			return !handler.Active
		},
		genEventType(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: required-builtin-functions, Property 2: Freeze/Activateラウンドトリップはハンドラ状態を保持する
// 任意の登録済みEventHandlerと任意のCurrentPC値およびWaitCounter値に対して、
// FreezeMesで一時停止してからActivateMesで再開した場合、ハンドラのActiveフラグはtrue、
// CurrentPCとWaitCounterは停止前の値と同一である。
// **Validates: Requirements 1.3, 2.1, 2.2**
func TestProperty2_FreezeActivateRoundTripPreservesState(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: After Freeze then Activate, Active is true and CurrentPC/WaitCounter are preserved
	properties.Property("Freeze/Activate round-trip preserves handler state", prop.ForAll(
		func(eventType EventType, currentPC int, waitCounter int) bool {
			vm := New([]opcode.OpCode{})
			handler := NewEventHandler("", eventType, []opcode.OpCode{}, vm, nil)
			vm.handlerRegistry.Register(handler)

			// Set arbitrary state
			handler.CurrentPC = currentPC
			handler.WaitCounter = waitCounter

			// Freeze
			_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
			if err != nil {
				return false
			}

			// Verify frozen state: Active must be false, but PC/WaitCounter preserved
			if handler.Active {
				return false
			}
			if handler.CurrentPC != currentPC {
				return false
			}
			if handler.WaitCounter != waitCounter {
				return false
			}

			// Activate
			_, err = vm.builtins["ActivateMes"](vm, []any{int64(1)})
			if err != nil {
				return false
			}

			// Verify reactivated state
			if !handler.Active {
				return false
			}
			if handler.CurrentPC != currentPC {
				return false
			}
			if handler.WaitCounter != waitCounter {
				return false
			}

			return true
		},
		genEventType(),
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
	))

	// Property: Multiple Freeze/Activate cycles preserve state
	properties.Property("Multiple Freeze/Activate cycles preserve CurrentPC and WaitCounter", prop.ForAll(
		func(eventType EventType, currentPC int, waitCounter int, cycles int) bool {
			vm := New([]opcode.OpCode{})
			handler := NewEventHandler("", eventType, []opcode.OpCode{}, vm, nil)
			vm.handlerRegistry.Register(handler)

			// Set arbitrary state
			handler.CurrentPC = currentPC
			handler.WaitCounter = waitCounter

			for i := 0; i < cycles; i++ {
				// Freeze
				_, err := vm.builtins["FreezeMes"](vm, []any{int64(1)})
				if err != nil {
					return false
				}
				if handler.Active {
					return false
				}

				// Activate
				_, err = vm.builtins["ActivateMes"](vm, []any{int64(1)})
				if err != nil {
					return false
				}
				if !handler.Active {
					return false
				}
			}

			// After all cycles, state must be preserved
			if handler.CurrentPC != currentPC {
				return false
			}
			if handler.WaitCounter != waitCounter {
				return false
			}

			return true
		},
		genEventType(),
		gen.IntRange(0, 1000),
		gen.IntRange(0, 1000),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
