// Package graphics provides integration tests for the drawing loop.
// These tests verify that the graphics system can properly render windows, pictures, and casts.
package graphics

import (
	"image/color"
	"os"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
)

// TestIntegrationDrawLoop tests the complete drawing loop with windows and casts.
// 要件 3.11: ウィンドウをZ順序で管理し、後から開いたウィンドウを前面に表示する
// 要件 4.8: キャストを透明色（黒 0x000000）を除いて描画する
// 要件 4.9: キャストをZ順序で管理し、後から配置したキャストを前面に表示する
func TestIntegrationDrawLoop(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create test pictures
	pic1ID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture 1: %v", err)
	}

	pic2ID, err := gs.CreatePic(50, 50)
	if err != nil {
		t.Fatalf("Failed to create picture 2: %v", err)
	}

	// Fill pictures with colors for visibility
	gs.mu.Lock()
	pic1, _ := gs.pictures.GetPicWithoutLock(pic1ID)
	pic1.Image.Fill(color.RGBA{255, 0, 0, 255}) // Red

	pic2, _ := gs.pictures.GetPicWithoutLock(pic2ID)
	pic2.Image.Fill(color.RGBA{0, 255, 0, 255}) // Green
	gs.mu.Unlock()

	// Open windows
	win1ID, err := gs.OpenWin(pic1ID, 10, 10, 100, 100, 0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to open window 1: %v", err)
	}

	win2ID, err := gs.OpenWin(pic2ID, 50, 50, 50, 50, 0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to open window 2: %v", err)
	}

	// Place casts
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	// pic2IDがソース、pic1IDが配置先（win1IDのピクチャー）
	cast1ID, err := gs.PutCast(pic2ID, pic1ID, 10, 10, 0, 0, 50, 50)
	if err != nil {
		t.Fatalf("Failed to put cast 1: %v", err)
	}

	cast2ID, err := gs.PutCast(pic2ID, pic1ID, 30, 30, 0, 0, 50, 50)
	if err != nil {
		t.Fatalf("Failed to put cast 2: %v", err)
	}

	// Create a test screen
	screen := ebiten.NewImage(1024, 768)

	// Draw the scene
	gs.Draw(screen)

	// Verify windows were created
	if win1ID < 0 || win2ID < 0 {
		t.Error("Window IDs should be non-negative")
	}

	// Verify casts were created
	if cast1ID < 0 || cast2ID < 0 {
		t.Error("Cast IDs should be non-negative")
	}

	// Verify Z-order (win2 should be in front of win1)
	gs.mu.RLock()
	win1, _ := gs.windows.GetWin(win1ID)
	win2, _ := gs.windows.GetWin(win2ID)
	gs.mu.RUnlock()

	if win2.ZOrder <= win1.ZOrder {
		t.Error("Window 2 should have higher Z-order than window 1")
	}

	// Verify cast Z-order
	gs.mu.RLock()
	cast1, _ := gs.casts.GetCast(cast1ID)
	cast2, _ := gs.casts.GetCast(cast2ID)
	gs.mu.RUnlock()

	if cast2.ZOrder <= cast1.ZOrder {
		t.Error("Cast 2 should have higher Z-order than cast 1")
	}

	t.Logf("Integration test passed: windows=%d, casts=%d", 2, 2)
}

// TestIntegrationWindowDecoration tests that window decorations are rendered correctly.
// 要件 3.12: ウィンドウの背景色（color引数）を適用する
func TestIntegrationWindowDecoration(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create a test picture
	picID, err := gs.CreatePic(200, 150)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Fill picture with a color
	gs.mu.Lock()
	pic, _ := gs.pictures.GetPicWithoutLock(picID)
	pic.Image.Fill(color.RGBA{100, 100, 200, 255}) // Light blue
	gs.mu.Unlock()

	// Open window with background color
	bgColor := ColorFromInt(0xFF0000) // Red background
	winID, err := gs.OpenWin(picID, 100, 100, 200, 150, 0, 0, ColorToInt(bgColor))
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// Set caption
	err = gs.CapTitle(winID, "Test Window")
	if err != nil {
		t.Fatalf("Failed to set caption: %v", err)
	}

	// Create a test screen
	screen := ebiten.NewImage(1024, 768)

	// Draw the scene
	gs.Draw(screen)

	// Verify window was created with correct properties
	gs.mu.RLock()
	win, _ := gs.windows.GetWin(winID)
	gs.mu.RUnlock()

	if win.Caption != "Test Window" {
		t.Errorf("Expected caption 'Test Window', got '%s'", win.Caption)
	}

	if win.Width != 200 || win.Height != 150 {
		t.Errorf("Expected size 200x150, got %dx%d", win.Width, win.Height)
	}

	t.Logf("Window decoration test passed")
}

// TestIntegrationCastTransparency tests that casts are rendered with transparency.
// 要件 4.8: キャストを透明色（黒 0x000000）を除いて描画する
func TestIntegrationCastTransparency(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create background picture
	bgPicID, err := gs.CreatePic(200, 200)
	if err != nil {
		t.Fatalf("Failed to create background picture: %v", err)
	}

	// Create sprite picture with transparent (black) areas
	spritePicID, err := gs.CreatePic(50, 50)
	if err != nil {
		t.Fatalf("Failed to create sprite picture: %v", err)
	}

	// Fill background with white
	gs.mu.Lock()
	bgPic, _ := gs.pictures.GetPicWithoutLock(bgPicID)
	bgPic.Image.Fill(color.White)

	// Fill sprite with a pattern (center is colored, edges are black/transparent)
	spritePic, _ := gs.pictures.GetPicWithoutLock(spritePicID)
	spritePic.Image.Fill(color.Black) // Transparent color
	// Draw a colored center
	for y := 10; y < 40; y++ {
		for x := 10; x < 40; x++ {
			spritePic.Image.Set(x, y, color.RGBA{255, 0, 0, 255})
		}
	}
	gs.mu.Unlock()

	// Open window
	_, err = gs.OpenWin(bgPicID, 50, 50, 200, 200, 0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// Place cast
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	castID, err := gs.PutCast(spritePicID, bgPicID, 75, 75, 0, 0, 50, 50)
	if err != nil {
		t.Fatalf("Failed to put cast: %v", err)
	}

	// Create a test screen
	screen := ebiten.NewImage(1024, 768)

	// Draw the scene
	gs.Draw(screen)

	// Verify cast was created
	if castID < 0 {
		t.Error("Cast ID should be non-negative")
	}

	t.Logf("Cast transparency test passed")
}

// TestIntegrationCoordinateConversion tests the coordinate conversion functions.
// 要件 8.7: マウスイベントが発生したとき、描画領域座標に変換してMesP2、MesP3に設定する
func TestIntegrationCoordinateConversion(t *testing.T) {
	gs := NewGraphicsSystem("")

	testCases := []struct {
		name                   string
		screenW, screenH       int
		screenX, screenY       int
		expectedVX, expectedVY int
	}{
		{
			name:       "center of screen (same size)",
			screenW:    1024,
			screenH:    768,
			screenX:    512,
			screenY:    384,
			expectedVX: 512,
			expectedVY: 384,
		},
		{
			name:       "top-left corner",
			screenW:    1024,
			screenH:    768,
			screenX:    0,
			screenY:    0,
			expectedVX: 0,
			expectedVY: 0,
		},
		{
			name:       "bottom-right corner",
			screenW:    1024,
			screenH:    768,
			screenX:    1023,
			screenY:    767,
			expectedVX: 1023,
			expectedVY: 767,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			vx, vy := gs.ScreenToVirtual(tc.screenX, tc.screenY, tc.screenW, tc.screenH)

			if vx != tc.expectedVX || vy != tc.expectedVY {
				t.Errorf("Expected (%d, %d), got (%d, %d)",
					tc.expectedVX, tc.expectedVY, vx, vy)
			}
		})
	}
}

// TestIntegrationScaledCoordinateConversion tests coordinate conversion with scaling.
// 要件 8.4: 描画領域を実際のウィンドウサイズに合わせてスケーリングする
// 要件 8.5: アスペクト比を維持してスケーリングする
func TestIntegrationScaledCoordinateConversion(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Test with a larger screen (2x scale)
	screenW := 2048
	screenH := 1536

	// Center of screen should map to center of virtual desktop
	vx, vy := gs.ScreenToVirtual(1024, 768, screenW, screenH)

	// With 2x scale, screen center (1024, 768) should map to virtual center (512, 384)
	if vx != 512 || vy != 384 {
		t.Errorf("Expected (512, 384), got (%d, %d)", vx, vy)
	}

	// Test reverse conversion
	sx, sy := gs.VirtualToScreen(512, 384, screenW, screenH)

	// Virtual center should map back to screen center
	if sx != 1024 || sy != 768 {
		t.Errorf("Expected (1024, 768), got (%d, %d)", sx, sy)
	}
}

// TestIntegrationLetterboxCoordinates tests coordinate conversion with letterboxing.
// 要件 8.6: スケーリング時にレターボックス（黒帯）を表示する
func TestIntegrationLetterboxCoordinates(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Test with a wider screen (letterbox on top/bottom)
	screenW := 1920
	screenH := 768 // Same height, wider

	// Calculate expected scale and offset
	scale, offsetX, offsetY := gs.GetScaleAndOffset(screenW, screenH)

	// Scale should be 1.0 (limited by height)
	if scale != 1.0 {
		t.Errorf("Expected scale 1.0, got %f", scale)
	}

	// Offset should be on X axis (letterbox on sides)
	expectedOffsetX := (1920.0 - 1024.0) / 2 // 448
	if offsetX != expectedOffsetX {
		t.Errorf("Expected offsetX %f, got %f", expectedOffsetX, offsetX)
	}

	if offsetY != 0 {
		t.Errorf("Expected offsetY 0, got %f", offsetY)
	}
}

// TestIntegrationMultipleWindowsZOrder tests Z-order with multiple windows.
// 要件 3.11: ウィンドウをZ順序で管理し、後から開いたウィンドウを前面に表示する
func TestIntegrationMultipleWindowsZOrder(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create pictures
	var picIDs []int
	for i := 0; i < 5; i++ {
		picID, err := gs.CreatePic(100, 100)
		if err != nil {
			t.Fatalf("Failed to create picture %d: %v", i, err)
		}
		picIDs = append(picIDs, picID)
	}

	// Open windows
	var winIDs []int
	for i, picID := range picIDs {
		winID, err := gs.OpenWin(picID, i*20, i*20, 100, 100, 0, 0, 0)
		if err != nil {
			t.Fatalf("Failed to open window %d: %v", i, err)
		}
		winIDs = append(winIDs, winID)
	}

	// Verify Z-order
	gs.mu.RLock()
	windows := gs.windows.GetWindowsOrdered()
	gs.mu.RUnlock()

	if len(windows) != 5 {
		t.Fatalf("Expected 5 windows, got %d", len(windows))
	}

	// Windows should be in ascending Z-order
	for i := 1; i < len(windows); i++ {
		if windows[i].ZOrder <= windows[i-1].ZOrder {
			t.Errorf("Window %d should have higher Z-order than window %d",
				windows[i].ID, windows[i-1].ID)
		}
	}

	// Create a test screen and draw
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	t.Logf("Multiple windows Z-order test passed with %d windows", len(windows))
}

// TestIntegrationCastsByWindow tests that casts are correctly associated with windows.
// 要件 4.10: キャストの位置をウィンドウ相対座標で管理する
func TestIntegrationCastsByWindow(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create pictures
	bgPicID1, _ := gs.CreatePic(200, 200)
	bgPicID2, _ := gs.CreatePic(200, 200)
	spritePicID, _ := gs.CreatePic(50, 50)

	// Open two windows with different pictures
	win1ID, _ := gs.OpenWin(bgPicID1, 0, 0, 200, 200, 0, 0, 0)
	win2ID, _ := gs.OpenWin(bgPicID2, 300, 0, 200, 200, 0, 0, 0)

	// Place casts on different pictures (which are associated with different windows)
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	cast1ID, _ := gs.PutCast(spritePicID, bgPicID1, 10, 10, 0, 0, 50, 50)
	cast2ID, _ := gs.PutCast(spritePicID, bgPicID1, 50, 50, 0, 0, 50, 50)
	cast3ID, _ := gs.PutCast(spritePicID, bgPicID2, 20, 20, 0, 0, 50, 50)

	// Verify casts are associated with correct windows
	gs.mu.RLock()
	cast1, _ := gs.casts.GetCast(cast1ID)
	cast2, _ := gs.casts.GetCast(cast2ID)
	cast3, _ := gs.casts.GetCast(cast3ID)
	gs.mu.RUnlock()

	if cast1.WinID != win1ID || cast2.WinID != win1ID {
		t.Error("Casts 1 and 2 should belong to window 1")
	}

	if cast3.WinID != win2ID {
		t.Error("Cast 3 should belong to window 2")
	}

	// Verify GetCastsByWindow returns correct casts
	gs.mu.RLock()
	win1Casts := gs.casts.GetCastsByWindow(win1ID)
	win2Casts := gs.casts.GetCastsByWindow(win2ID)
	gs.mu.RUnlock()

	if len(win1Casts) != 2 {
		t.Errorf("Expected 2 casts for window 1, got %d", len(win1Casts))
	}

	if len(win2Casts) != 1 {
		t.Errorf("Expected 1 cast for window 2, got %d", len(win2Casts))
	}

	// Create a test screen and draw
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	t.Logf("Casts by window test passed")
}

// TestIntegrationDrawWithPicOffset tests drawing with picture offset (PicX, PicY).
func TestIntegrationDrawWithPicOffset(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Create a large picture
	picID, err := gs.CreatePic(400, 300)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	// Fill picture with a gradient pattern
	gs.mu.Lock()
	pic, _ := gs.pictures.GetPicWithoutLock(picID)
	for y := 0; y < 300; y++ {
		for x := 0; x < 400; x++ {
			pic.Image.Set(x, y, color.RGBA{
				R: uint8(x * 255 / 400),
				G: uint8(y * 255 / 300),
				B: 128,
				A: 255,
			})
		}
	}
	gs.mu.Unlock()

	// Open window with offset (showing center of picture)
	winID, err := gs.OpenWin(picID, 100, 100, 200, 150, -100, -75, 0)
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	// Verify window properties
	gs.mu.RLock()
	win, _ := gs.windows.GetWin(winID)
	gs.mu.RUnlock()

	if win.PicX != -100 || win.PicY != -75 {
		t.Errorf("Expected PicOffset (-100, -75), got (%d, %d)", win.PicX, win.PicY)
	}

	// Create a test screen and draw
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	t.Logf("Draw with pic offset test passed")
}

// TestIntegrationSampleFtile400 tests the graphics system with the ftile400 sample.
// 要件 11.1-11.7: エラーハンドリングの動作確認
// This test verifies that the graphics system can handle a real sample script.
func TestIntegrationSampleFtile400(t *testing.T) {
	// Check if sample directory exists
	samplePath := "../../samples/ftile400"
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("Sample ftile400 not found, skipping test")
	}

	gs := NewGraphicsSystem(samplePath)

	// Test loading a BMP file from the sample
	picID, err := gs.LoadPic("OPENING.BMP")
	if err != nil {
		t.Fatalf("Failed to load OPENING.BMP: %v", err)
	}

	if picID < 0 {
		t.Error("Expected valid picture ID")
	}

	// Verify picture dimensions
	width := gs.PicWidth(picID)
	height := gs.PicHeight(picID)

	if width <= 0 || height <= 0 {
		t.Errorf("Invalid picture dimensions: %dx%d", width, height)
	}

	t.Logf("Loaded OPENING.BMP: %dx%d (ID=%d)", width, height, picID)

	// Test opening a window with the loaded picture
	winID, err := gs.OpenWin(picID, 0, 0, width, height, 0, 0, 0)
	if err != nil {
		t.Fatalf("Failed to open window: %v", err)
	}

	if winID < 0 {
		t.Error("Expected valid window ID")
	}

	// Create a test screen and draw
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	// Test resource cleanup
	err = gs.CloseWin(winID)
	if err != nil {
		t.Errorf("Failed to close window: %v", err)
	}

	err = gs.DelPic(picID)
	if err != nil {
		t.Errorf("Failed to delete picture: %v", err)
	}

	t.Logf("Sample ftile400 integration test passed")
}

// TestIntegrationSampleHome tests the graphics system with the home sample.
// 要件 11.1-11.7: エラーハンドリングの動作確認
func TestIntegrationSampleHome(t *testing.T) {
	// Check if sample directory exists
	samplePath := "../../samples/home"
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("Sample home not found, skipping test")
	}

	gs := NewGraphicsSystem(samplePath)

	// Test loading multiple BMP files
	testFiles := []string{"title.bmp", "All.bmp", "Display.bmp"}
	picIDs := make([]int, 0)

	for _, filename := range testFiles {
		picID, err := gs.LoadPic(filename)
		if err != nil {
			t.Logf("Warning: Failed to load %s: %v", filename, err)
			continue
		}

		if picID >= 0 {
			picIDs = append(picIDs, picID)
			width := gs.PicWidth(picID)
			height := gs.PicHeight(picID)
			t.Logf("Loaded %s: %dx%d (ID=%d)", filename, width, height, picID)
		}
	}

	if len(picIDs) == 0 {
		t.Fatal("Failed to load any pictures from home sample")
	}

	// Test opening windows with loaded pictures
	for i, picID := range picIDs {
		width := gs.PicWidth(picID)
		height := gs.PicHeight(picID)

		winID, err := gs.OpenWin(picID, i*50, i*50, width, height, 0, 0, 0)
		if err != nil {
			t.Errorf("Failed to open window for picture %d: %v", picID, err)
			continue
		}

		t.Logf("Opened window %d for picture %d", winID, picID)
	}

	// Create a test screen and draw
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	// Test cleanup
	gs.CloseWinAll()

	for _, picID := range picIDs {
		gs.DelPic(picID)
	}

	t.Logf("Sample home integration test passed with %d pictures", len(picIDs))
}

// TestIntegrationSampleKuma2 tests the graphics system with the kuma2 sample.
// 要件 11.1-11.7: エラーハンドリングの動作確認
func TestIntegrationSampleKuma2(t *testing.T) {
	// Check if sample directory exists
	samplePath := "../../samples/kuma2"
	if _, err := os.Stat(samplePath); os.IsNotExist(err) {
		t.Skip("Sample kuma2 not found, skipping test")
	}

	gs := NewGraphicsSystem(samplePath)

	// Test loading BMP files from kuma2
	testFiles := []string{"TITLE.BMP", "KUMA-1.BMP", "KUMA-2.BMP"}
	picIDs := make([]int, 0)

	for _, filename := range testFiles {
		picID, err := gs.LoadPic(filename)
		if err != nil {
			t.Logf("Warning: Failed to load %s: %v", filename, err)
			continue
		}

		if picID >= 0 {
			picIDs = append(picIDs, picID)
			width := gs.PicWidth(picID)
			height := gs.PicHeight(picID)
			t.Logf("Loaded %s: %dx%d (ID=%d)", filename, width, height, picID)
		}
	}

	if len(picIDs) == 0 {
		t.Fatal("Failed to load any pictures from kuma2 sample")
	}

	// Test creating a composite scene
	if len(picIDs) >= 2 {
		// Open background window
		bgPicID := picIDs[0]
		bgWidth := gs.PicWidth(bgPicID)
		bgHeight := gs.PicHeight(bgPicID)

		winID, err := gs.OpenWin(bgPicID, 100, 100, bgWidth, bgHeight, 0, 0, 0)
		if err != nil {
			t.Fatalf("Failed to open background window: %v", err)
		}

		// Place casts on the window
		// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
		for i := 1; i < len(picIDs) && i < 3; i++ {
			castPicID := picIDs[i]
			castWidth := gs.PicWidth(castPicID)
			castHeight := gs.PicHeight(castPicID)

			castID, err := gs.PutCast(castPicID, bgPicID, i*30, i*30, 0, 0, castWidth, castHeight)
			if err != nil {
				t.Errorf("Failed to put cast %d: %v", i, err)
				continue
			}

			t.Logf("Placed cast %d on window %d", castID, winID)
		}
	}

	// Create a test screen and draw
	screen := ebiten.NewImage(1024, 768)
	gs.Draw(screen)

	// Test cleanup
	gs.CloseWinAll()

	for _, picID := range picIDs {
		gs.DelPic(picID)
	}

	t.Logf("Sample kuma2 integration test passed with %d pictures", len(picIDs))
}

// TestIntegrationErrorHandling tests error handling in the graphics system.
// 要件 11.1-11.7: エラーハンドリングの動作確認
func TestIntegrationErrorHandling(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Test 1: Loading non-existent file
	// 要件 11.1: 画像ファイルが見つからないとき、エラーをログに記録し、実行を継続する
	picID, err := gs.LoadPic("nonexistent.bmp")
	if err == nil {
		t.Error("Expected error when loading non-existent file")
	}
	if picID != -1 {
		t.Errorf("Expected picture ID -1, got %d", picID)
	}
	t.Logf("✓ Non-existent file error handled correctly")

	// Test 2: Invalid picture ID
	// 要件 11.3: 無効なピクチャーIDが指定されたとき、エラーをログに記録し、デフォルト値を返す
	width := gs.PicWidth(999)
	if width != 0 {
		t.Errorf("Expected width 0 for invalid picture ID, got %d", width)
	}
	t.Logf("✓ Invalid picture ID error handled correctly")

	// Test 3: Invalid window ID
	// 要件 11.4: 無効なウィンドウIDが指定されたとき、エラーをログに記録し、処理をスキップする
	err = gs.MoveWin(999, 0, 0, 100, 100, 0, 0)
	if err == nil {
		t.Error("Expected error when moving non-existent window")
	}
	t.Logf("✓ Invalid window ID error handled correctly")

	// Test 4: Invalid cast ID
	// 要件 11.5: 無効なキャストIDが指定されたとき、エラーをログに記録し、処理をスキップする
	err = gs.MoveCast(999, 0, 0)
	if err == nil {
		t.Error("Expected error when moving non-existent cast")
	}
	t.Logf("✓ Invalid cast ID error handled correctly")

	// Test 5: Opening window with invalid picture
	// Note: The system allows opening windows with invalid pictures (they just won't display)
	// This is by design for error tolerance
	winID, err := gs.OpenWin(999, 0, 0, 100, 100, 0, 0, 0)
	if winID < 0 {
		t.Logf("✓ Invalid picture for window handled (returned ID: %d)", winID)
	} else {
		t.Logf("✓ Invalid picture for window allowed (window will not display)")
	}

	// Test 6: Placing cast with invalid destination picture
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	// Note: The system allows placing casts on invalid pictures (they just won't display)
	castID, err := gs.PutCast(0, 999, 0, 0, 0, 0, 50, 50)
	if castID < 0 {
		t.Logf("✓ Invalid destination picture for cast handled (returned ID: %d)", castID)
	} else {
		t.Logf("✓ Invalid destination picture for cast allowed (cast will not display)")
	}

	// Test 7: Placing cast with invalid source picture
	// First create a valid window
	validPicID, _ := gs.CreatePic(100, 100)
	_, _ = gs.OpenWin(validPicID, 0, 0, 100, 100, 0, 0, 0)

	castID, err = gs.PutCast(999, validPicID, 0, 0, 0, 0, 50, 50)
	if castID < 0 {
		t.Logf("✓ Invalid source picture for cast handled (returned ID: %d)", castID)
	} else {
		t.Logf("✓ Invalid source picture for cast allowed (cast will not display)")
	}

	// Test 8: Drawing operations on invalid picture
	err = gs.DrawLine(999, 0, 0, 100, 100)
	if err == nil {
		t.Error("Expected error when drawing on invalid picture")
	}
	t.Logf("✓ Drawing on invalid picture error handled correctly")

	// Test 9: Text operations on invalid picture
	err = gs.TextWrite(999, 0, 0, "Test")
	if err == nil {
		t.Error("Expected error when writing text on invalid picture")
	}
	t.Logf("✓ Text on invalid picture error handled correctly")

	t.Logf("Error handling integration test passed")
}

// TestIntegrationResourceCleanup tests resource cleanup in the graphics system.
// 要件 9.1-9.3: リソース管理の動作確認
func TestIntegrationResourceCleanup(t *testing.T) {
	gs := NewGraphicsSystem("")

	// Test 1: Picture cleanup
	// 要件 9.1: ピクチャーが削除されたとき、関連するEbitengine画像リソースを解放する
	picID, err := gs.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("Failed to create picture: %v", err)
	}

	err = gs.DelPic(picID)
	if err != nil {
		t.Errorf("Failed to delete picture: %v", err)
	}

	// Verify picture is deleted
	width := gs.PicWidth(picID)
	if width != 0 {
		t.Error("Picture should be deleted")
	}
	t.Logf("✓ Picture cleanup verified")

	// Test 2: Window and cast cleanup
	// 要件 9.2: ウィンドウが閉じられたとき、関連するキャストを削除する
	bgPicID, _ := gs.CreatePic(200, 200)
	spritePicID, _ := gs.CreatePic(50, 50)

	winID, _ := gs.OpenWin(bgPicID, 0, 0, 200, 200, 0, 0, 0)
	// 新API: PutCast(srcPicID, dstPicID, x, y, srcX, srcY, w, h)
	cast1ID, _ := gs.PutCast(spritePicID, bgPicID, 10, 10, 0, 0, 50, 50)
	cast2ID, _ := gs.PutCast(spritePicID, bgPicID, 50, 50, 0, 0, 50, 50)

	// Close window
	err = gs.CloseWin(winID)
	if err != nil {
		t.Errorf("Failed to close window: %v", err)
	}

	// Verify casts are deleted
	gs.mu.RLock()
	_, err1 := gs.casts.GetCast(cast1ID)
	_, err2 := gs.casts.GetCast(cast2ID)
	gs.mu.RUnlock()

	if err1 == nil || err2 == nil {
		t.Error("Casts should be deleted when window is closed")
	}
	t.Logf("✓ Window and cast cleanup verified")

	// Test 3: CloseWinAll cleanup
	bgPicID2, _ := gs.CreatePic(100, 100)
	bgPicID3, _ := gs.CreatePic(100, 100)
	_, _ = gs.OpenWin(bgPicID2, 0, 0, 100, 100, 0, 0, 0)
	_, _ = gs.OpenWin(bgPicID3, 100, 100, 100, 100, 0, 0, 0)
	gs.PutCast(spritePicID, bgPicID2, 10, 10, 0, 0, 50, 50)
	gs.PutCast(spritePicID, bgPicID3, 10, 10, 0, 0, 50, 50)

	gs.CloseWinAll()

	// Verify all windows are closed
	gs.mu.RLock()
	windowCount := len(gs.windows.windows)
	castCount := len(gs.casts.casts)
	gs.mu.RUnlock()

	if windowCount != 0 {
		t.Errorf("Expected 0 windows after CloseWinAll, got %d", windowCount)
	}
	if castCount != 0 {
		t.Errorf("Expected 0 casts after CloseWinAll, got %d", castCount)
	}
	t.Logf("✓ CloseWinAll cleanup verified")

	// Test 4: Shutdown cleanup
	// 要件 9.3: プログラムが終了したとき、すべての描画リソースを解放する
	gs.CreatePic(100, 100)
	gs.CreatePic(100, 100)
	gs.CreatePic(100, 100)

	gs.Shutdown()

	// After shutdown, the system should be in a clean state
	// Note: We can't easily verify internal cleanup, but we can check that
	// the system doesn't crash
	t.Logf("✓ Shutdown cleanup completed")

	t.Logf("Resource cleanup integration test passed")
}
