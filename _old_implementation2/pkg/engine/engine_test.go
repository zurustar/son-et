package engine

import (
	"testing"
	"time"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

func TestEngine_Creation(t *testing.T) {
	renderer := &MockRenderer{}
	assetLoader := &MockAssetLoader{Files: make(map[string][]byte)}
	imageDecoder := &MockImageDecoder{Width: 640, Height: 480}

	engine := NewEngine(renderer, assetLoader, imageDecoder)

	if engine == nil {
		t.Fatal("NewEngine returned nil")
	}
	if engine.state == nil {
		t.Error("Engine state not initialized")
	}
	if engine.logger == nil {
		t.Error("Engine logger not initialized")
	}
	if engine.headless {
		t.Error("Headless mode should be disabled by default")
	}
}

func TestEngine_HeadlessMode(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})

	if engine.IsHeadless() {
		t.Error("Headless mode should be disabled by default")
	}

	engine.SetHeadless(true)
	if !engine.IsHeadless() {
		t.Error("Headless mode should be enabled")
	}

	engine.SetHeadless(false)
	if engine.IsHeadless() {
		t.Error("Headless mode should be disabled")
	}
}

func TestEngine_Termination(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	if engine.IsTerminated() {
		t.Error("Engine should not be terminated initially")
	}

	engine.Terminate()
	if !engine.IsTerminated() {
		t.Error("Engine should be terminated after Terminate()")
	}

	// Check that CheckTermination returns true
	if !engine.CheckTermination() {
		t.Error("CheckTermination should return true after Terminate()")
	}
}

func TestEngine_Timeout(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.SetTimeout(50 * time.Millisecond)
	engine.Start()

	if engine.IsTerminated() {
		t.Error("Engine should not be terminated initially")
	}

	// Wait for timeout
	time.Sleep(60 * time.Millisecond)

	if !engine.CheckTermination() {
		t.Error("Engine should be terminated after timeout")
	}
	if !engine.IsTerminated() {
		t.Error("IsTerminated should return true after timeout")
	}
}

func TestEngine_NoTimeout(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.SetTimeout(0) // No timeout
	engine.Start()

	time.Sleep(10 * time.Millisecond)

	if engine.CheckTermination() {
		t.Error("Engine should not terminate without timeout")
	}
}

func TestEngine_Update(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a sequence that never completes (infinite wait)
	// We'll manually keep it active by not letting it reach the end
	opcodes := make([]interpreter.OpCode, 1000)
	for i := range opcodes {
		opcodes[i] = interpreter.OpCode{Cmd: interpreter.OpWait, Args: []any{int64(1)}}
	}
	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	initialTick := engine.state.GetTickCount()

	err := engine.Update()
	if err != nil {
		t.Errorf("Update failed: %v", err)
	}

	if engine.state.GetTickCount() != initialTick+1 {
		t.Errorf("Expected tick count %d, got %d", initialTick+1, engine.state.GetTickCount())
	}
}

func TestEngine_UpdateAfterTermination(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a sequence that never completes (infinite wait)
	opcodes := make([]interpreter.OpCode, 1000)
	for i := range opcodes {
		opcodes[i] = interpreter.OpCode{Cmd: interpreter.OpWait, Args: []any{int64(1)}}
	}
	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	engine.Terminate()

	initialTick := engine.state.GetTickCount()

	err := engine.Update()
	if err == nil {
		t.Error("Expected error after termination")
	}
	if err != ErrTerminated {
		t.Errorf("Expected ErrTerminated, got %v", err)
	}

	// Tick should not increment after termination
	if engine.state.GetTickCount() != initialTick {
		t.Error("Tick count should not increment after termination")
	}
}

func TestEngine_Render(t *testing.T) {
	renderer := &MockRenderer{}
	engine := NewEngine(renderer, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})

	// Test normal render (does nothing in current implementation)
	engine.Render()

	// Test headless render
	engine.SetHeadless(true)
	engine.Render()
}

func TestEngine_DebugLevel(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})

	engine.SetDebugLevel(DebugLevelDebug)
	if engine.logger.GetLevel() != DebugLevelDebug {
		t.Errorf("Expected debug level %d, got %d", DebugLevelDebug, engine.logger.GetLevel())
	}
}

func TestEngine_Shutdown(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Should not panic
	engine.Shutdown()
}

func TestEngine_MultipleUpdates(t *testing.T) {
	engine := NewEngine(&MockRenderer{}, &MockAssetLoader{Files: make(map[string][]byte)}, &MockImageDecoder{})
	engine.Start()

	// Create a sequence that never completes (infinite wait)
	opcodes := make([]interpreter.OpCode, 1000)
	for i := range opcodes {
		opcodes[i] = interpreter.OpCode{Cmd: interpreter.OpWait, Args: []any{int64(1)}}
	}
	seq := NewSequencer(opcodes, TIME, nil)
	engine.RegisterSequence(seq, 0)

	for i := 0; i < 100; i++ {
		err := engine.Update()
		if err != nil {
			t.Errorf("Update %d failed: %v", i, err)
		}
	}

	if engine.state.GetTickCount() != 100 {
		t.Errorf("Expected tick count 100, got %d", engine.state.GetTickCount())
	}
}
