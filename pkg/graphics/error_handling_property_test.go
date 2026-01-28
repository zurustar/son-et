package graphics

import (
	"image/color"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// ============================================================
// Property 10: エラーハンドリング
// *任意の*無効なID（存在しないウィンドウID、レイヤーID）に対して:
// - エラーがログに記録される
// - 処理がスキップされる（クラッシュしない）
// - 実行が継続される
// **Validates: Requirements 10.1, 10.2, 10.3, 10.4**
// ============================================================

// ErrorHandlingTestParams はエラーハンドリングテストのパラメータを表す
type ErrorHandlingTestParams struct {
	InvalidWinID   int
	InvalidLayerID int
	InvalidCastID  int
	ValidWinID     int
	ValidWidth     int
	ValidHeight    int
}

// Generate はErrorHandlingTestParamsのランダム生成を実装する
func (ErrorHandlingTestParams) Generate(rand *rand.Rand, size int) reflect.Value {
	params := ErrorHandlingTestParams{
		InvalidWinID:   rand.Intn(1000) + 1000, // 存在しないウィンドウID
		InvalidLayerID: rand.Intn(1000) + 1000, // 存在しないレイヤーID
		InvalidCastID:  rand.Intn(1000) + 1000, // 存在しないキャストID
		ValidWinID:     rand.Intn(100),
		ValidWidth:     rand.Intn(500) + 100,
		ValidHeight:    rand.Intn(500) + 100,
	}
	return reflect.ValueOf(params)
}

// InvalidSizeParams は無効なサイズパラメータを表す
type InvalidSizeParams struct {
	Width  int
	Height int
}

// Generate はInvalidSizeParamsのランダム生成を実装する
func (InvalidSizeParams) Generate(rand *rand.Rand, size int) reflect.Value {
	// 無効なサイズを生成（0以下の値）
	params := InvalidSizeParams{
		Width:  rand.Intn(100) - 100, // -100 から -1
		Height: rand.Intn(100) - 100, // -100 から -1
	}
	// 50%の確率で幅を0にする
	if rand.Intn(2) == 0 {
		params.Width = 0
	}
	// 50%の確率で高さを0にする
	if rand.Intn(2) == 0 {
		params.Height = 0
	}
	return reflect.ValueOf(params)
}

// ============================================================
// Property 10.1: 存在しないウィンドウIDでの操作がクラッシュしない
// **Validates: Requirements 10.1**
// ============================================================

// TestProperty10_NonExistentWindowID_NoCrash は存在しないウィンドウIDでの操作がクラッシュしないことをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.1**
func TestProperty10_NonExistentWindowID_NoCrash(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 存在しないウィンドウIDでGetWindowLayerSetを呼び出してもクラッシュしない
	property := func(params ErrorHandlingTestParams) bool {
		lm := NewLayerManager()

		// 存在しないウィンドウIDでWindowLayerSetを取得
		// nilが返されるべきで、クラッシュしてはならない
		wls := lm.GetWindowLayerSet(params.InvalidWinID)

		// nilが返されることを確認
		return wls == nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.1 (Non-Existent WindowID No Crash) failed: %v", err)
	}
}

// TestProperty10_NonExistentWindowID_CastOperations は存在しないウィンドウIDでのキャスト操作がクラッシュしないことをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.1**
func TestProperty10_NonExistentWindowID_CastOperations(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 存在しないウィンドウIDでPutCastを呼び出してもクラッシュしない
	property := func(params ErrorHandlingTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// 存在しないウィンドウIDでPutCastを呼び出す
		// WindowLayerSetが存在しないため、フォールバック処理が行われる
		// クラッシュしてはならない
		castID, err := cm.PutCast(params.InvalidWinID, 0, 0, 0, 0, 0, 32, 32)

		// エラーがなく、キャストIDが返されることを確認
		// （フォールバック処理でPictureLayerSetが使用される）
		return err == nil && castID >= 0
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.1 (Non-Existent WindowID Cast Operations) failed: %v", err)
	}
}

// ============================================================
// Property 10.2: 存在しないレイヤーIDでの操作がクラッシュしない
// **Validates: Requirements 10.2**
// ============================================================

// TestProperty10_NonExistentLayerID_NoCrash は存在しないレイヤーIDでの操作がクラッシュしないことをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.2**
func TestProperty10_NonExistentLayerID_NoCrash(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 存在しないレイヤーIDでGetLayerを呼び出してもクラッシュしない
	property := func(params ErrorHandlingTestParams) bool {
		lm := NewLayerManager()

		// 有効なWindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.ValidWinID, params.ValidWidth, params.ValidHeight, bgColor)
		if wls == nil {
			return false
		}

		// 存在しないレイヤーIDでGetLayerを呼び出す
		// nilが返されるべきで、クラッシュしてはならない
		layer := wls.GetLayer(params.InvalidLayerID)

		// nilが返されることを確認
		return layer == nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.2 (Non-Existent LayerID No Crash) failed: %v", err)
	}
}

// TestProperty10_NonExistentLayerID_RemoveLayer は存在しないレイヤーIDでのRemoveLayerがクラッシュしないことをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.2**
func TestProperty10_NonExistentLayerID_RemoveLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 存在しないレイヤーIDでRemoveLayerを呼び出してもクラッシュしない
	property := func(params ErrorHandlingTestParams) bool {
		lm := NewLayerManager()

		// 有効なWindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.ValidWinID, params.ValidWidth, params.ValidHeight, bgColor)
		if wls == nil {
			return false
		}

		// 存在しないレイヤーIDでRemoveLayerを呼び出す
		// falseが返されるべきで、クラッシュしてはならない
		removed := wls.RemoveLayer(params.InvalidLayerID)

		// falseが返されることを確認
		return !removed
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.2 (Non-Existent LayerID RemoveLayer) failed: %v", err)
	}
}

// TestProperty10_NonExistentCastID_Operations は存在しないキャストIDでの操作がクラッシュしないことをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.2**
func TestProperty10_NonExistentCastID_Operations(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 存在しないキャストIDでMoveCast/DelCastを呼び出してもクラッシュしない
	property := func(params ErrorHandlingTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// 存在しないキャストIDでMoveCastを呼び出す
		// エラーが返されるべきで、クラッシュしてはならない
		err := cm.MoveCast(params.InvalidCastID, WithCastPosition(100, 100))
		if err == nil {
			return false // エラーが返されるべき
		}

		// 存在しないキャストIDでDelCastを呼び出す
		// エラーが返されるべきで、クラッシュしてはならない
		err = cm.DelCast(params.InvalidCastID)
		if err == nil {
			return false // エラーが返されるべき
		}

		// 存在しないキャストIDでGetCastを呼び出す
		// エラーが返されるべきで、クラッシュしてはならない
		_, err = cm.GetCast(params.InvalidCastID)
		return err != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.2 (Non-Existent CastID Operations) failed: %v", err)
	}
}

// ============================================================
// Property 10.3: レイヤー作成失敗時にnilを返す
// **Validates: Requirements 10.3**
// ============================================================

// TestProperty10_LayerCreationFailure_ReturnsNil はレイヤー作成失敗時にnilが返されることをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.3**
func TestProperty10_LayerCreationFailure_ReturnsNil(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 無効なパラメータでNewPictureLayerを呼び出すとnilが返される
	property := func(params InvalidSizeParams) bool {
		// 無効なサイズでPictureLayerを作成
		layer := NewPictureLayer(1, params.Width, params.Height)

		// 幅または高さが0以下の場合、nilが返されるべき
		if params.Width <= 0 || params.Height <= 0 {
			return layer == nil
		}
		// 有効なサイズの場合、nilでないことを確認
		return layer != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.3 (Layer Creation Failure Returns Nil) failed: %v", err)
	}
}

// TestProperty10_WindowLayerSetCreationFailure_ReturnsNil はWindowLayerSet作成失敗時にnilが返されることをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.3**
func TestProperty10_WindowLayerSetCreationFailure_ReturnsNil(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 無効なパラメータでNewWindowLayerSetを呼び出すとnilが返される
	property := func(params InvalidSizeParams) bool {
		// 無効なサイズでWindowLayerSetを作成
		wls := NewWindowLayerSet(1, params.Width, params.Height, color.White)

		// 幅または高さが0以下の場合、nilが返されるべき
		if params.Width <= 0 || params.Height <= 0 {
			return wls == nil
		}
		// 有効なサイズの場合、nilでないことを確認
		return wls != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.3 (WindowLayerSet Creation Failure Returns Nil) failed: %v", err)
	}
}

// TestProperty10_CastLayerCreationFailure_ReturnsNil はCastLayer作成失敗時にnilが返されることをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.3**
func TestProperty10_CastLayerCreationFailure_ReturnsNil(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 無効なパラメータでNewCastLayerを呼び出すとnilが返される
	property := func(params InvalidSizeParams) bool {
		// 無効なサイズでCastLayerを作成
		layer := NewCastLayer(1, 1, 0, 0, 0, 0, 0, 0, params.Width, params.Height, 0)

		// 幅または高さが0以下の場合、nilが返されるべき
		if params.Width <= 0 || params.Height <= 0 {
			return layer == nil
		}
		// 有効なサイズの場合、nilでないことを確認
		return layer != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.3 (CastLayer Creation Failure Returns Nil) failed: %v", err)
	}
}

// TestProperty10_DrawingLayerCreationFailure_ReturnsNil はDrawingLayer作成失敗時にnilが返されることをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.3**
func TestProperty10_DrawingLayerCreationFailure_ReturnsNil(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 無効なパラメータでNewDrawingLayerを呼び出すとnilが返される
	property := func(params InvalidSizeParams) bool {
		// 無効なサイズでDrawingLayerを作成
		layer := NewDrawingLayer(1, 0, params.Width, params.Height)

		// 幅または高さが0以下の場合、nilが返されるべき
		if params.Width <= 0 || params.Height <= 0 {
			return layer == nil
		}
		// 有効なサイズの場合、nilでないことを確認
		return layer != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.3 (DrawingLayer Creation Failure Returns Nil) failed: %v", err)
	}
}

// ============================================================
// Property 10.4: 致命的でないエラーの後も実行が継続される
// **Validates: Requirements 10.4**
// ============================================================

// TestProperty10_ContinueAfterError_LayerManager はエラー後もLayerManagerが正常に動作することをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.4**
func TestProperty10_ContinueAfterError_LayerManager(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: エラー後もLayerManagerが正常に動作する
	property := func(params ErrorHandlingTestParams) bool {
		lm := NewLayerManager()

		// 1. 存在しないウィンドウIDでGetWindowLayerSetを呼び出す（エラー）
		wls := lm.GetWindowLayerSet(params.InvalidWinID)
		if wls != nil {
			return false
		}

		// 2. エラー後も有効なWindowLayerSetを作成できることを確認
		bgColor := color.RGBA{0, 0, 0, 255}
		validWls := lm.GetOrCreateWindowLayerSet(params.ValidWinID, params.ValidWidth, params.ValidHeight, bgColor)
		if validWls == nil {
			return false
		}

		// 3. 存在しないレイヤーIDでGetLayerを呼び出す（エラー）
		layer := validWls.GetLayer(params.InvalidLayerID)
		if layer != nil {
			return false
		}

		// 4. エラー後も有効なレイヤーを追加できることを確認
		pictureLayer := NewPictureLayer(lm.GetNextLayerID(), params.ValidWidth, params.ValidHeight)
		if pictureLayer == nil {
			return false
		}
		validWls.AddLayer(pictureLayer)

		// 5. 追加したレイヤーが取得できることを確認
		retrievedLayer := validWls.GetLayer(pictureLayer.GetID())
		return retrievedLayer != nil && retrievedLayer.GetID() == pictureLayer.GetID()
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.4 (Continue After Error LayerManager) failed: %v", err)
	}
}

// TestProperty10_ContinueAfterError_CastManager はエラー後もCastManagerが正常に動作することをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.4**
func TestProperty10_ContinueAfterError_CastManager(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: エラー後もCastManagerが正常に動作する
	property := func(params ErrorHandlingTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// 1. 存在しないキャストIDでMoveCastを呼び出す（エラー）
		err := cm.MoveCast(params.InvalidCastID, WithCastPosition(100, 100))
		if err == nil {
			return false // エラーが返されるべき
		}

		// 2. 存在しないキャストIDでDelCastを呼び出す（エラー）
		err = cm.DelCast(params.InvalidCastID)
		if err == nil {
			return false // エラーが返されるべき
		}

		// 3. エラー後も有効なキャストを作成できることを確認
		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.ValidWinID, params.ValidWidth, params.ValidHeight, bgColor)
		if wls == nil {
			return false
		}

		castID, err := cm.PutCast(params.ValidWinID, 0, 10, 10, 0, 0, 32, 32)
		if err != nil {
			return false
		}

		// 4. 作成したキャストが取得できることを確認
		cast, err := cm.GetCast(castID)
		if err != nil || cast == nil {
			return false
		}

		// 5. キャストの移動が正常に動作することを確認
		err = cm.MoveCast(castID, WithCastPosition(50, 50))
		if err != nil {
			return false
		}

		// 6. キャストの削除が正常に動作することを確認
		err = cm.DelCast(castID)
		return err == nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.4 (Continue After Error CastManager) failed: %v", err)
	}
}

// TestProperty10_ContinueAfterError_LayerCreation はレイヤー作成失敗後も実行が継続されることをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.4**
func TestProperty10_ContinueAfterError_LayerCreation(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: レイヤー作成失敗後も実行が継続される
	property := func(params ErrorHandlingTestParams) bool {
		// 1. 無効なパラメータでPictureLayerを作成（失敗）
		invalidLayer := NewPictureLayer(1, -100, -100)
		if invalidLayer != nil {
			return false // nilが返されるべき
		}

		// 2. 無効なパラメータでCastLayerを作成（失敗）
		invalidCast := NewCastLayer(2, 1, 0, 0, 0, 0, 0, 0, -50, -50, 0)
		if invalidCast != nil {
			return false // nilが返されるべき
		}

		// 3. 無効なパラメータでDrawingLayerを作成（失敗）
		invalidDrawing := NewDrawingLayer(3, 0, -100, -100)
		if invalidDrawing != nil {
			return false // nilが返されるべき
		}

		// 4. 無効なパラメータでWindowLayerSetを作成（失敗）
		invalidWls := NewWindowLayerSet(1, -100, -100, color.White)
		if invalidWls != nil {
			return false // nilが返されるべき
		}

		// 5. エラー後も有効なレイヤーを作成できることを確認
		validLayer := NewPictureLayer(10, params.ValidWidth, params.ValidHeight)
		if validLayer == nil {
			return false
		}

		validCast := NewCastLayer(11, 1, 0, 0, 0, 0, 0, 0, 100, 100, 0)
		if validCast == nil {
			return false
		}

		validDrawing := NewDrawingLayer(12, 0, params.ValidWidth, params.ValidHeight)
		if validDrawing == nil {
			return false
		}

		validWls := NewWindowLayerSet(2, params.ValidWidth, params.ValidHeight, color.White)
		return validWls != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10.4 (Continue After Error Layer Creation) failed: %v", err)
	}
}

// ============================================================
// 統合プロパティテスト: エラーハンドリング全体
// **Validates: Requirements 10.1, 10.2, 10.3, 10.4**
// ============================================================

// TestProperty10_IntegratedErrorHandling はエラーハンドリングの統合動作をテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.1, 10.2, 10.3, 10.4**
func TestProperty10_IntegratedErrorHandling(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 様々なエラー状況でもシステムがクラッシュせず、正常に動作を継続する
	property := func(params ErrorHandlingTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// 1. 存在しないウィンドウIDでの操作（要件 10.1）
		wls := lm.GetWindowLayerSet(params.InvalidWinID)
		if wls != nil {
			return false
		}

		// 2. 有効なWindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		validWls := lm.GetOrCreateWindowLayerSet(params.ValidWinID, params.ValidWidth, params.ValidHeight, bgColor)
		if validWls == nil {
			return false
		}

		// 3. 存在しないレイヤーIDでの操作（要件 10.2）
		layer := validWls.GetLayer(params.InvalidLayerID)
		if layer != nil {
			return false
		}

		removed := validWls.RemoveLayer(params.InvalidLayerID)
		if removed {
			return false
		}

		// 4. 存在しないキャストIDでの操作（要件 10.2）
		err := cm.MoveCast(params.InvalidCastID, WithCastPosition(100, 100))
		if err == nil {
			return false
		}

		err = cm.DelCast(params.InvalidCastID)
		if err == nil {
			return false
		}

		// 5. 無効なパラメータでのレイヤー作成（要件 10.3）
		invalidLayer := NewPictureLayer(1, -100, -100)
		if invalidLayer != nil {
			return false
		}

		// 6. エラー後も正常に動作することを確認（要件 10.4）
		castID, err := cm.PutCast(params.ValidWinID, 0, 10, 10, 0, 0, 32, 32)
		if err != nil {
			return false
		}

		cast, err := cm.GetCast(castID)
		if err != nil || cast == nil {
			return false
		}

		err = cm.MoveCast(castID, WithCastPosition(50, 50))
		if err != nil {
			return false
		}

		err = cm.DelCast(castID)
		return err == nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10 (Integrated Error Handling) failed: %v", err)
	}
}

// TestProperty10_MultipleErrorsInSequence は連続したエラーでもシステムが安定していることをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.1, 10.2, 10.3, 10.4**
func TestProperty10_MultipleErrorsInSequence(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 連続したエラーでもシステムが安定している
	property := func(numErrors uint8) bool {
		// エラー数を制限（5-20）
		n := int(numErrors%16) + 5

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		// 有効なWindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		validWls := lm.GetOrCreateWindowLayerSet(0, 640, 480, bgColor)
		if validWls == nil {
			return false
		}

		// 連続してエラーを発生させる
		for i := 0; i < n; i++ {
			switch i % 5 {
			case 0:
				// 存在しないウィンドウIDでGetWindowLayerSet
				_ = lm.GetWindowLayerSet(1000 + i)
			case 1:
				// 存在しないレイヤーIDでGetLayer
				_ = validWls.GetLayer(1000 + i)
			case 2:
				// 存在しないレイヤーIDでRemoveLayer
				_ = validWls.RemoveLayer(1000 + i)
			case 3:
				// 存在しないキャストIDでMoveCast
				_ = cm.MoveCast(1000+i, WithCastPosition(100, 100))
			case 4:
				// 無効なパラメータでレイヤー作成
				_ = NewPictureLayer(i, -100, -100)
			}
		}

		// エラー後も正常に動作することを確認
		castID, err := cm.PutCast(0, 0, 10, 10, 0, 0, 32, 32)
		if err != nil {
			return false
		}

		cast, err := cm.GetCast(castID)
		if err != nil || cast == nil {
			return false
		}

		err = cm.DelCast(castID)
		return err == nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10 (Multiple Errors In Sequence) failed: %v", err)
	}
}

// TestProperty10_PictureLayerSet_ErrorHandling はPictureLayerSetのエラーハンドリングをテスト
// Property 10: エラーハンドリング
// **Validates: Requirements 10.2**
func TestProperty10_PictureLayerSet_ErrorHandling(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: PictureLayerSetでの存在しないレイヤーIDでの操作がクラッシュしない
	property := func(params ErrorHandlingTestParams) bool {
		lm := NewLayerManager()

		// PictureLayerSetを作成
		pls := lm.GetOrCreatePictureLayerSet(params.ValidWinID)
		if pls == nil {
			return false
		}

		// 存在しないレイヤーIDでGetCastLayerByIDを呼び出す
		castLayer := pls.GetCastLayerByID(params.InvalidLayerID)
		if castLayer != nil {
			return false // nilが返されるべき
		}

		// 存在しないレイヤーIDでRemoveCastLayerByIDを呼び出す
		removed := pls.RemoveCastLayerByID(params.InvalidLayerID)
		if removed {
			return false // falseが返されるべき
		}

		// 存在しないレイヤーIDでGetTextLayerを呼び出す
		textLayer := pls.GetTextLayer(params.InvalidLayerID)
		if textLayer != nil {
			return false // nilが返されるべき
		}

		// 存在しないレイヤーIDでRemoveTextLayerを呼び出す
		removed = pls.RemoveTextLayer(params.InvalidLayerID)
		if removed {
			return false // falseが返されるべき
		}

		// エラー後も正常に動作することを確認
		layerID := lm.GetNextLayerID()
		castLayer2 := NewCastLayer(layerID, 1, params.ValidWinID, 0, 0, 0, 0, 0, 32, 32, 0)
		if castLayer2 == nil {
			return false
		}
		pls.AddCastLayer(castLayer2)

		// 追加したキャストレイヤーが取得できることを確認
		retrievedCast := pls.GetCastLayer(1)
		return retrievedCast != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 10 (PictureLayerSet Error Handling) failed: %v", err)
	}
}
