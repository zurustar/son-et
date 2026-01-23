package main

import (
	"fmt"
	"os"

	"github.com/zurustar/son-et/pkg/engine"
)

func main() {
	data, err := os.ReadFile("samples/kuma2/KUMA.MID")
	if err != nil {
		fmt.Printf("Failed to read MIDI: %v\n", err)
		os.Exit(1)
	}

	// Create engine to use its parseMIDITempo function
	eng := engine.NewEngine(nil, engine.NewFilesystemAssetLoader("samples/kuma2"), nil)

	// We need to access the internal functions, so let's replicate the logic
	// Parse header
	ppq := int(data[12])<<8 | int(data[13])
	fmt.Printf("PPQ: %d\n", ppq)

	// Find tempo events manually
	offset := 14
	if string(data[offset:offset+4]) != "MTrk" {
		fmt.Println("Track not found")
		os.Exit(1)
	}

	trackLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
	trackData := data[offset+8 : offset+8+trackLen]

	trackOffset := 0
	currentTick := 0
	lastStatus := byte(0)

	type TempoEvent struct {
		Tick          int
		MicrosPerBeat int
	}

	tempoEvents := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // Default 120 BPM

	for trackOffset < len(trackData) {
		deltaTime, consumed := readVarInt(trackData[trackOffset:])
		trackOffset += consumed
		currentTick += deltaTime

		if trackOffset >= len(trackData) {
			break
		}

		eventByte := trackData[trackOffset]

		if eventByte < 0x80 {
			eventByte = lastStatus
		} else {
			trackOffset++
			if eventByte >= 0x80 && eventByte < 0xF0 {
				lastStatus = eventByte
			}
		}

		if eventByte == 0xFF {
			lastStatus = 0
			if trackOffset >= len(trackData) {
				break
			}
			metaType := trackData[trackOffset]
			trackOffset++
			length, consumed := readVarInt(trackData[trackOffset:])
			trackOffset += consumed

			if metaType == 0x51 && length == 3 {
				tempoMicros := int(trackData[trackOffset])<<16 | int(trackData[trackOffset+1])<<8 | int(trackData[trackOffset+2])
				bpm := 60000000 / tempoMicros
				tempoEvents = append(tempoEvents, TempoEvent{
					Tick:          currentTick,
					MicrosPerBeat: tempoMicros,
				})
				fmt.Printf("Tempo event at tick %d: %d BPM (%d microseconds/beat)\n", currentTick, bpm, tempoMicros)
			}

			trackOffset += length
		} else if eventByte == 0xF0 || eventByte == 0xF7 {
			lastStatus = 0
			length, consumed := readVarInt(trackData[trackOffset:])
			trackOffset += consumed + length
		} else if eventByte >= 0x80 {
			if eventByte >= 0xC0 && eventByte < 0xE0 {
				trackOffset++
			} else {
				trackOffset += 2
			}
		}
	}

	totalTicks := currentTick
	fmt.Printf("\nTotal ticks: %d\n", totalTicks)
	fmt.Printf("Tempo events found: %d\n\n", len(tempoEvents))

	// Calculate duration using tempo map (like the engine does)
	durationSeconds := 0.0
	lastTick := 0
	lastTempo := 500000

	for i, tempoEvent := range tempoEvents {
		if i > 0 {
			ticksInSegment := tempoEvent.Tick - lastTick
			segmentDuration := float64(ticksInSegment) / float64(ppq) * float64(lastTempo) / 1000000.0
			fmt.Printf("Segment %d: ticks %d-%d (%d ticks), tempo %d BPM, duration %.2fs\n",
				i, lastTick, tempoEvent.Tick, ticksInSegment, 60000000/lastTempo, segmentDuration)
			durationSeconds += segmentDuration
		}
		lastTick = tempoEvent.Tick
		lastTempo = tempoEvent.MicrosPerBeat
	}

	if totalTicks > lastTick {
		ticksInSegment := totalTicks - lastTick
		segmentDuration := float64(ticksInSegment) / float64(ppq) * float64(lastTempo) / 1000000.0
		fmt.Printf("Final segment: ticks %d-%d (%d ticks), tempo %d BPM, duration %.2fs\n",
			lastTick, totalTicks, ticksInSegment, 60000000/lastTempo, segmentDuration)
		durationSeconds += segmentDuration
	}

	fmt.Printf("\nTotal duration (with tempo changes): %.2f seconds\n", durationSeconds)

	// Also calculate with single tempo (for comparison)
	singleTempoDuration := float64(totalTicks) / float64(ppq) * float64(lastTempo) / 1000000.0
	fmt.Printf("Total duration (single tempo): %.2f seconds\n", singleTempoDuration)

	_ = eng
}

func readVarInt(data []byte) (int, int) {
	value := 0
	bytesRead := 0

	for i := 0; i < len(data) && i < 4; i++ {
		b := data[i]
		bytesRead++
		value = (value << 7) | int(b&0x7F)
		if b&0x80 == 0 {
			break
		}
	}

	return value, bytesRead
}
