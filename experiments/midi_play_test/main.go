// Simple MIDI playback test
package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/zurustar/son-et/pkg/vm"
	vmAudio "github.com/zurustar/son-et/pkg/vm/audio"
)

type Game struct {
	audioSystem  *vmAudio.AudioSystem
	eventQueue   *vm.EventQueue
	midiFile     string
	started      bool
	startTime    time.Time
	quarterCount int // 4分音符のカウント
	lastBeatTick int // 最後に4分音符をカウントしたFILLYティック
}

func (g *Game) Update() error {
	if !g.started {
		// Start MIDI playback
		err := g.audioSystem.PlayMIDI(g.midiFile)
		if err != nil {
			fmt.Printf("Error playing MIDI: %v\n", err)
			return err
		}
		fmt.Println("MIDI playback started!")
		fmt.Println("Counting quarter notes (4分音符)...")
		fmt.Println()
		g.started = true
		g.startTime = time.Now()
		g.quarterCount = 0
		g.lastBeatTick = -8 // FILLYでは1四分音符 = 8ティック
	}

	// Update audio system (generates MIDI_TIME events)
	g.audioSystem.Update()

	// Process events from the queue
	for {
		event, ok := g.eventQueue.Pop()
		if !ok {
			break
		}

		switch event.Type {
		case vm.EventMIDI_TIME:
			// FILLYティックを取得 (1四分音符 = 8ティック)
			tick, ok := event.Params["Tick"].(int)
			if ok {
				// 8ティックごと（4分音符ごと）にカウント
				if tick >= g.lastBeatTick+8 {
					g.quarterCount++
					g.lastBeatTick = (tick / 8) * 8 // 8の倍数に正規化

					// 拍子を表示（4拍子を想定）
					beat := ((g.quarterCount - 1) % 4) + 1
					measure := ((g.quarterCount - 1) / 4) + 1

					elapsed := time.Since(g.startTime)
					fmt.Printf("♩ Beat %d (Measure %d, Beat %d) - Tick: %d - Time: %.2fs\n",
						g.quarterCount, measure, beat, tick, elapsed.Seconds())
				}
			}

		case vm.EventMIDI_END:
			fmt.Println()
			fmt.Printf("MIDI playback finished! Total quarter notes: %d\n", g.quarterCount)
			return ebiten.Termination
		}
	}

	// Check if still playing
	if g.started && !g.audioSystem.IsMIDIPlaying() {
		fmt.Println()
		fmt.Printf("MIDI playback finished! Total quarter notes: %d\n", g.quarterCount)
		return ebiten.Termination
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	// Nothing to draw
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	// コマンドライン引数の解析
	soundFontPath := flag.String("sf", "GeneralUser-GS.sf2", "Path to SoundFont file (.sf2)")
	flag.Parse()

	// MIDIファイルは位置引数から取得
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: midi_test [-sf soundfont.sf2] <midi_file>")
		fmt.Println()
		fmt.Println("Options:")
		fmt.Println("  -sf string    Path to SoundFont file (default: GeneralUser-GS.sf2)")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  midi_test samples/y_saru/FLYINSKY.MID")
		fmt.Println("  midi_test -sf /path/to/soundfont.sf2 music.mid")
		os.Exit(1)
	}
	midiFile := args[0]

	fmt.Printf("MIDI Playback Test - %s\n", midiFile)
	fmt.Println("Displaying quarter note (4分音符) timing")
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	// Create event queue
	eventQueue := vm.NewEventQueue()

	// Create audio context
	audioCtx := audio.NewContext(vmAudio.SampleRate)

	// Create audio system with SoundFont
	audioSystem, err := vmAudio.NewAudioSystemWithContext(*soundFontPath, eventQueue, audioCtx)
	if err != nil {
		fmt.Printf("Failed to create audio system: %v\n", err)
		os.Exit(1)
	}

	game := &Game{
		audioSystem: audioSystem,
		eventQueue:  eventQueue,
		midiFile:    midiFile,
	}

	// Set up Ebiten
	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("MIDI Playback Test")

	// Run the game
	if err := ebiten.RunGame(game); err != nil {
		fmt.Printf("Error: %v\n", err)
	}

	// Cleanup
	audioSystem.Shutdown()
	fmt.Println("Done!")
}
