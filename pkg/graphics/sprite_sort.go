// sprite_sort.go はスプライトのソート処理を提供する
// Z順序に基づくスプライトのソートアルゴリズムを含む
package graphics

// spriteItem はスプライトとそのタイプを保持する
type spriteItem struct {
	sprite     *Sprite
	spriteType string
	castSprite *CastSprite // キャストスプライトの場合のみ設定（透明色処理用）
	textSprite *TextSprite // テキストスプライトの場合のみ設定（背景ブレンド用）
}

// sortSpritesByZOrder はスプライトをZ順序でソートする
// 階層的Z順序システム: Z_Pathを使用してソートする
func sortSpritesByZOrder(items []spriteItem) {
	for i := 1; i < len(items); i++ {
		key := items[i]
		j := i - 1
		for j >= 0 {
			if !compareSpritesForSort(items[j], key) {
				break
			}
			items[j+1] = items[j]
			j--
		}
		items[j+1] = key
	}
}

// compareSpritesForSort は2つのスプライトを比較する（ソート用）
// aがbより後に描画されるべき場合（aがbより大きい場合）trueを返す
func compareSpritesForSort(a, b spriteItem) bool {
	// 両方ともZ_Pathを持つ場合は辞書順比較
	if a.sprite != nil && b.sprite != nil {
		aZPath := a.sprite.GetZPath()
		bZPath := b.sprite.GetZPath()

		if aZPath != nil && bZPath != nil {
			// aがbより大きい（後に描画される）場合true
			return aZPath.Compare(bZPath) > 0
		}

		// 片方だけZ_Pathを持つ場合
		if aZPath == nil && bZPath != nil {
			return false // aは先に描画される
		}
		if aZPath != nil && bZPath == nil {
			return true // aは後に描画される
		}
	}

	// 両方ともZ_Pathを持たない場合はIDで比較（安定ソート）
	aID := 0
	bID := 0
	if a.sprite != nil {
		aID = a.sprite.ID()
	}
	if b.sprite != nil {
		bID = b.sprite.ID()
	}
	return aID > bID
}

// sortCastSpritesByZPath はCastSpriteをZ_Pathでソートする
func sortCastSpritesByZPath(sprites []*CastSprite) {
	for i := 1; i < len(sprites); i++ {
		key := sprites[i]
		keyZPath := (*ZPath)(nil)
		if key.GetSprite() != nil {
			keyZPath = key.GetSprite().GetZPath()
		}
		j := i - 1
		for j >= 0 {
			jZPath := (*ZPath)(nil)
			if sprites[j].GetSprite() != nil {
				jZPath = sprites[j].GetSprite().GetZPath()
			}
			if jZPath == nil || (keyZPath != nil && !keyZPath.Less(jZPath)) {
				break
			}
			sprites[j+1] = sprites[j]
			j--
		}
		sprites[j+1] = key
	}
}

// sortTextSpritesByZPath はTextSpriteをZ_Pathでソートする
func sortTextSpritesByZPath(sprites []*TextSprite) {
	for i := 1; i < len(sprites); i++ {
		key := sprites[i]
		keyZPath := (*ZPath)(nil)
		if key.GetSprite() != nil {
			keyZPath = key.GetSprite().GetZPath()
		}
		j := i - 1
		for j >= 0 {
			jZPath := (*ZPath)(nil)
			if sprites[j].GetSprite() != nil {
				jZPath = sprites[j].GetSprite().GetZPath()
			}
			if jZPath == nil || (keyZPath != nil && !keyZPath.Less(jZPath)) {
				break
			}
			sprites[j+1] = sprites[j]
			j--
		}
		sprites[j+1] = key
	}
}
