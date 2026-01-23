package main

import (
	"fmt"
	"os"
)

func main() {
	data, err := os.ReadFile("samples/kuma2/KUMA.MID")
	if err != nil {
		fmt.Printf("Failed to read MIDI: %v\n", err)
		os.Exit(1)
	}

	// Parse header
	if string(data[0:4]) != "MThd" {
		fmt.Println("Invalid MIDI header")
		os.Exit(1)
	}

	ppq := int(data[12])<<8 | int(data[13])
	fmt.Printf("PPQ: %d\n", ppq)

	// Find track chunk
	offset := 14
	if string(data[offset:offset+4]) != "MTrk" {
		fmt.Println("Track not found")
		os.Exit(1)
	}

	trackLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
	fmt.Printf("Track length: %d bytes\n", trackLen)

	trackData := data[offset+8 : offset+8+trackLen]

	// Parse track to find last event
	trackOffset := 0
	currentTick := 0
	eventCount := 0
	lastStatus := byte(0)
	tempoMicros := 500000 // Default 120 BPM

	for trackOffset < len(trackData) {
		// Read delta time
		deltaTime, consumed := readVarInt(trackData[trackOffset:])
		trackOffset += consumed
		currentTick += deltaTime

		if trackOffset >= len(trackData) {
			break
		}

		// Read event
		eventByte := trackData[trackOffset]

		// Handle running status
		if eventByte < 0x80 {
			eventByte = lastStatus
		} else {
			trackOffset++
			if eventByte >= 0x80 && eventByte < 0xF0 {
				lastStatus = eventByte
			}
		}

		eventCount++

		if eventByte == 0xFF {
			// Meta event
			lastStatus = 0
			if trackOffset >= len(trackData) {
				break
			}
			metaType := trackData[trackOffset]
			trackOffset++
			length, consumed := readVarInt(trackData[trackOffset:])
			trackOffset += consumed

			// Check for tempo event
			if metaType == 0x51 && length == 3 {
				tempoMicros = int(trackData[trackOffset])<<16 | int(trackData[trackOffset+1])<<8 | int(trackData[trackOffset+2])
				bpm := 60000000 / tempoMicros
				fmt.Printf("Tempo at tick %d: %d BPM (%d microseconds/beat)\n", currentTick, bpm, tempoMicros)
			}

			trackOffset += length
		} else if eventByte == 0xF0 || eventByte == 0xF7 {
			// SysEx
			lastStatus = 0
			length, consumed := readVarInt(trackData[trackOffset:])
			trackOffset += consumed + length
		} else if eventByte >= 0x80 {
			// Channel event
			if eventByte >= 0xC0 && eventByte < 0xE0 {
				trackOffset++
			} else {
				trackOffset += 2
			}
		}
	}

	fmt.Printf("\nTotal events: %d\n", eventCount)
	fmt.Printf("Total ticks: %d\n", currentTick)

	// Calculate duration
	durationSeconds := float64(currentTick) / float64(ppq) * float64(tempoMicros) / 1000000.0
	fmt.Printf("Duration: %.2f seconds (%.2f minutes)\n", durationSeconds, durationSeconds/60.0)
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
