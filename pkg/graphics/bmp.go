// Package graphics provides BMP image decoding with RLE compression support.
// Go標準ライブラリの image/bmp はRLE圧縮をサポートしていないため、
// カスタムデコーダーを実装する。
//
// BMP圧縮方式:
//   - BI_RGB (0): 非圧縮
//   - BI_RLE8 (1): 8ビットRLE圧縮
//   - BI_RLE4 (2): 4ビットRLE圧縮
//
// 要件 1.10.1: RLE圧縮されたBMP形式（RLE8、RLE4）をサポートする
// 要件 1.10.2: 非圧縮BMP形式をサポートする
package graphics

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
	"io"
)

// BMP圧縮方式の定数
const (
	biRGB  = 0 // 非圧縮
	biRLE8 = 1 // 8ビットRLE圧縮
	biRLE4 = 2 // 4ビットRLE圧縮
)

// BMPファイルヘッダー (14バイト)
type bmpFileHeader struct {
	Signature  [2]byte // "BM"
	FileSize   uint32  // ファイルサイズ
	Reserved1  uint16  // 予約
	Reserved2  uint16  // 予約
	DataOffset uint32  // 画像データへのオフセット
}

// BMP情報ヘッダー (BITMAPINFOHEADER, 40バイト)
type bmpInfoHeader struct {
	HeaderSize      uint32 // ヘッダーサイズ (40)
	Width           int32  // 画像の幅
	Height          int32  // 画像の高さ (負の場合はトップダウン)
	Planes          uint16 // プレーン数 (常に1)
	BitCount        uint16 // ビット深度 (1, 4, 8, 24, 32)
	Compression     uint32 // 圧縮方式
	ImageSize       uint32 // 画像データサイズ
	XPixelsPerMeter int32  // 水平解像度
	YPixelsPerMeter int32  // 垂直解像度
	ColorsUsed      uint32 // 使用色数
	ColorsImportant uint32 // 重要な色数
}

// DecodeBMP はBMPファイルをデコードする（RLE圧縮対応）
func DecodeBMP(r io.Reader) (image.Image, error) {
	// ファイルヘッダーを読み込む
	var fileHeader bmpFileHeader
	if err := binary.Read(r, binary.LittleEndian, &fileHeader); err != nil {
		return nil, fmt.Errorf("failed to read BMP file header: %w", err)
	}

	// シグネチャを確認
	if fileHeader.Signature[0] != 'B' || fileHeader.Signature[1] != 'M' {
		return nil, fmt.Errorf("invalid BMP signature: %c%c", fileHeader.Signature[0], fileHeader.Signature[1])
	}

	// 情報ヘッダーを読み込む
	var infoHeader bmpInfoHeader
	if err := binary.Read(r, binary.LittleEndian, &infoHeader); err != nil {
		return nil, fmt.Errorf("failed to read BMP info header: %w", err)
	}

	// サポートするビット深度を確認
	if infoHeader.BitCount != 8 && infoHeader.BitCount != 4 && infoHeader.BitCount != 24 {
		return nil, fmt.Errorf("unsupported bit depth: %d", infoHeader.BitCount)
	}

	// 圧縮方式を確認
	switch infoHeader.Compression {
	case biRGB:
		// 非圧縮
	case biRLE8:
		if infoHeader.BitCount != 8 {
			return nil, fmt.Errorf("RLE8 compression requires 8-bit depth, got %d", infoHeader.BitCount)
		}
	case biRLE4:
		if infoHeader.BitCount != 4 {
			return nil, fmt.Errorf("RLE4 compression requires 4-bit depth, got %d", infoHeader.BitCount)
		}
	default:
		return nil, fmt.Errorf("unsupported compression: %d", infoHeader.Compression)
	}

	// 画像サイズを計算
	width := int(infoHeader.Width)
	height := int(infoHeader.Height)
	topDown := false
	if height < 0 {
		height = -height
		topDown = true
	}

	// カラーパレットを読み込む（8ビット、4ビットの場合）
	var palette color.Palette
	if infoHeader.BitCount <= 8 {
		paletteSize := int(infoHeader.ColorsUsed)
		if paletteSize == 0 {
			paletteSize = 1 << infoHeader.BitCount
		}
		palette = make(color.Palette, paletteSize)
		for i := 0; i < paletteSize; i++ {
			var entry [4]byte // BGRA
			if _, err := io.ReadFull(r, entry[:]); err != nil {
				return nil, fmt.Errorf("failed to read palette entry %d: %w", i, err)
			}
			palette[i] = color.RGBA{
				R: entry[2],
				G: entry[1],
				B: entry[0],
				A: 255,
			}
		}
	}

	// 画像データの開始位置までスキップ
	// 既に読み込んだバイト数: 14 (ファイルヘッダー) + 40 (情報ヘッダー) + パレットサイズ
	currentPos := 14 + 40 + len(palette)*4
	skipBytes := int(fileHeader.DataOffset) - currentPos
	if skipBytes > 0 {
		if _, err := io.CopyN(io.Discard, r, int64(skipBytes)); err != nil {
			return nil, fmt.Errorf("failed to skip to image data: %w", err)
		}
	}

	// 画像を作成
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 圧縮方式に応じてデコード
	switch infoHeader.Compression {
	case biRGB:
		if err := decodeRGB(r, img, width, height, int(infoHeader.BitCount), palette, topDown); err != nil {
			return nil, err
		}
	case biRLE8:
		if err := decodeRLE8(r, img, width, height, palette, topDown); err != nil {
			return nil, err
		}
	case biRLE4:
		if err := decodeRLE4(r, img, width, height, palette, topDown); err != nil {
			return nil, err
		}
	}

	return img, nil
}

// decodeRGB は非圧縮BMPをデコードする
func decodeRGB(r io.Reader, img *image.RGBA, width, height, bitCount int, palette color.Palette, topDown bool) error {
	// 行のパディングを計算（4バイト境界）
	var rowSize int
	switch bitCount {
	case 8:
		rowSize = (width + 3) &^ 3
	case 4:
		rowSize = ((width + 1) / 2)
		rowSize = (rowSize + 3) &^ 3
	case 24:
		rowSize = (width*3 + 3) &^ 3
	}

	rowData := make([]byte, rowSize)

	for y := 0; y < height; y++ {
		// BMPはボトムアップ形式（topDownでない場合）
		destY := y
		if !topDown {
			destY = height - 1 - y
		}

		if _, err := io.ReadFull(r, rowData); err != nil {
			return fmt.Errorf("failed to read row %d: %w", y, err)
		}

		switch bitCount {
		case 8:
			for x := 0; x < width; x++ {
				idx := rowData[x]
				if int(idx) < len(palette) {
					img.Set(x, destY, palette[idx])
				}
			}
		case 4:
			for x := 0; x < width; x++ {
				byteIdx := x / 2
				var idx uint8
				if x%2 == 0 {
					idx = rowData[byteIdx] >> 4
				} else {
					idx = rowData[byteIdx] & 0x0F
				}
				if int(idx) < len(palette) {
					img.Set(x, destY, palette[idx])
				}
			}
		case 24:
			for x := 0; x < width; x++ {
				b := rowData[x*3]
				g := rowData[x*3+1]
				r := rowData[x*3+2]
				img.Set(x, destY, color.RGBA{R: r, G: g, B: b, A: 255})
			}
		}
	}

	return nil
}

// decodeRLE8 はRLE8圧縮BMPをデコードする
// RLE8エンコーディング:
//   - 2バイトペアを読み取る
//   - 最初のバイトが0でない場合: 2番目のバイトを最初のバイト回繰り返す
//   - 最初のバイトが0の場合:
//   - 2番目のバイトが0: 行末 (End of Line)
//   - 2番目のバイトが1: ビットマップ終了 (End of Bitmap)
//   - 2番目のバイトが2: デルタ（位置移動）
//   - それ以外: 絶対モード（2番目のバイト個のピクセルをそのまま読み取る）
func decodeRLE8(r io.Reader, img *image.RGBA, width, height int, palette color.Palette, topDown bool) error {
	x, y := 0, 0

	for {
		var pair [2]byte
		if _, err := io.ReadFull(r, pair[:]); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read RLE8 data: %w", err)
		}

		count := int(pair[0])
		value := pair[1]

		if count > 0 {
			// エンコードモード: valueをcount回繰り返す
			for i := 0; i < count; i++ {
				if x < width && y < height {
					destY := y
					if !topDown {
						destY = height - 1 - y
					}
					if int(value) < len(palette) {
						img.Set(x, destY, palette[value])
					}
				}
				x++
			}
		} else {
			// エスケープモード
			switch value {
			case 0:
				// 行末 (End of Line)
				x = 0
				y++
			case 1:
				// ビットマップ終了 (End of Bitmap)
				return nil
			case 2:
				// デルタ（位置移動）
				var delta [2]byte
				if _, err := io.ReadFull(r, delta[:]); err != nil {
					return fmt.Errorf("failed to read RLE8 delta: %w", err)
				}
				x += int(delta[0])
				y += int(delta[1])
			default:
				// 絶対モード: value個のピクセルをそのまま読み取る
				absCount := int(value)
				absData := make([]byte, absCount)
				if _, err := io.ReadFull(r, absData); err != nil {
					return fmt.Errorf("failed to read RLE8 absolute data: %w", err)
				}

				for i := 0; i < absCount; i++ {
					if x < width && y < height {
						destY := y
						if !topDown {
							destY = height - 1 - y
						}
						idx := absData[i]
						if int(idx) < len(palette) {
							img.Set(x, destY, palette[idx])
						}
					}
					x++
				}

				// 絶対モードは2バイト境界にパディングされる
				if absCount%2 != 0 {
					var padding [1]byte
					if _, err := io.ReadFull(r, padding[:]); err != nil {
						return fmt.Errorf("failed to read RLE8 padding: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// decodeRLE4 はRLE4圧縮BMPをデコードする
// RLE4エンコーディング:
//   - 2バイトペアを読み取る
//   - 最初のバイトが0でない場合: 2番目のバイトの上位4ビットと下位4ビットを交互に繰り返す
//   - 最初のバイトが0の場合:
//   - 2番目のバイトが0: 行末 (End of Line)
//   - 2番目のバイトが1: ビットマップ終了 (End of Bitmap)
//   - 2番目のバイトが2: デルタ（位置移動）
//   - それ以外: 絶対モード（2番目のバイト個のピクセルをそのまま読み取る）
func decodeRLE4(r io.Reader, img *image.RGBA, width, height int, palette color.Palette, topDown bool) error {
	x, y := 0, 0

	for {
		var pair [2]byte
		if _, err := io.ReadFull(r, pair[:]); err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("failed to read RLE4 data: %w", err)
		}

		count := int(pair[0])
		value := pair[1]

		if count > 0 {
			// エンコードモード: 上位4ビットと下位4ビットを交互にcount回繰り返す
			highNibble := value >> 4
			lowNibble := value & 0x0F

			for i := 0; i < count; i++ {
				if x < width && y < height {
					destY := y
					if !topDown {
						destY = height - 1 - y
					}
					var idx uint8
					if i%2 == 0 {
						idx = highNibble
					} else {
						idx = lowNibble
					}
					if int(idx) < len(palette) {
						img.Set(x, destY, palette[idx])
					}
				}
				x++
			}
		} else {
			// エスケープモード
			switch value {
			case 0:
				// 行末 (End of Line)
				x = 0
				y++
			case 1:
				// ビットマップ終了 (End of Bitmap)
				return nil
			case 2:
				// デルタ（位置移動）
				var delta [2]byte
				if _, err := io.ReadFull(r, delta[:]); err != nil {
					return fmt.Errorf("failed to read RLE4 delta: %w", err)
				}
				x += int(delta[0])
				y += int(delta[1])
			default:
				// 絶対モード: value個のピクセルをそのまま読み取る
				absCount := int(value)
				// 必要なバイト数を計算（2ピクセルで1バイト）
				absBytes := (absCount + 1) / 2
				absData := make([]byte, absBytes)
				if _, err := io.ReadFull(r, absData); err != nil {
					return fmt.Errorf("failed to read RLE4 absolute data: %w", err)
				}

				for i := 0; i < absCount; i++ {
					if x < width && y < height {
						destY := y
						if !topDown {
							destY = height - 1 - y
						}
						byteIdx := i / 2
						var idx uint8
						if i%2 == 0 {
							idx = absData[byteIdx] >> 4
						} else {
							idx = absData[byteIdx] & 0x0F
						}
						if int(idx) < len(palette) {
							img.Set(x, destY, palette[idx])
						}
					}
					x++
				}

				// 絶対モードは2バイト境界にパディングされる
				if absBytes%2 != 0 {
					var padding [1]byte
					if _, err := io.ReadFull(r, padding[:]); err != nil {
						return fmt.Errorf("failed to read RLE4 padding: %w", err)
					}
				}
			}
		}
	}

	return nil
}

// IsBMPRLECompressed はBMPファイルがRLE圧縮されているかどうかを判定する
// ファイルの先頭を読み取り、圧縮方式を確認する
func IsBMPRLECompressed(r io.ReadSeeker) (bool, error) {
	// 現在位置を保存
	pos, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return false, err
	}

	// ファイルの先頭に移動
	if _, err := r.Seek(0, io.SeekStart); err != nil {
		return false, err
	}

	// ファイルヘッダーを読み込む
	var fileHeader bmpFileHeader
	if err := binary.Read(r, binary.LittleEndian, &fileHeader); err != nil {
		r.Seek(pos, io.SeekStart) // 元の位置に戻す
		return false, err
	}

	// シグネチャを確認
	if fileHeader.Signature[0] != 'B' || fileHeader.Signature[1] != 'M' {
		r.Seek(pos, io.SeekStart) // 元の位置に戻す
		return false, nil         // BMPファイルではない
	}

	// 情報ヘッダーを読み込む
	var infoHeader bmpInfoHeader
	if err := binary.Read(r, binary.LittleEndian, &infoHeader); err != nil {
		r.Seek(pos, io.SeekStart) // 元の位置に戻す
		return false, err
	}

	// 元の位置に戻す
	if _, err := r.Seek(pos, io.SeekStart); err != nil {
		return false, err
	}

	// RLE圧縮かどうかを判定
	return infoHeader.Compression == biRLE8 || infoHeader.Compression == biRLE4, nil
}

// IsBMPRLECompressedFromBytes はバイト配列からBMPがRLE圧縮されているかどうかを判定する
func IsBMPRLECompressedFromBytes(data []byte) (bool, error) {
	if len(data) < 54 { // 14 (file header) + 40 (info header)
		return false, fmt.Errorf("data too short for BMP header")
	}

	// シグネチャを確認
	if data[0] != 'B' || data[1] != 'M' {
		return false, nil // BMPファイルではない
	}

	// 圧縮方式を読み取る（オフセット 30 = 14 + 16）
	compression := binary.LittleEndian.Uint32(data[30:34])

	// RLE圧縮かどうかを判定
	return compression == biRLE8 || compression == biRLE4, nil
}

// DecodeBMPFromBytes はバイト配列からBMPをデコードする
func DecodeBMPFromBytes(data []byte) (image.Image, error) {
	return DecodeBMP(bytes.NewReader(data))
}
