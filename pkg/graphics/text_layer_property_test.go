package graphics

import (
	"image/color"
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// ============================================================
// Property 5: Text_Layerの新規作成
// *任意の*TextWrite呼び出しに対して、常に新しいText_Layerが作成され、
// 既存のレイヤーは再利用されない。
// **Validates: Requirements 2.6, 5.1, 5.2**
// ============================================================

// TextLayerTestParams はテキストレイヤーテストのパラメータを表す
type TextLayerTestParams struct {
	WinID     int
	WinWidth  uint16
	WinHeight uint16
	PicID     int
	X         int
	Y         int
	Text      string
}

// Generate はTextLayerTestParamsのランダム生成を実装する
func (TextLayerTestParams) Generate(rand *rand.Rand, size int) reflect.Value {
	// ランダムなテキストを生成
	textLen := rand.Intn(20) + 1
	textBytes := make([]byte, textLen)
	for i := range textBytes {
		textBytes[i] = byte('A' + rand.Intn(26))
	}

	params := TextLayerTestParams{
		WinID:     rand.Intn(100),
		WinWidth:  uint16(rand.Intn(500) + 100),
		WinHeight: uint16(rand.Intn(500) + 100),
		PicID:     rand.Intn(256),
		X:         rand.Intn(200),
		Y:         rand.Intn(200),
		Text:      string(textBytes),
	}
	return reflect.ValueOf(params)
}

// MultipleTextLayerTestParams は複数テキストレイヤーテストのパラメータを表す
type MultipleTextLayerTestParams struct {
	WinID     int
	WinWidth  uint16
	WinHeight uint16
	PicID     int
	NumTexts  uint8
}

// Generate はMultipleTextLayerTestParamsのランダム生成を実装する
func (MultipleTextLayerTestParams) Generate(rand *rand.Rand, size int) reflect.Value {
	params := MultipleTextLayerTestParams{
		WinID:     rand.Intn(100),
		WinWidth:  uint16(rand.Intn(500) + 100),
		WinHeight: uint16(rand.Intn(500) + 100),
		PicID:     rand.Intn(256),
		NumTexts:  uint8(rand.Intn(10) + 1), // 1-10個のテキスト
	}
	return reflect.ValueOf(params)
}

// ============================================================
// Property 5.1: TextWriteが呼び出されたとき、常に新しいText_Layerを作成する
// **Validates: Requirements 5.1**
// ============================================================

// TestProperty5_TextLayer_TextWriteCreatesNewLayer はTextWriteで新しいText_Layerが作成されることをテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 5.1**
func TestProperty5_TextLayer_TextWriteCreatesNewLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: TextWriteが呼び出されたとき、新しいText_Layerが作成される
	property := func(params TextLayerTestParams) bool {
		lm := NewLayerManager()

		// PictureLayerSetを取得または作成
		pls := lm.GetOrCreatePictureLayerSet(params.PicID)
		if pls == nil {
			return false
		}

		// 初期状態でテキストレイヤーがないことを確認
		initialTextLayerCount := pls.GetTextLayerCount()
		if initialTextLayerCount != 0 {
			return false
		}

		// TextLayerEntryを作成（TextWriteの内部動作をシミュレート）
		layerID := lm.GetNextLayerID()
		textLayerEntry := NewTextLayerEntry(layerID, params.PicID, params.X, params.Y, params.Text, pls.GetNextTextZOffset())

		// PictureLayerSetに追加
		pls.AddTextLayer(textLayerEntry)

		// テキストレイヤーが作成されたことを確認
		if pls.GetTextLayerCount() != 1 {
			return false
		}

		// テキストレイヤーが正しいプロパティを持つことを確認
		retrievedLayer := pls.GetTextLayer(layerID)
		if retrievedLayer == nil {
			return false
		}

		// レイヤータイプがTextであることを確認
		if retrievedLayer.GetLayerType() != LayerTypeText {
			return false
		}

		// 位置が正しいことを確認
		x, y := retrievedLayer.GetPosition()
		if x != params.X || y != params.Y {
			return false
		}

		// テキストが正しいことを確認
		if retrievedLayer.GetText() != params.Text {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5.1 (TextWrite Creates New Layer) failed: %v", err)
	}
}

// ============================================================
// Property 5.2: Text_Layerは既存のレイヤーを再利用しない
// **Validates: Requirements 5.2**
// ============================================================

// TestProperty5_TextLayer_NeverReusesExistingLayer はText_Layerが既存のレイヤーを再利用しないことをテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 5.2**
func TestProperty5_TextLayer_NeverReusesExistingLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 複数のTextWrite呼び出しで、各呼び出しが新しいレイヤーを作成する
	property := func(params MultipleTextLayerTestParams) bool {
		lm := NewLayerManager()

		// PictureLayerSetを取得または作成
		pls := lm.GetOrCreatePictureLayerSet(params.PicID)
		if pls == nil {
			return false
		}

		// テキスト数を制限（1-10）
		numTexts := int(params.NumTexts%10) + 1

		// 作成されたレイヤーIDを追跡
		layerIDs := make([]int, numTexts)

		// 複数のTextWriteをシミュレート
		for i := 0; i < numTexts; i++ {
			layerID := lm.GetNextLayerID()
			textLayerEntry := NewTextLayerEntry(layerID, params.PicID, i*10, i*10, "Test"+string(rune('A'+i)), pls.GetNextTextZOffset())
			pls.AddTextLayer(textLayerEntry)
			layerIDs[i] = layerID
		}

		// レイヤー数が正しいことを確認（各TextWriteで新しいレイヤーが作成される）
		if pls.GetTextLayerCount() != numTexts {
			return false
		}

		// すべてのレイヤーIDが一意であることを確認（再利用されていない）
		idSet := make(map[int]bool)
		for _, id := range layerIDs {
			if idSet[id] {
				// 重複するIDがある = 再利用されている
				return false
			}
			idSet[id] = true
		}

		// 各レイヤーが独立して存在することを確認
		for _, id := range layerIDs {
			layer := pls.GetTextLayer(id)
			if layer == nil {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5.2 (Never Reuses Existing Layer) failed: %v", err)
	}
}

// ============================================================
// Property 5.3: 同じ位置に複数のTextWriteを呼び出すと複数のレイヤーが作成される
// **Validates: Requirements 2.6, 5.1, 5.2**
// ============================================================

// TestProperty5_TextLayer_SamePositionCreatesMultipleLayers は同じ位置でのTextWriteが複数のレイヤーを作成することをテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 2.6, 5.1, 5.2**
func TestProperty5_TextLayer_SamePositionCreatesMultipleLayers(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 同じ位置に複数のTextWriteを呼び出すと、それぞれ新しいレイヤーが作成される
	property := func(params TextLayerTestParams, numCalls uint8) bool {
		lm := NewLayerManager()

		// PictureLayerSetを取得または作成
		pls := lm.GetOrCreatePictureLayerSet(params.PicID)
		if pls == nil {
			return false
		}

		// 呼び出し回数を制限（2-10）
		n := int(numCalls%9) + 2

		// 同じ位置に複数回TextWriteをシミュレート
		layerIDs := make([]int, n)
		for i := 0; i < n; i++ {
			layerID := lm.GetNextLayerID()
			// 同じ位置（params.X, params.Y）に異なるテキストを描画
			textLayerEntry := NewTextLayerEntry(layerID, params.PicID, params.X, params.Y, params.Text+string(rune('0'+i)), pls.GetNextTextZOffset())
			pls.AddTextLayer(textLayerEntry)
			layerIDs[i] = layerID
		}

		// レイヤー数が呼び出し回数と一致することを確認
		if pls.GetTextLayerCount() != n {
			return false
		}

		// すべてのレイヤーが同じ位置を持つことを確認
		for _, id := range layerIDs {
			layer := pls.GetTextLayer(id)
			if layer == nil {
				return false
			}
			x, y := layer.GetPosition()
			if x != params.X || y != params.Y {
				return false
			}
		}

		// すべてのレイヤーIDが一意であることを確認
		idSet := make(map[int]bool)
		for _, id := range layerIDs {
			if idSet[id] {
				return false
			}
			idSet[id] = true
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5.3 (Same Position Creates Multiple Layers) failed: %v", err)
	}
}

// ============================================================
// Property 5.4: Text_Layerは常にLayerTypeTextを返す
// **Validates: Requirements 2.6**
// ============================================================

// TestProperty5_TextLayer_AlwaysReturnsCorrectLayerType はText_LayerがLayerTypeTextを返すことをテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 2.6**
func TestProperty5_TextLayer_AlwaysReturnsCorrectLayerType(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のTextLayerEntryはGetLayerType()でLayerTypeTextを返す
	property := func(params TextLayerTestParams) bool {
		lm := NewLayerManager()

		// TextLayerEntryを作成
		layerID := lm.GetNextLayerID()
		textLayerEntry := NewTextLayerEntry(layerID, params.PicID, params.X, params.Y, params.Text, 0)

		// レイヤータイプがTextであることを確認
		if textLayerEntry.GetLayerType() != LayerTypeText {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5.4 (Always Returns Correct Layer Type) failed: %v", err)
	}
}

// ============================================================
// Property 5.5: 各TextWriteで新しいZ順序が割り当てられる
// **Validates: Requirements 2.6, 5.1**
// ============================================================

// TestProperty5_TextLayer_UniqueZOrderForEachLayer は各Text_Layerに一意のZ順序が割り当てられることをテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 2.6, 5.1**
func TestProperty5_TextLayer_UniqueZOrderForEachLayer(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 各TextWriteで新しいZ順序が割り当てられる
	property := func(params MultipleTextLayerTestParams) bool {
		lm := NewLayerManager()

		// PictureLayerSetを取得または作成
		pls := lm.GetOrCreatePictureLayerSet(params.PicID)
		if pls == nil {
			return false
		}

		// テキスト数を制限（2-10）
		numTexts := int(params.NumTexts%9) + 2

		// 複数のTextWriteをシミュレート
		zOrders := make([]int, numTexts)
		for i := 0; i < numTexts; i++ {
			layerID := lm.GetNextLayerID()
			textLayerEntry := NewTextLayerEntry(layerID, params.PicID, i*10, i*10, "Test"+string(rune('A'+i)), pls.GetNextTextZOffset())
			pls.AddTextLayer(textLayerEntry)
			zOrders[i] = textLayerEntry.GetZOrder()
		}

		// すべてのZ順序が一意であることを確認
		zOrderSet := make(map[int]bool)
		for _, z := range zOrders {
			if zOrderSet[z] {
				// 重複するZ順序がある
				return false
			}
			zOrderSet[z] = true
		}

		// Z順序が増加していることを確認
		for i := 1; i < len(zOrders); i++ {
			if zOrders[i] <= zOrders[i-1] {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5.5 (Unique Z Order For Each Layer) failed: %v", err)
	}
}

// ============================================================
// 統合プロパティテスト: Text_Layerの新規作成全体
// **Validates: Requirements 2.6, 5.1, 5.2**
// ============================================================

// TestProperty5_TextLayer_IntegratedBehavior はText_Layerの統合動作をテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 2.6, 5.1, 5.2**
func TestProperty5_TextLayer_IntegratedBehavior(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: TextWriteの一連の操作が正しく動作する
	property := func(params MultipleTextLayerTestParams) bool {
		lm := NewLayerManager()

		// PictureLayerSetを取得または作成
		pls := lm.GetOrCreatePictureLayerSet(params.PicID)
		if pls == nil {
			return false
		}

		// テキスト数を制限（1-10）
		numTexts := int(params.NumTexts%10) + 1

		// 作成されたレイヤーを追跡
		type layerInfo struct {
			id     int
			x, y   int
			text   string
			zOrder int
		}
		layers := make([]layerInfo, numTexts)

		// 複数のTextWriteをシミュレート
		for i := 0; i < numTexts; i++ {
			layerID := lm.GetNextLayerID()
			x := i * 10
			y := i * 10
			text := "Text" + string(rune('A'+i))

			textLayerEntry := NewTextLayerEntry(layerID, params.PicID, x, y, text, pls.GetNextTextZOffset())
			pls.AddTextLayer(textLayerEntry)

			layers[i] = layerInfo{
				id:     layerID,
				x:      x,
				y:      y,
				text:   text,
				zOrder: textLayerEntry.GetZOrder(),
			}
		}

		// 要件 5.1: 各TextWriteで新しいレイヤーが作成される
		if pls.GetTextLayerCount() != numTexts {
			return false
		}

		// 要件 5.2: 既存のレイヤーは再利用されない（すべてのIDが一意）
		idSet := make(map[int]bool)
		for _, info := range layers {
			if idSet[info.id] {
				return false
			}
			idSet[info.id] = true
		}

		// 要件 2.6: Text_Layerは常に新規作成される（各レイヤーが独立して存在）
		for _, info := range layers {
			layer := pls.GetTextLayer(info.id)
			if layer == nil {
				return false
			}

			// レイヤータイプがTextであることを確認
			if layer.GetLayerType() != LayerTypeText {
				return false
			}

			// 位置が正しいことを確認
			x, y := layer.GetPosition()
			if x != info.x || y != info.y {
				return false
			}

			// テキストが正しいことを確認
			if layer.GetText() != info.text {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5 (Integrated Behavior) failed: %v", err)
	}
}

// ============================================================
// 追加プロパティテスト: WindowLayerSetでのText_Layer管理
// **Validates: Requirements 2.6, 5.1, 5.2**
// ============================================================

// TestProperty5_TextLayer_WindowLayerSetIntegration はWindowLayerSetでのText_Layer管理をテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 2.6, 5.1, 5.2**
func TestProperty5_TextLayer_WindowLayerSetIntegration(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: WindowLayerSetでもText_Layerは常に新規作成される
	property := func(params TextLayerTestParams, numCalls uint8) bool {
		lm := NewLayerManager()
		bgColor := color.RGBA{0, 0, 0, 255}

		winWidth := int(params.WinWidth)
		winHeight := int(params.WinHeight)

		// WindowLayerSetを作成
		wls := lm.GetOrCreateWindowLayerSet(params.WinID, winWidth, winHeight, bgColor)
		if wls == nil {
			return false
		}

		// 呼び出し回数を制限（1-10）
		n := int(numCalls%10) + 1

		// 初期状態でレイヤーがないことを確認
		initialLayerCount := wls.GetLayerCount()
		if initialLayerCount != 0 {
			return false
		}

		// 複数のTextWriteをシミュレート（WindowLayerSetに追加）
		layerIDs := make([]int, n)
		for i := 0; i < n; i++ {
			layerID := lm.GetNextLayerID()
			textLayerEntry := NewTextLayerEntry(layerID, params.PicID, params.X+i*10, params.Y+i*10, params.Text+string(rune('0'+i)), 0)
			wls.AddLayer(textLayerEntry)
			layerIDs[i] = layerID
		}

		// レイヤー数が呼び出し回数と一致することを確認
		if wls.GetLayerCount() != n {
			return false
		}

		// すべてのレイヤーIDが一意であることを確認
		idSet := make(map[int]bool)
		for _, id := range layerIDs {
			if idSet[id] {
				return false
			}
			idSet[id] = true
		}

		// 各レイヤーが独立して存在することを確認
		for _, id := range layerIDs {
			layer := wls.GetLayer(id)
			if layer == nil {
				return false
			}
			// レイヤータイプがTextであることを確認
			if layer.GetLayerType() != LayerTypeText {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5 (WindowLayerSet Integration) failed: %v", err)
	}
}

// TestProperty5_TextLayer_ConsecutiveTextWriteAtSamePosition は同じ位置への連続TextWriteをテスト
// Property 5: Text_Layerの新規作成
// **Validates: Requirements 2.6, 5.1, 5.2**
func TestProperty5_TextLayer_ConsecutiveTextWriteAtSamePosition(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 同じ位置に連続してTextWriteを呼び出しても、各呼び出しで新しいレイヤーが作成される
	property := func(x, y int, numCalls uint8) bool {
		// 位置を正の値に制限
		if x < 0 {
			x = -x
		}
		if y < 0 {
			y = -y
		}

		lm := NewLayerManager()
		picID := 1

		// PictureLayerSetを取得または作成
		pls := lm.GetOrCreatePictureLayerSet(picID)
		if pls == nil {
			return false
		}

		// 呼び出し回数を制限（2-15）
		n := int(numCalls%14) + 2

		// 同じ位置に連続してTextWriteをシミュレート
		layerIDs := make([]int, n)
		for i := 0; i < n; i++ {
			layerID := lm.GetNextLayerID()
			// 同じ位置に異なるテキストを描画
			textLayerEntry := NewTextLayerEntry(layerID, picID, x, y, "Text"+string(rune('A'+i)), pls.GetNextTextZOffset())
			pls.AddTextLayer(textLayerEntry)
			layerIDs[i] = layerID
		}

		// 要件 5.1: 各TextWriteで新しいレイヤーが作成される
		if pls.GetTextLayerCount() != n {
			return false
		}

		// 要件 5.2: 既存のレイヤーは再利用されない
		idSet := make(map[int]bool)
		for _, id := range layerIDs {
			if idSet[id] {
				return false
			}
			idSet[id] = true
		}

		// すべてのレイヤーが同じ位置を持つことを確認
		for _, id := range layerIDs {
			layer := pls.GetTextLayer(id)
			if layer == nil {
				return false
			}
			lx, ly := layer.GetPosition()
			if lx != x || ly != y {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 5 (Consecutive TextWrite At Same Position) failed: %v", err)
	}
}
