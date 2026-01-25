package graphics

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: graphics-system, Property 4: ウィンドウZ順序
// **Validates: 要件 3.11**
func TestProperty4_WindowZOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("windows opened later have higher ZOrder", prop.ForAll(
		func(windowCount int) bool {
			if windowCount <= 0 || windowCount > 64 {
				return true
			}

			wm := NewWindowManager()

			// Open multiple windows
			windowIDs := make([]int, windowCount)
			for i := 0; i < windowCount; i++ {
				id, err := wm.OpenWin(i)
				if err != nil {
					return false
				}
				windowIDs[i] = id
			}

			// Get windows in Z order
			windows := wm.GetWindowsOrdered()

			// Verify windows are in ascending Z order
			for i := 0; i < len(windows)-1; i++ {
				if windows[i].ZOrder >= windows[i+1].ZOrder {
					return false
				}
			}

			// Verify windows are in creation order
			for i, win := range windows {
				if win.ID != windowIDs[i] {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 64),
	))

	properties.Property("later opened windows have strictly increasing ZOrder", prop.ForAll(
		func(windowCount int) bool {
			if windowCount <= 1 || windowCount > 64 {
				return true
			}

			wm := NewWindowManager()

			// Open windows and track their ZOrder
			zOrders := make([]int, windowCount)
			for i := 0; i < windowCount; i++ {
				id, err := wm.OpenWin(i)
				if err != nil {
					return false
				}

				win, err := wm.GetWin(id)
				if err != nil {
					return false
				}
				zOrders[i] = win.ZOrder
			}

			// Verify ZOrder is strictly increasing
			for i := 0; i < len(zOrders)-1; i++ {
				if zOrders[i] >= zOrders[i+1] {
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 64),
	))

	properties.Property("GetWindowsOrdered returns windows sorted by ZOrder", prop.ForAll(
		func(windowCount int) bool {
			if windowCount <= 0 || windowCount > 64 {
				return true
			}

			wm := NewWindowManager()

			// Open windows
			for i := 0; i < windowCount; i++ {
				_, err := wm.OpenWin(i)
				if err != nil {
					return false
				}
			}

			// Get windows in Z order
			windows := wm.GetWindowsOrdered()

			if len(windows) != windowCount {
				return false
			}

			// Verify sorting
			for i := 0; i < len(windows)-1; i++ {
				if windows[i].ZOrder >= windows[i+1].ZOrder {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: graphics-system, Property 10: リソース制限
// **Validates: 要件 9.6**
func TestProperty10_ResourceLimit(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("cannot exceed maximum window count (64)", prop.ForAll(
		func(attemptCount int) bool {
			if attemptCount <= 64 || attemptCount > 100 {
				return true
			}

			wm := NewWindowManager()

			// Try to open more than 64 windows
			successCount := 0
			for i := 0; i < attemptCount; i++ {
				_, err := wm.OpenWin(i)
				if err == nil {
					successCount++
				}
			}

			// Should only succeed for first 64 windows
			return successCount == 64
		},
		gen.IntRange(65, 100),
	))

	properties.Property("window count never exceeds 64", prop.ForAll(
		func(operations []bool) bool {
			if len(operations) == 0 || len(operations) > 200 {
				return true
			}

			wm := NewWindowManager()
			openWindows := make([]int, 0)

			for _, shouldOpen := range operations {
				if shouldOpen {
					// Try to open a window
					id, err := wm.OpenWin(len(openWindows))
					if err == nil {
						openWindows = append(openWindows, id)
					}
				} else if len(openWindows) > 0 {
					// Close a random window
					idx := len(openWindows) - 1
					_ = wm.CloseWin(openWindows[idx])
					openWindows = openWindows[:idx]
				}

				// Verify window count never exceeds 64
				windows := wm.GetWindowsOrdered()
				if len(windows) > 64 {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(100, gen.Bool()),
	))

	properties.Property("after opening 64 windows, all subsequent opens fail", prop.ForAll(
		func(extraAttempts int) bool {
			if extraAttempts <= 0 || extraAttempts > 20 {
				return true
			}

			wm := NewWindowManager()

			// Open exactly 64 windows
			for i := 0; i < 64; i++ {
				_, err := wm.OpenWin(i)
				if err != nil {
					return false
				}
			}

			// All subsequent attempts should fail
			for i := 0; i < extraAttempts; i++ {
				_, err := wm.OpenWin(64 + i)
				if err == nil {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	properties.Property("closing windows allows opening new ones", prop.ForAll(
		func(closeCount int) bool {
			if closeCount <= 0 || closeCount > 64 {
				return true
			}

			wm := NewWindowManager()

			// Open 64 windows
			windowIDs := make([]int, 64)
			for i := 0; i < 64; i++ {
				id, err := wm.OpenWin(i)
				if err != nil {
					return false
				}
				windowIDs[i] = id
			}

			// Close some windows
			for i := 0; i < closeCount; i++ {
				err := wm.CloseWin(windowIDs[i])
				if err != nil {
					return false
				}
			}

			// Should be able to open the same number of new windows
			for i := 0; i < closeCount; i++ {
				_, err := wm.OpenWin(100 + i)
				if err != nil {
					return false
				}
			}

			// Total window count should be 64
			windows := wm.GetWindowsOrdered()
			return len(windows) == 64
		},
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: graphics-system, Property 5: ウィンドウ削除時のキャスト削除
// **Validates: 要件 9.2**
// Note: This property test will be fully implemented when CastManager is available
func TestProperty5_WindowDeletionCascade(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("closing a window removes it from the manager", prop.ForAll(
		func(windowCount int) bool {
			if windowCount <= 0 || windowCount > 64 {
				return true
			}

			wm := NewWindowManager()

			// Open windows
			windowIDs := make([]int, windowCount)
			for i := 0; i < windowCount; i++ {
				id, err := wm.OpenWin(i)
				if err != nil {
					return false
				}
				windowIDs[i] = id
			}

			// Close all windows one by one
			for _, id := range windowIDs {
				err := wm.CloseWin(id)
				if err != nil {
					return false
				}

				// Verify window is gone
				_, err = wm.GetWin(id)
				if err == nil {
					return false
				}
			}

			// Verify all windows are gone
			windows := wm.GetWindowsOrdered()
			return len(windows) == 0
		},
		gen.IntRange(1, 64),
	))

	properties.Property("CloseWinAll removes all windows", prop.ForAll(
		func(windowCount int) bool {
			if windowCount <= 0 || windowCount > 64 {
				return true
			}

			wm := NewWindowManager()

			// Open windows
			for i := 0; i < windowCount; i++ {
				_, err := wm.OpenWin(i)
				if err != nil {
					return false
				}
			}

			// Close all windows
			wm.CloseWinAll()

			// Verify all windows are gone
			windows := wm.GetWindowsOrdered()
			return len(windows) == 0
		},
		gen.IntRange(1, 64),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
