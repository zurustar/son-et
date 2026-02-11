package vm

import (
	"os"
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// opType represents an operation on FileHandleTable.
type opType int

const (
	opOpen  opType = 0
	opClose opType = 1
)

// fhtOp represents a single operation in a sequence of Open/Close operations.
type fhtOp struct {
	op opType
	// For Close operations, this is the index into the list of currently open handles
	// to select which handle to close. Used as closeHandleIdx % len(openHandles).
	closeHandleIdx int
}

// genFHTOps generates a random sequence of Open/Close operations.
// The sequence always starts with at least one Open to ensure there's something to close.
func genFHTOps() gopter.Gen {
	return gen.SliceOfN(20, gen.IntRange(0, 1)).FlatMap(func(v interface{}) gopter.Gen {
		rawOps := v.([]int)
		return gen.SliceOfN(len(rawOps), gen.IntRange(0, 100)).Map(func(indices []int) []fhtOp {
			ops := make([]fhtOp, len(rawOps))
			for i, raw := range rawOps {
				ops[i] = fhtOp{
					op:             opType(raw),
					closeHandleIdx: indices[i],
				}
			}
			return ops
		})
	}, nil)
}

// Feature: required-builtin-functions, Property 3: FileHandleTableのハンドル割り当て不変条件
// 任意のOpen/Close操作のシーケンスに対して、FileHandleTableが返すハンドルは常に1以上の
// 正の整数であり、同時に開いているハンドルは全て一意である。また、Closeされたハンドルは
// 後続のOpen呼び出しで再利用される。
// **Validates: Requirements 3.2, 3.3**
func TestProperty3_FileHandleTableAllocationInvariants(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// Property: All handles returned by Open are positive integers (>= 1)
	// and all simultaneously open handles are unique.
	properties.Property("all handles are positive and simultaneously open handles are unique", prop.ForAll(
		func(ops []fhtOp) bool {
			fht := NewFileHandleTable()
			var tempFiles []*os.File
			defer func() {
				fht.CloseAll()
				for _, f := range tempFiles {
					f.Close()
					os.Remove(f.Name())
				}
			}()

			openHandles := make([]int, 0) // currently open handles

			for _, op := range ops {
				switch op.op {
				case opOpen:
					f, err := os.CreateTemp("", "fht-prop-*")
					if err != nil {
						// Skip if temp file creation fails
						continue
					}
					tempFiles = append(tempFiles, f)

					handle := fht.Open(f)

					// Invariant 1: handle must be >= 1
					if handle < minHandleID {
						return false
					}

					// Invariant 2: handle must be unique among currently open handles
					for _, h := range openHandles {
						if h == handle {
							return false
						}
					}

					openHandles = append(openHandles, handle)

				case opClose:
					if len(openHandles) == 0 {
						continue
					}
					idx := op.closeHandleIdx % len(openHandles)
					handle := openHandles[idx]

					err := fht.Close(handle)
					if err != nil {
						return false
					}

					// Remove from open handles
					openHandles = append(openHandles[:idx], openHandles[idx+1:]...)
				}
			}

			return true
		},
		genFHTOps(),
	))

	// Property: Closed handles are reused by subsequent Open calls.
	// When a handle is closed and a new Open is called, the returned handle
	// should be the smallest unused handle, which means closed handles get reused.
	properties.Property("closed handles are reused by subsequent Open calls", prop.ForAll(
		func(numInitialOpens int) bool {
			fht := NewFileHandleTable()
			var tempFiles []*os.File
			defer func() {
				fht.CloseAll()
				for _, f := range tempFiles {
					f.Close()
					os.Remove(f.Name())
				}
			}()

			// Open numInitialOpens files to get handles 1..numInitialOpens
			handles := make([]int, numInitialOpens)
			for i := 0; i < numInitialOpens; i++ {
				f, err := os.CreateTemp("", "fht-prop-reuse-*")
				if err != nil {
					return true // skip on temp file error
				}
				tempFiles = append(tempFiles, f)
				handles[i] = fht.Open(f)
			}

			// Close the first handle (handle 1, the smallest)
			if numInitialOpens == 0 {
				return true
			}
			closedHandle := handles[0]
			err := fht.Close(closedHandle)
			if err != nil {
				return false
			}

			// Open a new file — it should reuse the closed handle
			f, err := os.CreateTemp("", "fht-prop-reuse-*")
			if err != nil {
				return true // skip on temp file error
			}
			tempFiles = append(tempFiles, f)
			newHandle := fht.Open(f)

			// The new handle should equal the closed handle (smallest unused)
			if newHandle != closedHandle {
				return false
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	// Property: Closing a middle handle and reopening reuses that handle.
	// For any sequence of N opens, closing handle at position K, the next Open
	// returns that closed handle (since it becomes the smallest unused).
	properties.Property("closing middle handle and reopening reuses it when it is smallest unused", prop.ForAll(
		func(numOpens int, closeIdx int) bool {
			fht := NewFileHandleTable()
			var tempFiles []*os.File
			defer func() {
				fht.CloseAll()
				for _, f := range tempFiles {
					f.Close()
					os.Remove(f.Name())
				}
			}()

			// Open numOpens files
			handles := make([]int, numOpens)
			for i := 0; i < numOpens; i++ {
				f, err := os.CreateTemp("", "fht-prop-mid-*")
				if err != nil {
					return true
				}
				tempFiles = append(tempFiles, f)
				handles[i] = fht.Open(f)
			}

			// Close all handles from index 0 to closeIdx (inclusive)
			// so that the smallest closed handle is handles[0]
			actualCloseIdx := closeIdx % numOpens
			for i := 0; i <= actualCloseIdx; i++ {
				err := fht.Close(handles[i])
				if err != nil {
					return false
				}
			}

			// The next Open should return the smallest closed handle
			f, err := os.CreateTemp("", "fht-prop-mid-*")
			if err != nil {
				return true
			}
			tempFiles = append(tempFiles, f)
			newHandle := fht.Open(f)

			// Find the smallest closed handle
			smallestClosed := handles[0]
			for i := 1; i <= actualCloseIdx; i++ {
				if handles[i] < smallestClosed {
					smallestClosed = handles[i]
				}
			}

			return newHandle == smallestClosed
		},
		gen.IntRange(1, 15),
		gen.IntRange(0, 14),
	))

	// Property: After a random sequence of Open/Close, the set of open handles
	// is always a valid set (positive, unique) and the next Open returns the
	// smallest positive integer not in the set.
	properties.Property("next Open always returns smallest unused positive integer", prop.ForAll(
		func(ops []fhtOp) bool {
			fht := NewFileHandleTable()
			var tempFiles []*os.File
			defer func() {
				fht.CloseAll()
				for _, f := range tempFiles {
					f.Close()
					os.Remove(f.Name())
				}
			}()

			openHandleSet := make(map[int]bool) // track currently open handles

			for _, op := range ops {
				switch op.op {
				case opOpen:
					f, err := os.CreateTemp("", "fht-prop-smallest-*")
					if err != nil {
						continue
					}
					tempFiles = append(tempFiles, f)

					// Compute expected smallest unused handle
					expected := minHandleID
					for openHandleSet[expected] {
						expected++
					}

					handle := fht.Open(f)

					// Verify it matches expected
					if handle != expected {
						return false
					}

					openHandleSet[handle] = true

				case opClose:
					if len(openHandleSet) == 0 {
						continue
					}
					// Pick a handle to close
					var handleList []int
					for h := range openHandleSet {
						handleList = append(handleList, h)
					}
					idx := op.closeHandleIdx % len(handleList)
					handle := handleList[idx]

					err := fht.Close(handle)
					if err != nil {
						return false
					}
					delete(openHandleSet, handle)
				}
			}

			return true
		},
		genFHTOps(),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
