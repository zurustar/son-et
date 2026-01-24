package engine

import (
	"image"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2"
)

// AssetLoader abstracts the embedded filesystem for testing
// This interface allows mock implementations to be injected during tests
type AssetLoader interface {
	// ReadFile reads the named file from the embedded assets
	ReadFile(name string) ([]byte, error)

	// ReadDir reads the named directory from the embedded assets
	ReadDir(name string) ([]fs.DirEntry, error)
}

// ImageDecoder abstracts BMP decoding for testing
// This interface allows mock implementations to be injected during tests
type ImageDecoder interface {
	// Decode decodes a BMP image from the provided byte data
	Decode(data []byte) (image.Image, error)
}

// Renderer abstracts rendering operations for testing
// This interface allows headless testing without Ebitengine initialization
type Renderer interface {
	// RenderFrame renders the complete frame to the screen
	RenderFrame(screen *ebiten.Image, state *EngineState)
}
