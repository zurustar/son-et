package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zurustar/son-et/pkg/engine"
)

func main() {
	// Create NON-headless engine
	eng := engine.NewEngine(nil, engine.NewFilesystemAssetLoader("samples/kuma2"), nil)
	eng.SetHeadless(false) // NOT headless
	eng.SetDebugLevel(engine.DebugLevelInfo)

	// Load soundfont (absolute path)
	if err := eng.LoadSoundFont("/Users/oumi/Documents/GitHub/son-et/GeneralUser-GS.sf2"); err != nil {
		fmt.Printf("Failed to load soundfont: %v\n", err)
		os.Exit(1)
	}

	// Start engine
	eng.Start()

	// Play MIDI
	if err := eng.PlayMIDI("KUMA.MID"); err != nil {
		fmt.Printf("Failed to play MIDI: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("MIDI playback started, monitoring for 150 seconds...")
	fmt.Println("Expected duration: ~139 seconds")

	// Monitor for 150 seconds (longer than the MIDI duration)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	lastPlaying := true

	for i := 0; i < 30; i++ { // 30 * 5 = 150 seconds
		<-ticker.C
		isPlaying := eng.IsMIDIPlaying()
		elapsed := time.Since(startTime).Seconds()

		if isPlaying != lastPlaying {
			fmt.Printf("[%.1fs] *** MIDI state changed: %v -> %v ***\n", elapsed, lastPlaying, isPlaying)
			lastPlaying = isPlaying
		} else {
			fmt.Printf("[%.1fs] MIDI IsPlaying: %v\n", elapsed, isPlaying)
		}

		if !isPlaying && elapsed > 140 {
			fmt.Println("MIDI stopped playing after expected duration!")
			break
		}
	}

	eng.Shutdown()
}
