package graphics

import "errors"

var (
	// ErrPictureNotFound はピクチャーが見つからない場合のエラー
	ErrPictureNotFound = errors.New("picture not found")

	// ErrWindowNotFound はウィンドウが見つからない場合のエラー
	ErrWindowNotFound = errors.New("window not found")

	// ErrCastNotFound はキャストが見つからない場合のエラー
	ErrCastNotFound = errors.New("cast not found")

	// ErrResourceLimitExceeded はリソース制限を超えた場合のエラー
	ErrResourceLimitExceeded = errors.New("resource limit exceeded")
)
