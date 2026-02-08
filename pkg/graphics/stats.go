// Package graphics provides sprite-based rendering system.
package graphics

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// SpriteStats はスプライト数の統計情報を保持する
// タスク 7.1: スプライト数の測定機能
type SpriteStats struct {
	// 総スプライト数（SpriteManagerに登録されているすべてのスプライト）
	TotalSprites int `json:"total_sprites"`

	// 各スプライトタイプ別の数
	WindowSprites  int `json:"window_sprites"`
	PictureSprites int `json:"picture_sprites"`
	CastSprites    int `json:"cast_sprites"`
	TextSprites    int `json:"text_sprites"`
	ShapeSprites   int `json:"shape_sprites"`

	// その他の統計
	Pictures int `json:"pictures"` // ロードされているピクチャー数
	Windows  int `json:"windows"`  // 開いているウィンドウ数
	Casts    int `json:"casts"`    // アクティブなキャスト数（CastManager）
}

// String はSpriteStatsの文字列表現を返す
func (ss *SpriteStats) String() string {
	var sb strings.Builder
	sb.WriteString("=== Sprite Statistics ===\n")
	sb.WriteString(fmt.Sprintf("Total Sprites: %d\n", ss.TotalSprites))
	sb.WriteString(fmt.Sprintf("  Window Sprites:  %d\n", ss.WindowSprites))
	sb.WriteString(fmt.Sprintf("  Picture Sprites: %d\n", ss.PictureSprites))
	sb.WriteString(fmt.Sprintf("  Cast Sprites:    %d\n", ss.CastSprites))
	sb.WriteString(fmt.Sprintf("  Text Sprites:    %d\n", ss.TextSprites))
	sb.WriteString(fmt.Sprintf("  Shape Sprites:   %d\n", ss.ShapeSprites))
	sb.WriteString(fmt.Sprintf("Pictures: %d\n", ss.Pictures))
	sb.WriteString(fmt.Sprintf("Windows:  %d\n", ss.Windows))
	sb.WriteString(fmt.Sprintf("Casts:    %d\n", ss.Casts))
	sb.WriteString("=========================\n")
	return sb.String()
}

// Compact はSpriteStatsのコンパクトな文字列表現を返す
// デバッグオーバーレイ表示用
func (ss *SpriteStats) Compact() string {
	return fmt.Sprintf("Sprites:%d (W:%d P:%d C:%d T:%d S:%d)",
		ss.TotalSprites,
		ss.WindowSprites,
		ss.PictureSprites,
		ss.CastSprites,
		ss.TextSprites,
		ss.ShapeSprites)
}

// GetSpriteStats はスプライト数の統計情報を取得する
// タスク 7.1: スプライト数の測定機能
func (gs *GraphicsSystem) GetSpriteStats() *SpriteStats {
	gs.mu.RLock()
	defer gs.mu.RUnlock()

	stats := &SpriteStats{}

	// SpriteManagerからの総スプライト数
	if gs.spriteManager != nil {
		stats.TotalSprites = gs.spriteManager.Count()
	}

	// 各スプライトマネージャーからの個別カウント
	if gs.windowSpriteManager != nil {
		stats.WindowSprites = gs.windowSpriteManager.Count()
	}

	if gs.pictureSpriteManager != nil {
		stats.PictureSprites = gs.pictureSpriteManager.Count()
	}

	if gs.castSpriteManager != nil {
		stats.CastSprites = gs.castSpriteManager.Count()
	}

	if gs.textSpriteManager != nil {
		stats.TextSprites = gs.textSpriteManager.Count()
	}

	if gs.shapeSpriteManager != nil {
		stats.ShapeSprites = gs.shapeSpriteManager.Count()
	}

	// その他の統計
	if gs.pictures != nil {
		stats.Pictures = gs.pictures.Count()
	}

	if gs.windows != nil {
		stats.Windows = gs.windows.Count()
	}

	if gs.casts != nil {
		stats.Casts = gs.casts.Count()
	}

	return stats
}

// SpriteStatsCollector はスプライト統計を定期的に収集する
// パフォーマンス監視用
type SpriteStatsCollector struct {
	gs       *GraphicsSystem
	history  []*SpriteStats
	maxSize  int
	mu       sync.RWMutex
	enabled  bool
	interval int // 収集間隔（フレーム数）
	counter  int // 現在のフレームカウンタ
}

// デフォルトの履歴サイズ
const defaultStatsHistorySize = 60

// デフォルトの収集間隔（フレーム数）
const defaultStatsInterval = 60

// NewSpriteStatsCollector は新しいSpriteStatsCollectorを作成する
func NewSpriteStatsCollector(gs *GraphicsSystem) *SpriteStatsCollector {
	return &SpriteStatsCollector{
		gs:       gs,
		history:  make([]*SpriteStats, 0, defaultStatsHistorySize),
		maxSize:  defaultStatsHistorySize,
		enabled:  false,
		interval: defaultStatsInterval,
		counter:  0,
	}
}

// SetEnabled は統計収集の有効/無効を設定する
func (ssc *SpriteStatsCollector) SetEnabled(enabled bool) {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()
	ssc.enabled = enabled
}

// IsEnabled は統計収集が有効かどうかを返す
func (ssc *SpriteStatsCollector) IsEnabled() bool {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()
	return ssc.enabled
}

// SetInterval は収集間隔を設定する（フレーム数）
func (ssc *SpriteStatsCollector) SetInterval(frames int) {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()
	if frames > 0 {
		ssc.interval = frames
	}
}

// Update はフレームごとに呼び出され、必要に応じて統計を収集する
// GraphicsSystem.Update()から呼び出すことを想定
func (ssc *SpriteStatsCollector) Update() {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()

	if !ssc.enabled {
		return
	}

	ssc.counter++
	if ssc.counter >= ssc.interval {
		ssc.counter = 0
		ssc.collectLocked()
	}
}

// collectLocked は統計を収集する（ロック済み）
func (ssc *SpriteStatsCollector) collectLocked() {
	stats := ssc.gs.GetSpriteStats()

	// 履歴に追加
	ssc.history = append(ssc.history, stats)

	// 最大サイズを超えた場合、古いものを削除
	if len(ssc.history) > ssc.maxSize {
		ssc.history = ssc.history[1:]
	}
}

// Collect は即座に統計を収集する
func (ssc *SpriteStatsCollector) Collect() *SpriteStats {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()

	stats := ssc.gs.GetSpriteStats()
	ssc.history = append(ssc.history, stats)

	if len(ssc.history) > ssc.maxSize {
		ssc.history = ssc.history[1:]
	}

	return stats
}

// GetLatest は最新の統計を返す
func (ssc *SpriteStatsCollector) GetLatest() *SpriteStats {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	if len(ssc.history) == 0 {
		return nil
	}
	return ssc.history[len(ssc.history)-1]
}

// GetHistory は統計履歴を返す
func (ssc *SpriteStatsCollector) GetHistory() []*SpriteStats {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	result := make([]*SpriteStats, len(ssc.history))
	copy(result, ssc.history)
	return result
}

// GetPeakSprites は履歴中の最大スプライト数を返す
func (ssc *SpriteStatsCollector) GetPeakSprites() int {
	ssc.mu.RLock()
	defer ssc.mu.RUnlock()

	peak := 0
	for _, stats := range ssc.history {
		if stats.TotalSprites > peak {
			peak = stats.TotalSprites
		}
	}
	return peak
}

// Clear は履歴をクリアする
func (ssc *SpriteStatsCollector) Clear() {
	ssc.mu.Lock()
	defer ssc.mu.Unlock()
	ssc.history = make([]*SpriteStats, 0, ssc.maxSize)
	ssc.counter = 0
}

// FPSCounter はFPS（フレームレート）を測定する
// タスク 7.2: FPS測定機能
type FPSCounter struct {
	// フレーム時間の履歴（ナノ秒）
	frameTimes []time.Duration
	// 履歴の最大サイズ
	maxSize int
	// 最後のフレーム時刻
	lastFrameTime time.Time
	// 現在のFPS
	currentFPS float64
	// 平均FPS
	averageFPS float64
	// 最小FPS（履歴内）
	minFPS float64
	// 最大FPS（履歴内）
	maxFPS float64
	// 有効フラグ
	enabled bool
	// ミューテックス
	mu sync.RWMutex
}

// デフォルトのFPS履歴サイズ（60フレーム = 約1秒分）
const defaultFPSHistorySize = 60

// NewFPSCounter は新しいFPSCounterを作成する
func NewFPSCounter() *FPSCounter {
	return &FPSCounter{
		frameTimes:    make([]time.Duration, 0, defaultFPSHistorySize),
		maxSize:       defaultFPSHistorySize,
		lastFrameTime: time.Time{},
		enabled:       false,
	}
}

// SetEnabled はFPS測定の有効/無効を設定する
func (fc *FPSCounter) SetEnabled(enabled bool) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.enabled = enabled
	if enabled && fc.lastFrameTime.IsZero() {
		fc.lastFrameTime = time.Now()
	}
}

// IsEnabled はFPS測定が有効かどうかを返す
func (fc *FPSCounter) IsEnabled() bool {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.enabled
}

// Update はフレームごとに呼び出され、FPSを更新する
// GraphicsSystem.Update()またはDraw()から呼び出すことを想定
func (fc *FPSCounter) Update() {
	fc.mu.Lock()
	defer fc.mu.Unlock()

	if !fc.enabled {
		return
	}

	now := time.Now()

	// 初回呼び出しの場合は時刻を記録するのみ
	if fc.lastFrameTime.IsZero() {
		fc.lastFrameTime = now
		return
	}

	// フレーム時間を計算
	frameTime := now.Sub(fc.lastFrameTime)
	fc.lastFrameTime = now

	// 履歴に追加
	fc.frameTimes = append(fc.frameTimes, frameTime)

	// 最大サイズを超えた場合、古いものを削除
	if len(fc.frameTimes) > fc.maxSize {
		fc.frameTimes = fc.frameTimes[1:]
	}

	// FPSを計算
	fc.calculateFPSLocked()
}

// calculateFPSLocked はFPSを計算する（ロック済み）
func (fc *FPSCounter) calculateFPSLocked() {
	if len(fc.frameTimes) == 0 {
		fc.currentFPS = 0
		fc.averageFPS = 0
		fc.minFPS = 0
		fc.maxFPS = 0
		return
	}

	// 現在のFPS（最新のフレーム時間から）
	lastFrameTime := fc.frameTimes[len(fc.frameTimes)-1]
	if lastFrameTime > 0 {
		fc.currentFPS = float64(time.Second) / float64(lastFrameTime)
	} else {
		fc.currentFPS = 0
	}

	// 平均FPS
	var totalTime time.Duration
	for _, ft := range fc.frameTimes {
		totalTime += ft
	}
	if totalTime > 0 {
		fc.averageFPS = float64(len(fc.frameTimes)) * float64(time.Second) / float64(totalTime)
	} else {
		fc.averageFPS = 0
	}

	// 最小・最大FPS
	fc.minFPS = fc.currentFPS
	fc.maxFPS = fc.currentFPS
	for _, ft := range fc.frameTimes {
		if ft > 0 {
			fps := float64(time.Second) / float64(ft)
			if fps < fc.minFPS {
				fc.minFPS = fps
			}
			if fps > fc.maxFPS {
				fc.maxFPS = fps
			}
		}
	}
}

// GetCurrentFPS は現在のFPSを返す
func (fc *FPSCounter) GetCurrentFPS() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.currentFPS
}

// GetAverageFPS は平均FPSを返す
func (fc *FPSCounter) GetAverageFPS() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.averageFPS
}

// GetMinFPS は履歴内の最小FPSを返す
func (fc *FPSCounter) GetMinFPS() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.minFPS
}

// GetMaxFPS は履歴内の最大FPSを返す
func (fc *FPSCounter) GetMaxFPS() float64 {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fc.maxFPS
}

// GetFrameTime は最新のフレーム時間を返す
func (fc *FPSCounter) GetFrameTime() time.Duration {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	if len(fc.frameTimes) == 0 {
		return 0
	}
	return fc.frameTimes[len(fc.frameTimes)-1]
}

// GetAverageFrameTime は平均フレーム時間を返す
func (fc *FPSCounter) GetAverageFrameTime() time.Duration {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	if len(fc.frameTimes) == 0 {
		return 0
	}
	var total time.Duration
	for _, ft := range fc.frameTimes {
		total += ft
	}
	return total / time.Duration(len(fc.frameTimes))
}

// Clear は履歴をクリアする
func (fc *FPSCounter) Clear() {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	fc.frameTimes = make([]time.Duration, 0, fc.maxSize)
	fc.lastFrameTime = time.Time{}
	fc.currentFPS = 0
	fc.averageFPS = 0
	fc.minFPS = 0
	fc.maxFPS = 0
}

// SetHistorySize は履歴サイズを設定する
func (fc *FPSCounter) SetHistorySize(size int) {
	fc.mu.Lock()
	defer fc.mu.Unlock()
	if size > 0 {
		fc.maxSize = size
		// 現在の履歴が新しいサイズを超えている場合は切り詰める
		if len(fc.frameTimes) > size {
			fc.frameTimes = fc.frameTimes[len(fc.frameTimes)-size:]
		}
	}
}

// String はFPSCounterの文字列表現を返す
func (fc *FPSCounter) String() string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	var sb strings.Builder
	sb.WriteString("=== FPS Statistics ===\n")
	sb.WriteString(fmt.Sprintf("Current FPS:  %.1f\n", fc.currentFPS))
	sb.WriteString(fmt.Sprintf("Average FPS:  %.1f\n", fc.averageFPS))
	sb.WriteString(fmt.Sprintf("Min FPS:      %.1f\n", fc.minFPS))
	sb.WriteString(fmt.Sprintf("Max FPS:      %.1f\n", fc.maxFPS))
	sb.WriteString(fmt.Sprintf("Frame Time:   %.2fms\n", float64(fc.GetFrameTimeLocked())/float64(time.Millisecond)))
	sb.WriteString("======================\n")
	return sb.String()
}

// GetFrameTimeLocked は最新のフレーム時間を返す（ロック済み）
func (fc *FPSCounter) GetFrameTimeLocked() time.Duration {
	if len(fc.frameTimes) == 0 {
		return 0
	}
	return fc.frameTimes[len(fc.frameTimes)-1]
}

// Compact はFPSCounterのコンパクトな文字列表現を返す
// デバッグオーバーレイ表示用
func (fc *FPSCounter) Compact() string {
	fc.mu.RLock()
	defer fc.mu.RUnlock()
	return fmt.Sprintf("FPS:%.1f (avg:%.1f)", fc.currentFPS, fc.averageFPS)
}

// FPSStats はFPS統計情報を保持する構造体
type FPSStats struct {
	CurrentFPS       float64       `json:"current_fps"`
	AverageFPS       float64       `json:"average_fps"`
	MinFPS           float64       `json:"min_fps"`
	MaxFPS           float64       `json:"max_fps"`
	FrameTime        time.Duration `json:"frame_time"`
	AverageFrameTime time.Duration `json:"average_frame_time"`
}

// GetStats はFPS統計情報を返す
func (fc *FPSCounter) GetStats() *FPSStats {
	fc.mu.RLock()
	defer fc.mu.RUnlock()

	var frameTime time.Duration
	if len(fc.frameTimes) > 0 {
		frameTime = fc.frameTimes[len(fc.frameTimes)-1]
	}

	var avgFrameTime time.Duration
	if len(fc.frameTimes) > 0 {
		var total time.Duration
		for _, ft := range fc.frameTimes {
			total += ft
		}
		avgFrameTime = total / time.Duration(len(fc.frameTimes))
	}

	return &FPSStats{
		CurrentFPS:       fc.currentFPS,
		AverageFPS:       fc.averageFPS,
		MinFPS:           fc.minFPS,
		MaxFPS:           fc.maxFPS,
		FrameTime:        frameTime,
		AverageFrameTime: avgFrameTime,
	}
}
