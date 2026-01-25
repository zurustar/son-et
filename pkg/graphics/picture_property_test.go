package graphics

import (
	"testing"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// TestProperty_PictureIDUniqueness tests that all picture IDs are unique
// **Validates: 要件 1.2**
func TestProperty_PictureIDUniqueness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("任意の2つのピクチャーについて、それらのIDは異なる", prop.ForAll(
		func(count int) bool {
			pm := NewPictureManager("")

			// Create multiple pictures
			ids := make(map[int]bool)
			for i := 0; i < count; i++ {
				id, err := pm.CreatePic(10, 10)
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
		gen.IntRange(1, 100),
	))

	properties.TestingRun(t)
}

// TestProperty_PictureSizeAccuracy tests that PicWidth/PicHeight return accurate sizes
// **Validates: 要件 1.7, 1.8**
func TestProperty_PictureSizeAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("任意のピクチャーについて、PicWidth/PicHeightは実際の画像サイズと一致する", prop.ForAll(
		func(width, height int) bool {
			pm := NewPictureManager("")

			// Create a picture with specified dimensions
			id, err := pm.CreatePic(width, height)
			if err != nil {
				return true // Skip if creation fails (e.g., invalid dimensions)
			}

			// Verify dimensions match
			actualWidth := pm.PicWidth(id)
			actualHeight := pm.PicHeight(id)

			return actualWidth == width && actualHeight == height
		},
		gen.IntRange(1, 1000),
		gen.IntRange(1, 1000),
	))

	properties.TestingRun(t)
}

// TestProperty_DeletedPictureAccess tests that accessing deleted pictures returns errors
// **Validates: 要件 1.9**
func TestProperty_DeletedPictureAccess(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("任意の削除されたピクチャーIDについて、アクセス時にエラーが返される", prop.ForAll(
		func(count int) bool {
			pm := NewPictureManager("")

			// Create pictures
			ids := make([]int, 0, count)
			for i := 0; i < count; i++ {
				id, err := pm.CreatePic(10, 10)
				if err != nil {
					break
				}
				ids = append(ids, id)
			}

			if len(ids) == 0 {
				return true // Skip if no pictures created
			}

			// Delete all pictures
			for _, id := range ids {
				if err := pm.DelPic(id); err != nil {
					return false // Delete should succeed
				}
			}

			// Try to access deleted pictures
			for _, id := range ids {
				_, err := pm.GetPic(id)
				if err == nil {
					return false // Should return error
				}

				// PicWidth and PicHeight should return 0
				if pm.PicWidth(id) != 0 {
					return false
				}
				if pm.PicHeight(id) != 0 {
					return false
				}
			}

			return true
		},
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// TestProperty_ResourceLimit tests that picture count never exceeds 256
// **Validates: 要件 9.5**
func TestProperty_ResourceLimit(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("任意の時点で、ピクチャー数は256以下", prop.ForAll(
		func(operations []bool) bool {
			pm := NewPictureManager("")
			currentCount := 0
			ids := make([]int, 0)

			for _, shouldCreate := range operations {
				if shouldCreate {
					// Try to create a picture
					id, err := pm.CreatePic(10, 10)
					if err == nil {
						ids = append(ids, id)
						currentCount++

						// Check that we never exceed the limit
						if currentCount > 256 {
							return false
						}
					} else {
						// If creation fails, it should be because we hit the limit
						if currentCount < 256 {
							return false // Should not fail before limit
						}
					}
				} else if len(ids) > 0 {
					// Delete a random picture
					idx := len(ids) - 1
					id := ids[idx]
					if err := pm.DelPic(id); err == nil {
						ids = ids[:idx]
						currentCount--
					}
				}

				// Verify count is within limit
				if currentCount > 256 {
					return false
				}
			}

			return true
		},
		gen.SliceOf(gen.Bool()),
	))

	properties.TestingRun(t)
}

// TestProperty_CreatePicFromPreservesSize tests that CreatePicFrom creates a copy with same dimensions
func TestProperty_CreatePicFromPreservesSize(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	properties.Property("CreatePicFromで作成されたピクチャーは元のピクチャーと同じサイズを持つ", prop.ForAll(
		func(width, height int) bool {
			pm := NewPictureManager("")

			// Create source picture
			srcID, err := pm.CreatePic(width, height)
			if err != nil {
				return true // Skip if creation fails
			}

			// Create copy
			copyID, err := pm.CreatePicFrom(srcID)
			if err != nil {
				return false // Copy should succeed
			}

			// Verify dimensions match
			srcWidth := pm.PicWidth(srcID)
			srcHeight := pm.PicHeight(srcID)
			copyWidth := pm.PicWidth(copyID)
			copyHeight := pm.PicHeight(copyID)

			return srcWidth == copyWidth && srcHeight == copyHeight &&
				srcWidth == width && srcHeight == height
		},
		gen.IntRange(1, 500),
		gen.IntRange(1, 500),
	))

	properties.TestingRun(t)
}
