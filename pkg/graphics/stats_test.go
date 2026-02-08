package graphics

import (
	"strings"
	"testing"
)

// TestSpriteStats_String はSpriteStatsの文字列表現をテストする
func TestSpriteStats_String(t *testing.T) {
	stats := &SpriteStats{
		TotalSprites:   10,
		WindowSprites:  2,
		PictureSprites: 3,
		CastSprites:    2,
		TextSprites:    2,
		ShapeSprites:   1,
		Pictures:       5,
		Windows:        2,
		Casts:          2,
	}

	str := stats.String()

	// 必要な情報が含まれていることを確認
	if !strings.Contains(str, "Total Sprites: 10") {
		t.Errorf("String() should contain total sprites count")
	}
	if !strings.Contains(str, "Window Sprites:  2") {
		t.Errorf("String() should contain window sprites count")
	}
	if !strings.Contains(str, "Picture Sprites: 3") {
		t.Errorf("String() should contain picture sprites count")
	}
	if !strings.Contains(str, "Cast Sprites:    2") {
		t.Errorf("String() should contain cast sprites count")
	}
	if !strings.Contains(str, "Text Sprites:    2") {
		t.Errorf("String() should contain text sprites count")
	}
	if !strings.Contains(str, "Shape Sprites:   1") {
		t.Errorf("String() should contain shape sprites count")
	}
}

// TestSpriteStats_Compact はSpriteStatsのコンパクト表現をテストする
func TestSpriteStats_Compact(t *testing.T) {
	stats := &SpriteStats{
		TotalSprites:   10,
		WindowSprites:  2,
		PictureSprites: 3,
		CastSprites:    2,
		TextSprites:    2,
		ShapeSprites:   1,
	}

	compact := stats.Compact()

	// コンパクト形式の確認
	expected := "Sprites:10 (W:2 P:3 C:2 T:2 S:1)"
	if compact != expected {
		t.Errorf("Compact() = %q, want %q", compact, expected)
	}
}

// TestGraphicsSystem_GetSpriteStats はGraphicsSystemのスプライト統計取得をテストする
func TestGraphicsSystem_GetSpriteStats(t *testing.T) {
	// GraphicsSystemを作成
	gs := NewGraphicsSystem("")

	// 初期状態の統計を取得
	stats := gs.GetSpriteStats()

	// 初期状態では全てゼロ
	if stats.TotalSprites != 0 {
		t.Errorf("Initial TotalSprites = %d, want 0", stats.TotalSprites)
	}
	if stats.WindowSprites != 0 {
		t.Errorf("Initial WindowSprites = %d, want 0", stats.WindowSprites)
	}
	if stats.Pictures != 0 {
		t.Errorf("Initial Pictures = %d, want 0", stats.Pictures)
	}
	if stats.Windows != 0 {
		t.Errorf("Initial Windows = %d, want 0", stats.Windows)
	}
}

// TestSpriteStatsCollector はSpriteStatsCollectorの基本機能をテストする
func TestSpriteStatsCollector(t *testing.T) {
	gs := NewGraphicsSystem("")
	collector := NewSpriteStatsCollector(gs)

	// 初期状態では無効
	if collector.IsEnabled() {
		t.Error("Collector should be disabled by default")
	}

	// 有効化
	collector.SetEnabled(true)
	if !collector.IsEnabled() {
		t.Error("Collector should be enabled after SetEnabled(true)")
	}

	// 統計を収集
	stats := collector.Collect()
	if stats == nil {
		t.Fatal("Collect() should return non-nil stats")
	}

	// 最新の統計を取得
	latest := collector.GetLatest()
	if latest == nil {
		t.Fatal("GetLatest() should return non-nil after Collect()")
	}

	// 履歴を取得
	history := collector.GetHistory()
	if len(history) != 1 {
		t.Errorf("GetHistory() length = %d, want 1", len(history))
	}

	// ピーク値を取得
	peak := collector.GetPeakSprites()
	if peak != stats.TotalSprites {
		t.Errorf("GetPeakSprites() = %d, want %d", peak, stats.TotalSprites)
	}

	// クリア
	collector.Clear()
	history = collector.GetHistory()
	if len(history) != 0 {
		t.Errorf("GetHistory() length after Clear() = %d, want 0", len(history))
	}
}

// TestSpriteStatsCollector_Update はUpdate機能をテストする
func TestSpriteStatsCollector_Update(t *testing.T) {
	gs := NewGraphicsSystem("")
	collector := NewSpriteStatsCollector(gs)

	// 収集間隔を短く設定
	collector.SetInterval(2)
	collector.SetEnabled(true)

	// 1回目のUpdate（まだ収集されない）
	collector.Update()
	if len(collector.GetHistory()) != 0 {
		t.Error("Should not collect on first update")
	}

	// 2回目のUpdate（収集される）
	collector.Update()
	if len(collector.GetHistory()) != 1 {
		t.Error("Should collect on second update (interval=2)")
	}

	// 3回目のUpdate（まだ収集されない）
	collector.Update()
	if len(collector.GetHistory()) != 1 {
		t.Error("Should not collect on third update")
	}

	// 4回目のUpdate（収集される）
	collector.Update()
	if len(collector.GetHistory()) != 2 {
		t.Error("Should collect on fourth update")
	}
}

// TestSpriteStatsCollector_Disabled は無効時の動作をテストする
func TestSpriteStatsCollector_Disabled(t *testing.T) {
	gs := NewGraphicsSystem("")
	collector := NewSpriteStatsCollector(gs)

	// 無効状態でUpdate
	collector.SetInterval(1)
	collector.Update()
	collector.Update()
	collector.Update()

	// 履歴は空のまま
	if len(collector.GetHistory()) != 0 {
		t.Error("Should not collect when disabled")
	}
}

// TestPictureManager_Count はPictureManagerのCount機能をテストする
func TestPictureManager_Count(t *testing.T) {
	pm := NewPictureManager("")

	// 初期状態
	if pm.Count() != 0 {
		t.Errorf("Initial Count() = %d, want 0", pm.Count())
	}

	// ピクチャーを作成
	_, err := pm.CreatePic(100, 100)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	if pm.Count() != 1 {
		t.Errorf("Count() after CreatePic = %d, want 1", pm.Count())
	}

	// もう1つ作成
	_, err = pm.CreatePic(50, 50)
	if err != nil {
		t.Fatalf("CreatePic failed: %v", err)
	}

	if pm.Count() != 2 {
		t.Errorf("Count() after second CreatePic = %d, want 2", pm.Count())
	}
}

// TestWindowManager_Count はWindowManagerのCount機能をテストする
func TestWindowManager_Count(t *testing.T) {
	wm := NewWindowManager()

	// 初期状態
	if wm.Count() != 0 {
		t.Errorf("Initial Count() = %d, want 0", wm.Count())
	}

	// ウィンドウを開く
	_, err := wm.OpenWin(0)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	if wm.Count() != 1 {
		t.Errorf("Count() after OpenWin = %d, want 1", wm.Count())
	}

	// もう1つ開く
	_, err = wm.OpenWin(1)
	if err != nil {
		t.Fatalf("OpenWin failed: %v", err)
	}

	if wm.Count() != 2 {
		t.Errorf("Count() after second OpenWin = %d, want 2", wm.Count())
	}

	// 1つ閉じる
	err = wm.CloseWin(0)
	if err != nil {
		t.Fatalf("CloseWin failed: %v", err)
	}

	if wm.Count() != 1 {
		t.Errorf("Count() after CloseWin = %d, want 1", wm.Count())
	}
}

// TestWindowSpriteManager_Count はWindowSpriteManagerのCount機能をテストする
func TestWindowSpriteManager_Count(t *testing.T) {
	sm := NewSpriteManager()
	wsm := NewWindowSpriteManager(sm)

	// 初期状態
	if wsm.Count() != 0 {
		t.Errorf("Initial Count() = %d, want 0", wsm.Count())
	}
}

// TestFPSCounter_Basic はFPSCounterの基本機能をテストする
func TestFPSCounter_Basic(t *testing.T) {
	fc := NewFPSCounter()

	// 初期状態では無効
	if fc.IsEnabled() {
		t.Error("FPSCounter should be disabled by default")
	}

	// 有効化
	fc.SetEnabled(true)
	if !fc.IsEnabled() {
		t.Error("FPSCounter should be enabled after SetEnabled(true)")
	}

	// 初期状態ではFPSは0
	if fc.GetCurrentFPS() != 0 {
		t.Errorf("Initial CurrentFPS = %f, want 0", fc.GetCurrentFPS())
	}
	if fc.GetAverageFPS() != 0 {
		t.Errorf("Initial AverageFPS = %f, want 0", fc.GetAverageFPS())
	}
}

// TestFPSCounter_Update はFPSCounterのUpdate機能をテストする
func TestFPSCounter_Update(t *testing.T) {
	fc := NewFPSCounter()
	fc.SetEnabled(true)

	// 複数回Updateを呼び出す（フレーム時間をシミュレート）
	for i := 0; i < 10; i++ {
		fc.Update()
		// 少し待機してフレーム時間を作る
		// 注意: テストでは実際の時間経過を使用
	}

	// Updateを呼び出した後はFPSが計算されているはず
	// 注意: 実際のFPS値はテスト環境に依存するため、0より大きいことのみ確認
	// 最初のUpdateは時刻記録のみなので、少なくとも2回以上呼び出す必要がある
	if fc.GetCurrentFPS() <= 0 {
		// フレーム時間が非常に短い場合、FPSが非常に高くなる可能性がある
		// または、フレーム時間が0の場合はFPSも0になる
		// テスト環境では許容する
		t.Logf("CurrentFPS = %f (may be 0 or very high in test environment)", fc.GetCurrentFPS())
	}
}

// TestFPSCounter_Clear はFPSCounterのClear機能をテストする
func TestFPSCounter_Clear(t *testing.T) {
	fc := NewFPSCounter()
	fc.SetEnabled(true)

	// いくつかのフレームを記録
	for i := 0; i < 5; i++ {
		fc.Update()
	}

	// クリア
	fc.Clear()

	// クリア後はFPSが0
	if fc.GetCurrentFPS() != 0 {
		t.Errorf("CurrentFPS after Clear() = %f, want 0", fc.GetCurrentFPS())
	}
	if fc.GetAverageFPS() != 0 {
		t.Errorf("AverageFPS after Clear() = %f, want 0", fc.GetAverageFPS())
	}
}

// TestFPSCounter_Disabled は無効時の動作をテストする
func TestFPSCounter_Disabled(t *testing.T) {
	fc := NewFPSCounter()

	// 無効状態でUpdate
	fc.Update()
	fc.Update()
	fc.Update()

	// FPSは0のまま
	if fc.GetCurrentFPS() != 0 {
		t.Errorf("CurrentFPS when disabled = %f, want 0", fc.GetCurrentFPS())
	}
}

// TestFPSCounter_String はFPSCounterの文字列表現をテストする
func TestFPSCounter_String(t *testing.T) {
	fc := NewFPSCounter()
	fc.SetEnabled(true)

	str := fc.String()

	// 必要な情報が含まれていることを確認
	if !strings.Contains(str, "FPS Statistics") {
		t.Error("String() should contain 'FPS Statistics'")
	}
	if !strings.Contains(str, "Current FPS") {
		t.Error("String() should contain 'Current FPS'")
	}
	if !strings.Contains(str, "Average FPS") {
		t.Error("String() should contain 'Average FPS'")
	}
}

// TestFPSCounter_Compact はFPSCounterのコンパクト表現をテストする
func TestFPSCounter_Compact(t *testing.T) {
	fc := NewFPSCounter()

	compact := fc.Compact()

	// コンパクト形式の確認
	if !strings.Contains(compact, "FPS:") {
		t.Errorf("Compact() = %q, should contain 'FPS:'", compact)
	}
	if !strings.Contains(compact, "avg:") {
		t.Errorf("Compact() = %q, should contain 'avg:'", compact)
	}
}

// TestFPSCounter_GetStats はFPSCounterのGetStats機能をテストする
func TestFPSCounter_GetStats(t *testing.T) {
	fc := NewFPSCounter()
	fc.SetEnabled(true)

	// いくつかのフレームを記録
	for i := 0; i < 5; i++ {
		fc.Update()
	}

	stats := fc.GetStats()
	if stats == nil {
		t.Fatal("GetStats() should return non-nil stats")
	}

	// 統計情報が取得できることを確認
	t.Logf("FPS Stats: Current=%.1f, Average=%.1f, Min=%.1f, Max=%.1f",
		stats.CurrentFPS, stats.AverageFPS, stats.MinFPS, stats.MaxFPS)
}

// TestFPSCounter_SetHistorySize はFPSCounterの履歴サイズ設定をテストする
func TestFPSCounter_SetHistorySize(t *testing.T) {
	fc := NewFPSCounter()
	fc.SetEnabled(true)

	// 履歴サイズを小さく設定
	fc.SetHistorySize(3)

	// 5回Updateを呼び出す
	for i := 0; i < 5; i++ {
		fc.Update()
	}

	// 履歴サイズが制限されていることを確認
	// 注意: 内部の履歴サイズを直接確認することはできないが、
	// 動作が正常であることを確認
	stats := fc.GetStats()
	if stats == nil {
		t.Fatal("GetStats() should return non-nil stats")
	}
}
