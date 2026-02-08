package sprite

import (
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
)

// プロパティ9: ピクチャースプライトの作成
// 任意のLoadPic呼び出しについて、対応するPictureSpriteが作成される
// **Validates: Requirements 13.1**
//
// 要件 13.1: WHEN LoadPicが呼び出されたとき THEN THE System SHALL 非表示のPictureSpriteを作成する
//
// このプロパティは以下を検証する:
// 1. LoadPic（CreatePictureSpriteOnLoad）が呼び出されると、PictureSpriteが作成される
// 2. 作成されたPictureSpriteは非表示（Visible = false）である
// 3. 作成されたPictureSpriteは「未関連付け」（Unattached）状態である
// 4. 作成されたPictureSpriteはピクチャ番号で取得できる
func TestProperty_PictureSpriteCreationOnLoad(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// テスト用の画像サイズの定数
	const (
		minImageSize = 1
		maxImageSize = 1000
		minPicID     = 0
		maxPicID     = 255
	)

	properties.Property("任意のLoadPic呼び出しについて、対応するPictureSpriteが作成される", prop.ForAll(
		func(picID int, width, height int) bool {
			sm := NewSpriteManager()
			psm := NewPictureSpriteManager(sm)

			// テスト用の画像を作成
			srcImg := ebiten.NewImage(width, height)

			// LoadPic時にPictureSpriteを作成（CreatePictureSpriteOnLoadを呼び出す）
			ps := psm.CreatePictureSpriteOnLoad(srcImg, picID, width, height)

			// 1. PictureSpriteが作成されることを確認
			if ps == nil {
				return false
			}

			// 2. 作成されたPictureSpriteは非表示であることを確認
			if ps.GetSprite().Visible() {
				return false
			}

			// 3. 作成されたPictureSpriteは「未関連付け」状態であることを確認
			if ps.GetState() != PictureSpriteUnattached {
				return false
			}

			// 4. ピクチャ番号で取得できることを確認
			retrievedPS := psm.GetPictureSpriteByPictureID(picID)
			if retrievedPS != ps {
				return false
			}

			// 5. IsEffectivelyVisibleがfalseであることを確認（未関連付け状態では描画しない）
			if ps.IsEffectivelyVisible() {
				return false
			}

			// 6. ウインドウIDが-1（未関連付け）であることを確認
			if ps.GetWindowID() != -1 {
				return false
			}

			return true
		},
		gen.IntRange(minPicID, maxPicID),
		gen.IntRange(minImageSize, maxImageSize),
		gen.IntRange(minImageSize, maxImageSize),
	))

	properties.TestingRun(t)
}

// プロパティ9の追加テスト: 複数のLoadPic呼び出しでそれぞれPictureSpriteが作成される
// **Validates: Requirements 13.1**
func TestProperty_MultiplePictureSpriteCreationOnLoad(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// テスト用の定数
	const (
		minImageSize = 10
		maxImageSize = 100
		minCount     = 1
		maxCount     = 20
	)

	properties.Property("複数のLoadPic呼び出しについて、それぞれ対応するPictureSpriteが作成される", prop.ForAll(
		func(count int) bool {
			sm := NewSpriteManager()
			psm := NewPictureSpriteManager(sm)

			createdSprites := make(map[int]*PictureSprite)

			// 複数のピクチャをロード
			for i := 0; i < count; i++ {
				picID := i
				srcImg := ebiten.NewImage(minImageSize, minImageSize)
				ps := psm.CreatePictureSpriteOnLoad(srcImg, picID, minImageSize, minImageSize)

				if ps == nil {
					return false
				}

				createdSprites[picID] = ps
			}

			// すべてのPictureSpriteが正しく作成されていることを確認
			for picID, expectedPS := range createdSprites {
				retrievedPS := psm.GetPictureSpriteByPictureID(picID)
				if retrievedPS != expectedPS {
					return false
				}

				// 非表示であることを確認
				if retrievedPS.GetSprite().Visible() {
					return false
				}

				// 未関連付け状態であることを確認
				if retrievedPS.GetState() != PictureSpriteUnattached {
					return false
				}
			}

			return true
		},
		gen.IntRange(minCount, maxCount),
	))

	properties.TestingRun(t)
}

// プロパティ9の追加テスト: 同じピクチャIDで再度LoadPicを呼び出すと、既存のPictureSpriteが置き換えられる
// **Validates: Requirements 13.1**
func TestProperty_PictureSpriteReplacementOnLoad(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// テスト用の定数
	const (
		minImageSize = 10
		maxImageSize = 100
		minPicID     = 0
		maxPicID     = 50
	)

	properties.Property("同じピクチャIDで再度LoadPicを呼び出すと、既存のPictureSpriteが置き換えられる", prop.ForAll(
		func(picID int, width1, height1, width2, height2 int) bool {
			sm := NewSpriteManager()
			psm := NewPictureSpriteManager(sm)

			// 最初のLoadPic
			srcImg1 := ebiten.NewImage(width1, height1)
			ps1 := psm.CreatePictureSpriteOnLoad(srcImg1, picID, width1, height1)

			if ps1 == nil {
				return false
			}

			// 同じピクチャIDで再度LoadPic
			srcImg2 := ebiten.NewImage(width2, height2)
			ps2 := psm.CreatePictureSpriteOnLoad(srcImg2, picID, width2, height2)

			if ps2 == nil {
				return false
			}

			// 新しいPictureSpriteが取得できることを確認
			retrievedPS := psm.GetPictureSpriteByPictureID(picID)
			if retrievedPS != ps2 {
				return false
			}

			// 新しいPictureSpriteのサイズが正しいことを確認
			if ps2.GetWidth() != width2 || ps2.GetHeight() != height2 {
				return false
			}

			// 新しいPictureSpriteも非表示・未関連付け状態であることを確認
			if ps2.GetSprite().Visible() {
				return false
			}
			if ps2.GetState() != PictureSpriteUnattached {
				return false
			}

			return true
		},
		gen.IntRange(minPicID, maxPicID),
		gen.IntRange(minImageSize, maxImageSize),
		gen.IntRange(minImageSize, maxImageSize),
		gen.IntRange(minImageSize, maxImageSize),
		gen.IntRange(minImageSize, maxImageSize),
	))

	properties.TestingRun(t)
}

// プロパティ10: 未関連付けピクチャの非表示
// 任意の未関連付けPictureSpriteについて、そのスプライトは描画されない
// **Validates: Requirements 14.2**
//
// 要件 14.2: WHEN PictureSpriteが「未関連付け」状態のとき THEN THE System SHALL そのスプライトを描画しない
//
// このプロパティは以下を検証する:
// 1. 未関連付け状態のPictureSpriteはIsEffectivelyVisible()がfalseを返す
// 2. 未関連付け状態のPictureSpriteの基盤スプライトはVisible()がfalseを返す
// 3. 未関連付け状態のPictureSpriteは状態がPictureSpriteUnattachedである
func TestProperty_UnattachedPictureSpriteNotRendered(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100

	properties := gopter.NewProperties(parameters)

	// テスト用の定数
	const (
		minImageSize = 1
		maxImageSize = 1000
		minPicID     = 0
		maxPicID     = 255
	)

	properties.Property("任意の未関連付けPictureSpriteについて、そのスプライトは描画されない", prop.ForAll(
		func(picID int, width, height int) bool {
			sm := NewSpriteManager()
			psm := NewPictureSpriteManager(sm)

			// テスト用の画像を作成
			srcImg := ebiten.NewImage(width, height)

			// LoadPic時にPictureSpriteを作成（未関連付け状態で作成される）
			ps := psm.CreatePictureSpriteOnLoad(srcImg, picID, width, height)

			// 1. PictureSpriteが作成されることを確認
			if ps == nil {
				return false
			}

			// 2. 状態がPictureSpriteUnattachedであることを確認
			if ps.GetState() != PictureSpriteUnattached {
				return false
			}

			// 3. IsEffectivelyVisible()がfalseを返すことを確認
			// 要件 14.2: 未関連付け状態ではスプライトを描画しない
			if ps.IsEffectivelyVisible() {
				return false
			}

			// 4. 基盤スプライトのVisible()がfalseを返すことを確認
			if ps.GetSprite().Visible() {
				return false
			}

			// 5. ウインドウIDが-1（未関連付け）であることを確認
			if ps.GetWindowID() != -1 {
				return false
			}

			return true
		},
		gen.IntRange(minPicID, maxPicID),
		gen.IntRange(minImageSize, maxImageSize),
		gen.IntRange(minImageSize, maxImageSize),
	))

	properties.TestingRun(t)
}

// プロパティ10の追加テスト: 複数の未関連付けPictureSpriteがすべて非表示であることを確認
// **Validates: Requirements 14.2**
func TestProperty_MultipleUnattachedPictureSpritesNotRendered(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// テスト用の定数
	const (
		minImageSize = 10
		maxImageSize = 100
		minCount     = 1
		maxCount     = 20
	)

	properties.Property("複数の未関連付けPictureSpriteについて、すべてのスプライトは描画されない", prop.ForAll(
		func(count int) bool {
			sm := NewSpriteManager()
			psm := NewPictureSpriteManager(sm)

			createdSprites := make(map[int]*PictureSprite)

			// 複数のピクチャをロード（すべて未関連付け状態）
			for i := range count {
				picID := i
				srcImg := ebiten.NewImage(minImageSize, minImageSize)
				ps := psm.CreatePictureSpriteOnLoad(srcImg, picID, minImageSize, minImageSize)

				if ps == nil {
					return false
				}

				createdSprites[picID] = ps
			}

			// すべてのPictureSpriteが未関連付け状態で非表示であることを確認
			for _, ps := range createdSprites {
				// 状態がPictureSpriteUnattachedであることを確認
				if ps.GetState() != PictureSpriteUnattached {
					return false
				}

				// IsEffectivelyVisible()がfalseを返すことを確認
				if ps.IsEffectivelyVisible() {
					return false
				}

				// 基盤スプライトのVisible()がfalseを返すことを確認
				if ps.GetSprite().Visible() {
					return false
				}
			}

			return true
		},
		gen.IntRange(minCount, maxCount),
	))

	properties.TestingRun(t)
}

// プロパティ10の追加テスト: 関連付け後は表示されることを確認（対照テスト）
// **Validates: Requirements 14.2, 14.3**
func TestProperty_AttachedPictureSpriteIsRendered(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 50

	properties := gopter.NewProperties(parameters)

	// テスト用の定数
	const (
		minImageSize = 10
		maxImageSize = 100
		minPicID     = 0
		maxPicID     = 50
	)

	properties.Property("関連付け後のPictureSpriteは描画される", prop.ForAll(
		func(picID int, width, height int) bool {
			sm := NewSpriteManager()
			psm := NewPictureSpriteManager(sm)

			// テスト用の画像を作成
			srcImg := ebiten.NewImage(width, height)

			// LoadPic時にPictureSpriteを作成（未関連付け状態）
			ps := psm.CreatePictureSpriteOnLoad(srcImg, picID, width, height)

			if ps == nil {
				return false
			}

			// 未関連付け状態では非表示であることを確認
			if ps.IsEffectivelyVisible() {
				return false
			}

			// ウインドウスプライトを作成
			windowImg := ebiten.NewImage(width, height)
			windowSprite := sm.CreateRootSprite(windowImg)
			windowSprite.SetVisible(true)

			// PictureSpriteをウインドウに関連付ける
			err := psm.AttachPictureSpriteToWindow(picID, windowSprite, 0)
			if err != nil {
				return false
			}

			// 関連付け後は状態がPictureSpriteAttachedであることを確認
			if ps.GetState() != PictureSpriteAttached {
				return false
			}

			// 関連付け後はIsEffectivelyVisible()がtrueを返すことを確認
			// 要件 14.3: 関連付け済み状態では親ウインドウの可視性に従って描画する
			if !ps.IsEffectivelyVisible() {
				return false
			}

			// 基盤スプライトのVisible()がtrueを返すことを確認
			if !ps.GetSprite().Visible() {
				return false
			}

			return true
		},
		gen.IntRange(minPicID, maxPicID),
		gen.IntRange(minImageSize, maxImageSize),
		gen.IntRange(minImageSize, maxImageSize),
	))

	properties.TestingRun(t)
}
