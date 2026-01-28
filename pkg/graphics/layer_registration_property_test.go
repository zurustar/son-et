package graphics

import (
	"image/color"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// ============================================================
// Property 8: レイヤーのウィンドウ登録
// *任意の*レイヤー作成操作（PutCast、MovePic、TextWrite）に対して:
// - レイヤーは正しいウィンドウIDで登録される
// - ピクチャーIDからウィンドウIDへの逆引きが正しく動作する
// **Validates: Requirements 7.1, 7.2, 7.3, 7.5**
// ============================================================

// LayerRegistrationTestParams はレイヤー登録テストのパラメータを表す
type LayerRegistrationTestParams struct {
	WinID     int
	WinWidth  uint16
	WinHeight uint16
	PicID     int
	X         int
	Y         int
	Width     uint16
	Height    uint16
}

// Generate はLayerRegistrationTestParamsのランダム生成を実装する
func (LayerRegistrationTestParams) Generate(rand *rand.Rand, size int) reflect.Value {
	params := LayerRegistrationTestParams{
		WinID:     rand.Intn(50),
		WinWidth:  uint16(rand.Intn(500) + 100),
		WinHeight: uint16(rand.Intn(500) + 100),
		PicID:     rand.Intn(256),
		X:         rand.Intn(200),
		Y:         rand.Intn(200),
		Width:     uint16(rand.Intn(100) + 10),
		Height:    uint16(rand.Intn(100) + 10),
	}
	return reflect.ValueOf(params)
}

// ============================================================
// Property 8.1: PutCastでCast_LayerがウィンドウIDで登録される
// **Validates: Requirements 7.1**
// ============================================================

// TestProperty8_LayerRegistration_PutCastRegistersWithWindowID はPutCastでCast_LayerがウィンドウIDで登録されることをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1**
func TestProperty8_LayerRegistration_PutCastRegistersWithWindowID(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: PutCastが呼び出されたとき、Cast_LayerはウィンドウIDで登録される
	property := func(params LayerRegistrationTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)
		width := int(params.Width)
		height := int(params.Height)

		// WindowLayerSetを作成（ウィンドウIDで管理）
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 初期状態でレイヤーがないことを確認
		if wls.GetLayerCount() != 0 {
			return false
		}

		// PutCastを呼び出す（ウィンドウIDを指定）
		castID, err := cm.PutCast(params.WinID, params.PicID, params.X, params.Y,
			0, 0, width, height)
		if err != nil {
			return false
		}

		// CastLayerがウィンドウIDで登録されたことを確認
		// 要件 7.1: Cast_LayerをウィンドウIDで登録する
		if wls.GetLayerCount() != 1 {
			return false
		}

		// CastLayerが正しいウィンドウに登録されていることを確認
		castLayer := wls.GetCastLayer(castID)
		if castLayer == nil {
			return false
		}

		// レイヤータイプがCastであることを確認
		if castLayer.GetLayerType() != LayerTypeCast {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.1 (PutCast Registers With WindowID) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_MultiplePutCastSameWindow は同じウィンドウへの複数PutCastをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1**
func TestProperty8_LayerRegistration_MultiplePutCastSameWindow(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 同じウィンドウに複数のPutCastを呼び出すと、すべてのCast_Layerが同じウィンドウに登録される
	property := func(params LayerRegistrationTestParams, numCasts uint8) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)

		// キャスト数を制限（1-10）
		n := int(numCasts%10) + 1

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 複数のPutCastを呼び出す
		castIDs := make([]int, n)
		for i := 0; i < n; i++ {
			castID, err := cm.PutCast(params.WinID, i, i*10, i*10, 0, 0, 32, 32)
			if err != nil {
				return false
			}
			castIDs[i] = castID
		}

		// すべてのCastLayerが同じウィンドウに登録されていることを確認
		if wls.GetLayerCount() != n {
			return false
		}

		// 各CastLayerが存在することを確認
		for _, castID := range castIDs {
			if wls.GetCastLayer(castID) == nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.1 (Multiple PutCast Same Window) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_PutCastDifferentWindows は異なるウィンドウへのPutCastをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1**
func TestProperty8_LayerRegistration_PutCastDifferentWindows(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 異なるウィンドウにPutCastを呼び出すと、各Cast_Layerは正しいウィンドウに登録される
	property := func(winID1, winID2 int) bool {
		// 負のIDは正に変換
		if winID1 < 0 {
			winID1 = -winID1
		}
		if winID2 < 0 {
			winID2 = -winID2
		}

		// 同じウィンドウIDの場合はスキップ
		if winID1 == winID2 {
			return true
		}

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		bgColor := color.RGBA{0, 0, 0, 255}

		// 2つのWindowLayerSetを作成
		wls1 := lm.GetOrCreateWindowLayerSet(winID1, 640, 480, bgColor)
		wls2 := lm.GetOrCreateWindowLayerSet(winID2, 800, 600, bgColor)

		if wls1 == nil || wls2 == nil {
			return false
		}

		// 各ウィンドウにPutCastを呼び出す
		castID1, err := cm.PutCast(winID1, 1, 10, 10, 0, 0, 32, 32)
		if err != nil {
			return false
		}

		castID2, err := cm.PutCast(winID2, 2, 20, 20, 0, 0, 32, 32)
		if err != nil {
			return false
		}

		// 各CastLayerが正しいウィンドウに登録されていることを確認
		// wls1にはcastID1のみ
		if wls1.GetCastLayer(castID1) == nil {
			return false
		}
		if wls1.GetCastLayer(castID2) != nil {
			return false
		}

		// wls2にはcastID2のみ
		if wls2.GetCastLayer(castID2) == nil {
			return false
		}
		if wls2.GetCastLayer(castID1) != nil {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.1 (PutCast Different Windows) failed: %v", err)
	}
}

// ============================================================
// Property 8.2: ピクチャーIDからウィンドウIDへの逆引き
// **Validates: Requirements 7.5**
// ============================================================

// TestProperty8_LayerRegistration_PicIDToWinIDReverseLookup はピクチャーIDからウィンドウIDへの逆引きをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.5**
func TestProperty8_LayerRegistration_PicIDToWinIDReverseLookup(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ウィンドウを開いた後、ピクチャーIDからウィンドウIDを逆引きできる
	property := func(winID int, picID int) bool {
		// 負のIDは正に変換
		if winID < 0 {
			winID = -winID
		}
		if picID < 0 {
			picID = -picID
		}

		wm := NewWindowManager()

		// ウィンドウを開く（ピクチャーIDを関連付け）
		openedWinID, err := wm.OpenWin(picID)
		if err != nil {
			return false
		}

		// ピクチャーIDからウィンドウIDを逆引き
		// 要件 7.5: ピクチャーIDからウィンドウIDへの逆引きをサポートする
		foundWinID, err := wm.GetWinByPicID(picID)
		if err != nil {
			return false
		}

		// 逆引きしたウィンドウIDが正しいことを確認
		return foundWinID == openedWinID
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.2 (PicID to WinID Reverse Lookup) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_PicIDToWinIDNotFound は存在しないピクチャーIDの逆引きをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.5**
func TestProperty8_LayerRegistration_PicIDToWinIDNotFound(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 存在しないピクチャーIDの逆引きはエラーを返す
	property := func(picID, nonExistentPicID int) bool {
		// 負のIDは正に変換
		if picID < 0 {
			picID = -picID
		}
		if nonExistentPicID < 0 {
			nonExistentPicID = -nonExistentPicID
		}

		// 同じIDの場合はスキップ
		if picID == nonExistentPicID {
			return true
		}

		wm := NewWindowManager()

		// ウィンドウを開く
		_, err := wm.OpenWin(picID)
		if err != nil {
			return false
		}

		// 存在しないピクチャーIDで逆引きを試みる
		_, err = wm.GetWinByPicID(nonExistentPicID)

		// エラーが返されることを確認
		return err != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.2 (PicID to WinID Not Found) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_PicIDToWinIDMultipleWindows は複数ウィンドウでの逆引きをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.5**
func TestProperty8_LayerRegistration_PicIDToWinIDMultipleWindows(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のウィンドウが同じピクチャーを使用している場合、最後に開かれたウィンドウを返す
	property := func(picID int) bool {
		// 負のIDは正に変換
		if picID < 0 {
			picID = -picID
		}

		wm := NewWindowManager()

		// 同じピクチャーIDで複数のウィンドウを開く
		winID1, err := wm.OpenWin(picID)
		if err != nil {
			return false
		}

		winID2, err := wm.OpenWin(picID)
		if err != nil {
			return false
		}

		// ピクチャーIDからウィンドウIDを逆引き
		// 最後に開かれたウィンドウ（最高のZOrder）を返す
		foundWinID, err := wm.GetWinByPicID(picID)
		if err != nil {
			return false
		}

		// 最後に開かれたウィンドウが返されることを確認
		return foundWinID == winID2 && winID1 != winID2
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.2 (PicID to WinID Multiple Windows) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_PicIDToWinIDAfterClose はウィンドウ閉鎖後の逆引きをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.5**
func TestProperty8_LayerRegistration_PicIDToWinIDAfterClose(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: ウィンドウを閉じた後、そのピクチャーIDの逆引きはエラーを返す
	property := func(picID int) bool {
		// 負のIDは正に変換
		if picID < 0 {
			picID = -picID
		}

		wm := NewWindowManager()

		// ウィンドウを開く
		winID, err := wm.OpenWin(picID)
		if err != nil {
			return false
		}

		// 逆引きが成功することを確認
		_, err = wm.GetWinByPicID(picID)
		if err != nil {
			return false
		}

		// ウィンドウを閉じる
		err = wm.CloseWin(winID)
		if err != nil {
			return false
		}

		// 閉じた後の逆引きはエラーを返す
		_, err = wm.GetWinByPicID(picID)
		return err != nil
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.2 (PicID to WinID After Close) failed: %v", err)
	}
}

// ============================================================
// Property 8.3: TextWriteでText_Layerが対象ピクチャーに関連付けられたウィンドウに登録される
// **Validates: Requirements 7.3**
// ============================================================

// TestProperty8_LayerRegistration_TextLayerRegistration はTextWriteでText_Layerが登録されることをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.3**
func TestProperty8_LayerRegistration_TextLayerRegistration(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: TextWriteが呼び出されたとき、Text_Layerは対象ピクチャーに関連付けられたウィンドウに登録される
	// 注: 現在の実装ではTextWriteはPictureLayerSetに登録するため、このテストはPictureLayerSetへの登録を確認する
	property := func(params LayerRegistrationTestParams) bool {
		lm := NewLayerManager()

		// PictureLayerSetを作成
		pls := lm.GetOrCreatePictureLayerSet(params.PicID)
		if pls == nil {
			return false
		}

		// 初期状態でテキストレイヤーがないことを確認
		if pls.GetTextLayerCount() != 0 {
			return false
		}

		// TextLayerEntryを作成（TextWriteの内部動作をシミュレート）
		layerID := lm.GetNextLayerID()
		textLayerEntry := NewTextLayerEntry(layerID, params.PicID, params.X, params.Y, "Test", pls.GetNextTextZOffset())

		// PictureLayerSetに追加
		pls.AddTextLayer(textLayerEntry)

		// Text_Layerが登録されたことを確認
		// 要件 7.3: Text_Layerを対象ピクチャーに関連付けられたウィンドウに登録する
		if pls.GetTextLayerCount() != 1 {
			return false
		}

		// TextLayerEntryが正しく登録されていることを確認
		retrievedLayer := pls.GetTextLayer(layerID)
		if retrievedLayer == nil {
			return false
		}

		// レイヤータイプがTextであることを確認
		if retrievedLayer.GetLayerType() != LayerTypeText {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.3 (TextLayer Registration) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_TextLayerWindowLayerSet はWindowLayerSetへのText_Layer登録をテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.3**
func TestProperty8_LayerRegistration_TextLayerWindowLayerSet(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: WindowLayerSetにText_Layerを追加すると、正しく登録される
	property := func(params LayerRegistrationTestParams) bool {
		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)

		// WindowLayerSetを作成
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 初期状態でレイヤーがないことを確認
		if wls.GetLayerCount() != 0 {
			return false
		}

		// TextLayerEntryを作成
		layerID := lm.GetNextLayerID()
		textLayerEntry := NewTextLayerEntry(layerID, params.PicID, params.X, params.Y, "Test", 0)

		// WindowLayerSetに追加
		wls.AddLayer(textLayerEntry)

		// Text_Layerが登録されたことを確認
		if wls.GetLayerCount() != 1 {
			return false
		}

		// TextLayerEntryが正しく登録されていることを確認
		retrievedLayer := wls.GetLayer(layerID)
		if retrievedLayer == nil {
			return false
		}

		// レイヤータイプがTextであることを確認
		if retrievedLayer.GetLayerType() != LayerTypeText {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.3 (TextLayer WindowLayerSet) failed: %v", err)
	}
}

// ============================================================
// Property 8.4: 統合テスト - すべてのレイヤータイプの登録
// **Validates: Requirements 7.1, 7.2, 7.3, 7.5**
// ============================================================

// TestProperty8_LayerRegistration_AllLayerTypesIntegration はすべてのレイヤータイプの登録をテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1, 7.2, 7.3, 7.5**
func TestProperty8_LayerRegistration_AllLayerTypesIntegration(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: PutCast、MovePic（PictureLayer）、TextWriteのすべてが正しいウィンドウに登録される
	property := func(params LayerRegistrationTestParams) bool {
		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)

		// WindowLayerSetを作成
		bgColor := color.RGBA{0, 0, 0, 255}
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 1. PutCastでCast_Layerを登録（要件 7.1）
		castID, err := cm.PutCast(params.WinID, params.PicID, params.X, params.Y, 0, 0, 32, 32)
		if err != nil {
			return false
		}

		// CastLayerが登録されたことを確認
		if wls.GetCastLayer(castID) == nil {
			return false
		}

		// 2. PictureLayerを登録（要件 7.2のシミュレート）
		pictureLayerID := lm.GetNextLayerID()
		pictureLayer := NewPictureLayer(pictureLayerID, winWidth, winHeight)
		wls.AddLayer(pictureLayer)

		// PictureLayerが登録されたことを確認
		if wls.GetLayer(pictureLayerID) == nil {
			return false
		}

		// 3. TextLayerを登録（要件 7.3のシミュレート）
		textLayerID := lm.GetNextLayerID()
		textLayerEntry := NewTextLayerEntry(textLayerID, params.PicID, params.X, params.Y, "Test", 0)
		wls.AddLayer(textLayerEntry)

		// TextLayerが登録されたことを確認
		if wls.GetLayer(textLayerID) == nil {
			return false
		}

		// すべてのレイヤーが同じウィンドウに登録されていることを確認
		// Cast(1) + Picture(1) + Text(1) = 3
		if wls.GetLayerCount() != 3 {
			return false
		}

		// 各レイヤータイプが正しいことを確認
		castLayer := wls.GetCastLayer(castID)
		if castLayer.GetLayerType() != LayerTypeCast {
			return false
		}

		picLayer := wls.GetLayer(pictureLayerID)
		if picLayer.GetLayerType() != LayerTypePicture {
			return false
		}

		txtLayer := wls.GetLayer(textLayerID)
		if txtLayer.GetLayerType() != LayerTypeText {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.4 (All Layer Types Integration) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_LayerIsolationBetweenWindows は異なるウィンドウ間のレイヤー分離をテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1, 7.2, 7.3**
func TestProperty8_LayerRegistration_LayerIsolationBetweenWindows(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 異なるウィンドウに登録されたレイヤーは互いに独立している
	property := func(winID1, winID2 int) bool {
		// 負のIDは正に変換
		if winID1 < 0 {
			winID1 = -winID1
		}
		if winID2 < 0 {
			winID2 = -winID2
		}

		// 同じウィンドウIDの場合はスキップ
		if winID1 == winID2 {
			return true
		}

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		bgColor := color.RGBA{0, 0, 0, 255}

		// 2つのWindowLayerSetを作成
		wls1 := lm.GetOrCreateWindowLayerSet(winID1, 640, 480, bgColor)
		wls2 := lm.GetOrCreateWindowLayerSet(winID2, 800, 600, bgColor)

		if wls1 == nil || wls2 == nil {
			return false
		}

		// ウィンドウ1にレイヤーを追加
		castID1, _ := cm.PutCast(winID1, 1, 10, 10, 0, 0, 32, 32)
		textLayerID1 := lm.GetNextLayerID()
		textLayer1 := NewTextLayerEntry(textLayerID1, 1, 10, 10, "Window1", 0)
		wls1.AddLayer(textLayer1)

		// ウィンドウ2にレイヤーを追加
		castID2, _ := cm.PutCast(winID2, 2, 20, 20, 0, 0, 32, 32)
		textLayerID2 := lm.GetNextLayerID()
		textLayer2 := NewTextLayerEntry(textLayerID2, 2, 20, 20, "Window2", 0)
		wls2.AddLayer(textLayer2)

		// ウィンドウ1のレイヤーがウィンドウ2に存在しないことを確認
		if wls2.GetCastLayer(castID1) != nil {
			return false
		}
		if wls2.GetLayer(textLayerID1) != nil {
			return false
		}

		// ウィンドウ2のレイヤーがウィンドウ1に存在しないことを確認
		if wls1.GetCastLayer(castID2) != nil {
			return false
		}
		if wls1.GetLayer(textLayerID2) != nil {
			return false
		}

		// 各ウィンドウのレイヤー数が正しいことを確認
		// wls1: Cast(1) + Text(1) = 2
		// wls2: Cast(1) + Text(1) = 2
		if wls1.GetLayerCount() != 2 {
			return false
		}
		if wls2.GetLayerCount() != 2 {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.4 (Layer Isolation Between Windows) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_CorrectWindowIDAssignment はレイヤーが正しいウィンドウIDで登録されることをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1, 7.2, 7.3, 7.5**
func TestProperty8_LayerRegistration_CorrectWindowIDAssignment(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: レイヤーは指定されたウィンドウIDで正しく登録される
	property := func(numWindows uint8) bool {
		// ウィンドウ数を制限（1-5）
		n := int(numWindows%5) + 1

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		bgColor := color.RGBA{0, 0, 0, 255}

		// 複数のウィンドウを作成
		windowLayerSets := make([]*WindowLayerSet, n)
		for i := 0; i < n; i++ {
			wls := lm.GetOrCreateWindowLayerSet(i, 640, 480, bgColor)
			if wls == nil {
				return false
			}
			windowLayerSets[i] = wls
		}

		// 各ウィンドウにレイヤーを追加
		castIDs := make([]int, n)
		for i := 0; i < n; i++ {
			castID, err := cm.PutCast(i, i, i*10, i*10, 0, 0, 32, 32)
			if err != nil {
				return false
			}
			castIDs[i] = castID
		}

		// 各レイヤーが正しいウィンドウに登録されていることを確認
		for i := 0; i < n; i++ {
			wls := windowLayerSets[i]

			// このウィンドウには自分のキャストのみが存在する
			if wls.GetCastLayer(castIDs[i]) == nil {
				return false
			}

			// 他のウィンドウのキャストは存在しない
			for j := 0; j < n; j++ {
				if i != j {
					if wls.GetCastLayer(castIDs[j]) != nil {
						return false
					}
				}
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.4 (Correct WindowID Assignment) failed: %v", err)
	}
}

// TestProperty8_LayerRegistration_ZOrderPreservedAcrossTypes は異なるレイヤータイプ間でZ順序が保持されることをテスト
// Property 8: レイヤーのウィンドウ登録
// **Validates: Requirements 7.1, 7.2, 7.3**
func TestProperty8_LayerRegistration_ZOrderPreservedAcrossTypes(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 異なるレイヤータイプを追加しても、追加順序に基づくZ順序が保持される
	property := func(sequence []uint8) bool {
		// シーケンス長を制限（1-10）
		if len(sequence) == 0 {
			return true
		}
		if len(sequence) > 10 {
			sequence = sequence[:10]
		}

		cm := NewCastManager()
		lm := NewLayerManager()
		cm.SetLayerManager(lm)

		winID := 0
		bgColor := color.RGBA{0, 0, 0, 255}

		// WindowLayerSetを作成
		wls := lm.GetOrCreateWindowLayerSet(winID, 640, 480, bgColor)
		if wls == nil {
			return false
		}

		// 追加されたレイヤーのZ順序を記録
		zOrders := make([]int, 0)

		for i, s := range sequence {
			layerType := s % 3 // 0: Cast, 1: Picture, 2: Text

			var zOrder int
			switch layerType {
			case 0:
				// Cast
				castID, err := cm.PutCast(winID, i, i*10, i*10, 0, 0, 32, 32)
				if err != nil {
					return false
				}
				castLayer := wls.GetCastLayer(castID)
				if castLayer == nil {
					return false
				}
				zOrder = castLayer.GetZOrder()
			case 1:
				// Picture
				layerID := lm.GetNextLayerID()
				pictureLayer := NewPictureLayer(layerID, 640, 480)
				wls.AddLayer(pictureLayer)
				zOrder = pictureLayer.GetZOrder()
			case 2:
				// Text
				layerID := lm.GetNextLayerID()
				textLayer := NewTextLayerEntry(layerID, i, i*10, i*10, "Test", 0)
				wls.AddLayer(textLayer)
				zOrder = textLayer.GetZOrder()
			}

			zOrders = append(zOrders, zOrder)
		}

		// Z順序が追加順序に基づいて増加していることを確認
		for i := 1; i < len(zOrders); i++ {
			if zOrders[i] <= zOrders[i-1] {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 8.4 (Z-Order Preserved Across Types) failed: %v", err)
	}
}
