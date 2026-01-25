package graphics

import (
	"fmt"
	"image"
	_ "image/png" // PNG デコーダを登録
	"log/slog"
	"os"
	"path/filepath"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	_ "golang.org/x/image/bmp" // BMP デコーダを登録

	"github.com/zurustar/son-et/pkg/fileutil"
)

// Picture はメモリ上の画像データを表す
type Picture struct {
	ID     int
	Image  *ebiten.Image
	Width  int
	Height int
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
// 要件 1.1, 1.2, 1.3, 1.10, 1.11, 1.12
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

	// 画像をデコード（BMP/PNG対応、要件 1.10, 1.11）
	img, _, err := image.Decode(file)
	if err != nil {
		pm.log.Error("LoadPic: failed to decode image", "filename", filename, "error", err)
		return -1, fmt.Errorf("failed to decode image: %w", err)
	}

	// Ebiten画像に変換
	ebitenImg := ebiten.NewImageFromImage(img)

	// ピクチャーIDを割り当て（要件 1.2）
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:     picID,
		Image:  ebitenImg,
		Width:  ebitenImg.Bounds().Dx(),
		Height: ebitenImg.Bounds().Dy(),
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

	// ピクチャーIDを割り当て（要件 1.5）
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:     picID,
		Image:  ebitenImg,
		Width:  width,
		Height: height,
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

	// ピクチャーIDを割り当て
	picID := pm.nextID
	pm.nextID++

	// Pictureを作成
	pic := &Picture{
		ID:     picID,
		Image:  ebitenImg,
		Width:  srcPic.Width,
		Height: srcPic.Height,
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
