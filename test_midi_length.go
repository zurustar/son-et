package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/zurustar/son-et/pkg/engine"
)

func main() {
	files := []string{
		"samples/kuma2/KUMA.MID",
		"samples/kuma2/KUMAEND.MID",
	}

	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			fmt.Printf("Error reading %s: %v\n", file, err)
			continue
		}

		// Parse tempo and PPQ
		tempoMap, ppq, err := engine.ParseMIDITempo(data)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", file, err)
			continue
		}

		// Calculate total ticks
		totalTicks := engine.CalculateMIDILength(data, ppq)

		// Calculate duration in seconds
		// Assuming 120 BPM (500000 microseconds per beat)
		microsPerBeat := 500000
		if len(tempoMap) > 0 {
			microsPerBeat = tempoMap[0].MicrosPerBeat
		}

		// Duration = (totalTicks / ppq) * (microsPerBeat / 1000000)
		durationSeconds := float64(totalTicks) / float64(ppq) * float64(microsPerBeat) / 1000000.0

		fmt.Printf("%s:\n", filepath.Base(file))
		fmt.Printf("  PPQ: %d\n", ppq)
		fmt.Printf("  Total ticks: %d\n", totalTicks)
		fmt.Printf("  Tempo: %d BPM\n", 60000000/microsPerBeat)
		fmt.Printf("  Duration: %.2f seconds\n\n", durationSeconds)
	}
}
