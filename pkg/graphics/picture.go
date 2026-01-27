package graphics

import (
	"fmt"
	"image"
	"image/draw"
	_ "image/png" // PNG デコーダを登録
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	_ "golang.org/x/image/bmp" // BMP デコーダを登録（非圧縮BMP用）

	"github.com/zurustar/son-et/pkg/fileutil"
)

// isBMPFile はファイルパスがBMPファイルかどうかを判定する
func isBMPFile(path string) bool {
	ext := strings.ToLower(filepath.Ext(path))
	return ext == ".bmp"
}

// Picture はメモリ上の画像データを表す
type Picture struct {
	ID            int
	Image         *ebiten.Image // 現在の画像（テキスト描画後）
	OriginalImage *image.RGBA   // 元の背景画像（テキスト描画前）
	BackBuffer    *ebiten.Image // ダブルバッファリング用（キャスト再描画時に使用）
	Width         int
	Height        int
}

// PictureManager はピクチャーを管理する
type PictureManager struct {
	pictures map[int]*Picture
	nextID   int
	maxID    int    // 最大256（要件 9.5）
	basePath string // 画像ファイルの基準パス
	log      *slog.Logger
	mu       sync.RWMutex
}

// NewPictureManager は新しい PictureManager を作成する
func NewPictureManager(basePath string) *PictureManager {
	return &PictureManager{
		pictures: make(map[int]*Picture),
		nextID:   0,
		maxID:    256,
		basePath: basePath,
		log:      slog.Default(),
	}
}

// LoadPic は指定されたファイルから画像を読み込み、ピクチャーIDを返す
// 要件 1.1, 1.2, 1.3, 1.10, 1.10.1, 1.10.2, 1.11, 1.12
func (pm *PictureManager) LoadPic(filename string) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// リソース制限チェック（要件 9.5, 9.8）
	if len(pm.pictures) >= pm.maxID {
		err := fmt.Errorf("picture limit reached: %d", pm.maxID)
		pm.log.Error("LoadPic: resource limit exceeded", "filename", filename, "limit", pm.maxID)
		return -1, err
	}

	// ファイルパスを解決（大文字小文字非依存、要件 1.12）
	var fullPath string
	var err error

	if filepath.IsAbs(filename) {
		// 絶対パスの場合
		fullPath = filename
	} else if pm.basePath != "" {
		// 相対パスの場合、basePathから検索
		fullPath, err = fileutil.FindFileCaseInsensitive(pm.basePath, filename)
		if err != nil {
			pm.log.Error("LoadPic: file not found", "filename", filename, "basePath", pm.basePath)
			return -1, fmt.Errorf("file not found: %s", filename)
		}
	} else {
		// basePathが設定されていない場合、カレントディレクトリから検索
		fullPath, err = fileutil.FindFileCaseInsensitive(".", filename)
		if err != nil {
			pm.log.Error("LoadPic: file not found", "filename", filename)
			return -1, fmt.Errorf("file not found: %s", filename)
		}
	}

	// ファイルを開く
	file, err := os.Open(fullPath)
	if err != nil {
		pm.log.Error("LoadPic: failed to open file", "filename", filename, "path", fullPath, "error", err)
		return -1, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	// 画像をデコード（BMP/PNG対応、要件 1.10, 1.10.1, 1.10.2, 1.11）
	var img image.Image

	// BMPファイルの場合、RLE圧縮かどうかを確認
	if isBMPFile(fullPath) {
		isRLE, err := IsBMPRLECompressed(file)
		if err != nil {
			pm.log.Warn("LoadPic: failed to check RLE compression, falling back to standard decoder", "filename", filename, "error", err)
		}

		// ファイルの先頭に戻す
		if _, err := file.Seek(0, 0); err != nil {
			pm.log.Error("LoadPic: failed to seek file", "filename", filename, "error", err)
			return -1, fmt.Errorf("failed to seek file: %w", err)
		}

		if isRLE {
			// RLE圧縮BMPの場合、カスタムデコーダーを使用（要件 1.10.1）
			pm.log.Info("LoadPic: using custom RLE BMP decoder", "filename", filename)
			img, err = DecodeBMP(file)
			if err != nil {
				pm.log.Error("LoadPic: failed to decode RLE BMP", "filename", filename, "error", err)
				return -1, fmt.Errorf("failed to decode RLE BMP: %w", err)
			}
		} else {
			// 非圧縮BMPの場合、標準デコーダーを使用（要件 1.10.2）
			img, _, err = image.Decode(file)
			if err != nil {
				pm.log.Error("LoadPic: failed to decode image", "filename", filename, "error", err)
				return -1, fmt.Errorf("failed to decode image: %w", err)
			}
		}
	} else {
		// BMP以外の場合、標準デコーダーを使用
		img, _, err = image.Decode(file)
		if err != nil {
			pm.log.Error("LoadPic: failed to decode image", "filename", filename, "error", err)
			return -1, fmt.Errorf("failed to decode image: %w", err)
		}
	}

	// Ebiten画像に変換
	ebitenImg := ebiten.NewImageFromImage(img)

	// 元の背景画像をRGBAとして保存（テキスト描画用）
	bounds := img.Bounds()
	originalRGBA := image.NewRGBA(bounds)
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			originalRGBA.Set(x, y, img.At(x, y))
		}
	}

	// ピクチャーIDを割り当て（要件 1.2）
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:            picID,
		Image:         ebitenImg,
		OriginalImage: originalRGBA,
		Width:         ebitenImg.Bounds().Dx(),
		Height:        ebitenImg.Bounds().Dy(),
	}

	pm.pictures[picID] = pic

	pm.log.Info("LoadPic: loaded picture",
		"filename", filename,
		"pictureID", picID,
		"width", pic.Width,
		"height", pic.Height)

	return picID, nil
}

// CreatePic は指定されたサイズの空のピクチャーを生成する
// 要件 1.4, 1.5
func (pm *PictureManager) CreatePic(width, height int) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// リソース制限チェック（要件 9.5, 9.8）
	if len(pm.pictures) >= pm.maxID {
		err := fmt.Errorf("picture limit reached: %d", pm.maxID)
		pm.log.Error("CreatePic: resource limit exceeded", "limit", pm.maxID)
		return -1, err
	}

	// 空の画像を作成
	ebitenImg := ebiten.NewImage(width, height)

	// 元の背景画像をRGBAとして保存（空の画像 = 透明）
	originalRGBA := image.NewRGBA(image.Rect(0, 0, width, height))

	// ピクチャーIDを割り当て（要件 1.5）
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:            picID,
		Image:         ebitenImg,
		OriginalImage: originalRGBA,
		Width:         width,
		Height:        height,
	}

	pm.pictures[picID] = pic

	pm.log.Info("CreatePic: created picture",
		"pictureID", picID,
		"width", width,
		"height", height)

	return picID, nil
}

// CreatePicFrom は既存のピクチャーからコピーして新しいピクチャーを生成する
func (pm *PictureManager) CreatePicFrom(srcID int) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// リソース制限チェック（要件 9.5, 9.8）
	if len(pm.pictures) >= pm.maxID {
		err := fmt.Errorf("picture limit reached: %d", pm.maxID)
		pm.log.Error("CreatePicFrom: resource limit exceeded", "limit", pm.maxID)
		return -1, err
	}

	// ソースピクチャーを取得
	srcPic, exists := pm.pictures[srcID]
	if !exists {
		err := fmt.Errorf("source picture not found: %d", srcID)
		pm.log.Error("CreatePicFrom: source picture not found", "srcID", srcID)
		return -1, err
	}

	// 新しい画像を作成してコピー
	ebitenImg := ebiten.NewImage(srcPic.Width, srcPic.Height)
	ebitenImg.DrawImage(srcPic.Image, nil)

	// 元の背景画像もコピー
	var originalRGBA *image.RGBA
	if srcPic.OriginalImage != nil {
		originalRGBA = image.NewRGBA(srcPic.OriginalImage.Bounds())
		draw.Draw(originalRGBA, originalRGBA.Bounds(), srcPic.OriginalImage, image.Point{}, draw.Src)
	}

	// ピクチャーIDを割り当て
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:            picID,
		Image:         ebitenImg,
		OriginalImage: originalRGBA,
		Width:         srcPic.Width,
		Height:        srcPic.Height,
	}

	pm.pictures[picID] = pic

	pm.log.Info("CreatePicFrom: created picture from source",
		"pictureID", picID,
		"srcID", srcID,
		"width", pic.Width,
		"height", pic.Height)

	return picID, nil
}

// DelPic は指定されたピクチャーを削除し、メモリを解放する
// 要件 1.6, 9.1, 9.4
func (pm *PictureManager) DelPic(id int) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	pic, exists := pm.pictures[id]
	if !exists {
		err := fmt.Errorf("picture not found: %d", id)
		pm.log.Warn("DelPic: picture not found", "pictureID", id)
		return err
	}

	// Ebiten画像リソースを解放（要件 9.1）
	if pic.Image != nil {
		pic.Image.Deallocate()
	}

	// マップから削除（要件 9.4: ID再利用を許可）
	delete(pm.pictures, id)

	pm.log.Info("DelPic: deleted picture", "pictureID", id)

	return nil
}

// GetPic は指定されたピクチャーを取得する
func (pm *PictureManager) GetPic(id int) (*Picture, error) {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pic, exists := pm.pictures[id]
	if !exists {
		err := fmt.Errorf("picture not found: %d", id)
		pm.log.Warn("GetPic: picture not found", "pictureID", id)
		return nil, err
	}

	return pic, nil
}

// GetPicWithoutLock はロックなしでピクチャーを取得する（内部用）
// 呼び出し元でロックを取得している場合に使用
func (pm *PictureManager) GetPicWithoutLock(id int) (*Picture, error) {
	pic, exists := pm.pictures[id]
	if !exists {
		err := fmt.Errorf("picture not found: %d", id)
		pm.log.Warn("GetPicWithoutLock: picture not found", "pictureID", id)
		return nil, err
	}

	return pic, nil
}

// PicWidth は指定されたピクチャーの幅を返す
// 要件 1.7, 1.9
func (pm *PictureManager) PicWidth(id int) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pic, exists := pm.pictures[id]
	if !exists {
		pm.log.Error("PicWidth: picture not found", "pictureID", id)
		return 0 // 要件 1.9: エラー時は0を返す
	}

	return pic.Width
}

// PicHeight は指定されたピクチャーの高さを返す
// 要件 1.8, 1.9
func (pm *PictureManager) PicHeight(id int) int {
	pm.mu.RLock()
	defer pm.mu.RUnlock()

	pic, exists := pm.pictures[id]
	if !exists {
		pm.log.Error("PicHeight: picture not found", "pictureID", id)
		return 0 // 要件 1.9: エラー時は0を返す
	}

	return pic.Height
}

// CreatePicWithSize は指定されたサイズの空のピクチャーを生成する
// srcID: 参照用のソースピクチャーID（存在確認のみ）
// width, height: 新しいピクチャーのサイズ
// 戻り値: 新しいピクチャーID、エラー
// 要件 2.1, 2.2, 2.3, 2.4
func (pm *PictureManager) CreatePicWithSize(srcID, width, height int) (int, error) {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	// ソースピクチャーIDの存在確認（要件 2.3）
	_, exists := pm.pictures[srcID]
	if !exists {
		err := fmt.Errorf("source picture not found: %d", srcID)
		pm.log.Error("CreatePicWithSize: source picture not found", "srcID", srcID)
		return -1, err
	}

	// 幅・高さのバリデーション（要件 2.4）
	if width <= 0 {
		err := fmt.Errorf("invalid width: %d (must be greater than 0)", width)
		pm.log.Error("CreatePicWithSize: invalid width", "width", width)
		return -1, err
	}
	if height <= 0 {
		err := fmt.Errorf("invalid height: %d (must be greater than 0)", height)
		pm.log.Error("CreatePicWithSize: invalid height", "height", height)
		return -1, err
	}

	// リソース制限チェック
	if len(pm.pictures) >= pm.maxID {
		err := fmt.Errorf("picture limit reached: %d", pm.maxID)
		pm.log.Error("CreatePicWithSize: resource limit exceeded", "limit", pm.maxID)
		return -1, err
	}

	// 指定サイズの空のピクチャーを作成（要件 2.1, 2.2）
	// ソースピクチャーの内容はコピーしない
	ebitenImg := ebiten.NewImage(width, height)

	// 元の背景画像をRGBAとして保存（空の画像 = 透明）
	originalRGBA := image.NewRGBA(image.Rect(0, 0, width, height))

	// ピクチャーIDを割り当て
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:            picID,
		Image:         ebitenImg,
		OriginalImage: originalRGBA,
		Width:         width,
		Height:        height,
	}

	pm.pictures[picID] = pic

	pm.log.Info("CreatePicWithSize: created picture with specified size",
		"pictureID", picID,
		"srcID", srcID,
		"width", width,
		"height", height)

	return picID, nil
}
