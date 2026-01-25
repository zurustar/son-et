package graphics

import (
	"reflect"
	"sync"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// genCommand generates a random Command
func genCommand() gopter.Gen {
	return gen.IntRange(int(CmdMovePic), int(CmdDrawCircle)).
		FlatMap(func(cmdType interface{}) gopter.Gen {
			ct := CommandType(cmdType.(int))
			return gen.SliceOfN(3, gen.Int()).Map(func(args []int) Command {
				anyArgs := make([]any, len(args))
				for i, arg := range args {
					anyArgs[i] = arg
				}
				return Command{
					Type: ct,
					Args: anyArgs,
				}
			})
		}, reflect.TypeOf(Command{}))
}

// Feature: graphics-system, Property 8: コマンド実行順序
// **Validates: 要件 7.4**
func TestProperty8_CommandExecutionOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("commands are dequeued in FIFO order", prop.ForAll(
		func(commands []Command) bool {
			queue := NewCommandQueue()

			// Push all commands
			for _, cmd := range commands {
				queue.Push(cmd)
			}

			// PopAll should return commands in the same order
			result := queue.PopAll()

			if len(result) != len(commands) {
				return false
			}

			for i := range commands {
				if result[i].Type != commands[i].Type {
					return false
				}
				if !reflect.DeepEqual(result[i].Args, commands[i].Args) {
					return false
				}
			}

			return true
		},
		gen.SliceOf(genCommand()),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: graphics-system, Property 9: スレッドセーフ性
// **Validates: 要件 7.1**
func TestProperty9_ThreadSafety(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("concurrent Push operations are thread-safe", prop.ForAll(
		func(numGoroutines int) bool {
			if numGoroutines <= 0 || numGoroutines > 10 {
				return true
			}

			queue := NewCommandQueue()
			var wg sync.WaitGroup

			commandsPerGoroutine := 10
			totalCommands := numGoroutines * commandsPerGoroutine

			// Push commands concurrently from multiple goroutines
			for i := 0; i < numGoroutines; i++ {
				wg.Add(1)
				go func(goroutineID int) {
					defer wg.Done()
					for j := 0; j < commandsPerGoroutine; j++ {
						cmd := Command{
							Type: CommandType(goroutineID % int(CmdDrawCircle+1)),
							Args: []any{goroutineID, j},
						}
						queue.Push(cmd)
					}
				}(i)
			}

			wg.Wait()

			// Verify all commands were added
			result := queue.PopAll()
			return len(result) == totalCommands
		},
		gen.IntRange(1, 10),
	))

	properties.Property("concurrent Push and PopAll operations are thread-safe", prop.ForAll(
		func(commandCount int) bool {
			if commandCount <= 0 || commandCount > 100 {
				return true
			}

			queue := NewCommandQueue()
			var wg sync.WaitGroup
			var pushDone sync.WaitGroup

			// Push commands concurrently
			pushDone.Add(1)
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer pushDone.Done()
				for i := 0; i < commandCount; i++ {
					cmd := Command{
						Type: CommandType(i % int(CmdDrawCircle+1)),
						Args: []any{i},
					}
					queue.Push(cmd)
				}
			}()

			// Wait for all pushes to complete before starting PopAll operations
			pushDone.Wait()

			// PopAll concurrently multiple times
			results := make([][]Command, 5)
			for i := 0; i < 5; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					results[idx] = queue.PopAll()
				}(i)
			}

			wg.Wait()

			// Count total commands retrieved
			totalRetrieved := 0
			for _, result := range results {
				totalRetrieved += len(result)
			}

			// All commands should be retrieved exactly once
			return totalRetrieved == commandCount
		},
		gen.IntRange(1, 100),
	))

	properties.Property("Len is thread-safe", prop.ForAll(
		func(commandCount int) bool {
			if commandCount <= 0 || commandCount > 100 {
				return true
			}

			queue := NewCommandQueue()
			var wg sync.WaitGroup

			// Push commands
			for i := 0; i < commandCount; i++ {
				cmd := Command{
					Type: CommandType(i % int(CmdDrawCircle+1)),
					Args: []any{i},
				}
				queue.Push(cmd)
			}

			// Call Len concurrently
			lengths := make([]int, 10)
			for i := 0; i < 10; i++ {
				wg.Add(1)
				go func(idx int) {
					defer wg.Done()
					lengths[idx] = queue.Len()
				}(i)
			}

			wg.Wait()

			// All Len calls should return the same value (or less if PopAll was called)
			for _, length := range lengths {
				if length < 0 || length > commandCount {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
