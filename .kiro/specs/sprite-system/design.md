# 設計ドキュメント: スプライトシステム (Sprite System)

## 概要

このドキュメントは、FILLYエミュレータのグラフィックスシステムを簡素化するためのスプライトシステムの設計を定義します。すべての描画要素を統一的なスプライトとして扱い、Ebitengineの機能をラップします。

## アーキテクチャ

### 全体構成

```
┌─────────────────────────────────────────────────────────────┐
│ SpriteManager                                                │
│   sprites: map[int]*Sprite                                  │
│   sorted: []*Sprite (Z順序キャッシュ)                       │
├─────────────────────────────────────────────────────────────┤
│ Sprite                                                       │
│   - ID, Image, X, Y, ZOrder, Visible, Alpha, Parent, Dirty  │
└─────────────────────────────────────────────────────────────┘
```

### 描画要素とスプライトの対応

| 描画要素 | スプライトとしての実装 |
|----------|------------------------|
| 仮想ウインドウ | 背景色で塗りつぶしたスプライト（子の親となる） |
| ピクチャ | BMPを読み込んだスプライト |
| キャスト | 透明色処理済みのスプライト |
| 文字 | 差分抽出で生成した透過スプライト |
| 図形 | 描画結果をスプライト化 |


## コンポーネント

### Sprite構造体

```go
type Sprite struct {
    id      int            // 一意のID
    image   *ebiten.Image  // 描画画像
    x, y    float64        // 位置
    zOrder  int            // Z順序（小さいほど背面）
    visible bool           // 可視性
    alpha   float64        // 透明度（0.0〜1.0）
    parent  *Sprite        // 親スプライト（オプション）
    dirty   bool           // 再描画フラグ
}
```

### SpriteManager構造体

```go
type SpriteManager struct {
    mu       sync.RWMutex
    sprites  map[int]*Sprite  // ID→スプライト
    nextID   int              // 次のID
    sorted   []*Sprite        // Z順序ソート済みキャッシュ
    needSort bool             // ソート必要フラグ
}
```

## 親子関係の処理

### 絶対位置の計算

```go
func (s *Sprite) AbsolutePosition() (float64, float64) {
    x, y := s.x, s.y
    if s.parent != nil {
        px, py := s.parent.AbsolutePosition()
        x += px
        y += py
    }
    return x, y
}
```

### 実効透明度の計算

```go
func (s *Sprite) EffectiveAlpha() float64 {
    alpha := s.alpha
    if s.parent != nil {
        alpha *= s.parent.EffectiveAlpha()
    }
    return alpha
}
```

### 実効可視性の計算

```go
func (s *Sprite) IsEffectivelyVisible() bool {
    if !s.visible {
        return false
    }
    if s.parent != nil {
        return s.parent.IsEffectivelyVisible()
    }
    return true
}
```


## テキストスプライト（差分抽出方式）

### 処理フロー

```
1. 背景色で塗りつぶした一時画像を作成
2. 一時画像にテキストを描画（アンチエイリアスあり）
3. 背景色と異なるピクセルのみを抽出
4. 抽出結果を透過画像としてスプライト化
```

### 実装

```go
func CreateTextSprite(bg color.Color, text string, textColor color.Color, face font.Face) *image.RGBA {
    // 1. 背景画像を作成
    bgImg := image.NewRGBA(bounds)
    draw.Draw(bgImg, bgImg.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
    
    // 2. テキストを描画
    tempImg := image.NewRGBA(bounds)
    draw.Draw(tempImg, tempImg.Bounds(), bgImg, image.Point{}, draw.Src)
    drawer := &font.Drawer{Dst: tempImg, Src: image.NewUniform(textColor), Face: face}
    drawer.DrawString(text)
    
    // 3. 差分を抽出
    result := image.NewRGBA(bounds)
    for y := 0; y < height; y++ {
        for x := 0; x < width; x++ {
            if bgImg.At(x, y) != tempImg.At(x, y) {
                result.Set(x, y, tempImg.At(x, y))
            } else {
                result.Set(x, y, color.RGBA{0, 0, 0, 0}) // 透明
            }
        }
    }
    return result
}
```

## 描画処理

### Z順序ソート

```go
func (sm *SpriteManager) sortSprites() {
    sm.sorted = make([]*Sprite, 0, len(sm.sprites))
    for _, s := range sm.sprites {
        sm.sorted = append(sm.sorted, s)
    }
    sort.Slice(sm.sorted, func(i, j int) bool {
        return sm.sorted[i].zOrder < sm.sorted[j].zOrder
    })
    sm.needSort = false
}
```

### 描画ループ

```go
func (sm *SpriteManager) Draw(screen *ebiten.Image) {
    if sm.needSort {
        sm.sortSprites()
    }
    
    for _, s := range sm.sorted {
        if !s.IsEffectivelyVisible() || s.image == nil {
            continue
        }
        
        op := &ebiten.DrawImageOptions{}
        x, y := s.AbsolutePosition()
        op.GeoM.Translate(x, y)
        
        alpha := s.EffectiveAlpha()
        if alpha < 1.0 {
            op.ColorScale.ScaleAlpha(float32(alpha))
        }
        
        screen.DrawImage(s.image, op)
    }
}
```


## 正しさの性質

### Property 1: スプライトID管理

*任意の*スプライト作成に対して、作成されたスプライトは一意のIDを持ち、そのIDで取得できる。

**Validates: Requirements 1.1, 3.1, 3.3**

### Property 2: 親子関係の位置計算

*任意の*親子関係を持つスプライトに対して、子の絶対位置は親の位置と子の相対位置の和である。

**Validates: Requirements 2.1**

### Property 3: 親子関係の透明度計算

*任意の*親子関係を持つスプライトに対して、子の実効透明度は親の透明度と子の透明度の積である。

**Validates: Requirements 2.2**

### Property 4: 親子関係の可視性

*任意の*親子関係を持つスプライトに対して、親が非表示なら子も非表示として扱われる。

**Validates: Requirements 2.3, 4.3**

### Property 5: Z順序による描画順

*任意の*スプライト集合に対して、描画はZ順序の小さい順に行われる。

**Validates: Requirements 4.1**

### Property 6: テキスト差分抽出

*任意の*テキスト描画に対して、差分抽出後の画像は背景色のピクセルを含まない（透明になる）。

**Validates: Requirements 5.1, 5.2**

### Property 7: スプライト削除

*任意の*スプライト削除に対して、削除後はそのIDでスプライトを取得できない。

**Validates: Requirements 3.4**

## エラーハンドリング

| エラー | 処理 |
|--------|------|
| 存在しないIDでの取得 | nilを返す |
| 存在しないIDでの削除 | 何もしない |
| nil画像でのスプライト作成 | 許可（後で画像を設定可能） |

## テスト戦略

### 単体テスト

- スプライトの作成と属性設定
- 親子関係の位置・透明度・可視性計算
- SpriteManagerのCRUD操作
- Z順序ソート

### プロパティベーステスト

- Property 1〜7の検証
- ランダムな親子関係での位置計算
- ランダムなZ順序での描画順

