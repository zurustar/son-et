package graphics

import (
	"math/rand"
	"reflect"
	"testing"
	"testing/quick"
)

// ============================================================
// Property 3: レイヤータイプの識別
// *任意の*レイヤーに対して、GetLayerType()は正しいレイヤータイプ（Picture、Text、Cast）を返す。
// **Validates: Requirements 2.4**
// ============================================================

// ============================================================
// PictureLayerのプロパティテスト
// ============================================================

// TestProperty3_PictureLayerType_AlwaysReturnsPicture はPictureLayerが常にLayerTypePictureを返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_PictureLayerType_AlwaysReturnsPicture(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のパラメータでPictureLayerを作成した場合、
	// GetLayerType()は常にLayerTypePictureを返す
	property := func(id int, width, height uint16) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		// サイズは最小1以上、最大2000以下
		w := int(width%2000) + 1
		h := int(height%2000) + 1

		layer := NewPictureLayer(id, w, h)
		if layer == nil {
			return false
		}

		// GetLayerType()がLayerTypePictureを返すことを確認
		if layer.GetLayerType() != LayerTypePicture {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (PictureLayer Type) failed: %v", err)
	}
}

// TestProperty3_PictureLayerType_ConsistentAcrossMultipleCalls はGetLayerType()が複数回呼び出しても一貫した結果を返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_PictureLayerType_ConsistentAcrossMultipleCalls(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: GetLayerType()を複数回呼び出しても、常に同じ結果を返す
	property := func(id int, width, height uint16, numCalls uint8) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		// サイズは最小1以上
		w := int(width%1000) + 1
		h := int(height%1000) + 1

		// 呼び出し回数は1-50回
		calls := int(numCalls%50) + 1

		layer := NewPictureLayer(id, w, h)
		if layer == nil {
			return false
		}

		// 複数回呼び出して一貫性を確認
		firstResult := layer.GetLayerType()
		for i := 0; i < calls; i++ {
			if layer.GetLayerType() != firstResult {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (PictureLayer Consistency) failed: %v", err)
	}
}

// ============================================================
// CastLayerのプロパティテスト
// ============================================================

// TestProperty3_CastLayerType_AlwaysReturnsCast はCastLayerが常にLayerTypeCastを返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_CastLayerType_AlwaysReturnsCast(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のパラメータでCastLayerを作成した場合、
	// GetLayerType()は常にLayerTypeCastを返す
	property := func(id, castID, picID, srcPicID int, x, y, srcX, srcY int, width, height uint16, zOrderOffset int) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}
		if castID < 0 {
			castID = -castID
		}
		if picID < 0 {
			picID = -picID
		}
		if srcPicID < 0 {
			srcPicID = -srcPicID
		}
		if zOrderOffset < 0 {
			zOrderOffset = -zOrderOffset
		}

		// サイズは最小1以上
		w := int(width%1000) + 1
		h := int(height%1000) + 1

		layer := NewCastLayer(id, castID, picID, srcPicID, x, y, srcX, srcY, w, h, zOrderOffset)
		if layer == nil {
			return false
		}

		// GetLayerType()がLayerTypeCastを返すことを確認
		if layer.GetLayerType() != LayerTypeCast {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (CastLayer Type) failed: %v", err)
	}
}

// TestProperty3_CastLayerType_ConsistentAcrossMultipleCalls はGetLayerType()が複数回呼び出しても一貫した結果を返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_CastLayerType_ConsistentAcrossMultipleCalls(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: GetLayerType()を複数回呼び出しても、常に同じ結果を返す
	property := func(id int, numCalls uint8) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		// 呼び出し回数は1-50回
		calls := int(numCalls%50) + 1

		layer := NewCastLayer(id, 1, 1, 1, 0, 0, 0, 0, 100, 100, 0)
		if layer == nil {
			return false
		}

		// 複数回呼び出して一貫性を確認
		firstResult := layer.GetLayerType()
		for i := 0; i < calls; i++ {
			if layer.GetLayerType() != firstResult {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (CastLayer Consistency) failed: %v", err)
	}
}

// ============================================================
// TextLayerEntryのプロパティテスト
// ============================================================

// TestProperty3_TextLayerType_AlwaysReturnsText はTextLayerEntryが常にLayerTypeTextを返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_TextLayerType_AlwaysReturnsText(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のパラメータでTextLayerEntryを作成した場合、
	// GetLayerType()は常にLayerTypeTextを返す
	property := func(id, picID, x, y, zOrderOffset int, text string) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}
		if picID < 0 {
			picID = -picID
		}
		if zOrderOffset < 0 {
			zOrderOffset = -zOrderOffset
		}

		layer := NewTextLayerEntry(id, picID, x, y, text, zOrderOffset)
		if layer == nil {
			return false
		}

		// GetLayerType()がLayerTypeTextを返すことを確認
		if layer.GetLayerType() != LayerTypeText {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (TextLayerEntry Type) failed: %v", err)
	}
}

// TestProperty3_TextLayerType_ConsistentAcrossMultipleCalls はGetLayerType()が複数回呼び出しても一貫した結果を返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_TextLayerType_ConsistentAcrossMultipleCalls(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: GetLayerType()を複数回呼び出しても、常に同じ結果を返す
	property := func(id int, text string, numCalls uint8) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		// 呼び出し回数は1-50回
		calls := int(numCalls%50) + 1

		layer := NewTextLayerEntry(id, 1, 0, 0, text, 0)
		if layer == nil {
			return false
		}

		// 複数回呼び出して一貫性を確認
		firstResult := layer.GetLayerType()
		for i := 0; i < calls; i++ {
			if layer.GetLayerType() != firstResult {
				return false
			}
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (TextLayerEntry Consistency) failed: %v", err)
	}
}

// ============================================================
// 全レイヤータイプの統合プロパティテスト
// ============================================================

// LayerTypeTestCase はレイヤータイプテストのケースを表す
type LayerTypeTestCase struct {
	LayerKind int // 0: Picture, 1: Cast, 2: Text
	ID        int
	Width     uint16
	Height    uint16
}

// Generate はLayerTypeTestCaseのランダム生成を実装する
func (LayerTypeTestCase) Generate(rand *rand.Rand, size int) reflect.Value {
	tc := LayerTypeTestCase{
		LayerKind: rand.Intn(3),
		ID:        rand.Intn(10000),
		Width:     uint16(rand.Intn(1000) + 1),
		Height:    uint16(rand.Intn(1000) + 1),
	}
	return reflect.ValueOf(tc)
}

// TestProperty3_AllLayerTypes_CorrectTypeIdentification は全レイヤータイプが正しく識別されることをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_AllLayerTypes_CorrectTypeIdentification(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のレイヤータイプに対して、GetLayerType()は正しいタイプを返す
	property := func(tc LayerTypeTestCase) bool {
		var layer Layer
		var expectedType LayerType

		switch tc.LayerKind {
		case 0: // PictureLayer
			layer = NewPictureLayer(tc.ID, int(tc.Width), int(tc.Height))
			expectedType = LayerTypePicture
		case 1: // CastLayer
			layer = NewCastLayer(tc.ID, 1, 1, 1, 0, 0, 0, 0, int(tc.Width), int(tc.Height), 0)
			expectedType = LayerTypeCast
		case 2: // TextLayerEntry
			layer = NewTextLayerEntry(tc.ID, 1, 0, 0, "test", 0)
			expectedType = LayerTypeText
		default:
			return true // 無効なケースはスキップ
		}

		if layer == nil {
			return false
		}

		// GetLayerType()が期待されるタイプを返すことを確認
		if layer.GetLayerType() != expectedType {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (All Layer Types) failed: %v", err)
	}
}

// TestProperty3_LayerTypeDistinct_DifferentTypesReturnDifferentValues は異なるレイヤータイプが異なる値を返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_LayerTypeDistinct_DifferentTypesReturnDifferentValues(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 異なるレイヤータイプは異なるGetLayerType()値を返す
	property := func(id int) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		// 各タイプのレイヤーを作成
		pictureLayer := NewPictureLayer(id, 100, 100)
		castLayer := NewCastLayer(id+1, 1, 1, 1, 0, 0, 0, 0, 100, 100, 0)
		textLayer := NewTextLayerEntry(id+2, 1, 0, 0, "test", 0)

		if pictureLayer == nil || castLayer == nil || textLayer == nil {
			return false
		}

		// 各タイプが異なる値を返すことを確認
		pictureType := pictureLayer.GetLayerType()
		castType := castLayer.GetLayerType()
		textType := textLayer.GetLayerType()

		// すべて異なることを確認
		if pictureType == castType || pictureType == textType || castType == textType {
			return false
		}

		// 期待される値であることを確認
		if pictureType != LayerTypePicture {
			return false
		}
		if castType != LayerTypeCast {
			return false
		}
		if textType != LayerTypeText {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (Layer Type Distinct) failed: %v", err)
	}
}

// ============================================================
// レイヤー状態変更後のタイプ一貫性テスト
// ============================================================

// TestProperty3_PictureLayerType_ConsistentAfterStateChanges はPictureLayerの状態変更後もタイプが一貫していることをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_PictureLayerType_ConsistentAfterStateChanges(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 状態変更後もGetLayerType()は同じ値を返す
	property := func(id int, newZOrder int, visible bool) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		layer := NewPictureLayer(id, 640, 480)
		if layer == nil {
			return false
		}

		// 初期タイプを確認
		initialType := layer.GetLayerType()
		if initialType != LayerTypePicture {
			return false
		}

		// 状態を変更
		layer.SetZOrder(newZOrder)
		layer.SetVisible(visible)
		layer.SetDirty(true)
		layer.Invalidate()

		// 状態変更後もタイプが同じであることを確認
		if layer.GetLayerType() != initialType {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (PictureLayer State Changes) failed: %v", err)
	}
}

// TestProperty3_CastLayerType_ConsistentAfterStateChanges はCastLayerの状態変更後もタイプが一貫していることをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_CastLayerType_ConsistentAfterStateChanges(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 状態変更後もGetLayerType()は同じ値を返す
	property := func(id int, newX, newY int, visible bool) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		layer := NewCastLayer(id, 1, 1, 1, 0, 0, 0, 0, 100, 100, 0)
		if layer == nil {
			return false
		}

		// 初期タイプを確認
		initialType := layer.GetLayerType()
		if initialType != LayerTypeCast {
			return false
		}

		// 状態を変更
		layer.SetPosition(newX, newY)
		layer.SetVisible(visible)
		layer.SetDirty(true)
		layer.Invalidate()

		// 状態変更後もタイプが同じであることを確認
		if layer.GetLayerType() != initialType {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (CastLayer State Changes) failed: %v", err)
	}
}

// TestProperty3_TextLayerType_ConsistentAfterStateChanges はTextLayerEntryの状態変更後もタイプが一貫していることをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_TextLayerType_ConsistentAfterStateChanges(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 状態変更後もGetLayerType()は同じ値を返す
	property := func(id int, newX, newY int, newText string, visible bool) bool {
		// 負のIDは正に変換
		if id < 0 {
			id = -id
		}

		layer := NewTextLayerEntry(id, 1, 0, 0, "initial", 0)
		if layer == nil {
			return false
		}

		// 初期タイプを確認
		initialType := layer.GetLayerType()
		if initialType != LayerTypeText {
			return false
		}

		// 状態を変更
		layer.SetPosition(newX, newY)
		layer.SetText(newText)
		layer.SetVisible(visible)
		layer.SetDirty(true)
		layer.Invalidate()

		// 状態変更後もタイプが同じであることを確認
		if layer.GetLayerType() != initialType {
			return false
		}

		return true
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (TextLayerEntry State Changes) failed: %v", err)
	}
}

// ============================================================
// LayerType定数の一貫性テスト
// ============================================================

// TestProperty3_LayerTypeConstants_AreDistinct はLayerType定数が互いに異なることをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_LayerTypeConstants_AreDistinct(t *testing.T) {
	// LayerType定数が互いに異なることを確認
	types := []LayerType{LayerTypePicture, LayerTypeText, LayerTypeCast}

	for i := 0; i < len(types); i++ {
		for j := i + 1; j < len(types); j++ {
			if types[i] == types[j] {
				t.Errorf("LayerType constants are not distinct: %v == %v", types[i], types[j])
			}
		}
	}
}

// TestProperty3_LayerTypeString_ReturnsValidString はLayerType.String()が有効な文字列を返すことをテストする
// Property 3: レイヤータイプの識別
// **Validates: Requirements 2.4**
func TestProperty3_LayerTypeString_ReturnsValidString(t *testing.T) {
	config := &quick.Config{
		MaxCount: 100,
	}

	// プロパティ: 任意のLayerTypeに対して、String()は空でない文字列を返す
	property := func(typeVal int) bool {
		lt := LayerType(typeVal % 4) // 0-3の範囲（3はUnknown）

		str := lt.String()
		if str == "" {
			return false
		}

		// 既知のタイプは特定の文字列を返す
		switch lt {
		case LayerTypePicture:
			return str == "Picture"
		case LayerTypeText:
			return str == "Text"
		case LayerTypeCast:
			return str == "Cast"
		default:
			return str == "Unknown"
		}
	}

	if err := quick.Check(property, config); err != nil {
		t.Errorf("Property 3 (LayerType String) failed: %v", err)
	}
}
