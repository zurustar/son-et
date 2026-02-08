# スプライトシステム

## 概要

FILLYエミュレータのスプライトシステムについて説明します。
スプライトシステムは、すべての描画要素（ウインドウ、ピクチャ、キャスト、テキスト、図形）を統一的に管理し、親子関係と階層（レイヤー）に基づいた描画順序を実現します。

### 設計原則

- **操作順序優先**: スプライトタイプに関係なく、操作順序（PutCast、TextWrite、MovePicの呼び出し順）でスライスに追加
- **スライス順序 = 描画順序**: 数値によるZ順序管理を廃止し、スライスの順序がそのまま描画順序
- **焼き付け禁止**: すべての描画要素はスプライトとして独立して管理される（ピクチャー画像への焼き付けは行わない）
- **既存互換性**: 既存のFILLYスクリプトAPIを維持

---

## 1. 階層定義（Layer 0-4）と描画順序ルール

### 階層定義

FILLYの描画システムでは、スプライトタイプごとに固定の階層（Layer）が割り当てられています。

| 階層 | 定数名 | 対象 | 説明 |
|------|--------|------|------|
| **Layer 0** | `LayerDesktop` | 仮想デスクトップ | 背景レイヤー |
| **Layer 1** | `LayerWindow` | WindowSprite | 仮想ウインドウ |
| **Layer 2** | `LayerPicture` | PictureSprite | ピクチャ（MovePicで作成） |
| **Layer 3** | `LayerText` | TextSprite, ShapeSprite | テキスト、プリミティブ図形（線、矩形、円など） |
| **Layer 4** | `LayerCast` | CastSprite | キャスト（アニメーションスプライト） |

### 描画順序ルール

描画順序は以下の3つのルールで決定されます。

| ルール | 説明 | 優先度 |
|--------|------|--------|
| **ルール1（同一階層内）** | 同じ階層内では、より新しく追加されたスプライトが前面に来る | **最優先** |
| **ルール2（階層間）** | 階層が深い（数値が大きい）ほど前面に来る | 次点 |
| **ルール3（優先順位）** | ルール1とルール2が競合する場合、ルール1が優先される | — |

**重要**: ルール1がルール2より優先されるため、後から追加されたピクチャ（Layer 2）が、先に追加されたキャスト（Layer 4）より前面に来ることがあります。

### 描画順序の具体例

ピクチャ → キャスト → ピクチャ の順で追加した場合:

```
追加順序:
  1. ピクチャA（Layer 2, 順序0）
  2. キャスト（Layer 4, 順序0）
  3. ピクチャB（Layer 2, 順序1）

描画順序（背面→前面）:
  ピクチャA → ピクチャB → キャスト
  ※ただし、ルール1により、ピクチャBはキャストより後に追加されたため、
    キャストはピクチャBに覆い隠される
```

### 階層別の子スプライト管理

各スプライトは `childrenByLayer` マップで階層ごとに子スプライトのスライスを管理します。

```go
// 各階層ごとにスライスを持ち、同じ階層内では追加順序で描画
childrenByLayer map[int][]*Sprite
```

描画時は Layer 0 → 1 → 2 → 3 → 4 の順に各階層のスライスを走査し、スライス内では先頭から末尾の順（追加順）で描画します。

### 親子関係に基づく描画

- 親スプライトが先に描画され、その後に子スプライトが階層順で描画される
- 親スプライトが非表示の場合、子スプライトも描画されない
- 親子関係の深さに制限はない（n次元の階層をサポート）

### ウインドウ間の描画順序

ウインドウはルートスプライトとして管理されます。

- ルートスプライトのスライス順序がウインドウの描画順序を決定
- 前面のウインドウのすべての子スプライトは、背面のウインドウのすべての子スプライトより後に描画される

```
デスクトップ（仮想ルート）
├── ウインドウ0 (WindowSprite)  ← 最背面
│   └── 背景ピクチャー (PictureSprite)
│       ├── [Layer 2] MovePicで追加されたピクチャー
│       ├── [Layer 3] テキスト (TextSprite)
│       ├── [Layer 3] 図形 (ShapeSprite)
│       └── [Layer 4] キャスト (CastSprite)
├── ウインドウ1 (WindowSprite)
│   └── ...
└── ウインドウ2 (WindowSprite)  ← 最前面
    └── ...
```

### 動的な描画順序変更

スプライトの描画順序は動的に変更できます。

| メソッド | 動作 |
|----------|------|
| `BringToFront()` | 同一階層スライスの末尾に移動（最前面） |
| `SendToBack()` | 同一階層スライスの先頭に移動（最背面） |

---

## 2. PictureとSpriteの関係

### 基本概念

FILLYでは、PictureとSpriteは明確に区別されます。

| 概念 | 説明 | 画面表示 |
|------|------|----------|
| **Picture** | メモリ上の画像データ | 直接表示されない |
| **Sprite** | 画面に表示される描画単位（位置・可視性・透明度を持つ） | 表示される |

### PictureからSpriteへの変換

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
                             │ OpenWin, SetPic, PutCast, TextWrite, MovePic
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

### スプライトタイプ一覧

| スプライトタイプ | 階層 | 作成契機 | 説明 |
|------------------|------|----------|------|
| **WindowSprite** | Layer 1 | `OpenWin()` | 仮想ウインドウ。子スプライトの親となる |
| **PictureSprite** | Layer 2 | `LoadPic()`, `MovePic()` | ピクチャの表示。ウインドウに関連付けて表示 |
| **TextSprite** | Layer 3 | `TextWrite()` | テキスト描画。差分抽出方式でアンチエイリアスを除去 |
| **ShapeSprite** | Layer 3 | `DrawRect()`, `DrawLine()` 等 | 図形描画（線、矩形、円など） |
| **CastSprite** | Layer 4 | `PutCast()` | キャスト（アニメーションスプライト）。透明色処理をサポート |

### スプライトの親決定ロジック

キャストやテキストは、配置先として指定されたピクチャーの PictureSprite の子として追加されます。
各関数の引数仕様については [言語仕様](language-spec.md) を参照してください。

| 関数 | 親となるPictureSprite |
|------|----------------------|
| `PutCast(src_pic, dst_pic, ...)` | dst_pic のPictureSprite |
| `TextWrite(text, pic_no, ...)` | pic_no のPictureSprite |
| `MovePic(src_pic, ..., dst_pic, ...)` | dst_pic のPictureSprite |

```filly
// 使用例
base_pic = CreatePic(25);           // 新しいピクチャーを作成（ID=27が返される）
OpenWin(base_pic, ...);             // ウインドウを開く → base_pic の PictureSprite が親候補に
PutCast(25, base_pic, 0, 0);        // base_pic の PictureSprite の子として CastSprite を作成
cast = PutCast(17, base_pic, ...);  // 同じく base_pic の PictureSprite の子
```

### Sprite構造体の主要フィールド

```go
type Sprite struct {
    id       int            // 一意のID
    image    *ebiten.Image  // 画像データ
    x, y     float64        // 位置（親からの相対座標）
    visible  bool           // 可視性フラグ
    alpha    float64        // 透明度（0.0〜1.0）
    parent   *Sprite        // 親スプライトへのポインタ（nilの場合はルート）
    layer    int            // 階層（0〜4）
    childrenByLayer map[int][]*Sprite  // 階層別の子スプライト
}
```

### 親子関係の座標・透明度計算

| 計算 | ルール |
|------|--------|
| **絶対位置** | 親の位置 + 自身の位置（再帰的に加算） |
| **実効透明度** | 親の透明度 × 自身の透明度（再帰的に乗算） |
| **可視性** | 親が非表示なら子も非表示 |

---

## 3. PictureSpriteの状態遷移

### 状態定義

PictureSpriteは以下の2つの状態を持ちます。

| 状態 | 定数名 | 説明 |
|------|--------|------|
| **未関連付け** | `PictureSpriteUnattached` | ウインドウに関連付けられていない。非表示 |
| **関連付け済み** | `PictureSpriteAttached` | ウインドウに関連付けられている。親の可視性に従って表示 |

### 状態遷移図

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

### 各遷移の詳細

#### LoadPic → 未関連付け状態

`LoadPic()` が呼び出されると、非表示の PictureSprite が作成されます。

- スプライトは `visible=false` で作成
- 親スプライトなし（ルートにも追加されない）
- `pictureSpriteMap` にピクチャー番号をキーとして登録

この時点でも、キャストやテキストの親として機能できます。

#### SetPic → 関連付け済み状態

`SetPic()` が呼び出されると、PictureSprite がウインドウの子として関連付けられます。

- ウインドウスプライトの子として追加
- `visible=true` に変更
- 状態を `PictureSpriteAttached` に更新

#### FreePic → 削除

`FreePic()` が呼び出されると、PictureSprite とその子スプライトが削除されます。

- PictureSprite 自体を削除
- 子スプライト（CastSprite、TextSprite等）も再帰的に削除
- `pictureSpriteMap` からエントリを削除

### 未関連付け状態でのスプライト管理

ウインドウに関連付けられていないピクチャでも、キャストやテキストの親として機能します。

| 操作 | 動作 |
|------|------|
| `PutCast(src, unattached_pic, ...)` | キャストを未関連付けピクチャの子として管理 |
| `TextWrite(unattached_pic, ...)` | テキストを未関連付けピクチャの子として管理 |

これにより、ウインドウに関連付ける前にピクチャ上にコンテンツを構築し、後から `SetPic()` や `MovePic()` で転送するパターンが可能になります。

---

## 4. MovePicによる覆い隠し・融合・子スプライト転送

### 4.1 MovePicによる覆い隠し

#### 背景

FILLYでは以下のパターンが頻繁に使用されます。

1. 黒文字でテキストを描画
2. `MovePic` で黒文字を別の位置に転送
3. 白文字で同じテキストを描画（元の位置を「消す」）
4. `MovePic` で全体を転送（白文字を覆い隠す）

このパターンでは、4番目の `MovePic` で作成される PictureSprite が3番目の白文字 TextSprite を覆い隠す必要があります。

#### 覆い隠しの仕組み

`MovePic` で作成された PictureSprite には、その時点で最新のZ-orderが割り当てられます。

| 対象 | 動作 |
|------|------|
| **TextSprite** | MovePicで作成されたPictureSpriteがTextSpriteより前面に描画される |
| **ShapeSprite** | MovePicで作成されたPictureSpriteがShapeSpriteより前面に描画される |

**重要**: スプライトを削除せずに、Z-orderによる覆い隠しで描画順序を制御します。

### 4.2 PictureSpriteの融合

同じピクチャーIDに対して `MovePic` が複数回呼び出された場合、既存の PictureSprite に画像を合成（融合）できます。

#### 融合の動作

| 項目 | 説明 |
|------|------|
| **検索** | PictureSpriteManager が融合可能なスプライトを検索 |
| **合成** | 新しい画像を既存の画像に合成 |
| **領域拡張** | 融合によりスプライトの領域を適切に拡張 |
| **Z-order更新** | 融合後のスプライトは、融合前に存在していた子スプライト（TextSprite等）より後に描画される |
| **スプライト数削減** | 融合により全体のスプライト数を削減し、パフォーマンスを向上 |

### 4.3 子スプライト転送

`MovePic` はピクチャの画像だけでなく、そのピクチャに属する子スプライト（キャスト、テキスト等）も転送先に移動させます。

```
転送前:
  ソースピクチャ (PictureSprite)
  ├── テキスト1 (TextSprite)
  └── キャスト1 (CastSprite)

MovePic(src, ..., dst, ...) 実行後:
  転送先ピクチャ (PictureSprite)
  ├── テキスト1 (TextSprite)  ← 転送された
  └── キャスト1 (CastSprite)  ← 転送された
```

---

## 5. レースコンディション対策（非表示作成パターン）

### 問題の背景

スプライトシステムでは、描画スレッドとスプライト作成スレッドが並行して動作します。スプライトを作成してから `Z_Path` を設定するまでの間に描画が行われると、`Z_Path` が未設定のスプライトが誤った位置（最背面）に描画される可能性があります。

### 問題のシナリオ

```
時刻T1: スプライト作成（visible=true）
時刻T2: 描画スレッドがsortSprites()を実行
        → Z_Pathが未設定のスプライトは最背面にソートされる
時刻T3: Z_Pathを設定
        → 既に描画が完了しているため、スプライトは背景の後ろに表示される
```

### 解決策: 非表示作成パターン

スプライトを非表示状態（`visible=false`）で作成し、`Z_Path` を設定した後に表示状態にすることで、レースコンディションを防ぎます。

```go
// 悪い例: レースコンディションが発生する可能性がある
sprite := CreateSprite(image)  // visible=true で作成
sprite.SetZPath(zPath)         // この間に描画が行われる可能性

// 良い例: レースコンディションを防ぐ
sprite := CreateSpriteHidden(image)  // visible=false で作成
sprite.SetZPath(zPath)               // Z_Pathを設定
sprite.SetVisible(true)              // 最後に表示状態にする
```

### 実装パターン

#### CastSprite作成時

```go
func CreateCastSpriteWithParent(img *ebiten.Image, parent *Sprite) *Sprite {
    // 1. 非表示状態で作成
    sprite := spriteManager.CreateSpriteHidden(img, LayerCast, nil)

    // 2. 親を設定（Z_Pathが継承される）
    if parent != nil {
        parent.AddChild(sprite)
    }

    // 3. Z_Pathが設定された後に表示状態にする
    sprite.SetVisible(true)

    return sprite
}
```

#### PictureSprite作成時

```go
func CreateBackgroundPictureSprite(img *ebiten.Image) *Sprite {
    // 1. 非表示状態で作成
    sprite := spriteManager.CreateSpriteHidden(img, LayerPicture, nil)

    // 2. Z_Pathを設定
    sprite.SetZPath(zPath)

    // 3. 最後に表示状態にする
    sprite.SetVisible(true)

    return sprite
}
```

### 適用箇所

以下の関数でこのパターンを適用します。

| 関数 | ファイル | 説明 |
|------|----------|------|
| `CreateCastSprite` | `pkg/graphics/cast_sprite.go` | キャストスプライト作成 |
| `CreateCastSpriteWithParent` | `pkg/graphics/cast_sprite.go` | 親付きキャストスプライト作成 |
| `CreateCastSpriteWithTransColor` | `pkg/graphics/cast_sprite.go` | 透明色付きキャストスプライト作成 |
| `CreateCastSpriteWithTransColorAndParent` | `pkg/graphics/cast_sprite.go` | 親・透明色付きキャストスプライト作成 |
| `CreateBackgroundPictureSprite` | `pkg/graphics/picture_sprite.go` | 背景ピクチャースプライト作成 |
| `OpenWin` | `pkg/graphics/graphics.go` | ウインドウオープン時の背景スプライト作成 |

### スレッドセーフティ

SpriteManager は `sync.RWMutex` で保護されており、描画中のスプライト変更を防ぎます。

| 操作 | ロック種別 |
|------|-----------|
| スプライト作成・削除・変更 | 書き込みロック（`Lock`） |
| 描画・検索 | 読み取りロック（`RLock`） |

---

## パッケージ構成

スプライトシステムは独立したパッケージ `pkg/sprite/` として実装されています。

```
pkg/sprite/                    ← スプライトシステム本体
├── sprite.go                  # Sprite, SpriteManager
├── sprite_test.go             # ユニットテスト
├── sprite_property_test.go    # プロパティベーステスト
├── window_sprite.go           # WindowSprite
├── picture_sprite.go          # PictureSprite
├── cast_sprite.go             # CastSprite
├── shape_sprite.go            # ShapeSprite
└── errors.go                  # スプライト関連エラー

pkg/graphics/                  ← 描画システム（spriteを使用）
├── graphics.go                # GraphicsSystem（VM統合）
├── picture.go                 # Picture管理
├── text_sprite.go             # TextSprite
├── transfer.go                # MovePic等の転送
└── ...
```

### 依存関係

```
pkg/graphics → pkg/sprite  （graphicsがspriteを使う）
pkg/vm → pkg/graphics      （VMがgraphicsを使う）
```

`pkg/sprite` は `pkg/graphics` に依存せず、独立して動作します。
