// Ebitengine + go-meltysynth でMIDI再生とMIDI_TIMEイベント生成
// テンポ変更対応版
package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/sinshu/go-meltysynth/meltysynth"
)

const sampleRate = 44100

// TempoEvent represents a tempo change
type TempoEvent struct {
	Tick          int
	MicrosPerBeat int // microseconds per quarter note
}

// TickCalculator calculates ticks from sample count considering tempo changes
type TickCalculator struct {
	ppq      int
	tempoMap []TempoEvent
	// Pre-calculated: sample count at each tempo change point
	sampleAtTempo []int64
}

func NewTickCalculator(ppq int, tempoMap []TempoEvent) *TickCalculator {
	tc := &TickCalculator{
		ppq:      ppq,
		tempoMap: tempoMap,
	}
	tc.precalculate()
	return tc
}

func (tc *TickCalculator) precalculate() {
	// Calculate sample count at each tempo change point
	tc.sampleAtTempo = make([]int64, len(tc.tempoMap))
	tc.sampleAtTempo[0] = 0

	for i := 1; i < len(tc.tempoMap); i++ {
		prevTempo := tc.tempoMap[i-1]
		currTempo := tc.tempoMap[i]

		// Ticks in previous tempo segment
		ticksInSegment := currTempo.Tick - prevTempo.Tick

		// Samples per tick at previous tempo
		// 1 quarter note = ppq ticks = microsPerBeat microseconds
		// 1 tick = microsPerBeat / ppq microseconds
		// samples per tick = sampleRate * microsPerBeat / ppq / 1000000
		samplesPerTick := float64(sampleRate) * float64(prevTempo.MicrosPerBeat) / float64(tc.ppq) / 1000000.0

		tc.sampleAtTempo[i] = tc.sampleAtTempo[i-1] + int64(float64(ticksInSegment)*samplesPerTick)
	}
}

// TickFromSamples converts sample count to MIDI tick (PPQ units)
func (tc *TickCalculator) TickFromSamples(samples int64) int {
	// Find which tempo segment we're in
	segmentIdx := 0
	for i := len(tc.tempoMap) - 1; i >= 0; i-- {
		if samples >= tc.sampleAtTempo[i] {
			segmentIdx = i
			break
		}
	}

	tempo := tc.tempoMap[segmentIdx]
	samplesIntoSegment := samples - tc.sampleAtTempo[segmentIdx]

	// Convert samples to ticks
	samplesPerTick := float64(sampleRate) * float64(tempo.MicrosPerBeat) / float64(tc.ppq) / 1000000.0
	ticksIntoSegment := int(float64(samplesIntoSegment) / samplesPerTick)

	return tempo.Tick + ticksIntoSegment
}

// FillyTickFromSamples converts sample count to FILLY tick (32nd note units)
func (tc *TickCalculator) FillyTickFromSamples(samples int64) int {
	midiTick := tc.TickFromSamples(samples)
	// 1 quarter note = ppq MIDI ticks = 8 FILLY ticks (32nd notes)
	return midiTick * 8 / tc.ppq
}

// MIDIStream implements io.Reader for Ebitengine audio
type MIDIStream struct {
	sequencer   *meltysynth.MidiFileSequencer
	sampleCount int64
	mu          sync.Mutex
}

func (s *MIDIStream) Read(p []byte) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	samples := len(p) / 4
	left := make([]float32, samples)
	right := make([]float32, samples)

	s.sequencer.Render(left, right)
	s.sampleCount += int64(samples)

	for i := 0; i < samples; i++ {
		l := int16(clamp(left[i], -1, 1) * 32767)
		r := int16(clamp(right[i], -1, 1) * 32767)
		binary.LittleEndian.PutUint16(p[i*4:], uint16(l))
		binary.LittleEndian.PutUint16(p[i*4+2:], uint16(r))
	}

	return len(p), nil
}

func (s *MIDIStream) GetSampleCount() int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.sampleCount
}

func clamp(v, min, max float32) float32 {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// Game implements ebiten.Game
type Game struct {
	audioContext   *audio.Context
	player         *audio.Player
	stream         *MIDIStream
	tickCalculator *TickCalculator
	duration       time.Duration
	totalSamples   int64
	startTime      time.Time

	lastFillyTick int
	midiTimeCount int
	finished      bool
}

func (g *Game) Update() error {
	if g.finished {
		return nil
	}

	// Use audio player's actual position (not wall clock)
	// This ensures tick calculation stays in sync with audio playback
	position := g.player.Position()
	currentSamples := int64(position.Seconds() * float64(sampleRate))
	currentFillyTick := g.tickCalculator.FillyTickFromSamples(currentSamples)

	// Generate MIDI_TIME events
	for g.lastFillyTick < currentFillyTick {
		g.lastFillyTick++
		g.midiTimeCount++

		fmt.Printf("[MIDI_TIME #%3d] filly_tick=%3d position=%v\n",
			g.midiTimeCount, g.lastFillyTick, position)
	}

	// Check if finished
	if position >= g.duration {
		fmt.Printf("\n[MIDI_END] Total MIDI_TIME events: %d\n", g.midiTimeCount)
		g.finished = true
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	currentSamples := g.stream.GetSampleCount()
	elapsed := time.Duration(float64(currentSamples) / float64(sampleRate) * float64(time.Second))
	msg := fmt.Sprintf("MIDI Playback (Tempo-aware)\nElapsed: %.2fs / %.2fs\nMIDI_TIME events: %d",
		elapsed.Seconds(), g.duration.Seconds(), g.midiTimeCount)
	ebitenutil.DebugPrint(screen, msg)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return 320, 240
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run ebiten_midi_v2.go <midi_file>")
		os.Exit(1)
	}

	midiFile := os.Args[1]
	soundFontFile := "../../GeneralUser-GS.sf2"

	fmt.Printf("MIDI: %s\nSoundFont: %s\n", midiFile, soundFontFile)

	// Load SoundFont
	sf2Data, _ := os.ReadFile(soundFontFile)
	soundFont, _ := meltysynth.NewSoundFont(bytes.NewReader(sf2Data))

	// Load MIDI file
	midiData, _ := os.ReadFile(midiFile)
	midi, _ := meltysynth.NewMidiFile(bytes.NewReader(midiData))

	// Parse tempo map and PPQ
	tempoMap, ppq := parseMIDITempoMap(midiData)
	fmt.Printf("PPQ: %d, Tempo events: %d\n", ppq, len(tempoMap))
	for i, t := range tempoMap {
		bpm := 60000000.0 / float64(t.MicrosPerBeat)
		fmt.Printf("  [%d] tick=%d BPM=%.1f\n", i, t.Tick, bpm)
	}

	// Create tick calculator
	tickCalc := NewTickCalculator(ppq, tempoMap)

	// Create synthesizer and sequencer
	settings := meltysynth.NewSynthesizerSettings(sampleRate)
	synth, _ := meltysynth.NewSynthesizer(soundFont, settings)
	sequencer := meltysynth.NewMidiFileSequencer(synth)
	sequencer.Play(midi, false)

	duration := midi.GetLength()
	totalSamples := int64(duration.Seconds() * float64(sampleRate))
	fmt.Printf("Duration: %v, Total samples: %d\n", duration, totalSamples)

	stream := &MIDIStream{sequencer: sequencer}

	audioCtx := audio.NewContext(sampleRate)
	player, _ := audioCtx.NewPlayer(stream)
	player.Play()

	game := &Game{
		audioContext:   audioCtx,
		player:         player,
		stream:         stream,
		tickCalculator: tickCalc,
		duration:       duration,
		totalSamples:   totalSamples,
		startTime:      time.Now(),
		lastFillyTick:  0,
	}

	ebiten.SetWindowSize(320, 240)
	ebiten.SetWindowTitle("MIDI Test (Tempo-aware)")

	if err := ebiten.RunGame(game); err != nil && err != io.EOF {
		fmt.Printf("Error: %v\n", err)
	}
}

// parseMIDITempoMap extracts all tempo events and PPQ from MIDI data
func parseMIDITempoMap(data []byte) ([]TempoEvent, int) {
	ppq := 480
	var events []TempoEvent

	if len(data) < 14 || string(data[0:4]) != "MThd" {
		return []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}, ppq
	}

	// PPQ from header
	timeDivision := int(data[12])<<8 | int(data[13])
	if timeDivision&0x8000 == 0 {
		ppq = timeDivision
	}

	// Scan all tracks for tempo events
	offset := 14
	for offset < len(data) {
		if offset+8 > len(data) || string(data[offset:offset+4]) != "MTrk" {
			break
		}

		trackLen := int(data[offset+4])<<24 | int(data[offset+5])<<16 | int(data[offset+6])<<8 | int(data[offset+7])
		trackEnd := offset + 8 + trackLen
		pos := offset + 8
		currentTick := 0
		lastStatus := byte(0)

		for pos < trackEnd {
			// Read delta time
			delta, n := readVarLen(data[pos:])
			pos += n
			currentTick += delta

			if pos >= trackEnd {
				break
			}

			eventByte := data[pos]

			if eventByte < 0x80 {
				eventByte = lastStatus
			} else {
				pos++
				if eventByte >= 0x80 && eventByte < 0xF0 {
					lastStatus = eventByte
				}
			}

			if eventByte == 0xFF { // Meta event
				if pos >= trackEnd {
					break
				}
				metaType := data[pos]
				pos++
				length, n := readVarLen(data[pos:])
				pos += n

				if metaType == 0x51 && length == 3 && pos+3 <= trackEnd { // Tempo
					microsPerBeat := int(data[pos])<<16 | int(data[pos+1])<<8 | int(data[pos+2])
					events = append(events, TempoEvent{Tick: currentTick, MicrosPerBeat: microsPerBeat})
				}
				pos += length
			} else if eventByte == 0xF0 || eventByte == 0xF7 { // SysEx
				length, n := readVarLen(data[pos:])
				pos += n + length
			} else if eventByte >= 0x80 {
				if eventByte >= 0xC0 && eventByte < 0xE0 {
					pos++
				} else {
					pos += 2
				}
			}
		}
		offset = trackEnd
	}

	// Ensure we have at least one tempo event at tick 0
	if len(events) == 0 {
		events = []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}
	} else if events[0].Tick > 0 {
		events = append([]TempoEvent{{Tick: 0, MicrosPerBeat: 500000}}, events...)
	}

	return events, ppq
}

func readVarLen(data []byte) (int, int) {
	value := 0
	n := 0
	for i := 0; i < len(data) && i < 4; i++ {
		n++
		value = (value << 7) | int(data[i]&0x7F)
		if data[i]&0x80 == 0 {
			break
		}
	}
	return value, n
}
