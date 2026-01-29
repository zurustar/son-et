# 設計書: 階層的Z順序システム (Hierarchical Z-Order System)

## 概要

階層的Z順序システムは、FILLYエミュレータのスプライトシステムにおける描画順序を、1次元（フラット）なZ順序から親子関係に基づいた階層的（n次元）なZ順序に変更します。これにより、より柔軟で直感的な描画順序制御を実現します。

### 主要機能

1. **Z_Path（Zパス）**: スプライトの階層的Z順序を整数配列で表現
2. **操作順序追跡**: PutCast、TextWrite、MovePicの呼び出し順でZ順序を決定
3. **親子関係管理**: 親スプライトの子が親の描画後に描画される
4. **辞書順比較**: Z_Pathの効率的な比較とソート
5. **動的Z順序変更**: スプライトの前面/背面移動

### 設計原則

- **操作順序優先**: タイプに関係なく、操作順序でZ順序を決定
- **階層的継承**: 子スプライトは親のZ_Pathを継承
- **効率的比較**: Z_Pathの辞書順比較でO(depth)の計算量
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
│   - 描画: 1                                                     │
│   - キャスト: 100-999  ← タイプ別に固定範囲                     │
│   - テキスト: 1000-    ← 常にキャストより前面                   │
│                                                                 │
│ 問題: テキストは常にキャストより前面になってしまう              │
└─────────────────────────────────────────────────────────────────┘

新しい実装（階層的Z順序）:
┌─────────────────────────────────────────────────────────────────┐
│ Z_Path = [windowZOrder, operationOrder]                         │
│                                                                 │
│ ウインドウ0:                                                    │
│   - 背景: [0, 0]                                                │
│   - PutCast: [0, 1]    ← 操作順序1                              │
│   - TextWrite: [0, 2]  ← 操作順序2（キャストより前面）          │
│   - PutCast: [0, 3]    ← 操作順序3（テキストより前面）          │
│                                                                 │
│ 解決: 操作順序でZ順序が決まる                                   │
└─────────────────────────────────────────────────────────────────┘
```

### Z_Pathの構造

```
Z_Path = [level0, level1, level2, ...]

例:
- ウインドウ0: [0]
- ウインドウ0の背景: [0, 0]
- ウインドウ0のキャスト1: [0, 1]
- ウインドウ0のテキスト1: [0, 2]
- ウインドウ0のキャスト2: [0, 3]
- ウインドウ1: [1]
- ウインドウ1の背景: [1, 0]
- ウインドウ1のキャスト1: [1, 1]

辞書順比較:
[0, 1] < [0, 2] < [0, 3] < [1, 0] < [1, 1]
```

### システム構成図

```
┌─────────────────────────────────────────────────────────────────┐
│                      SpriteManager                               │
│                   (pkg/graphics/sprite.go)                       │
│  ┌──────────────────────────────────────────────────────────┐   │
│  │                    ZPathManager                           │   │
│  │  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐       │   │
│  │  │ Z_Path      │  │ Z_Order     │  │ Sorted      │       │   │
│  │  │ Storage     │  │ Counter     │  │ Cache       │       │   │
│  │  └─────────────┘  └─────────────┘  └─────────────┘       │   │
│  └──────────────────────────────────────────────────────────┘   │
└────────────────────────────┬────────────────────────────────────┘
                             │
         ┌───────────────────┼───────────────────┐
         │                   │                   │
         ▼                   ▼                   ▼
┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐
│  WindowSprite   │  │  CastSprite     │  │  TextSprite     │
│                 │  │                 │  │                 │
│ Z_Path: [n]     │  │ Z_Path: [n, m]  │  │ Z_Path: [n, m]  │
│ Counter: 0      │  │                 │  │                 │
└─────────────────┘  └─────────────────┘  └─────────────────┘
```

## コンポーネントとインターフェース

### 1. ZPath (pkg/graphics/zpath.go)

#### 構造体

```go
// ZPath は階層的Z順序を表す
// 要件 1.1: スプライトはZ_Pathを持つ（整数配列として表現）
type ZPath struct {
    path []int
}

// NewZPath は新しいZPathを作成する
func NewZPath(path ...int) *ZPath {
    return &ZPath{path: append([]int{}, path...)}
}

// NewZPathFromParent は親のZPathに子のローカルZ順序を追加したZPathを作成する
// 要件 1.2: Z_Pathは親のZ_Pathに自身のLocal_Z_Orderを追加した形式
func NewZPathFromParent(parent *ZPath, localZOrder int) *ZPath {
    if parent == nil {
        return NewZPath(localZOrder)
    }
    newPath := make([]int, len(parent.path)+1)
    copy(newPath, parent.path)
    newPath[len(parent.path)] = localZOrder
    return &ZPath{path: newPath}
}

// Path はZ_Pathの配列を返す
func (z *ZPath) Path() []int {
    return z.path
}

// Depth はZ_Pathの深さを返す
func (z *ZPath) Depth() int {
    return len(z.path)
}

// LocalZOrder は最後の要素（ローカルZ順序）を返す
func (z *ZPath) LocalZOrder() int {
    if len(z.path) == 0 {
        return 0
    }
    return z.path[len(z.path)-1]
}

// Parent は親のZ_Pathを返す
func (z *ZPath) Parent() *ZPath {
    if len(z.path) <= 1 {
        return nil
    }
    return &ZPath{path: z.path[:len(z.path)-1]}
}

// String はZ_Pathの文字列表現を返す
// 要件 10.1: スプライトのZ_Pathを文字列として取得できる
func (z *ZPath) String() string {
    return fmt.Sprintf("%v", z.path)
}
```

#### 比較関数

```go
// Compare はZ_Pathを辞書順で比較する
// 要件 5.1: Z_Pathを辞書順（lexicographic order）で比較する
// 戻り値: -1 (z < other), 0 (z == other), 1 (z > other)
func (z *ZPath) Compare(other *ZPath) int {
    if other == nil {
        return 1
    }
    
    minLen := len(z.path)
    if len(other.path) < minLen {
        minLen = len(other.path)
    }
    
    // 要件 5.2: Z_Path Aの先頭がZ_Path Bの先頭と一致するとき、次の要素を比較する
    for i := 0; i < minLen; i++ {
        if z.path[i] < other.path[i] {
            return -1
        }
        if z.path[i] > other.path[i] {
            return 1
        }
    }
    
    // 要件 5.3: Z_Path AがZ_Path Bのプレフィックスであるとき、AをBより前（背面）と判定する
    if len(z.path) < len(other.path) {
        return -1
    }
    if len(z.path) > len(other.path) {
        return 1
    }
    
    return 0
}

// Less は z < other かどうかを返す（sort.Interface用）
func (z *ZPath) Less(other *ZPath) bool {
    return z.Compare(other) < 0
}

// IsPrefix は z が other のプレフィックスかどうかを返す
func (z *ZPath) IsPrefix(other *ZPath) bool {
    if other == nil || len(z.path) > len(other.path) {
        return false
    }
    for i := 0; i < len(z.path); i++ {
        if z.path[i] != other.path[i] {
            return false
        }
    }
    return true
}
```

### 2. ZOrderCounter (pkg/graphics/zpath.go)

#### 構造体

```go
// ZOrderCounter は操作順序を追跡するカウンター
// 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
type ZOrderCounter struct {
    counters map[int]int  // parentSpriteID -> counter
    mu       sync.RWMutex
}

// NewZOrderCounter は新しいZOrderCounterを作成する
func NewZOrderCounter() *ZOrderCounter {
    return &ZOrderCounter{
        counters: make(map[int]int),
    }
}

// GetNext は指定された親スプライトの次のZ順序を取得し、カウンターをインクリメントする
// 要件 2.5: スプライトが作成されたとき、Z_Order_Counterをインクリメントする
func (c *ZOrderCounter) GetNext(parentID int) int {
    c.mu.Lock()
    defer c.mu.Unlock()
    
    current := c.counters[parentID]
    c.counters[parentID] = current + 1
    return current
}

// Reset は指定された親スプライトのカウンターをリセットする
func (c *ZOrderCounter) Reset(parentID int) {
    c.mu.Lock()
    defer c.mu.Unlock()
    delete(c.counters, parentID)
}

// ResetAll はすべてのカウンターをリセットする
func (c *ZOrderCounter) ResetAll() {
    c.mu.Lock()
    defer c.mu.Unlock()
    c.counters = make(map[int]int)
}
```

### 3. 拡張されたSprite (pkg/graphics/sprite.go)

#### 構造体の変更

```go
// Sprite は汎用スプライト（階層的Z順序対応）
type Sprite struct {
    id       int
    image    *ebiten.Image
    x, y     float64
    visible  bool
    alpha    float64
    parent   *Sprite
    dirty    bool
    
    // 階層的Z順序
    // 要件 1.1: スプライトはZ_Pathを持つ
    zPath    *ZPath
    
    // 子スプライト管理
    // 要件 9.1: PictureSpriteは子スプライトを持てる
    children []*Sprite
    
    // ソートキャッシュ
    // 要件 5.4: 比較結果をキャッシュして再利用する
    sortKey  string  // Z_Pathの文字列表現（キャッシュ用）
}

// ZPath はスプライトのZ_Pathを返す
func (s *Sprite) ZPath() *ZPath {
    return s.zPath
}

// SetZPath はスプライトのZ_Pathを設定する
// 要件 8.2: Local_Z_Orderが変更されたとき、Z_Pathを再計算する
func (s *Sprite) SetZPath(zPath *ZPath) {
    s.zPath = zPath
    s.sortKey = zPath.String()
    s.dirty = true
}

// GetChildren は子スプライトのリストを返す
func (s *Sprite) GetChildren() []*Sprite {
    return s.children
}

// AddChild は子スプライトを追加する
// 要件 1.4: 子スプライトが作成されたとき、親のZ_Pathを継承し、自身のLocal_Z_Orderを追加する
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
```

### 4. 拡張されたSpriteManager (pkg/graphics/sprite.go)

#### 構造体の変更

```go
// SpriteManager はスプライトを管理する（階層的Z順序対応）
type SpriteManager struct {
    mu            sync.RWMutex
    sprites       map[int]*Sprite
    nextID        int
    
    // 階層的Z順序
    // 要件 2.1: 各親スプライトごとにZ_Order_Counterを管理する
    zOrderCounter *ZOrderCounter
    
    // ソートキャッシュ
    // 要件 7.1: Z_Pathのソート結果をキャッシュする
    sorted        []*Sprite
    needSort      bool
}

// NewSpriteManager は新しいSpriteManagerを作成する
func NewSpriteManager() *SpriteManager {
    return &SpriteManager{
        sprites:       make(map[int]*Sprite),
        nextID:        1,
        zOrderCounter: NewZOrderCounter(),
        needSort:      true,
    }
}

// CreateSpriteWithZPath は新しいスプライトを作成してZ_Pathを設定する
// 要件 2.2, 2.3, 2.4: 操作時にZ_Order_Counterを使用してLocal_Z_Orderを割り当てる
func (sm *SpriteManager) CreateSpriteWithZPath(img *ebiten.Image, parent *Sprite) *Sprite {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    s := NewSprite(sm.nextID, img)
    sm.sprites[s.id] = s
    sm.nextID++
    
    // 親のZ_Pathを継承してZ_Pathを設定
    var parentID int
    var parentZPath *ZPath
    if parent != nil {
        parentID = parent.id
        parentZPath = parent.zPath
        parent.AddChild(s)
    }
    
    // 要件 2.5: Z_Order_Counterをインクリメント
    localZOrder := sm.zOrderCounter.GetNext(parentID)
    s.SetZPath(NewZPathFromParent(parentZPath, localZOrder))
    
    // 要件 7.2: スプライトの変更時にソートが必要であることをマークする
    sm.needSort = true
    
    return s
}

// CreateRootSprite はルートスプライト（ウインドウ）を作成する
// 要件 1.3: Root_Spriteは単一要素のZ_Path（例: [0]）を持つ
func (sm *SpriteManager) CreateRootSprite(img *ebiten.Image, windowZOrder int) *Sprite {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    s := NewSprite(sm.nextID, img)
    sm.sprites[s.id] = s
    sm.nextID++
    
    // ルートスプライトは単一要素のZ_Path
    s.SetZPath(NewZPath(windowZOrder))
    
    sm.needSort = true
    return s
}
```

#### ソートと描画

```go
// sortSprites はスプライトをZ_Pathの辞書順でソートする
// 要件 1.5: Z_Pathの辞書順比較でスプライトの描画順序を決定する
// 要件 7.1: Z_Pathのソート結果をキャッシュする
func (sm *SpriteManager) sortSprites() {
    sm.sorted = make([]*Sprite, 0, len(sm.sprites))
    for _, s := range sm.sprites {
        sm.sorted = append(sm.sorted, s)
    }
    
    sort.Slice(sm.sorted, func(i, j int) bool {
        return sm.sorted[i].zPath.Less(sm.sorted[j].zPath)
    })
    
    sm.needSort = false
}

// Draw はすべての可視スプライトをZ_Path順で描画する
// 要件 3.1: 親スプライトを先に描画し、その後に子スプライトを描画する
// 要件 3.2: 同じ親を持つ子スプライトをLocal_Z_Order順で描画する
func (sm *SpriteManager) Draw(screen *ebiten.Image) {
    sm.mu.Lock()
    if sm.needSort {
        sm.sortSprites()
    }
    sorted := sm.sorted
    sm.mu.Unlock()
    
    for _, s := range sorted {
        // 要件 3.4: 親スプライトが非表示のとき、子スプライトも描画しない
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

### 5. 動的Z順序変更 (pkg/graphics/sprite.go)

```go
// BringToFront はスプライトを最前面に移動する
// 要件 8.4: スプライトを最前面に移動するメソッドを提供する
func (sm *SpriteManager) BringToFront(spriteID int) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    s := sm.sprites[spriteID]
    if s == nil {
        return fmt.Errorf("sprite not found: %d", spriteID)
    }
    
    // 親のIDを取得
    var parentID int
    if s.parent != nil {
        parentID = s.parent.id
    }
    
    // 新しいZ順序を取得
    newLocalZOrder := sm.zOrderCounter.GetNext(parentID)
    
    // Z_Pathを再計算
    // 要件 8.2: Local_Z_Orderが変更されたとき、Z_Pathを再計算する
    var parentZPath *ZPath
    if s.parent != nil {
        parentZPath = s.parent.zPath
    }
    s.SetZPath(NewZPathFromParent(parentZPath, newLocalZOrder))
    
    // 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
    sm.updateChildrenZPaths(s)
    
    sm.needSort = true
    return nil
}

// SendToBack はスプライトを最背面に移動する
// 要件 8.5: スプライトを最背面に移動するメソッドを提供する
func (sm *SpriteManager) SendToBack(spriteID int) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    s := sm.sprites[spriteID]
    if s == nil {
        return fmt.Errorf("sprite not found: %d", spriteID)
    }
    
    // 最小のZ順序を見つける
    minZOrder := 0
    for _, other := range sm.sprites {
        if other.parent == s.parent && other.id != s.id {
            if other.zPath != nil && other.zPath.LocalZOrder() < minZOrder {
                minZOrder = other.zPath.LocalZOrder()
            }
        }
    }
    
    // 新しいZ順序を設定（最小値 - 1）
    newLocalZOrder := minZOrder - 1
    
    var parentZPath *ZPath
    if s.parent != nil {
        parentZPath = s.parent.zPath
    }
    s.SetZPath(NewZPathFromParent(parentZPath, newLocalZOrder))
    
    sm.updateChildrenZPaths(s)
    sm.needSort = true
    return nil
}

// updateChildrenZPaths は子スプライトのZ_Pathを再帰的に更新する
// 要件 8.3: 親スプライトが変更されたとき、子スプライトのZ_Pathを再計算する
func (sm *SpriteManager) updateChildrenZPaths(parent *Sprite) {
    for _, child := range parent.children {
        localZOrder := child.zPath.LocalZOrder()
        child.SetZPath(NewZPathFromParent(parent.zPath, localZOrder))
        sm.updateChildrenZPaths(child)
    }
}
```

### 6. ウインドウZ順序の更新 (pkg/graphics/window_sprite.go)

```go
// UpdateWindowZOrder はウインドウのZ順序を更新し、子スプライトのZ_Pathも更新する
// 要件 4.3: ウインドウのZ順序変更時に、そのウインドウの子スプライトのZ_Pathを更新する
// 要件 4.4: ウインドウが前面に移動したとき、そのウインドウのZ_Pathを更新する
func (ws *WindowSprite) UpdateWindowZOrder(newZOrder int, sm *SpriteManager) {
    ws.window.ZOrder = newZOrder
    
    // ウインドウスプライトのZ_Pathを更新
    ws.sprite.SetZPath(NewZPath(newZOrder))
    
    // 子スプライトのZ_Pathを再帰的に更新
    sm.updateChildrenZPaths(ws.sprite)
    
    sm.MarkNeedSort()
}
```

## データモデル

### Z_Pathの表現

```go
// Z_Pathの例
// ウインドウ0: [0]
// ウインドウ0の背景: [0, 0]
// ウインドウ0のキャスト1: [0, 1]
// ウインドウ0のテキスト1: [0, 2]
// ウインドウ0のキャスト2: [0, 3]
// ウインドウ1: [1]
// ウインドウ1の背景: [1, 0]

// 辞書順比較の例
// [0] < [0, 0] < [0, 1] < [0, 2] < [1] < [1, 0]
```

### 親子関係

**重要**: FILLYの仕様では、キャストやテキストは**ピクチャーに対して配置される**。
ウインドウではなく、ピクチャーが親となる。

```
デスクトップ（仮想ルート）
├── ウインドウ0 (WindowSprite) [0]
│   └── 背景ピクチャー (PictureSprite) [0, 0]
│       ├── キャスト1 (CastSprite) [0, 0, 0]
│       ├── テキスト1 (TextSprite) [0, 0, 1]
│       ├── MovePicで追加された画像 (PictureSprite) [0, 0, 2] ← 焼き付け対象
│       └── キャスト2 (CastSprite) [0, 0, 3]
├── ウインドウ1 (WindowSprite) [1]
│   └── 背景ピクチャー (PictureSprite) [1, 0]
│       └── キャスト1 (CastSprite) [1, 0, 0]
└── ウインドウ2 (WindowSprite) [2]
    └── 背景ピクチャー (PictureSprite) [2, 0]
```

**根拠**:
- PutCast: `PutCast(Cast, Pic, X, Y, ...)` - 第2引数は**ピクチャー番号**
- TextWrite: `TextWrite(pic_no, x, y, text)` - 第1引数は**ピクチャー番号**
- MovePic: `MovePic(src_pic, ..., dst_pic, ...)` - 転送先は**ピクチャー番号**

**焼き付けロジック**:
- 連続するPictureSprite（MovePicで作成）は、直前のPictureSpriteに焼き付ける
- CastSpriteやTextSpriteの後にMovePicが呼ばれた場合は、新しいPictureSpriteを作成

## 正確性プロパティ

### Z_Pathのプロパティ

**プロパティ1: Z_Pathの一意性**
*任意の*2つのスプライトについて、それらのZ_Pathは異なる
**検証: 要件 1.1, 1.2**

**プロパティ2: Z_Pathの継承**
*任意の*子スプライトについて、そのZ_Pathは親のZ_Pathをプレフィックスとして持つ
**検証: 要件 1.2, 1.4**

**プロパティ3: ルートスプライトのZ_Path**
*任意の*ルートスプライトについて、そのZ_Pathは単一要素である
**検証: 要件 1.3**

### 操作順序のプロパティ

**プロパティ4: 操作順序の反映**
*任意の*同じ親を持つ2つのスプライトについて、後から作成されたスプライトのLocal_Z_Orderは大きい
**検証: 要件 2.2, 2.3, 2.4, 2.5**

**プロパティ5: タイプ非依存性**
*任意の*スプライトについて、そのZ順序はタイプ（キャスト、テキスト、ピクチャ）に依存しない
**検証: 要件 2.6**

### 描画順序のプロパティ

**プロパティ6: 親子描画順序**
*任意の*親子関係について、親スプライトは子スプライトより先に描画される
**検証: 要件 3.1**

**プロパティ7: 兄弟描画順序**
*任意の*同じ親を持つ2つのスプライトについて、Local_Z_Orderが小さいスプライトが先に描画される
**検証: 要件 3.2**

**プロパティ8: 可視性の継承**
*任意の*親が非表示のスプライトについて、その子スプライトも描画されない
**検証: 要件 3.4**

### ウインドウ間のプロパティ

**プロパティ9: ウインドウ間の描画順序**
*任意の*2つのウインドウについて、前面のウインドウのすべての子スプライトは背面のウインドウのすべての子スプライトより後に描画される
**検証: 要件 4.1, 4.2**

**プロパティ10: ウインドウZ順序更新の伝播**
*任意の*ウインドウについて、Z順序が変更されたとき、そのすべての子スプライトのZ_Pathが更新される
**検証: 要件 4.3, 4.4**

### 比較のプロパティ

**プロパティ11: 辞書順比較の正確性**
*任意の*2つのZ_Pathについて、Compare関数は正しい辞書順を返す
**検証: 要件 5.1, 5.2, 5.3**

### パフォーマンスのプロパティ

**プロパティ12: ソートキャッシュの有効性**
*任意の*変更がない状態で、ソートは再実行されない
**検証: 要件 7.1, 7.2**

**プロパティ13: 比較の計算量**
*任意の*Z_Path比較について、計算量はO(depth)である
**検証: 要件 7.4**

### 動的変更のプロパティ

**プロパティ14: 最前面移動**
*任意の*スプライトについて、BringToFront後にそのスプライトは同じ親を持つ兄弟の中で最大のLocal_Z_Orderを持つ
**検証: 要件 8.4**

**プロパティ15: 最背面移動**
*任意の*スプライトについて、SendToBack後にそのスプライトは同じ親を持つ兄弟の中で最小のLocal_Z_Orderを持つ
**検証: 要件 8.5**

## エラーハンドリング

### エラーの分類

#### 致命的エラー
- Z_Pathの循環参照（親子関係のループ）

#### 非致命的エラー
- 存在しないスプライトIDへのアクセス
- 無効なZ_Path操作

### エラーログフォーマット

```
[timestamp] [level] [component] function_name: message (args)
```

例:
```
[19:43:20.577] [ERROR] [ZPath] BringToFront: sprite not found (spriteID=999)
[19:43:20.578] [WARN] [ZPath] UpdateWindowZOrder: window has no children (windowID=1)
```

## テスト戦略

### ユニットテスト

- ZPath: 作成、比較、文字列変換
- ZOrderCounter: カウンター取得、リセット
- Sprite: Z_Path設定、子スプライト管理
- SpriteManager: ソート、描画順序

### プロパティベーステスト

- プロパティ1-15（上記参照）
- テストライブラリ: gopter

### 統合テスト

- 既存のサンプルスクリプトの実行
- 描画順序の視覚的確認

## 実装の優先順位

### フェーズ1: 基盤
1. ZPath構造体の実装
2. ZOrderCounter構造体の実装
3. 比較関数の実装

### フェーズ2: Sprite拡張
1. SpriteにZ_Pathフィールドを追加
2. 子スプライト管理の実装
3. SpriteManagerの拡張

### フェーズ3: ソートと描画
1. Z_Pathによるソートの実装
2. 描画順序の変更
3. キャッシュの実装

### フェーズ4: 動的変更
1. BringToFrontの実装
2. SendToBackの実装
3. ウインドウZ順序更新の実装

### フェーズ5: 既存システムとの統合
1. WindowSpriteの更新
2. CastSpriteの更新
3. TextSpriteの更新
4. PictureSpriteの更新

### フェーズ6: デバッグ支援
1. Z_Pathの文字列表現
2. 階層ツリーの出力
3. デバッグオーバーレイ

## 技術的な考慮事項

### 並行性

- ZOrderCounterはsync.RWMutexで保護
- SpriteManagerのソートキャッシュはsync.RWMutexで保護
- Z_Path自体はイミュータブル（変更時は新しいインスタンスを作成）

### パフォーマンス

- Z_Pathの比較はO(depth)で効率的
- ソート結果はキャッシュされ、変更がない限り再利用
- 変更のあったサブツリーのみを再ソート（将来の最適化）

### メモリ管理

- Z_Pathは小さな整数配列なのでメモリ効率が良い
- ソートキャッシュはスプライト数に比例
- 子スプライト削除時に親からの参照も削除

## 既存システムとの互換性

### 移行戦略

1. **段階的移行**: 既存のCalculateGlobalZOrder関数は維持しつつ、新しいZ_Pathシステムを並行して実装
2. **テストによる検証**: 既存のサンプルスクリプトで動作確認

### 移行完了

**注意**: 2026年1月29日に、従来の`zOrder`フィールドは完全に削除されました。

以下の変更が行われました：
- `Sprite`構造体から`zOrder`フィールドを削除
- `Sprite.ZOrder()`および`Sprite.SetZOrder()`メソッドを削除
- `SpriteManager.SetSpriteZOrder()`および`SpriteManager.GetSpriteByZOrder()`メソッドを削除
- 各スプライトマネージャー（CastSprite, TextSprite, ShapeSprite, PictureSprite, WindowSprite）から`SetZOrder`呼び出しを削除
- ソート時のフォールバックをzOrderからスプライトIDに変更

現在、すべてのZ順序管理はZ_Pathシステムで行われています。

### 互換性関数

```go
// 要件 6.4: 1次元Z順序から階層的Z順序への移行パスを提供する
// この関数は既存コードからの移行を支援するために残されています
func ConvertFlatZOrderToZPath(windowZOrder, localZOrder int) *ZPath {
    return NewZPath(windowZOrder, localZOrder)
}

// CalculateGlobalZOrder は既存のグローバルZ順序計算関数
// 一部の既存コードとの互換性のために残されています
func CalculateGlobalZOrder(windowZOrder, localZOrder int) int {
    return windowZOrder*ZOrderWindowRange + localZOrder
}
```

## デバッグ支援

### Z_Pathの可視化

```go
// 要件 10.1: スプライトのZ_Pathを文字列として取得できる
func (s *Sprite) ZPathString() string {
    if s.zPath == nil {
        return "nil"
    }
    return s.zPath.String()
}

// 要件 10.2: スプライト階層をツリー形式で出力できる
func (sm *SpriteManager) PrintHierarchy() string {
    var sb strings.Builder
    
    // ルートスプライトを見つける
    roots := make([]*Sprite, 0)
    for _, s := range sm.sprites {
        if s.parent == nil {
            roots = append(roots, s)
        }
    }
    
    // Z_Path順でソート
    sort.Slice(roots, func(i, j int) bool {
        return roots[i].zPath.Less(roots[j].zPath)
    })
    
    // ツリー形式で出力
    for _, root := range roots {
        sm.printSpriteTree(&sb, root, 0)
    }
    
    return sb.String()
}

func (sm *SpriteManager) printSpriteTree(sb *strings.Builder, s *Sprite, depth int) {
    indent := strings.Repeat("  ", depth)
    sb.WriteString(fmt.Sprintf("%s- Sprite %d: %s\n", indent, s.id, s.zPath.String()))
    
    for _, child := range s.children {
        sm.printSpriteTree(sb, child, depth+1)
    }
}

// 要件 10.3: 描画順序のリストを出力できる
func (sm *SpriteManager) PrintDrawOrder() string {
    sm.mu.Lock()
    if sm.needSort {
        sm.sortSprites()
    }
    sorted := sm.sorted
    sm.mu.Unlock()
    
    var sb strings.Builder
    sb.WriteString("Draw Order:\n")
    for i, s := range sorted {
        sb.WriteString(fmt.Sprintf("  %d. Sprite %d: %s\n", i+1, s.id, s.zPath.String()))
    }
    return sb.String()
}
```

### デバッグオーバーレイ

```go
// 要件 10.4: デバッグモードが有効なとき、Z_Pathをオーバーレイ表示できる
func (sm *SpriteManager) DrawDebugOverlay(screen *ebiten.Image, enabled bool) {
    if !enabled {
        return
    }
    
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    
    for _, s := range sm.sprites {
        if !s.IsEffectivelyVisible() {
            continue
        }
        
        x, y := s.AbsolutePosition()
        label := s.zPath.String()
        
        // 半透明黒背景 + 白テキストでZ_Pathを表示
        // ...
    }
}
```

## パッケージ構成

```
pkg/graphics/
├── sprite.go           # Sprite, SpriteManager（拡張）
├── zpath.go            # ZPath, ZOrderCounter（新規）
├── zpath_test.go       # ZPathのユニットテスト（新規）
├── window_sprite.go    # WindowSprite（更新）
├── cast_sprite.go      # CastSprite（更新）
├── text_sprite.go      # TextSprite（更新）
├── picture_sprite.go   # PictureSprite（更新）
└── ...
```

## 依存関係

### 外部ライブラリ

- **github.com/hajimehoshi/ebiten/v2**: 描画エンジン（既存）

### 内部パッケージ

- **pkg/graphics**: 既存のスプライトシステム
- **pkg/logger**: ログ出力

## ピクチャのスプライト化設計

### 概要

ピクチャをロード（LoadPic）した時点で非表示のPictureSpriteを作成し、ウインドウに関連付けられる前でもキャストやテキストの親として機能できるようにする。

### 現在の問題点

```
現在の実装:
┌─────────────────────────────────────────────────────────────────┐
│ 1. LoadPic(35, "image.bmp")  → Picture構造体のみ作成            │
│ 2. TextWrite(35, 10, 10, "Hello")  → 親が見つからない！         │
│ 3. SetPic(0, 35)  → ここで初めてPictureSpriteが作成される       │
│                                                                 │
│ 問題: TextWriteの時点でPictureSpriteが存在しないため、          │
│       TextSpriteの親が設定できない                              │
└─────────────────────────────────────────────────────────────────┘
```

### 新しい設計

```
新しい実装:
┌─────────────────────────────────────────────────────────────────┐
│ 1. LoadPic(35, "image.bmp")                                     │
│    → Picture構造体を作成                                        │
│    → 非表示のPictureSpriteを作成（状態: 未関連付け）            │
│    → pictureSpriteMap[35] = pictureSprite                       │
│                                                                 │
│ 2. TextWrite(35, 10, 10, "Hello")                               │
│    → pictureSpriteMap[35]を親として取得                         │
│    → TextSpriteを作成し、PictureSpriteの子として追加            │
│    → 親が未関連付けなので、TextSpriteも描画されない             │
│                                                                 │
│ 3. SetPic(0, 35)                                                │
│    → PictureSpriteをWindowSpriteの子として関連付け              │
│    → PictureSpriteの状態を「関連付け済み」に変更                │
│    → PictureSpriteとその子（TextSprite）のZ_Pathを更新          │
│    → PictureSpriteが表示状態になり、子も描画される              │
└─────────────────────────────────────────────────────────────────┘
```

### PictureSpriteの状態遷移

```
┌─────────────────┐     SetPic()      ┌─────────────────┐
│   未関連付け    │ ─────────────────→ │   関連付け済み  │
│  (Unattached)   │                    │   (Attached)    │
│                 │                    │                 │
│ - 非表示        │                    │ - 親の可視性に  │
│ - 親なし        │                    │   従って表示    │
│ - Z_Path: nil   │                    │ - 親: Window    │
│   または仮の値  │                    │ - Z_Path: 有効  │
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

### データ構造

```go
// PictureSpriteState はPictureSpriteの状態を表す
type PictureSpriteState int

const (
    // PictureSpriteUnattached は未関連付け状態
    // 要件 12.1: 「未関連付け」状態
    PictureSpriteUnattached PictureSpriteState = iota
    
    // PictureSpriteAttached は関連付け済み状態
    // 要件 12.1: 「関連付け済み」状態
    PictureSpriteAttached
)

// PictureSprite はピクチャのスプライト表現
type PictureSprite struct {
    sprite    *Sprite
    pictureID int
    state     PictureSpriteState
    windowID  int  // 関連付けられたウインドウID（-1 = 未関連付け）
}

// SpriteManager の拡張
type SpriteManager struct {
    // ... 既存フィールド ...
    
    // 要件 12.4: ピクチャ番号からPictureSpriteを効率的に検索できる
    pictureSpriteMap map[int]*PictureSprite  // pictureID -> PictureSprite
}
```

### 主要メソッド

```go
// CreatePictureSpriteOnLoad はLoadPic時に非表示のPictureSpriteを作成する
// 要件 11.1: LoadPicが呼び出されたとき、非表示のPictureSpriteを作成する
func (sm *SpriteManager) CreatePictureSpriteOnLoad(pictureID int, img *ebiten.Image) *PictureSprite {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    // 既存のPictureSpriteがあれば削除
    if existing, ok := sm.pictureSpriteMap[pictureID]; ok {
        sm.removePictureSprite(existing)
    }
    
    // 新しいスプライトを作成（非表示）
    s := NewSprite(sm.nextID, img)
    s.visible = false  // 未関連付けなので非表示
    sm.sprites[s.id] = s
    sm.nextID++
    
    // PictureSpriteを作成
    ps := &PictureSprite{
        sprite:    s,
        pictureID: pictureID,
        state:     PictureSpriteUnattached,
        windowID:  -1,
    }
    
    // 要件 11.2: ピクチャ番号をキーとして管理される
    sm.pictureSpriteMap[pictureID] = ps
    
    sm.needSort = true
    return ps
}

// AttachPictureSpriteToWindow はSetPic時にPictureSpriteをウインドウに関連付ける
// 要件 11.3: SetPicが呼び出されたとき、既存のPictureSpriteをウインドウの子として関連付ける
// 要件 11.4: SetPicが呼び出されたとき、PictureSpriteを表示状態にする
func (sm *SpriteManager) AttachPictureSpriteToWindow(pictureID int, windowSprite *WindowSprite) error {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    ps, ok := sm.pictureSpriteMap[pictureID]
    if !ok {
        return fmt.Errorf("picture sprite not found: %d", pictureID)
    }
    
    // ウインドウの子として追加
    windowSprite.sprite.AddChild(ps.sprite)
    
    // Z_Pathを設定
    localZOrder := sm.zOrderCounter.GetNext(windowSprite.sprite.id)
    ps.sprite.SetZPath(NewZPathFromParent(windowSprite.sprite.zPath, localZOrder))
    
    // 状態を更新
    ps.state = PictureSpriteAttached
    ps.windowID = windowSprite.window.ID
    ps.sprite.visible = true  // 表示状態に
    
    // 要件 11.7: 既存の子スプライトのZ_Pathを更新
    sm.updateChildrenZPaths(ps.sprite)
    
    sm.needSort = true
    return nil
}

// GetPictureSpriteByPictureID はピクチャ番号からPictureSpriteを取得する
// 要件 11.5, 11.6: ウインドウに関連付けられていないピクチャに対するCastSet/TextWriteで使用
// 要件 12.4: ピクチャ番号からPictureSpriteを効率的に検索できる
func (sm *SpriteManager) GetPictureSpriteByPictureID(pictureID int) *PictureSprite {
    sm.mu.RLock()
    defer sm.mu.RUnlock()
    return sm.pictureSpriteMap[pictureID]
}

// FreePictureSprite はピクチャ解放時にPictureSpriteを削除する
// 要件 11.8: ピクチャが解放されたとき、対応するPictureSpriteとその子スプライトを削除する
func (sm *SpriteManager) FreePictureSprite(pictureID int) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    ps, ok := sm.pictureSpriteMap[pictureID]
    if !ok {
        return
    }
    
    sm.removePictureSprite(ps)
    delete(sm.pictureSpriteMap, pictureID)
    sm.needSort = true
}

// removePictureSprite はPictureSpriteとその子を削除する（内部メソッド）
func (sm *SpriteManager) removePictureSprite(ps *PictureSprite) {
    // 子スプライトを再帰的に削除
    for _, child := range ps.sprite.children {
        sm.removeSprite(child)
    }
    
    // 親から削除
    if ps.sprite.parent != nil {
        ps.sprite.parent.RemoveChild(ps.sprite.id)
    }
    
    // スプライトを削除
    delete(sm.sprites, ps.sprite.id)
}
```

### 描画時の状態チェック

```go
// IsEffectivelyVisible は実効的な可視性を返す
// 要件 12.2: PictureSpriteが「未関連付け」状態のとき、そのスプライトを描画しない
// 要件 12.3: PictureSpriteが「関連付け済み」状態のとき、親ウインドウの可視性に従って描画する
func (ps *PictureSprite) IsEffectivelyVisible() bool {
    if ps.state == PictureSpriteUnattached {
        return false
    }
    return ps.sprite.IsEffectivelyVisible()
}
```

### MovePicとの連携

```go
// UpdatePictureSpriteImage はMovePic時にPictureSpriteの画像を更新する
// 要件 12.5: MovePicが呼び出されたとき、転送先ピクチャのPictureSpriteの画像を更新する
func (sm *SpriteManager) UpdatePictureSpriteImage(pictureID int, img *ebiten.Image) {
    sm.mu.Lock()
    defer sm.mu.Unlock()
    
    ps, ok := sm.pictureSpriteMap[pictureID]
    if !ok {
        return
    }
    
    ps.sprite.image = img
}
```

### TextWriteとの連携

**重要**: `TextWrite`は内部で`pic.Image`を新しい`*ebiten.Image`に置き換える。
これにより、`LoadPic`時に`CreatePictureSpriteOnLoad`に渡した画像参照と、
`TextWrite`後の`pic.Image`が異なるオブジェクトになる。

```go
// TextRenderer.Write() の内部処理:
// 1. pic.Imageから背景を取得
// 2. 新しいRGBA画像を作成
// 3. 背景とテキストを合成
// 4. pic.Image = ebiten.NewImageFromImage(rgba)  ← 新しい画像に置き換え！

// 解決策: TextWrite後にPictureSpriteの画像を更新する
func (gs *GraphicsSystem) TextWrite(picID, x, y int, text string) error {
    // ... 従来のTextRenderer.TextWrite()を呼び出し ...
    
    // TextWriteでpic.Imageが新しい画像に置き換えられるため、
    // PictureSpriteの画像も更新する必要がある
    if gs.pictureSpriteManager != nil {
        gs.pictureSpriteManager.UpdatePictureSpriteImage(picID, pic.Image)
    }
    
    // ... TextSpriteの作成 ...
}
```

この問題は、以下のシナリオで発生する:
1. `LoadPic(0, "image.bmp")` → `PictureSprite`が`pic.Image`への参照を保持
2. `TextWrite(0, 10, 10, "Hello")` → `pic.Image`が新しい画像に置き換えられる
3. `OpenWin(0, ...)` → ウインドウを開く
4. `MovePic(1, ..., 0, ...)` → `pic.Image`に描画するが、`PictureSprite`は古い画像を参照
5. 結果: アニメーションが画面に反映されない

### 親子関係の階層

```
新しい階層構造:
┌─────────────────────────────────────────────────────────────────┐
│ デスクトップ（仮想ルート）                                       │
│ ├── ウインドウ0 (WindowSprite) [0]                              │
│ │   └── ピクチャ35 (PictureSprite) [0, 0] ← SetPicで関連付け    │
│ │       ├── テキスト1 (TextSprite) [0, 0, 0] ← LoadPic後に作成  │
│ │       └── キャスト1 (CastSprite) [0, 0, 1]                    │
│ │                                                               │
│ └── ピクチャ99 (PictureSprite) [未関連付け] ← まだSetPicされていない
│     └── テキスト2 (TextSprite) [未関連付け] ← 親が未関連付けなので非表示
└─────────────────────────────────────────────────────────────────┘
```

### 正確性プロパティ（追加）

**プロパティ16: ピクチャスプライトの作成**
*任意の*LoadPic呼び出しについて、対応するPictureSpriteが作成される
**検証: 要件 11.1**

**プロパティ17: ピクチャスプライトの関連付け**
*任意の*SetPic呼び出しについて、PictureSpriteがウインドウの子として関連付けられる
**検証: 要件 11.3, 11.4**

**プロパティ18: 未関連付けピクチャの非表示**
*任意の*未関連付けPictureSpriteについて、そのスプライトは描画されない
**検証: 要件 12.2**

**プロパティ19: 子スプライトのZ_Path更新**
*任意の*SetPic呼び出しについて、既存の子スプライトのZ_Pathが更新される
**検証: 要件 11.7**

## 参考資料

- 要件定義書: `.kiro/specs/hierarchical-z-order/requirements.md`
- 既存の設計書: `.kiro/specs/3_graphics-system/design.md`
- 現在の実装: `pkg/graphics/sprite.go`, `pkg/graphics/layer.go`
