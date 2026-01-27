package graphics

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// Feature: graphics-system, Property 6: キャストIDの一意性
// **Validates: 要件 4.2**
func TestProperty6_CastIDUniqueness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("任意の2つのキャストについて、それらのIDは異なる", prop.ForAll(
		func(count int) bool {
			cm := NewCastManager()

			// Create multiple casts
			ids := make(map[int]bool)
			for i := 0; i < count; i++ {
				id, err := cm.PutCast(0, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					// Resource limit reached, which is acceptable
					break
				}

				// Check if ID is unique
				if ids[id] {
					return false // Duplicate ID found
				}
				ids[id] = true
			}

			return true
		},
		gen.IntRange(1, 200),
	))

	properties.Property("連続して作成されたキャストは連番のIDを持つ", prop.ForAll(
		func(count int) bool {
			if count <= 0 || count > 1024 {
				return true
			}

			cm := NewCastManager()

			// Create casts and verify sequential IDs
			for i := 0; i < count; i++ {
				id, err := cm.PutCast(0, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}

				// ID should be sequential starting from 0
				if id != i {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: graphics-system, Property 7: キャスト位置の更新
// **Validates: 要件 4.3**
func TestProperty7_CastPositionUpdate(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("任意のキャストについて、MoveCast後の位置は指定された値と一致する", prop.ForAll(
		func(initialX, initialY, newX, newY int) bool {
			cm := NewCastManager()

			// Create a cast
			id, err := cm.PutCast(0, 0, initialX, initialY, 0, 0, 32, 32)
			if err != nil {
				return false
			}

			// Move the cast
			err = cm.MoveCast(id, WithCastPosition(newX, newY))
			if err != nil {
				return false
			}

			// Verify position
			cast, err := cm.GetCast(id)
			if err != nil {
				return false
			}

			return cast.X == newX && cast.Y == newY
		},
		gen.IntRange(-1000, 1000),
		gen.IntRange(-1000, 1000),
		gen.IntRange(-1000, 1000),
		gen.IntRange(-1000, 1000),
	))

	properties.Property("MoveCastでソース領域を変更できる", prop.ForAll(
		func(srcX, srcY, width, height int) bool {
			cm := NewCastManager()

			// Create a cast
			id, err := cm.PutCast(0, 0, 0, 0, 0, 0, 32, 32)
			if err != nil {
				return false
			}

			// Move the cast with new source
			err = cm.MoveCast(id, WithCastSource(srcX, srcY, width, height))
			if err != nil {
				return false
			}

			// Verify source
			cast, err := cm.GetCast(id)
			if err != nil {
				return false
			}

			return cast.SrcX == srcX && cast.SrcY == srcY &&
				cast.Width == width && cast.Height == height
		},
		gen.IntRange(0, 500),
		gen.IntRange(0, 500),
		gen.IntRange(1, 200),
		gen.IntRange(1, 200),
	))

	properties.Property("MoveCastでピクチャーIDを変更できる", prop.ForAll(
		func(initialPicID, newPicID int) bool {
			cm := NewCastManager()

			// Create a cast
			id, err := cm.PutCast(0, initialPicID, 0, 0, 0, 0, 32, 32)
			if err != nil {
				return false
			}

			// Move the cast with new picture ID
			err = cm.MoveCast(id, WithCastPicID(newPicID))
			if err != nil {
				return false
			}

			// Verify picture ID
			cast, err := cm.GetCast(id)
			if err != nil {
				return false
			}

			return cast.PicID == newPicID
		},
		gen.IntRange(0, 255),
		gen.IntRange(0, 255),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: graphics-system, Property 5: ウィンドウ削除時のキャスト削除
// **Validates: 要件 9.2**
func TestProperty5_WindowDeletionCastCascade(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("DeleteCastsByWindow後にそのウィンドウに属するキャストは存在しない", prop.ForAll(
		func(winID, castCount int) bool {
			if castCount <= 0 || castCount > 100 {
				return true
			}

			cm := NewCastManager()

			// Create casts for the window
			castIDs := make([]int, 0, castCount)
			for i := 0; i < castCount; i++ {
				id, err := cm.PutCast(winID, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
				castIDs = append(castIDs, id)
			}

			// Verify casts exist
			casts := cm.GetCastsByWindow(winID)
			if len(casts) != castCount {
				return false
			}

			// Delete casts by window
			cm.DeleteCastsByWindow(winID)

			// Verify no casts remain for this window
			casts = cm.GetCastsByWindow(winID)
			if len(casts) != 0 {
				return false
			}

			// Verify casts are not accessible
			for _, id := range castIDs {
				_, err := cm.GetCast(id)
				if err == nil {
					return false // Should return error
				}
			}

			return true
		},
		gen.IntRange(0, 63),
		gen.IntRange(1, 50),
	))

	properties.Property("DeleteCastsByWindowは他のウィンドウのキャストに影響しない", prop.ForAll(
		func(winID1, winID2, castCount int) bool {
			if winID1 == winID2 || castCount <= 0 || castCount > 50 {
				return true
			}

			cm := NewCastManager()

			// Create casts for window 1
			for i := 0; i < castCount; i++ {
				_, err := cm.PutCast(winID1, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
			}

			// Create casts for window 2
			win2CastIDs := make([]int, 0, castCount)
			for i := 0; i < castCount; i++ {
				id, err := cm.PutCast(winID2, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
				win2CastIDs = append(win2CastIDs, id)
			}

			// Delete casts for window 1
			cm.DeleteCastsByWindow(winID1)

			// Verify window 2 casts still exist
			casts := cm.GetCastsByWindow(winID2)
			if len(casts) != castCount {
				return false
			}

			// Verify window 2 casts are accessible
			for _, id := range win2CastIDs {
				_, err := cm.GetCast(id)
				if err != nil {
					return false
				}
			}

			return true
		},
		gen.IntRange(0, 31),
		gen.IntRange(32, 63),
		gen.IntRange(1, 30),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// Feature: graphics-system, Property 10: リソース制限
// **Validates: 要件 9.7**
func TestProperty10_CastResourceLimit(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("キャスト数は1024を超えない", prop.ForAll(
		func(attemptCount int) bool {
			if attemptCount <= 1024 || attemptCount > 1100 {
				return true
			}

			cm := NewCastManager()

			// Try to create more than 1024 casts
			successCount := 0
			for i := 0; i < attemptCount; i++ {
				_, err := cm.PutCast(0, i%256, i*10, i*10, 0, 0, 32, 32)
				if err == nil {
					successCount++
				}
			}

			// Should only succeed for first 1024 casts
			return successCount == 1024
		},
		gen.IntRange(1025, 1100),
	))

	properties.Property("キャスト数は任意の時点で1024以下", prop.ForAll(
		func(operations []bool) bool {
			if len(operations) == 0 || len(operations) > 2000 {
				return true
			}

			cm := NewCastManager()
			openCasts := make([]int, 0)

			for _, shouldCreate := range operations {
				if shouldCreate {
					// Try to create a cast
					id, err := cm.PutCast(0, len(openCasts)%256, 0, 0, 0, 0, 32, 32)
					if err == nil {
						openCasts = append(openCasts, id)
					}
				} else if len(openCasts) > 0 {
					// Delete a cast
					idx := len(openCasts) - 1
					_ = cm.DelCast(openCasts[idx])
					openCasts = openCasts[:idx]
				}

				// Verify cast count never exceeds 1024
				if cm.Count() > 1024 {
					return false
				}
			}

			return true
		},
		gen.SliceOfN(1000, gen.Bool()),
	))

	properties.Property("1024キャスト作成後、追加作成は失敗する", prop.ForAll(
		func(extraAttempts int) bool {
			if extraAttempts <= 0 || extraAttempts > 20 {
				return true
			}

			cm := NewCastManager()

			// Create exactly 1024 casts
			for i := 0; i < 1024; i++ {
				_, err := cm.PutCast(0, i%256, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
			}

			// All subsequent attempts should fail
			for i := 0; i < extraAttempts; i++ {
				_, err := cm.PutCast(0, i%256, 0, 0, 0, 0, 32, 32)
				if err == nil {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 20),
	))

	properties.Property("キャスト削除後は新しいキャストを作成できる", prop.ForAll(
		func(deleteCount int) bool {
			if deleteCount <= 0 || deleteCount > 100 {
				return true
			}

			cm := NewCastManager()

			// Create 1024 casts
			castIDs := make([]int, 1024)
			for i := 0; i < 1024; i++ {
				id, err := cm.PutCast(0, i%256, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
				castIDs[i] = id
			}

			// Delete some casts
			for i := 0; i < deleteCount; i++ {
				err := cm.DelCast(castIDs[i])
				if err != nil {
					return false
				}
			}

			// Should be able to create the same number of new casts
			for i := 0; i < deleteCount; i++ {
				_, err := cm.PutCast(0, i%256, 0, 0, 0, 0, 32, 32)
				if err != nil {
					return false
				}
			}

			// Total cast count should be 1024
			return cm.Count() == 1024
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}

// キャストZ順序のプロパティテスト
// **Validates: 要件 4.9**
func TestProperty_CastZOrder(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("後から配置したキャストのZOrderは大きい", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 1 || castCount > 100 {
				return true
			}

			cm := NewCastManager()

			// Create casts and track their ZOrder
			zOrders := make([]int, castCount)
			for i := 0; i < castCount; i++ {
				id, err := cm.PutCast(0, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}

				cast, err := cm.GetCast(id)
				if err != nil {
					return false
				}
				zOrders[i] = cast.ZOrder
			}

			// Verify ZOrder is strictly increasing
			for i := 0; i < len(zOrders)-1; i++ {
				if zOrders[i] >= zOrders[i+1] {
					return false
				}
			}

			return true
		},
		gen.IntRange(2, 100),
	))

	properties.Property("GetCastsByWindowはZ順序でソートされたキャストを返す", prop.ForAll(
		func(castCount int) bool {
			if castCount <= 0 || castCount > 100 {
				return true
			}

			cm := NewCastManager()

			// Create casts
			for i := 0; i < castCount; i++ {
				_, err := cm.PutCast(0, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
			}

			// Get casts by window
			casts := cm.GetCastsByWindow(0)

			if len(casts) != castCount {
				return false
			}

			// Verify sorting
			for i := 0; i < len(casts)-1; i++ {
				if casts[i].ZOrder >= casts[i+1].ZOrder {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t, gopter.ConsoleReporter(false))
}
