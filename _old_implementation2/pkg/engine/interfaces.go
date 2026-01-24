package engine

import (
	"image"
	"io"
	"io/fs"
)

// Renderer abstracts rendering operations to enable headless mode and testing.
// Production implementation uses Ebiten, mock implementation logs operations.
type Renderer interface {
	// RenderFrame renders the current engine state to the screen.
	// screen is the target image to render to (typically from Ebiten).
	// state contains all graphics state (pictures, windows, casts).
	RenderFrame(screen image.Image, state *EngineState)

	// Clear clears the screen with the specified color.
	Clear(color uint32)
}

// AssetLoader abstracts asset loading to support both filesystem and embedded modes.
// FilesystemAssetLoader reads from disk, EmbedFSAssetLoader reads from embed.FS.
type AssetLoader interface {
	// ReadFile reads the entire file content.
	// path is relative to the project directory.
	// Returns file content or error if not found.
	ReadFile(path string) ([]byte, error)

	// Exists checks if a file exists.
	// path is relative to the project directory.
	Exists(path string) bool

	// ListFiles lists files matching a pattern.
	// pattern can use wildcards (* and ?).
	// Returns list of matching file paths.
	ListFiles(pattern string) ([]string, error)
}

// ImageDecoder abstracts image decoding to enable testing without actual images.
// Production implementation decodes BMP files, mock returns test images.
type ImageDecoder interface {
	// Decode decodes an image from a reader.
	// Returns the decoded image and format name.
	Decode(r io.Reader) (image.Image, string, error)

	// DecodeConfig decodes only the image configuration (dimensions, color model).
	// Useful for getting image size without full decoding.
	DecodeConfig(r io.Reader) (image.Config, string, error)
}

// TickGenerator abstracts tick generation for testing.
// Production implementation uses wall-clock time, mock uses manual control.
type TickGenerator interface {
	// CalculateTickFromTime calculates the current tick from elapsed time.
	// elapsed is in seconds since playback started.
	// Returns the current tick number.
	CalculateTickFromTime(elapsed float64) int

	// GetLastDeliveredTick returns the last tick that was delivered.
	GetLastDeliveredTick() int

	// SetLastDeliveredTick sets the last delivered tick (for testing).
	SetLastDeliveredTick(tick int)
}

// FileSystem abstracts filesystem operations for testing.
// Wraps fs.FS interface with additional methods.
type FileSystem interface {
	fs.FS

	// ReadDir reads the directory and returns directory entries.
	ReadDir(name string) ([]fs.DirEntry, error)
}
