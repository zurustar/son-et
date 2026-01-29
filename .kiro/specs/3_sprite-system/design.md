# 設計書: スプライトシステム (Sprite System)

## 概要

スプライトシステムは、FILLYエミュレータのすべての描画要素を統一的に管理します。描画順序は親子関係とスライスの順序で決定され、数値によるZ順序管理は使用しません。

### 主要機能

1. **スライスベースの描画順序**: 子スプライトのスライス順序がそのまま描画順序
2. **操作順序追跡**: PutCast、TextWrite、MovePicの呼び出し順でスライスに追加
3. **親子関係管理**: 親スプライトの子が親の描画後に描画される
4. **動的順序変更**: スプライトの前面/背面移動（スライス内の位置変更）

### 設計原則

- **操作順序優先**: タイプに関係なく、操作順序でスライスに追加
- **スライス順序 = 描画順序**: 数値管理を廃止し、シンプルな順序管理
- **既存互換性**: 既存のスプライトAPIを維持

## アーキテクチャ

### 現在の問題点と解決策

```
現在の実装（1次元Z順序）:
┌─────────────────────────────────────────────────────────────────┐
│ CalculateGlobalZOrder(windowZOrder, localZOrder)                │
│                                                                 │
│ ウインドウ0: 0 - 9999                                           │
│   - 背景: 0                                                     │
│   - キャスト: 100-999  ← タイプ別に固定範囲                     │
│   - テキスト: 1000-    ← 常にキャストより前面                   │
│                                                                 │
│ 問題: テキストは常にキャストより前面になってしまう              │
└─────────────────────────────────────────────────────────────────┘

新しい実装（スライスベースの描画順序）:
┌─────────────────────────────────────────────────────────────────┐
│ 親スプライト.children = []*Sprite                               │
│                                                                 │
│ ウインドウ0.children:                                           │
│   [0]: 背景ピクチャー                                           │
│   [1]: PutCastで作成したキャスト                                │
│   [2]: TextWriteで作成したテキスト                              │
│   [3]: PutCastで作成したキャスト                                │
│                                                                 │
│ 解決: スライスの順序 = 描画順序（先頭が最背面、末尾が最前面）   │
└─────────────────────────────────────────────────────────────────┘
```

### システム構成図

```
┌─────────────────────────────────────────────────────────────────┐
│                      SpriteManager                               │
│                   (pkg/graphics/sprite.go)                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    スプライト管理                         │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │ sprites     │  │ roots       │  │ picture     │       │   │
│  │  │ map[id]     │  │ []*Sprite   │  │ SpriteMap   │       │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  WindowSprite   │  │  CastSprite     │  │  TextSprite     │
│                 │  │                 │  │                 │
│ children:       │  │ parent:         │  │ parent:         │
│   []*Sprite     │  │   *Sprite       │  │   *Sprite       │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## コンポーネントとインターフェース

### 1. Sprite (pkg/sprite/sprite.go)

#### 構造体

```go
// Sprite は汎用スプライト（スライスベースの描画順序）
type Sprite struct {
    id       int
    image    *ebiten.Image
    x, y     float64
    visible  bool
    alpha    float64
    parent   *Sprite       // 親スプライトへのポインタ（nilの場合はルート）
    children []*Sprite     // 子スプライトのスライス（順序 = 描画順序）
    dirty    bool
}

// NewSprite は新しいスプライトを作成する
func NewSprite(id int, img *ebiten.Image) *Sprite {
    return &Sprite{
        id:       id,
        image:    img,
        visible:  true,
        alpha:    1.0,
        children: make([]*Sprite, 0),
    }
}

// AddChild は子スプライトをスライスの末尾に追加する（最前面に配置）
func (s *Sprite) AddChild(child *Sprite) {
    child.parent = s
    s.children = append(s.children, child)
}

// RemoveChild は子スプライトを削除する
func (s *Sprite) RemoveChild(childID int) {
    for i, child := range s.children {
        if child.id == childID {
            child.parent = nil
            s.children = append(s.children[:i], s.children[i+1:]...)
            return
        }
    }
}

// BringToFront はスプライトを最前面に移動する（スライス末尾に移動）
func (s *Sprite) BringToFront() {
    if s.parent == nil {
        return
    }
    parent := s.parent
    parent.RemoveChild(s.id)
    parent.AddChild(s)
}

// SendToBack はスプライトを最背面に移動する（スライス先頭に移動）
func (s *Sprite) SendToBack() {
    if s.parent == nil {
        return
    }
    parent := s.parent
    // 現在の位置を見つけて削除
    for i, child := range parent.children {
        if child.id == s.id {
            parent.children = append(parent.children[:i], parent.children[i+1:]...)
            break
        }
    }
    // 先頭に挿入
    parent.children = append([]*Sprite{s}, parent.children...)
}
```

### 2. SpriteManager (pkg/sprite/sprite.go)

#### 構造体

```go
// SpriteManager はスプライトを管理する
// 注: pictureSpriteMapはPictureSpriteManagerで管理される（責務の分離）
type SpriteManager struct {
    mu      sync.RWMutex
    sprites map[int]*Sprite // ID -> Sprite
    roots   []*Sprite       // ルートスプライト（ウインドウ）のスライス
    nextID  int
}

// NewSpriteManager は新しいSpriteManagerを作成する
func NewSpriteManager() *SpriteManager {
    return &SpriteManager{
        sprites: make(map[int]*Sprite),
        roots:   make([]*Sprite, 0),
        nextID:  1,
    }
}

// CreateSprite は新しいスプライトを作成する
func (sm *SpriteManager) CreateSprite(img *ebiten.Image, parent *Sprite) *Sprite {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    s := NewSprite(sm.nextID, img)
    sm.sprites[s.id] = s
    sm.nextID++
    
    if parent != nil {
        parent.AddChild(s)
    } else {
        sm.roots = append(sm.roots, s)
    }
    
    return s
}

// CreateRootSprite はルートスプライト（ウインドウ）を作成する
func (sm *SpriteManager) CreateRootSprite(img *ebiten.Image) *Sprite {
    return sm.CreateSprite(img, nil)
}
```

#### 描画メソッド

```go
// Draw はすべての可視スプライトを描画する
func (sm *SpriteManager) Draw(screen *ebiten.Image) {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    // ルートスプライトを順番に描画
    for _, root := range sm.roots {
        sm.drawSprite(screen, root, 0, 0, 1.0)
    }
}

// drawSprite はスプライトとその子を再帰的に描画する
func (sm *SpriteManager) drawSprite(screen *ebiten.Image, s *Sprite, parentX, parentY, parentAlpha float64) {
    if !s.visible || s.image == nil {
        return
    }
    
    // 絶対位置と実効透明度を計算
    absX := parentX + s.x
    absY := parentY + s.y
    effectiveAlpha := parentAlpha * s.alpha
    
    // スプライトを描画
    op := &ebiten.DrawImageOptions{}
    op.GeoM.Translate(absX, absY)
    if effectiveAlpha < 1.0 {
        op.ColorScale.ScaleAlpha(float32(effectiveAlpha))
    }
    screen.DrawImage(s.image, op)
    
    // 子スプライトを順番に描画（スライス順序 = 描画順序）
    for _, child := range s.children {
        sm.drawSprite(screen, child, absX, absY, effectiveAlpha)
    }
}
```

## データモデル

### 親子関係の階層

**重要**: FILLYの仕様では、キャストやテキストは**ピクチャーに対して配置される**。

```
デスクトップ（仮想ルート）
├── ウインドウ0 (WindowSprite)
│   └── 背景ピクチャー (PictureSprite)
│       ├── キャスト1 (CastSprite)     ← PutCast(src, dst_pic, ...)で作成
│       ├── テキスト1 (TextSprite)     ← TextWrite(pic, ...)で作成
│       ├── MovePicで追加された画像 (PictureSprite)
│       └── キャスト2 (CastSprite)
├── ウインドウ1 (WindowSprite)
│   └── 背景ピクチャー (PictureSprite)
│       └── キャスト1 (CastSprite)
└── ウインドウ2 (WindowSprite)
    └── 背景ピクチャー (PictureSprite)
```

**根拠**:
- PutCast: `PutCast(src_pic, dst_pic, X, Y, ...)` - 第2引数は**配置先ピクチャー番号**
- TextWrite: `TextWrite(pic_no, x, y, text)` - 第1引数は**ピクチャー番号**
- MovePic: `MovePic(src_pic, ..., dst_pic, ...)` - 転送先は**ピクチャー番号**

### PictureとSpriteの関係

```
┌─────────────────────────────────────────────────────────────────┐
│ Picture（メモリ上の画像データ）                                  │
│                                                                 │
│ LoadPic("image.bmp")  → Picture構造体を作成                     │
│ CreatePic(n, w, h)    → Picture構造体を作成                     │
│                                                                 │
│ ※ 画面には直接表示されない                                      │
└─────────────────────────────────────────────────────────────────┘
                             │
                             │ OpenWin, SetPic
                             ▼
┌─────────────────────────────────────────────────────────────────┐
│ Sprite（画面に表示される描画単位）                               │
│                                                                 │
│ OpenWin(pic, ...)     → WindowSprite + PictureSprite を作成     │
│ PutCast(src, dst, ...)→ CastSprite を作成                       │
│ TextWrite(pic, ...)   → TextSprite を作成                       │
│ MovePic(src, ..., dst)→ PictureSprite を作成                    │
│                                                                 │
│ ※ 画面に表示される                                              │
└─────────────────────────────────────────────────────────────────┘
```

## ピクチャースプライトの状態管理

### PictureSpriteの状態遷移

```
┌─────────────────┐     SetPic()      ┌─────────────────┐
│   未関連付け    │ ─────────────────→ │   関連付け済み  │
│  (Unattached)   │                    │   (Attached)    │
│                 │                    │                 │
│ - 非表示        │                    │ - 親の可視性に  │
│ - 親なし        │                    │   従って表示    │
└─────────────────┘                    └─────────────────┘
        │                                      │
        │ FreePic()                            │ FreePic()
        ▼                                      ▼
┌─────────────────────────────────────────────────────────────────┐
│                         削除                                     │
│  - PictureSpriteを削除                                          │
│  - 子スプライト（Cast, Text）も削除                             │
│  - pictureSpriteMapから削除                                     │
└─────────────────────────────────────────────────────────────────┘
```

### PictureSpriteの実装

```go
// PictureSpriteState はPictureSpriteの状態を表す
type PictureSpriteState int

const (
    PictureSpriteUnattached PictureSpriteState = iota  // 未関連付け
    PictureSpriteAttached                               // 関連付け済み
)

// PictureSprite はピクチャーのスプライト表現
type PictureSprite struct {
    sprite    *Sprite
    pictureID int
    state     PictureSpriteState
    windowID  int  // 関連付けられたウインドウID（-1 = 未関連付け）
}

// CreatePictureSpriteOnLoad はLoadPic時に非表示のPictureSpriteを作成する
func (sm *SpriteManager) CreatePictureSpriteOnLoad(pictureID int, img *ebiten.Image) *PictureSprite {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // 新しいスプライトを作成（非表示）
    s := NewSprite(sm.nextID, img)
    s.visible = false  // 未関連付けなので非表示
    sm.sprites[s.id] = s
    sm.nextID++
    
    ps := &PictureSprite{
        sprite:    s,
        pictureID: pictureID,
        state:     PictureSpriteUnattached,
        windowID:  -1,
    }
    
    sm.pictureSpriteMap[pictureID] = ps
    return ps
}

// AttachPictureSpriteToWindow はSetPic時にPictureSpriteをウインドウに関連付ける
func (sm *SpriteManager) AttachPictureSpriteToWindow(pictureID int, windowSprite *Sprite) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    ps, ok := sm.pictureSpriteMap[pictureID]
    if !ok {
        return fmt.Errorf("picture sprite not found: %d", pictureID)
    }
    
    // ウインドウの子として追加
    windowSprite.AddChild(ps.sprite)
    
    // 状態を更新
    ps.state = PictureSpriteAttached
    ps.sprite.visible = true
    
    return nil
}

// GetPictureSpriteByPictureID はピクチャー番号からPictureSpriteを取得する
func (sm *SpriteManager) GetPictureSpriteByPictureID(pictureID int) *PictureSprite {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    return sm.pictureSpriteMap[pictureID]
}
```

## 正確性プロパティ

### スライスベースの描画順序プロパティ

**プロパティ1: 追加順序の保持**
*任意の*子スプライトについて、後から追加されたスプライトはスライスの後方に位置する
**検証: 要件 9.3**

**プロパティ2: 描画順序の一貫性**
*任意の*親子関係について、親スプライトは子スプライトより先に描画される
**検証: 要件 10.1**

**プロパティ3: 兄弟描画順序**
*任意の*同じ親を持つ2つのスプライトについて、スライスの前方にあるスプライトが先に描画される
**検証: 要件 9.2**

**プロパティ4: 可視性の継承**
*任意の*親が非表示のスプライトについて、その子スプライトも描画されない
**検証: 要件 10.4**

### 動的変更のプロパティ

**プロパティ5: 最前面移動**
*任意の*スプライトについて、BringToFront後にそのスプライトはスライスの末尾に位置する
**検証: 要件 12.1**

**プロパティ6: 最背面移動**
*任意の*スプライトについて、SendToBack後にそのスプライトはスライスの先頭に位置する
**検証: 要件 12.2**

### ピクチャースプライトのプロパティ

**プロパティ7: ピクチャースプライトの作成**
*任意の*LoadPic呼び出しについて、対応するPictureSpriteが作成される
**検証: 要件 13.1**

**プロパティ8: 未関連付けピクチャの非表示**
*任意の*未関連付けPictureSpriteについて、そのスプライトは描画されない
**検証: 要件 14.2**

## エラーハンドリング

### エラーの分類

#### 致命的エラー
- 親子関係の循環参照（親子関係のループ）

#### 非致命的エラー
- 存在しないスプライトIDへのアクセス
- 存在しないピクチャーIDへのアクセス

### エラーログフォーマット

```
[timestamp] [level] [component] function_name: message (args)
```

例:
```
[19:43:20.577] [ERROR] [Sprite] BringToFront: sprite not found (spriteID=999)
[19:43:20.578] [WARN] [Sprite] AttachToWindow: picture sprite not found (pictureID=35)
```

## テスト戦略

### ユニットテスト

- Sprite: 作成、親子関係、描画順序変更
- SpriteManager: スプライト管理、描画
- PictureSprite: 状態管理、関連付け

### プロパティベーステスト

- プロパティ1-8（上記参照）
- テストライブラリ: gopter

### 統合テスト

- 既存のサンプルスクリプトの実行
- 描画順序の視覚的確認

## 実装の優先順位

### フェーズ1: 基盤
1. Sprite構造体の実装（スライスベース）
2. 親子関係管理の実装
3. SpriteManagerの実装

### フェーズ2: 描画
1. 再帰的描画の実装
2. 可視性・透明度の継承
3. 描画順序変更（BringToFront, SendToBack）

### フェーズ3: ピクチャースプライト
1. PictureSpriteの実装
2. 状態管理（未関連付け/関連付け済み）
3. pictureSpriteMapの実装

### フェーズ4: 既存システムとの統合
1. WindowSpriteの更新
2. CastSpriteの更新（PutCast引数修正）
3. TextSpriteの更新
4. PictureSpriteの更新

### フェーズ5: デバッグ支援
1. 階層ツリーの出力
2. 描画順序のリスト出力
3. デバッグオーバーレイ

## 技術的な考慮事項

### 並行性

- SpriteManagerはsync.RWMutexで保護
- 描画中のスプライト変更を防ぐ

### パフォーマンス

- スライスベースなのでソート不要
- 再帰的描画でO(n)の計算量
- 変更がない限り再描画をスキップ可能（将来の最適化）

### メモリ管理

- 子スプライト削除時に親からの参照も削除
- ピクチャー解放時に関連スプライトも削除

## パッケージ構成

スプライトシステムは独立したパッケージ `pkg/sprite/` として実装します。
これにより、設計書（3_sprite-system）とコードが1対1で対応します。

```
pkg/sprite/                    ← 3_sprite-system に対応
├── sprite.go                  # Sprite, SpriteManager
├── sprite_test.go             # スプライトのユニットテスト
├── sprite_property_test.go    # プロパティベーステスト
├── window_sprite.go           # WindowSprite
├── window_sprite_test.go
├── picture_sprite.go          # PictureSprite
├── picture_sprite_test.go
├── cast_sprite.go             # CastSprite
├── cast_sprite_test.go
├── text_sprite.go             # TextSprite
├── text_sprite_test.go
├── shape_sprite.go            # ShapeSprite
├── shape_sprite_test.go
└── errors.go                  # スプライト関連エラー

pkg/graphics/                  ← 4_graphics-system に対応（spriteを使用）
├── graphics.go                # GraphicsSystem（VM統合）
├── picture.go                 # Picture管理
├── window.go                  # Window管理
├── text.go                    # テキスト描画
├── transfer.go                # MovePic等の転送
├── primitives.go              # 描画プリミティブ
├── bmp.go                     # BMP読み込み
├── queue.go                   # 描画コマンドキュー
├── scene_change.go            # シーンチェンジ
└── ...
```

### 依存関係

```
pkg/graphics → pkg/sprite  （graphicsがspriteを使う）
pkg/vm → pkg/graphics      （VMがgraphicsを使う）
```

スプライトシステムは描画システムに依存せず、独立して動作します。

## 依存関係

### 外部ライブラリ

- **github.com/hajimehoshi/ebiten/v2**: 描画エンジン

### 内部パッケージ

- **pkg/logger**: ログ出力

**注意**: `pkg/sprite` は `pkg/graphics` に依存しません。逆方向の依存（graphics → sprite）のみ許可されます。

## 参考資料

- 要件定義書: `.kiro/specs/3_sprite-system/requirements.md`
- 描画システム要件: `.kiro/specs/4_graphics-system/requirements.md`
- 言語仕様: `_old_implementation2/.kiro/reference/language_spec.md`
- サンプル: `samples/y_saru/Y-SARU.TFY`（PutCast使用例）
