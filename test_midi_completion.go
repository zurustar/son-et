package main

import (
	"fmt"
	"os"
	"time"

	"github.com/zurustar/son-et/pkg/engine"
)

func main() {
	// Create headless engine
	eng := engine.NewEngine(nil, engine.NewFilesystemAssetLoader("samples/kuma2"), nil)
	eng.SetHeadless(true)
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

	fmt.Println("MIDI playback started, monitoring for 30 seconds...")

	// Monitor for 30 seconds
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	startTime := time.Now()
	for i := 0; i < 30; i++ {
		<-ticker.C
		isPlaying := eng.IsMIDIPlaying()
		elapsed := time.Since(startTime).Seconds()
		fmt.Printf("[%.1fs] MIDI IsPlaying: %v\n", elapsed, isPlaying)

		if !isPlaying && i > 5 {
			fmt.Println("MIDI stopped playing!")
			break
		}
	}

	eng.Shutdown()
}
