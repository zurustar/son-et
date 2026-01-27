package graphics

import (
	"image"
	"image/color"
	"math/rand"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
)

// SceneChangeMode はシーンチェンジのモードを表す
// 要件 13.1-13.11
type SceneChangeMode int

const (
	// SceneChangeNone は通常コピー（mode=0）
	SceneChangeNone SceneChangeMode = 0
	// SceneChangeTransparent は透明色除外（mode=1）
	SceneChangeTransparent SceneChangeMode = 1
	// SceneChangeWipeDown は上から下へのワイプ（mode=2）
	SceneChangeWipeDown SceneChangeMode = 2
	// SceneChangeWipeRight は左から右へのワイプ（mode=3）
	SceneChangeWipeRight SceneChangeMode = 3
	// SceneChangeWipeLeft は右から左へのワイプ（mode=4）
	SceneChangeWipeLeft SceneChangeMode = 4
	// SceneChangeWipeUp は下から上へのワイプ（mode=5）
	SceneChangeWipeUp SceneChangeMode = 5
	// SceneChangeWipeOut は中央から外側へのワイプ（mode=6）
	SceneChangeWipeOut SceneChangeMode = 6
	// SceneChangeWipeIn は外側から中央へのワイプ（mode=7）
	SceneChangeWipeIn SceneChangeMode = 7
	// SceneChangeRandom はランダムブロック（mode=8）
	SceneChangeRandom SceneChangeMode = 8
	// SceneChangeFade はフェード（mode=9）
	SceneChangeFade SceneChangeMode = 9
)

// DefaultSceneChangeSpeed はデフォルトのシーンチェンジ速度
// 値が大きいほど速い（1フレームあたりの進捗率）
const DefaultSceneChangeSpeed = 5

// RandomBlockSize はランダムブロックエフェクトのブロックサイズ
const RandomBlockSize = 16

// SceneChange はシーンチェンジを管理する
// 要件 13.11: シーンチェンジは非同期で実行される
type SceneChange struct {
	// ソースとデスティネーション
	srcImage *ebiten.Image
	dstImage *ebiten.Image
	srcRect  image.Rectangle
	dstPoint image.Point

	// エフェクト設定
	mode  SceneChangeMode
	speed int // 1-100の範囲、大きいほど速い

	// 進捗状態
	progress  float64 // 0.0 - 1.0
	completed bool

	// ランダムブロック用の状態
	blockOrder []int // ブロックの描画順序（シャッフル済み）
	blockCount int   // 総ブロック数

	mu sync.Mutex
}

// NewSceneChange は新しいSceneChangeを作成する
func NewSceneChange(
	srcImage *ebiten.Image,
	dstImage *ebiten.Image,
	srcRect image.Rectangle,
	dstPoint image.Point,
	mode SceneChangeMode,
	speed int,
) *SceneChange {
	if speed <= 0 {
		speed = DefaultSceneChangeSpeed
	}
	if speed > 100 {
		speed = 100
	}

	sc := &SceneChange{
		srcImage:  srcImage,
		dstImage:  dstImage,
		srcRect:   srcRect,
		dstPoint:  dstPoint,
		mode:      mode,
		speed:     speed,
		progress:  0.0,
		completed: false,
	}

	// ランダムブロックモードの場合、ブロック順序を初期化
	if mode == SceneChangeRandom {
		sc.initRandomBlocks()
	}

	return sc
}

// initRandomBlocks はランダムブロックの描画順序を初期化する
func (sc *SceneChange) initRandomBlocks() {
	width := sc.srcRect.Dx()
	height := sc.srcRect.Dy()

	// ブロック数を計算
	blocksX := (width + RandomBlockSize - 1) / RandomBlockSize
	blocksY := (height + RandomBlockSize - 1) / RandomBlockSize
	sc.blockCount = blocksX * blocksY

	// ブロックインデックスを作成
	sc.blockOrder = make([]int, sc.blockCount)
	for i := 0; i < sc.blockCount; i++ {
		sc.blockOrder[i] = i
	}

	// Fisher-Yatesシャッフル
	for i := sc.blockCount - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		sc.blockOrder[i], sc.blockOrder[j] = sc.blockOrder[j], sc.blockOrder[i]
	}
}

// Update はシーンチェンジの進捗を更新する
// 完了したらtrueを返す
func (sc *SceneChange) Update() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if sc.completed {
		return true
	}

	// 進捗を更新（speedに基づいて）
	// speed=1で1フレームあたり1%、speed=100で1フレームあたり100%
	sc.progress += float64(sc.speed) / 100.0

	if sc.progress >= 1.0 {
		sc.progress = 1.0
		sc.completed = true
	}

	return sc.completed
}

// Apply は現在の進捗に基づいてエフェクトを適用する
func (sc *SceneChange) Apply() {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	switch sc.mode {
	case SceneChangeWipeDown:
		sc.applyWipeDown()
	case SceneChangeWipeRight:
		sc.applyWipeRight()
	case SceneChangeWipeLeft:
		sc.applyWipeLeft()
	case SceneChangeWipeUp:
		sc.applyWipeUp()
	case SceneChangeWipeOut:
		sc.applyWipeOut()
	case SceneChangeWipeIn:
		sc.applyWipeIn()
	case SceneChangeRandom:
		sc.applyRandom()
	case SceneChangeFade:
		sc.applyFade()
	default:
		// 通常コピー（mode=0）または透明色除外（mode=1）
		// これらはSceneChangeではなく直接処理される
		sc.applyNormal()
	}
}

// applyNormal は通常コピーを適用する
func (sc *SceneChange) applyNormal() {
	subImg := sc.srcImage.SubImage(sc.srcRect).(*ebiten.Image)
	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y))
	sc.dstImage.DrawImage(subImg, opts)
}

// applyWipeDown は上から下へのワイプを適用する
// 要件 13.2
func (sc *SceneChange) applyWipeDown() {
	height := sc.srcRect.Dy()
	currentHeight := int(float64(height) * sc.progress)

	if currentHeight <= 0 {
		return
	}

	// 上から現在の高さまでの領域を描画
	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y
	width := sc.srcRect.Dx()

	clipRect := image.Rect(srcX, srcY, srcX+width, srcY+currentHeight)
	subImg := sc.srcImage.SubImage(clipRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y))
	sc.dstImage.DrawImage(subImg, opts)
}

// applyWipeRight は左から右へのワイプを適用する
// 要件 13.3
func (sc *SceneChange) applyWipeRight() {
	width := sc.srcRect.Dx()
	currentWidth := int(float64(width) * sc.progress)

	if currentWidth <= 0 {
		return
	}

	// 左から現在の幅までの領域を描画
	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y
	height := sc.srcRect.Dy()

	clipRect := image.Rect(srcX, srcY, srcX+currentWidth, srcY+height)
	subImg := sc.srcImage.SubImage(clipRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y))
	sc.dstImage.DrawImage(subImg, opts)
}

// applyWipeLeft は右から左へのワイプを適用する
// 要件 13.4
func (sc *SceneChange) applyWipeLeft() {
	width := sc.srcRect.Dx()
	currentWidth := int(float64(width) * sc.progress)

	if currentWidth <= 0 {
		return
	}

	// 右から現在の幅までの領域を描画
	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y
	height := sc.srcRect.Dy()

	// 右端から描画開始
	startX := srcX + width - currentWidth
	clipRect := image.Rect(startX, srcY, srcX+width, srcY+height)
	subImg := sc.srcImage.SubImage(clipRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	// デスティネーションも右端から
	dstStartX := sc.dstPoint.X + width - currentWidth
	opts.GeoM.Translate(float64(dstStartX), float64(sc.dstPoint.Y))
	sc.dstImage.DrawImage(subImg, opts)
}

// applyWipeUp は下から上へのワイプを適用する
// 要件 13.5
func (sc *SceneChange) applyWipeUp() {
	height := sc.srcRect.Dy()
	currentHeight := int(float64(height) * sc.progress)

	if currentHeight <= 0 {
		return
	}

	// 下から現在の高さまでの領域を描画
	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y
	width := sc.srcRect.Dx()

	// 下端から描画開始
	startY := srcY + height - currentHeight
	clipRect := image.Rect(srcX, startY, srcX+width, srcY+height)
	subImg := sc.srcImage.SubImage(clipRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	// デスティネーションも下端から
	dstStartY := sc.dstPoint.Y + height - currentHeight
	opts.GeoM.Translate(float64(sc.dstPoint.X), float64(dstStartY))
	sc.dstImage.DrawImage(subImg, opts)
}

// applyWipeOut は中央から外側へのワイプを適用する
// 要件 13.6
func (sc *SceneChange) applyWipeOut() {
	width := sc.srcRect.Dx()
	height := sc.srcRect.Dy()

	// 中央から外側に広がる矩形
	currentWidth := int(float64(width) * sc.progress)
	currentHeight := int(float64(height) * sc.progress)

	if currentWidth <= 0 || currentHeight <= 0 {
		return
	}

	// 中央を基準に計算
	centerX := width / 2
	centerY := height / 2
	halfW := currentWidth / 2
	halfH := currentHeight / 2

	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y

	// クリップ領域を計算
	clipX := srcX + centerX - halfW
	clipY := srcY + centerY - halfH
	clipW := currentWidth
	clipH := currentHeight

	// 境界チェック
	if clipX < srcX {
		clipX = srcX
	}
	if clipY < srcY {
		clipY = srcY
	}
	if clipX+clipW > srcX+width {
		clipW = srcX + width - clipX
	}
	if clipY+clipH > srcY+height {
		clipH = srcY + height - clipY
	}

	clipRect := image.Rect(clipX, clipY, clipX+clipW, clipY+clipH)
	subImg := sc.srcImage.SubImage(clipRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	dstX := sc.dstPoint.X + (clipX - srcX)
	dstY := sc.dstPoint.Y + (clipY - srcY)
	opts.GeoM.Translate(float64(dstX), float64(dstY))
	sc.dstImage.DrawImage(subImg, opts)
}

// applyWipeIn は外側から中央へのワイプを適用する
// 要件 13.7
func (sc *SceneChange) applyWipeIn() {
	width := sc.srcRect.Dx()
	height := sc.srcRect.Dy()

	// 外側から中央に向かって縮小する矩形の「外側」を描画
	// progress=0で全体、progress=1で中央のみ（何も描画しない）
	// 実際には、4つの矩形（上、下、左、右）を描画する

	// 現在の「穴」のサイズ
	holeWidth := int(float64(width) * sc.progress)
	holeHeight := int(float64(height) * sc.progress)

	centerX := width / 2
	centerY := height / 2
	halfHoleW := holeWidth / 2
	halfHoleH := holeHeight / 2

	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y

	// 上部の矩形
	if centerY-halfHoleH > 0 {
		topRect := image.Rect(srcX, srcY, srcX+width, srcY+centerY-halfHoleH)
		topImg := sc.srcImage.SubImage(topRect).(*ebiten.Image)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y))
		sc.dstImage.DrawImage(topImg, opts)
	}

	// 下部の矩形
	if centerY+halfHoleH < height {
		bottomY := srcY + centerY + halfHoleH
		bottomRect := image.Rect(srcX, bottomY, srcX+width, srcY+height)
		bottomImg := sc.srcImage.SubImage(bottomRect).(*ebiten.Image)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y+centerY+halfHoleH))
		sc.dstImage.DrawImage(bottomImg, opts)
	}

	// 左部の矩形（上下の間）
	if centerX-halfHoleW > 0 {
		leftY := centerY - halfHoleH
		leftH := holeHeight
		if leftY < 0 {
			leftY = 0
		}
		if leftY+leftH > height {
			leftH = height - leftY
		}
		if leftH > 0 {
			leftRect := image.Rect(srcX, srcY+leftY, srcX+centerX-halfHoleW, srcY+leftY+leftH)
			leftImg := sc.srcImage.SubImage(leftRect).(*ebiten.Image)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y+leftY))
			sc.dstImage.DrawImage(leftImg, opts)
		}
	}

	// 右部の矩形（上下の間）
	if centerX+halfHoleW < width {
		rightX := centerX + halfHoleW
		rightY := centerY - halfHoleH
		rightH := holeHeight
		if rightY < 0 {
			rightY = 0
		}
		if rightY+rightH > height {
			rightH = height - rightY
		}
		if rightH > 0 {
			rightRect := image.Rect(srcX+rightX, srcY+rightY, srcX+width, srcY+rightY+rightH)
			rightImg := sc.srcImage.SubImage(rightRect).(*ebiten.Image)
			opts := &ebiten.DrawImageOptions{}
			opts.GeoM.Translate(float64(sc.dstPoint.X+rightX), float64(sc.dstPoint.Y+rightY))
			sc.dstImage.DrawImage(rightImg, opts)
		}
	}
}

// applyRandom はランダムブロックエフェクトを適用する
// 要件 13.8
func (sc *SceneChange) applyRandom() {
	if sc.blockCount == 0 {
		return
	}

	width := sc.srcRect.Dx()
	height := sc.srcRect.Dy()
	blocksX := (width + RandomBlockSize - 1) / RandomBlockSize

	// 現在の進捗に基づいて描画するブロック数を計算
	blocksToShow := int(float64(sc.blockCount) * sc.progress)

	srcX := sc.srcRect.Min.X
	srcY := sc.srcRect.Min.Y

	// シャッフルされた順序でブロックを描画
	for i := 0; i < blocksToShow && i < len(sc.blockOrder); i++ {
		blockIdx := sc.blockOrder[i]

		// ブロックの位置を計算
		blockX := (blockIdx % blocksX) * RandomBlockSize
		blockY := (blockIdx / blocksX) * RandomBlockSize

		// ブロックのサイズを計算（端のブロックは小さくなる可能性）
		blockW := RandomBlockSize
		blockH := RandomBlockSize
		if blockX+blockW > width {
			blockW = width - blockX
		}
		if blockY+blockH > height {
			blockH = height - blockY
		}

		if blockW <= 0 || blockH <= 0 {
			continue
		}

		// ブロックを描画
		blockRect := image.Rect(srcX+blockX, srcY+blockY, srcX+blockX+blockW, srcY+blockY+blockH)
		blockImg := sc.srcImage.SubImage(blockRect).(*ebiten.Image)

		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(sc.dstPoint.X+blockX), float64(sc.dstPoint.Y+blockY))
		sc.dstImage.DrawImage(blockImg, opts)
	}
}

// applyFade はフェードエフェクトを適用する
// 要件 13.9
func (sc *SceneChange) applyFade() {
	subImg := sc.srcImage.SubImage(sc.srcRect).(*ebiten.Image)

	opts := &ebiten.DrawImageOptions{}
	opts.GeoM.Translate(float64(sc.dstPoint.X), float64(sc.dstPoint.Y))

	// アルファ値を進捗に基づいて設定
	// progress=0でアルファ=0（透明）、progress=1でアルファ=1（不透明）
	opts.ColorScale.ScaleAlpha(float32(sc.progress))

	sc.dstImage.DrawImage(subImg, opts)
}

// IsCompleted はシーンチェンジが完了したかどうかを返す
func (sc *SceneChange) IsCompleted() bool {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.completed
}

// GetProgress は現在の進捗を返す（0.0-1.0）
func (sc *SceneChange) GetProgress() float64 {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	return sc.progress
}

// GetMode はシーンチェンジモードを返す
func (sc *SceneChange) GetMode() SceneChangeMode {
	return sc.mode
}

// Complete はシーンチェンジを即座に完了させる
func (sc *SceneChange) Complete() {
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.progress = 1.0
	sc.completed = true
}

// SceneChangeManager はアクティブなシーンチェンジを管理する
type SceneChangeManager struct {
	activeChanges []*SceneChange
	mu            sync.Mutex
}

// NewSceneChangeManager は新しいSceneChangeManagerを作成する
func NewSceneChangeManager() *SceneChangeManager {
	return &SceneChangeManager{
		activeChanges: make([]*SceneChange, 0),
	}
}

// Add はシーンチェンジを追加する
func (scm *SceneChangeManager) Add(sc *SceneChange) {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	scm.activeChanges = append(scm.activeChanges, sc)
}

// Update はすべてのアクティブなシーンチェンジを更新する
// 完了したシーンチェンジは削除される
func (scm *SceneChangeManager) Update() {
	scm.mu.Lock()
	defer scm.mu.Unlock()

	// 各シーンチェンジを更新
	for _, sc := range scm.activeChanges {
		if !sc.IsCompleted() {
			sc.Update()
			sc.Apply()
		}
	}

	// 完了したシーンチェンジを削除
	active := make([]*SceneChange, 0, len(scm.activeChanges))
	for _, sc := range scm.activeChanges {
		if !sc.IsCompleted() {
			active = append(active, sc)
		}
	}
	scm.activeChanges = active
}

// HasActiveChanges はアクティブなシーンチェンジがあるかどうかを返す
func (scm *SceneChangeManager) HasActiveChanges() bool {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	return len(scm.activeChanges) > 0
}

// GetActiveCount はアクティブなシーンチェンジの数を返す
func (scm *SceneChangeManager) GetActiveCount() int {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	return len(scm.activeChanges)
}

// Clear はすべてのシーンチェンジをクリアする
func (scm *SceneChangeManager) Clear() {
	scm.mu.Lock()
	defer scm.mu.Unlock()
	scm.activeChanges = make([]*SceneChange, 0)
}

// ApplyImmediate はシーンチェンジを即座に適用する（アニメーションなし）
// mode=0,1の場合や、即座に完了させたい場合に使用
func ApplyImmediate(
	srcImage *ebiten.Image,
	dstImage *ebiten.Image,
	srcRect image.Rectangle,
	dstPoint image.Point,
	mode SceneChangeMode,
	transColor color.Color,
) {
	switch mode {
	case SceneChangeNone:
		// 通常コピー
		subImg := srcImage.SubImage(srcRect).(*ebiten.Image)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(dstPoint.X), float64(dstPoint.Y))
		dstImage.DrawImage(subImg, opts)

	case SceneChangeTransparent:
		// 透明色除外（簡易実装）
		// 完全な透明色処理はシェーダーが必要だが、
		// 多くの場合は黒が透明色として使用される
		subImg := srcImage.SubImage(srcRect).(*ebiten.Image)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(dstPoint.X), float64(dstPoint.Y))
		dstImage.DrawImage(subImg, opts)

	default:
		// シーンチェンジモード（2-9）は即座に完了
		subImg := srcImage.SubImage(srcRect).(*ebiten.Image)
		opts := &ebiten.DrawImageOptions{}
		opts.GeoM.Translate(float64(dstPoint.X), float64(dstPoint.Y))
		dstImage.DrawImage(subImg, opts)
	}
}
