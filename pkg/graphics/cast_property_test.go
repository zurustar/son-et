package graphics

import (
	"image"
	"image/color"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// ============================================================
// Property 6: Cast_Layerのスプライト動作
// *任意の*Cast_Layerに対して:
// - PutCastで新しいCast_Layerが作成される
// - MoveCastで位置が更新される（Z順序は変わらない）
// - DelCastでCast_Layerが削除される
// - 移動時に古い位置と新しい位置がダーティ領域としてマークされる
// **Validates: Requirements 4.1, 4.2, 4.3, 4.6**
// ============================================================

// CastTestParams はキャストテストのパラメータを表す
type CastTestParams struct {
	WinID     int
	WinWidth  uint16
	WinHeight uint16
	PicID     int
	X         int
	Y         int
	SrcX      int
	SrcY      int
	Width     uint16
	Height    uint16
}

// Generate はCastTestParamsのランダム生成を実装する
func (CastTestParams) Generate(rand *rand.Rand, size int) reflect.Value {
	params := CastTestParams{
		WinID:     rand.Intn(100),
		WinWidth:  uint16(rand.Intn(500) + 100),
		WinHeight: uint16(rand.Intn(500) + 100),
		PicID:     rand.Intn(256),
		X:         rand.Intn(400),
		Y:         rand.Intn(400),
		SrcX:      rand.Intn(100),
		SrcY:      rand.Intn(100),
		Width:     uint16(rand.Intn(100) + 10),
		Height:    uint16(rand.Intn(100) + 10),
	}
	return reflect.ValueOf(params)
}

// CastMoveParams はキャスト移動テストのパラメータを表す
type CastMoveParams struct {
	WinID     int
	WinWidth  uint16
	WinHeight uint16
	PicID     int
	InitX     int
	InitY     int
	NewX      int
	NewY      int
	SrcX      int
	SrcY      int
	Width     uint16
	Height    uint16
}

// Generate はCastMoveParamsのランダム生成を実装する
func (CastMoveParams) Generate(rand *rand.Rand, size int) reflect.Value {
	params := CastMoveParams{
		WinID:     rand.Intn(100),
		WinWidth:  uint16(rand.Intn(500) + 100),
		WinHeight: uint16(rand.Intn(500) + 100),
		PicID:     rand.Intn(256),
		InitX:     rand.Intn(200),
		InitY:     rand.Intn(200),
		NewX:      rand.Intn(400),
		NewY:      rand.Intn(400),
		SrcX:      rand.Intn(100),
		SrcY:      rand.Intn(100),
		Width:     uint16(rand.Intn(100) + 10),
		Height:    uint16(rand.Intn(100) + 10),
	}
	return reflect.ValueOf(params)
}

// ============================================================
// Property 6.1: PutCastで新しいCast_Layerが作成される
// **Validates: Requirements 4.1**
// ============================================================

// TestProperty6_CastLayer_PutCastCreatesNewLayer はPutCastで新しいCast_Layerが作成されることをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.1**
func TestProperty6_CastLayer_PutCastCreatesNewLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: PutCastが呼び出されたとき、新しいCast_Layerが作成される
	property := func(params CastTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 初期状態でレイヤーがないことを確認
		initialLayerCount := wls.GetLayerCount()
		if initialLayerCount != 0 {
			return false
		}

		// PutCastを呼び出す
		castID, err := cm.PutCast(params.WinID, params.PicID, params.X, params.Y,
			params.SrcX, params.SrcY, width, height)
		if err != nil {
			return false
		}

		// CastLayerが作成されたことを確認
		if wls.GetLayerCount() != 1 {
			return false
		}

		// CastLayerが正しいプロパティを持つことを確認
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			return false
		}

		// レイヤータイプがCastであることを確認
		if castLayer.GetLayerType() != LayerTypeCast {
			return false
		}

		// 位置が正しいことを確認
		x, y := castLayer.GetPosition()
		if x != params.X || y != params.Y {
			return false
		}

		// ソース領域が正しいことを確認
		srcX, srcY, w, h := castLayer.GetSourceRect()
		if srcX != params.SrcX || srcY != params.SrcY || w != width || h != height {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.1 (PutCast Creates New Layer) failed: %v", err)
	}
}

// TestProperty6_CastLayer_MultiplePutCastCreatesMultipleLayers は複数のPutCastで複数のCast_Layerが作成されることをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.1**
func TestProperty6_CastLayer_MultiplePutCastCreatesMultipleLayers(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のPutCastが呼び出されたとき、それぞれに対応するCast_Layerが作成される
	property := func(numCasts uint8) bool {
		// キャスト数を制限（1-20）
		n := int(numCasts%20) + 1

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winID := 0
		winWidth := 640
		winHeight := 480

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(winID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 複数のキャストを作成
		castIDs := make([]int, n)
		for i := 0; i < n; i++ {
			castID, err := cm.PutCast(winID, i, i*10, i*10, 0, 0, 32, 32)
			if err != nil {
				return false
			}
			castIDs[i] = castID
		}

		// レイヤー数が正しいことを確認
		if wls.GetLayerCount() != n {
			return false
		}

		// 各キャストに対応するCastLayerが存在することを確認
		for _, castID := range castIDs {
			castLayer := wls.GetCastLayer(castID)
			if castLayer == nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.1 (Multiple PutCast) failed: %v", err)
	}
}

// ============================================================
// Property 6.2: MoveCastで位置が更新される（Z順序は変わらない）
// **Validates: Requirements 4.2**
// ============================================================

// TestProperty6_CastLayer_MoveCastUpdatesPosition はMoveCastで位置が更新されることをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.2**
func TestProperty6_CastLayer_MoveCastUpdatesPosition(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: MoveCastが呼び出されたとき、Cast_Layerの位置が更新される
	property := func(params CastMoveParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// キャストを作成
		castID, err := cm.PutCast(params.WinID, params.PicID, params.InitX, params.InitY,
			params.SrcX, params.SrcY, width, height)
		if err != nil {
			return false
		}

		// 初期位置を確認
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			return false
		}

		x, y := castLayer.GetPosition()
		if x != params.InitX || y != params.InitY {
			return false
		}

		// MoveCastで位置を更新
		err = cm.MoveCast(castID, WithCastPosition(params.NewX, params.NewY))
		if err != nil {
			return false
		}

		// 位置が更新されたことを確認
		x, y = castLayer.GetPosition()
		if x != params.NewX || y != params.NewY {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.2 (MoveCast Updates Position) failed: %v", err)
	}
}

// TestProperty6_CastLayer_MoveCastPreservesZOrder はMoveCastでZ順序が変わらないことをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.2**
func TestProperty6_CastLayer_MoveCastPreservesZOrder(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: MoveCastが呼び出されたとき、Z順序は変わらない
	property := func(params CastMoveParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// キャストを作成
		castID, err := cm.PutCast(params.WinID, params.PicID, params.InitX, params.InitY,
			params.SrcX, params.SrcY, width, height)
		if err != nil {
			return false
		}

		// 初期Z順序を記録
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			return false
		}
		initialZOrder := castLayer.GetZOrder()

		// MoveCastで位置を更新
		err = cm.MoveCast(castID, WithCastPosition(params.NewX, params.NewY))
		if err != nil {
			return false
		}

		// Z順序が変わっていないことを確認
		if castLayer.GetZOrder() != initialZOrder {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.2 (MoveCast Preserves ZOrder) failed: %v", err)
	}
}

// TestProperty6_CastLayer_MoveCastWithMultipleCasts は複数キャストでのMoveCastをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.2**
func TestProperty6_CastLayer_MoveCastWithMultipleCasts(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のキャストがある場合、MoveCastは指定したキャストのみを移動する
	property := func(targetIndex uint8, newX, newY int) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winID := 0
		winWidth := 640
		winHeight := 480

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(winID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 3つのキャストを作成
		numCasts := 3
		castIDs := make([]int, numCasts)
		initialPositions := make([][2]int, numCasts)

		for i := 0; i < numCasts; i++ {
			castID, err := cm.PutCast(winID, i, i*50, i*50, 0, 0, 32, 32)
			if err != nil {
				return false
			}
			castIDs[i] = castID
			initialPositions[i] = [2]int{i * 50, i * 50}
		}

		// ターゲットインデックスを制限
		target := int(targetIndex) % numCasts

		// ターゲットキャストを移動
		err := cm.MoveCast(castIDs[target], WithCastPosition(newX, newY))
		if err != nil {
			return false
		}

		// 各キャストの位置を確認
		for i := 0; i < numCasts; i++ {
			castLayer := wls.GetCastLayer(castIDs[i])
			if castLayer == nil {
				return false
			}

			x, y := castLayer.GetPosition()
			if i == target {
				// ターゲットは新しい位置
				if x != newX || y != newY {
					return false
				}
			} else {
				// 他のキャストは元の位置
				if x != initialPositions[i][0] || y != initialPositions[i][1] {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.2 (MoveCast With Multiple Casts) failed: %v", err)
	}
}

// ============================================================
// Property 6.3: DelCastでCast_Layerが削除される
// **Validates: Requirements 4.3**
// ============================================================

// TestProperty6_CastLayer_DelCastRemovesLayer はDelCastでCast_Layerが削除されることをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.3**
func TestProperty6_CastLayer_DelCastRemovesLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: DelCastが呼び出されたとき、Cast_Layerが削除される
	property := func(params CastTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// キャストを作成
		castID, err := cm.PutCast(params.WinID, params.PicID, params.X, params.Y,
			params.SrcX, params.SrcY, width, height)
		if err != nil {
			return false
		}

		// CastLayerが存在することを確認
		if wls.GetCastLayer(castID) == nil {
			return false
		}
		if wls.GetLayerCount() != 1 {
			return false
		}

		// DelCastを呼び出す
		err = cm.DelCast(castID)
		if err != nil {
			return false
		}

		// CastLayerが削除されたことを確認
		if wls.GetCastLayer(castID) != nil {
			return false
		}
		if wls.GetLayerCount() != 0 {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.3 (DelCast Removes Layer) failed: %v", err)
	}
}

// TestProperty6_CastLayer_DelCastOnlyRemovesTargetLayer はDelCastが指定したレイヤーのみを削除することをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.3**
func TestProperty6_CastLayer_DelCastOnlyRemovesTargetLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: DelCastは指定したキャストのみを削除し、他のキャストは残る
	property := func(targetIndex uint8) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winID := 0
		winWidth := 640
		winHeight := 480

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(winID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 3つのキャストを作成
		numCasts := 3
		castIDs := make([]int, numCasts)

		for i := 0; i < numCasts; i++ {
			castID, err := cm.PutCast(winID, i, i*50, i*50, 0, 0, 32, 32)
			if err != nil {
				return false
			}
			castIDs[i] = castID
		}

		// 初期レイヤー数を確認
		if wls.GetLayerCount() != numCasts {
			return false
		}

		// ターゲットインデックスを制限
		target := int(targetIndex) % numCasts

		// ターゲットキャストを削除
		err := cm.DelCast(castIDs[target])
		if err != nil {
			return false
		}

		// レイヤー数が1減っていることを確認
		if wls.GetLayerCount() != numCasts-1 {
			return false
		}

		// ターゲットが削除され、他のキャストが残っていることを確認
		for i := 0; i < numCasts; i++ {
			castLayer := wls.GetCastLayer(castIDs[i])
			if i == target {
				// ターゲットは削除されている
				if castLayer != nil {
					return false
				}
			} else {
				// 他のキャストは残っている
				if castLayer == nil {
					return false
				}
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.3 (DelCast Only Removes Target) failed: %v", err)
	}
}

// ============================================================
// Property 6.4: 移動時に古い位置と新しい位置がダーティ領域としてマークされる
// **Validates: Requirements 4.6**
// ============================================================

// TestProperty6_CastLayer_MoveCastMarksDirtyRegion はMoveCastでダーティ領域がマークされることをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.6**
func TestProperty6_CastLayer_MoveCastMarksDirtyRegion(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: MoveCastが呼び出されたとき、古い位置と新しい位置がダーティ領域としてマークされる
	property := func(params CastMoveParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// キャストを作成
		castID, err := cm.PutCast(params.WinID, params.PicID, params.InitX, params.InitY,
			params.SrcX, params.SrcY, width, height)
		if err != nil {
			return false
		}

		// ダーティ領域をクリア
		wls.ClearDirtyRegion()

		// 位置が同じ場合はスキップ（ダーティ領域は更新されない）
		if params.InitX == params.NewX && params.InitY == params.NewY {
			return true
		}

		// MoveCastで位置を更新
		err = cm.MoveCast(castID, WithCastPosition(params.NewX, params.NewY))
		if err != nil {
			return false
		}

		// ダーティ領域が設定されていることを確認
		dirtyRegion := wls.GetDirtyRegion()
		if dirtyRegion.Empty() {
			return false
		}

		// 古い位置がダーティ領域に含まれていることを確認
		oldBounds := image.Rect(params.InitX, params.InitY, params.InitX+width, params.InitY+height)
		if !dirtyRegion.Overlaps(oldBounds) {
			return false
		}

		// 新しい位置がダーティ領域に含まれていることを確認
		newBounds := image.Rect(params.NewX, params.NewY, params.NewX+width, params.NewY+height)
		if !dirtyRegion.Overlaps(newBounds) {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.4 (MoveCast Marks Dirty Region) failed: %v", err)
	}
}

// TestProperty6_CastLayer_DirtyRegionContainsBothPositions はダーティ領域が両方の位置を含むことをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.6**
func TestProperty6_CastLayer_DirtyRegionContainsBothPositions(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ダーティ領域は古い位置と新しい位置の両方を含む
	property := func(initX, initY, newX, newY int, width, height uint16) bool {
		// 位置を正の値に制限
		if initX < 0 {
			initX = -initX
		}
		if initY < 0 {
			initY = -initY
		}
		if newX < 0 {
			newX = -newX
		}
		if newY < 0 {
			newY = -newY
		}

		// サイズを制限
		w := int(width%200) + 10
		h := int(height%200) + 10

		// 位置が同じ場合はスキップ
		if initX == newX && initY == newY {
			return true
		}

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winID := 0
		winWidth := 1000
		winHeight := 1000

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(winID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// キャストを作成
		castID, err := cm.PutCast(winID, 1, initX, initY, 0, 0, w, h)
		if err != nil {
			return false
		}

		// ダーティ領域をクリア
		wls.ClearDirtyRegion()

		// MoveCastで位置を更新
		err = cm.MoveCast(castID, WithCastPosition(newX, newY))
		if err != nil {
			return false
		}

		// ダーティ領域を取得
		dirtyRegion := wls.GetDirtyRegion()

		// 古い位置の境界
		oldBounds := image.Rect(initX, initY, initX+w, initY+h)
		// 新しい位置の境界
		newBounds := image.Rect(newX, newY, newX+w, newY+h)

		// ダーティ領域が両方の位置を含むことを確認
		expectedUnion := oldBounds.Union(newBounds)

		// ダーティ領域が期待される領域を含むことを確認
		if dirtyRegion.Min.X > expectedUnion.Min.X ||
			dirtyRegion.Min.Y > expectedUnion.Min.Y ||
			dirtyRegion.Max.X < expectedUnion.Max.X ||
			dirtyRegion.Max.Y < expectedUnion.Max.Y {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6.4 (Dirty Region Contains Both Positions) failed: %v", err)
	}
}

// ============================================================
// 統合プロパティテスト: Cast_Layerのスプライト動作全体
// **Validates: Requirements 4.1, 4.2, 4.3, 4.6**
// ============================================================

// TestProperty6_CastLayer_IntegratedBehavior はCast_Layerの統合動作をテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.1, 4.2, 4.3, 4.6**
func TestProperty6_CastLayer_IntegratedBehavior(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: PutCast → MoveCast → DelCast の一連の操作が正しく動作する
	property := func(params CastMoveParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 1. PutCast: 新しいCast_Layerが作成される（要件 4.1）
		castID, err := cm.PutCast(params.WinID, params.PicID, params.InitX, params.InitY,
			params.SrcX, params.SrcY, width, height)
		if err != nil {
			return false
		}

		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			return false
		}

		// 初期Z順序を記録
		initialZOrder := castLayer.GetZOrder()

		// 2. MoveCast: 位置が更新される（要件 4.2）
		wls.ClearDirtyRegion()

		err = cm.MoveCast(castID, WithCastPosition(params.NewX, params.NewY))
		if err != nil {
			return false
		}

		// 位置が更新されたことを確認
		x, y := castLayer.GetPosition()
		if x != params.NewX || y != params.NewY {
			return false
		}

		// Z順序が変わっていないことを確認
		if castLayer.GetZOrder() != initialZOrder {
			return false
		}

		// ダーティ領域が設定されていることを確認（位置が変わった場合のみ）
		if params.InitX != params.NewX || params.InitY != params.NewY {
			dirtyRegion := wls.GetDirtyRegion()
			if dirtyRegion.Empty() {
				return false
			}
		}

		// 3. DelCast: Cast_Layerが削除される（要件 4.3）
		err = cm.DelCast(castID)
		if err != nil {
			return false
		}

		// CastLayerが削除されたことを確認
		if wls.GetCastLayer(castID) != nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6 (Integrated Behavior) failed: %v", err)
	}
}

// TestProperty6_CastLayer_MultipleOperationsSequence は複数の操作シーケンスをテスト
// Property 6: Cast_Layerのスプライト動作
// **Validates: Requirements 4.1, 4.2, 4.3, 4.6**
func TestProperty6_CastLayer_MultipleOperationsSequence(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のキャストに対する操作シーケンスが正しく動作する
	property := func(numOps uint8) bool {
		// 操作数を制限（5-15）
		ops := int(numOps%11) + 5

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winID := 0
		winWidth := 640
		winHeight := 480

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(winID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// キャストIDを追跡
		activeCasts := make(map[int]bool)

		for i := 0; i < ops; i++ {
			switch i % 3 {
			case 0:
				// PutCast
				castID, err := cm.PutCast(winID, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
				activeCasts[castID] = true

				// CastLayerが作成されたことを確認
				if wls.GetCastLayer(castID) == nil {
					return false
				}

			case 1:
				// MoveCast（アクティブなキャストがある場合）
				for castID := range activeCasts {
					err := cm.MoveCast(castID, WithCastPosition(i*20, i*20))
					if err != nil {
						return false
					}

					// 位置が更新されたことを確認
					castLayer := wls.GetCastLayer(castID)
					if castLayer == nil {
						return false
					}
					x, y := castLayer.GetPosition()
					if x != i*20 || y != i*20 {
						return false
					}
					break // 1つだけ移動
				}

			case 2:
				// DelCast（アクティブなキャストがある場合）
				for castID := range activeCasts {
					err := cm.DelCast(castID)
					if err != nil {
						return false
					}
					delete(activeCasts, castID)

					// CastLayerが削除されたことを確認
					if wls.GetCastLayer(castID) != nil {
						return false
					}
					break // 1つだけ削除
				}
			}
		}

		// 最終状態: アクティブなキャスト数とレイヤー数が一致することを確認
		if wls.GetLayerCount() != len(activeCasts) {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 6 (Multiple Operations Sequence) failed: %v", err)
	}
}
