// Package fileutil provides unified file system access for both real and embedded file systems.
package fileutil

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// FileSystem は実ファイルシステムと埋め込みファイルシステムを統一的に扱うインターフェース
type FileSystem interface {
	// Open はファイルを開く（大文字小文字を無視）
	Open(name string) (fs.File, error)
	// ReadFile はファイルの内容を読み込む（大文字小文字を無視）
	ReadFile(name string) ([]byte, error)
	// ReadDir はディレクトリの内容を読み込む
	ReadDir(name string) ([]fs.DirEntry, error)
	// FindFile は大文字小文字を無視してファイルを検索し、実際のパスを返す
	FindFile(dir, filename string) (string, error)
	// BasePath はベースパスを返す
	BasePath() string
	// IsEmbedded は埋め込みファイルシステムかどうかを返す
	IsEmbedded() bool
}

// RealFS は実ファイルシステムへのアクセスを提供する
type RealFS struct {
	basePath string
}

// NewRealFS は実ファイルシステム用のFileSystemを作成する
func NewRealFS(basePath string) *RealFS {
	return &RealFS{basePath: basePath}
}

func (r *RealFS) Open(name string) (fs.File, error) {
	path := r.resolvePath(name)
	actualPath, err := r.findFileCaseInsensitive(path)
	if err != nil {
		return nil, err
	}
	return os.Open(actualPath)
}

func (r *RealFS) ReadFile(name string) ([]byte, error) {
	path := r.resolvePath(name)
	actualPath, err := r.findFileCaseInsensitive(path)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(actualPath)
}

func (r *RealFS) ReadDir(name string) ([]fs.DirEntry, error) {
	path := r.resolvePath(name)
	return os.ReadDir(path)
}

func (r *RealFS) FindFile(dir, filename string) (string, error) {
	searchDir := dir
	if r.basePath != "" && !filepath.IsAbs(dir) {
		searchDir = filepath.Join(r.basePath, dir)
	}
	return FindFileCaseInsensitive(searchDir, filename)
}

func (r *RealFS) BasePath() string {
	return r.basePath
}

func (r *RealFS) IsEmbedded() bool {
	return false
}

func (r *RealFS) resolvePath(name string) string {
	// 先頭の "/" や "\" を除去
	cleanName := strings.TrimPrefix(strings.TrimPrefix(name, "/"), "\\")
	if r.basePath != "" {
		return filepath.Join(r.basePath, cleanName)
	}
	return cleanName
}

func (r *RealFS) findFileCaseInsensitive(path string) (string, error) {
	// まず直接アクセスを試みる
	if _, err := os.Stat(path); err == nil {
		return path, nil
	}

	// 大文字小文字を無視して検索
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	return FindFileCaseInsensitive(dir, filename)
}

// EmbedFS は埋め込みファイルシステムへのアクセスを提供する
type EmbedFS struct {
	fsys     fs.FS
	basePath string
}

// NewEmbedFS は埋め込みファイルシステム用のFileSystemを作成する
func NewEmbedFS(fsys fs.FS, basePath string) *EmbedFS {
	return &EmbedFS{fsys: fsys, basePath: basePath}
}

func (e *EmbedFS) Open(name string) (fs.File, error) {
	path := e.resolvePath(name)
	actualPath, err := e.findFileCaseInsensitive(path)
	if err != nil {
		return nil, err
	}
	return e.fsys.Open(actualPath)
}

func (e *EmbedFS) ReadFile(name string) ([]byte, error) {
	path := e.resolvePath(name)
	actualPath, err := e.findFileCaseInsensitive(path)
	if err != nil {
		return nil, err
	}
	return fs.ReadFile(e.fsys, actualPath)
}

func (e *EmbedFS) ReadDir(name string) ([]fs.DirEntry, error) {
	path := e.resolvePath(name)
	return fs.ReadDir(e.fsys, path)
}

func (e *EmbedFS) FindFile(dir, filename string) (string, error) {
	searchDir := dir
	if e.basePath != "" {
		searchDir = e.basePath + "/" + dir
	}
	return FindFileCaseInsensitiveFS(e.fsys, searchDir, filename)
}

func (e *EmbedFS) BasePath() string {
	return e.basePath
}

func (e *EmbedFS) IsEmbedded() bool {
	return true
}

func (e *EmbedFS) resolvePath(name string) string {
	// 先頭の "/" や "\" を除去
	cleanName := strings.TrimPrefix(strings.TrimPrefix(name, "/"), "\\")
	// "." は現在のディレクトリを意味するので、basePathそのものを返す
	if cleanName == "." || cleanName == "" {
		if e.basePath != "" {
			return e.basePath
		}
		return "."
	}
	if e.basePath != "" {
		return e.basePath + "/" + cleanName
	}
	return cleanName
}

func (e *EmbedFS) findFileCaseInsensitive(path string) (string, error) {
	// まず直接アクセスを試みる
	if f, err := e.fsys.Open(path); err == nil {
		f.Close()
		return path, nil
	}

	// 大文字小文字を無視して検索
	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	// embed.FSでは "/" を使用
	dir = strings.ReplaceAll(dir, "\\", "/")
	return FindFileCaseInsensitiveFS(e.fsys, dir, filename)
}

// GetUnderlyingFS は内部のfs.FSを返す（embed.FSの場合のみ有効）
func (e *EmbedFS) GetUnderlyingFS() fs.FS {
	return e.fsys
}

// ReadFileWithReader はファイルを開いてReaderとして返す
// 呼び出し元でCloseする必要がある
func ReadFileWithReader(fsys FileSystem, name string) (io.ReadCloser, error) {
	return fsys.Open(name)
}

// WalkDir はディレクトリを再帰的に走査する
// 返されるパスはベースパスからの相対パス
func WalkDir(fsys FileSystem, root string, fn fs.WalkDirFunc) error {
	if embedFS, ok := fsys.(*EmbedFS); ok {
		path := root
		if embedFS.basePath != "" {
			// "." の場合はベースパスそのものを使用
			if root == "." || root == "" {
				path = embedFS.basePath
			} else if !strings.HasPrefix(root, embedFS.basePath) {
				path = embedFS.basePath + "/" + root
			}
		}
		basePath := embedFS.basePath
		return fs.WalkDir(embedFS.fsys, path, func(walkPath string, d fs.DirEntry, err error) error {
			// ベースパスからの相対パスに変換
			relPath := walkPath
			if basePath != "" && strings.HasPrefix(walkPath, basePath+"/") {
				relPath = strings.TrimPrefix(walkPath, basePath+"/")
			} else if basePath != "" && walkPath == basePath {
				relPath = "."
			}
			return fn(relPath, d, err)
		})
	}

	if realFS, ok := fsys.(*RealFS); ok {
		path := root
		if realFS.basePath != "" && !filepath.IsAbs(root) {
			path = filepath.Join(realFS.basePath, root)
		}
		basePath := realFS.basePath
		return filepath.WalkDir(path, func(walkPath string, d fs.DirEntry, err error) error {
			// ベースパスからの相対パスに変換
			relPath := walkPath
			if basePath != "" {
				rel, relErr := filepath.Rel(basePath, walkPath)
				if relErr == nil {
					relPath = rel
				}
			}
			return fn(relPath, d, err)
		})
	}

	return fmt.Errorf("unsupported file system type")
}
