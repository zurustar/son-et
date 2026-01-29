// Package sprite provides sprite-based rendering system with slice-based draw ordering.
package sprite

import "errors"

// スプライト関連のエラー定義
var (
	// ErrSpriteNotFound はスプライトが見つからない場合のエラー
	ErrSpriteNotFound = errors.New("sprite not found")

	// ErrRootSpriteNotFound はルートスプライトが見つからない場合のエラー
	ErrRootSpriteNotFound = errors.New("root sprite not found")

	// ErrInvalidParent は無効な親スプライトが指定された場合のエラー
	ErrInvalidParent = errors.New("invalid parent sprite")

	// ErrCircularReference は親子関係に循環参照がある場合のエラー
	ErrCircularReference = errors.New("circular reference detected in sprite hierarchy")

	// ErrPictureSpriteNotFound はピクチャースプライトが見つからない場合のエラー
	ErrPictureSpriteNotFound = errors.New("picture sprite not found")

	// ErrPictureSpriteAlreadyAttached はピクチャースプライトが既に関連付けられている場合のエラー
	ErrPictureSpriteAlreadyAttached = errors.New("picture sprite already attached to window")

	// ErrPictureSpriteNotAttached はピクチャースプライトが関連付けられていない場合のエラー
	ErrPictureSpriteNotAttached = errors.New("picture sprite not attached to any window")
)
