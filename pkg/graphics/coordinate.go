// Package graphics provides sprite-based rendering system.
package graphics

// 座標系の概要
//
// 本システムでは以下の座標系を使用する：
//
// 1. 仮想デスクトップ座標系 (Virtual Desktop Coordinates)
//    - 原点: 仮想デスクトップの左上 (0, 0)
//    - 範囲: (0, 0) から (virtualWidth-1, virtualHeight-1)
//    - 用途: ウィンドウの位置 (win.X, win.Y)
//
// 2. ウィンドウコンテンツ座標系 (Window Content Coordinates)
//    - 原点: ウィンドウのコンテンツ領域の左上
//    - 変換: contentX = win.X + borderThickness
//            contentY = win.Y + borderThickness + titleBarHeight
//    - 用途: ウィンドウ内の描画要素の基準位置
//
// 3. ピクチャー座標系 (Picture Coordinates)
//    - 原点: ピクチャー画像の左上 (0, 0)
//    - 範囲: (0, 0) から (picWidth-1, picHeight-1)
//    - 用途: キャスト、テキスト、図形の位置
//    - オフセット: PicX/PicYはピクチャーの表示オフセット
//      - PicX > 0: ピクチャーが左にシフト（右側が見える）
//      - PicX < 0: ピクチャーが右にシフト（左側が見える）
//
// 4. スクリーン座標系 (Screen Coordinates)
//    - 原点: 実際のウィンドウの左上 (0, 0)
//    - 範囲: (0, 0) から (screenWidth-1, screenHeight-1)
//    - 用途: 最終的な描画位置
//
// 座標変換の流れ:
//
//   ピクチャー座標 (x, y)
//         ↓
//   PicX/PicYオフセット適用: (x - PicX, y - PicY)
//         ↓
//   コンテンツ領域オフセット適用: (contentX + x - PicX, contentY + y - PicY)
//         ↓
//   スクリーン座標 (screenX, screenY)
//
// 注意事項:
// - スプライトの位置はピクチャー座標系で設定される
// - 描画時にPicX/PicYオフセットとコンテンツ領域オフセットが適用される
// - オフセットは一箇所（CoordinateConverter）で管理し、二重適用を防ぐ

// CoordinateConverter は座標変換を行うユーティリティ構造体
// 座標変換のロジックを一箇所に集約し、一貫性を保つ
type CoordinateConverter struct {
	// ウィンドウ装飾の定数
	borderThickness int
	titleBarHeight  int
}

// NewCoordinateConverter は新しいCoordinateConverterを作成する
func NewCoordinateConverter() *CoordinateConverter {
	return &CoordinateConverter{
		borderThickness: BorderThickness,
		titleBarHeight:  TitleBarHeight,
	}
}

// ウィンドウ装飾の定数
const (
	// BorderThickness はウィンドウ外枠の幅
	BorderThickness = 4

	// TitleBarHeight はタイトルバーの高さ
	TitleBarHeight = 20
)

// GetContentOffset はウィンドウのコンテンツ領域のオフセットを返す
// コンテンツ領域は、ウィンドウ装飾（枠、タイトルバー）の内側の領域
func (cc *CoordinateConverter) GetContentOffset() (int, int) {
	return cc.borderThickness, cc.borderThickness + cc.titleBarHeight
}

// WindowToContent はウィンドウ座標からコンテンツ領域座標に変換する
// winX, winY: ウィンドウの位置（仮想デスクトップ座標系）
// 戻り値: コンテンツ領域の左上座標（仮想デスクトップ座標系）
func (cc *CoordinateConverter) WindowToContent(winX, winY int) (int, int) {
	contentX := winX + cc.borderThickness
	contentY := winY + cc.borderThickness + cc.titleBarHeight
	return contentX, contentY
}

// PictureToScreen はピクチャー座標からスクリーン座標に変換する
// picX, picY: ピクチャー座標系での位置
// contentX, contentY: コンテンツ領域の左上座標（仮想デスクトップ座標系）
// picOffsetX, picOffsetY: ピクチャーの表示オフセット（PicX, PicY）
// 戻り値: スクリーン座標
//
// 変換式:
//
//	screenX = contentX + (picX - picOffsetX)
//	screenY = contentY + (picY - picOffsetY)
//
// 注意: picOffsetX/picOffsetYは正の値の場合、ピクチャーが左/上にシフトする
// つまり、描画位置は右/下にシフトする（-picOffsetX, -picOffsetY）
func (cc *CoordinateConverter) PictureToScreen(picX, picY int, contentX, contentY int, picOffsetX, picOffsetY int) (int, int) {
	screenX := contentX + picX - picOffsetX
	screenY := contentY + picY - picOffsetY
	return screenX, screenY
}

// PictureToScreenFloat はピクチャー座標（float64）からスクリーン座標に変換する
// スプライトの位置がfloat64で管理されている場合に使用
func (cc *CoordinateConverter) PictureToScreenFloat(picX, picY float64, contentX, contentY int, picOffsetX, picOffsetY int) (float64, float64) {
	screenX := float64(contentX) + picX - float64(picOffsetX)
	screenY := float64(contentY) + picY - float64(picOffsetY)
	return screenX, screenY
}

// ScreenToPicture はスクリーン座標からピクチャー座標に変換する（逆変換）
// screenX, screenY: スクリーン座標
// contentX, contentY: コンテンツ領域の左上座標（仮想デスクトップ座標系）
// picOffsetX, picOffsetY: ピクチャーの表示オフセット（PicX, PicY）
// 戻り値: ピクチャー座標系での位置
func (cc *CoordinateConverter) ScreenToPicture(screenX, screenY int, contentX, contentY int, picOffsetX, picOffsetY int) (int, int) {
	picX := screenX - contentX + picOffsetX
	picY := screenY - contentY + picOffsetY
	return picX, picY
}

// GetPicOffset はウィンドウのピクチャーオフセットを取得する
// PicX/PicYはピクチャーの表示オフセット
// - PicX > 0: ピクチャーが左にシフト（右側が見える）
// - PicX < 0: ピクチャーが右にシフト（左側が見える）
func GetPicOffset(win *Window) (int, int) {
	if win == nil {
		return 0, 0
	}
	return win.PicX, win.PicY
}

// CalculateDrawPosition はスプライトの描画位置を計算する
// これは座標変換の一連の処理をまとめたヘルパー関数
// win: ウィンドウ情報
// spriteX, spriteY: スプライトの位置（ピクチャー座標系）
// 戻り値: スクリーン座標
func (cc *CoordinateConverter) CalculateDrawPosition(win *Window, spriteX, spriteY float64) (float64, float64) {
	// 1. コンテンツ領域の位置を計算
	contentX, contentY := cc.WindowToContent(win.X, win.Y)

	// 2. ピクチャーオフセットを取得
	picOffsetX, picOffsetY := GetPicOffset(win)

	// 3. スクリーン座標に変換
	screenX, screenY := cc.PictureToScreenFloat(spriteX, spriteY, contentX, contentY, picOffsetX, picOffsetY)

	return screenX, screenY
}

// CalculateDrawPositionInt はスプライトの描画位置を計算する（int版）
func (cc *CoordinateConverter) CalculateDrawPositionInt(win *Window, spriteX, spriteY int) (int, int) {
	// 1. コンテンツ領域の位置を計算
	contentX, contentY := cc.WindowToContent(win.X, win.Y)

	// 2. ピクチャーオフセットを取得
	picOffsetX, picOffsetY := GetPicOffset(win)

	// 3. スクリーン座標に変換
	screenX, screenY := cc.PictureToScreen(spriteX, spriteY, contentX, contentY, picOffsetX, picOffsetY)

	return screenX, screenY
}

// defaultCoordinateConverter はデフォルトのCoordinateConverter
var defaultCoordinateConverter = NewCoordinateConverter()

// GetDefaultCoordinateConverter はデフォルトのCoordinateConverterを返す
func GetDefaultCoordinateConverter() *CoordinateConverter {
	return defaultCoordinateConverter
}
