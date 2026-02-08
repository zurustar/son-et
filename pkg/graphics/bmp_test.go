package graphics

import (
	"bytes"
	"image"
	"os"
	"path/filepath"
	"testing"
)

// TestDecodeBMP_RLE8 はRLE8圧縮BMPのデコードをテストする
// samples/robot/ROBOT001.BMP を使用（RLE8圧縮、8ビット）
func TestDecodeBMP_RLE8(t *testing.T) {
	// サンプルファイルのパスを取得
	samplePath := filepath.Join("..", "..", "samples", "robot", "ROBOT001.BMP")

	// ファイルが存在するか確認
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skipf("Sample file not found: %s", samplePath)
	}

	// ファイルを開く
	file, err := os.Open(samplePath)
	if err != nil {
		t.Fatalf("Failed to open file: %v", err)
	}
	defer file.Close()

	// RLE圧縮かどうかを確認
	isRLE, err := IsBMPRLECompressed(file)
	if err != nil {
		t.Fatalf("Failed to check RLE compression: %v", err)
	}
	if !isRLE {
		t.Fatalf("Expected RLE compressed BMP, but got non-RLE")
	}

	// ファイルの先頭に戻す
	if _, err := file.Seek(0, 0); err != nil {
		t.Fatalf("Failed to seek file: %v", err)
	}

	// デコード
	img, err := DecodeBMP(file)
	if err != nil {
		t.Fatalf("Failed to decode BMP: %v", err)
	}

	// 画像サイズを確認
	bounds := img.Bounds()
	t.Logf("Image size: %dx%d", bounds.Dx(), bounds.Dy())

	// 画像サイズが正の値であることを確認
	if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
		t.Errorf("Invalid image size: %dx%d", bounds.Dx(), bounds.Dy())
	}

	// 画像がRGBAであることを確認
	if _, ok := img.(*image.RGBA); !ok {
		t.Errorf("Expected *image.RGBA, got %T", img)
	}
}

// TestDecodeBMP_RLE8_AllRobotFiles はすべてのROBOT*.BMPファイルをテストする
func TestDecodeBMP_RLE8_AllRobotFiles(t *testing.T) {
	robotDir := filepath.Join("..", "..", "samples", "robot")

	// ディレクトリが存在するか確認
	if _, err := os.Stat(robotDir); os.IsNotExist(err) {
		t.Skipf("Robot sample directory not found: %s", robotDir)
	}

	// ROBOT*.BMPファイルを検索
	matches, err := filepath.Glob(filepath.Join(robotDir, "ROBOT*.BMP"))
	if err != nil {
		t.Fatalf("Failed to glob files: %v", err)
	}

	if len(matches) == 0 {
		t.Skip("No ROBOT*.BMP files found")
	}

	for _, path := range matches {
		t.Run(filepath.Base(path), func(t *testing.T) {
			file, err := os.Open(path)
			if err != nil {
				t.Fatalf("Failed to open file: %v", err)
			}
			defer file.Close()

			// RLE圧縮かどうかを確認
			isRLE, err := IsBMPRLECompressed(file)
			if err != nil {
				t.Fatalf("Failed to check RLE compression: %v", err)
			}

			// ファイルの先頭に戻す
			if _, err := file.Seek(0, 0); err != nil {
				t.Fatalf("Failed to seek file: %v", err)
			}

			if isRLE {
				// RLE圧縮の場合、カスタムデコーダーを使用
				img, err := DecodeBMP(file)
				if err != nil {
					t.Fatalf("Failed to decode RLE BMP: %v", err)
				}

				bounds := img.Bounds()
				t.Logf("RLE compressed, size: %dx%d", bounds.Dx(), bounds.Dy())

				if bounds.Dx() <= 0 || bounds.Dy() <= 0 {
					t.Errorf("Invalid image size: %dx%d", bounds.Dx(), bounds.Dy())
				}
			} else {
				// 非圧縮の場合、標準デコーダーを使用
				img, _, err := image.Decode(file)
				if err != nil {
					t.Fatalf("Failed to decode non-RLE BMP: %v", err)
				}

				bounds := img.Bounds()
				t.Logf("Non-RLE, size: %dx%d", bounds.Dx(), bounds.Dy())
			}
		})
	}
}

// TestDecodeBMP_NonRLE は非圧縮BMPのデコードをテストする
func TestDecodeBMP_NonRLE(t *testing.T) {
	// 非圧縮8ビットBMPを作成（テスト用）
	// BMPファイルヘッダー + 情報ヘッダー + パレット + 画像データ
	var buf bytes.Buffer

	// ファイルヘッダー (14バイト)
	buf.Write([]byte{'B', 'M'})               // シグネチャ
	buf.Write([]byte{0x76, 0x00, 0x00, 0x00}) // ファイルサイズ (118バイト)
	buf.Write([]byte{0x00, 0x00})             // 予約1
	buf.Write([]byte{0x00, 0x00})             // 予約2
	buf.Write([]byte{0x76, 0x00, 0x00, 0x00}) // データオフセット (118バイト)

	// 情報ヘッダー (40バイト)
	buf.Write([]byte{0x28, 0x00, 0x00, 0x00}) // ヘッダーサイズ (40)
	buf.Write([]byte{0x02, 0x00, 0x00, 0x00}) // 幅 (2)
	buf.Write([]byte{0x02, 0x00, 0x00, 0x00}) // 高さ (2)
	buf.Write([]byte{0x01, 0x00})             // プレーン数 (1)
	buf.Write([]byte{0x08, 0x00})             // ビット深度 (8)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 圧縮方式 (0 = BI_RGB)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 画像サイズ
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 水平解像度
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 垂直解像度
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 使用色数 (0 = 256)
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 重要な色数

	// パレット (256色 × 4バイト = 1024バイト)
	// 簡略化のため、最初の4色のみ設定
	buf.Write([]byte{0x00, 0x00, 0x00, 0x00}) // 色0: 黒
	buf.Write([]byte{0x00, 0x00, 0xFF, 0x00}) // 色1: 赤
	buf.Write([]byte{0x00, 0xFF, 0x00, 0x00}) // 色2: 緑
	buf.Write([]byte{0xFF, 0x00, 0x00, 0x00}) // 色3: 青
	// 残りの252色は0で埋める
	for i := 0; i < 252; i++ {
		buf.Write([]byte{0x00, 0x00, 0x00, 0x00})
	}

	// 画像データ (2x2ピクセル、各行4バイト境界)
	// BMPはボトムアップなので、下の行から
	buf.Write([]byte{0x02, 0x03, 0x00, 0x00}) // 行0: 緑、青 + パディング
	buf.Write([]byte{0x00, 0x01, 0x00, 0x00}) // 行1: 黒、赤 + パディング

	// デコード
	img, err := DecodeBMP(bytes.NewReader(buf.Bytes()))
	if err != nil {
		t.Fatalf("Failed to decode BMP: %v", err)
	}

	// 画像サイズを確認
	bounds := img.Bounds()
	if bounds.Dx() != 2 || bounds.Dy() != 2 {
		t.Errorf("Expected 2x2, got %dx%d", bounds.Dx(), bounds.Dy())
	}

	// ピクセル値を確認
	rgba := img.(*image.RGBA)

	// (0, 0) = 黒
	r, g, b, a := rgba.At(0, 0).RGBA()
	if r != 0 || g != 0 || b != 0 || a != 0xFFFF {
		t.Errorf("Pixel (0,0): expected black, got (%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}

	// (1, 0) = 赤
	r, g, b, a = rgba.At(1, 0).RGBA()
	if r != 0xFFFF || g != 0 || b != 0 || a != 0xFFFF {
		t.Errorf("Pixel (1,0): expected red, got (%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}

	// (0, 1) = 緑
	r, g, b, a = rgba.At(0, 1).RGBA()
	if r != 0 || g != 0xFFFF || b != 0 || a != 0xFFFF {
		t.Errorf("Pixel (0,1): expected green, got (%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}

	// (1, 1) = 青
	r, g, b, a = rgba.At(1, 1).RGBA()
	if r != 0 || g != 0 || b != 0xFFFF || a != 0xFFFF {
		t.Errorf("Pixel (1,1): expected blue, got (%d, %d, %d, %d)", r>>8, g>>8, b>>8, a>>8)
	}
}

// TestIsBMPRLECompressed はRLE圧縮判定をテストする
func TestIsBMPRLECompressed(t *testing.T) {
	robotDir := filepath.Join("..", "..", "samples", "robot")

	// ディレクトリが存在するか確認
	if _, err := os.Stat(robotDir); os.IsNotExist(err) {
		t.Skipf("Robot sample directory not found: %s", robotDir)
	}

	// ROBOT001.BMPはRLE圧縮されているはず
	file, err := os.Open(filepath.Join(robotDir, "ROBOT001.BMP"))
	if err != nil {
		t.Skipf("Failed to open ROBOT001.BMP: %v", err)
	}
	defer file.Close()

	isRLE, err := IsBMPRLECompressed(file)
	if err != nil {
		t.Fatalf("Failed to check RLE compression: %v", err)
	}

	if !isRLE {
		t.Errorf("Expected ROBOT001.BMP to be RLE compressed")
	}
}
