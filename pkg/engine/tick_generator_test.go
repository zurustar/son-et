package engine

import (
	"math"
	"testing"
	"testing/quick"
)

// TestTickGenerator_InvalidSampleRate tests that invalid sample rates are rejected
// Validates Requirements: 8.2
func TestTickGenerator_InvalidSampleRate(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate int
		ppq        int
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "zero sample rate",
			sampleRate: 0,
			ppq:        480,
			wantErr:    true,
			errMsg:     "invalid sample rate: 0 (must be positive)",
		},
		{
			name:       "negative sample rate",
			sampleRate: -44100,
			ppq:        480,
			wantErr:    true,
			errMsg:     "invalid sample rate: -44100 (must be positive)",
		},
		{
			name:       "sample rate too low",
			sampleRate: 1000,
			ppq:        480,
			wantErr:    true,
			errMsg:     "invalid sample rate: 1000 (must be between 8000 and 192000)",
		},
		{
			name:       "sample rate too high",
			sampleRate: 200000,
			ppq:        480,
			wantErr:    true,
			errMsg:     "invalid sample rate: 200000 (must be between 8000 and 192000)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
			tg, err := NewTickGenerator(tt.sampleRate, tt.ppq, tempoMap)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTickGenerator() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewTickGenerator() error = %v, want %v", err.Error(), tt.errMsg)
				}
				if tg != nil {
					t.Errorf("NewTickGenerator() expected nil TickGenerator on error, got %v", tg)
				}
			} else {
				if err != nil {
					t.Errorf("NewTickGenerator() unexpected error = %v", err)
				}
				if tg == nil {
					t.Errorf("NewTickGenerator() expected non-nil TickGenerator")
				}
			}
		})
	}
}

// TestTickGenerator_InvalidPPQ tests that invalid PPQ values are rejected
// Validates Requirements: 8.2
func TestTickGenerator_InvalidPPQ(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate int
		ppq        int
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "zero PPQ",
			sampleRate: 44100,
			ppq:        0,
			wantErr:    true,
			errMsg:     "invalid PPQ: 0 (must be positive)",
		},
		{
			name:       "negative PPQ",
			sampleRate: 44100,
			ppq:        -480,
			wantErr:    true,
			errMsg:     "invalid PPQ: -480 (must be positive)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
			tg, err := NewTickGenerator(tt.sampleRate, tt.ppq, tempoMap)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewTickGenerator() expected error but got nil")
					return
				}
				if err.Error() != tt.errMsg {
					t.Errorf("NewTickGenerator() error = %v, want %v", err.Error(), tt.errMsg)
				}
				if tg != nil {
					t.Errorf("NewTickGenerator() expected nil TickGenerator on error, got %v", tg)
				}
			} else {
				if err != nil {
					t.Errorf("NewTickGenerator() unexpected error = %v", err)
				}
				if tg == nil {
					t.Errorf("NewTickGenerator() expected non-nil TickGenerator")
				}
			}
		})
	}
}

// TestTickGenerator_ValidInputs tests that valid inputs are accepted
// This is a positive test to ensure we don't reject valid configurations
func TestTickGenerator_ValidInputs(t *testing.T) {
	tests := []struct {
		name       string
		sampleRate int
		ppq        int
	}{
		{
			name:       "standard CD quality",
			sampleRate: 44100,
			ppq:        480,
		},
		{
			name:       "high quality audio",
			sampleRate: 48000,
			ppq:        960,
		},
		{
			name:       "minimum valid sample rate",
			sampleRate: 8000,
			ppq:        120,
		},
		{
			name:       "maximum valid sample rate",
			sampleRate: 192000,
			ppq:        480,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: 500000}} // 120 BPM
			tg, err := NewTickGenerator(tt.sampleRate, tt.ppq, tempoMap)

			if err != nil {
				t.Errorf("NewTickGenerator() unexpected error = %v", err)
			}
			if tg == nil {
				t.Errorf("NewTickGenerator() expected non-nil TickGenerator")
				return
			}

			// Verify initial state
			if tg.sampleRate != tt.sampleRate {
				t.Errorf("sampleRate = %v, want %v", tg.sampleRate, tt.sampleRate)
			}
			if tg.ppq != tt.ppq {
				t.Errorf("ppq = %v, want %v", tg.ppq, tt.ppq)
			}
			if tg.GetCurrentTick() != 0 {
				t.Errorf("GetCurrentTick() = %v, want 0", tg.GetCurrentTick())
			}
			if tg.GetFractionalTick() != 0.0 {
				t.Errorf("GetFractionalTick() = %v, want 0.0", tg.GetFractionalTick())
			}
		})
	}
}

// TestProperty2_FractionalPrecisionPreservation verifies that fractional tick values
// are preserved across multiple ProcessSamples calls without precision loss
// **Validates: Requirements 1.3, 5.2**
// Feature: midi-timing-accuracy, Property 2: Fractional Precision Preservation
func TestProperty2_FractionalPrecisionPreservation(t *testing.T) {
	// Property: For any sequence of audio buffer processing calls,
	// the internal fractional tick value should be preserved without truncation,
	// and the cumulative error should remain below 0.01 ticks

	t.Run("Fractional precision across multiple buffers", func(t *testing.T) {
		property := func(bufferSizes []uint16, tempoBPM uint8, ppqValue uint16) bool {
			// Constrain inputs to reasonable ranges
			if len(bufferSizes) == 0 || len(bufferSizes) > 50 {
				return true // Skip invalid test cases or too many buffers
			}

			// Constrain tempo to reasonable range (60-240 BPM)
			tempo := 60.0 + float64(tempoBPM%181)

			// Constrain PPQ to common values (120-960)
			ppq := 120 + int(ppqValue%841)

			// Use standard sample rate
			sampleRate := 44100

			// Create tempo map with single tempo
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				t.Logf("Failed to create TickGenerator: %v", err)
				return false
			}

			// Process samples in multiple buffers and track fractional tick
			previousFractional := 0.0
			for _, size := range bufferSizes {
				bufferSize := 64 + int(size%8129)
				tg.ProcessSamples(bufferSize)

				currentFractional := tg.GetFractionalTick()

				// Verify fractional tick is monotonically increasing
				if currentFractional < previousFractional {
					t.Logf("Fractional tick decreased: %.6f -> %.6f", previousFractional, currentFractional)
					return false
				}

				// Verify fractional part is preserved (not truncated to integer)
				integerPart := float64(int(currentFractional))
				fractionalPart := currentFractional - integerPart

				// Fractional part should be in [0, 1)
				if fractionalPart < 0.0 || fractionalPart >= 1.0 {
					t.Logf("Invalid fractional part: %.6f", fractionalPart)
					return false
				}

				previousFractional = currentFractional
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Fractional part preservation without truncation", func(t *testing.T) {
		property := func(numSamples uint16, tempoBPM uint8) bool {
			// Constrain inputs
			samples := 100 + int(numSamples%8000) // 100-8099 samples
			tempo := 60.0 + float64(tempoBPM%181) // 60-240 BPM
			ppq := 480
			sampleRate := 44100

			// Create tempo map
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples
			tg.ProcessSamples(samples)

			// Get fractional tick
			fractionalTick := tg.GetFractionalTick()
			integerTick := tg.GetCurrentTick()

			// Verify fractional tick >= integer tick (no truncation)
			if fractionalTick < float64(integerTick) {
				t.Logf("Fractional tick truncated: fractional=%.6f < integer=%d", fractionalTick, integerTick)
				return false
			}

			// Verify fractional part is preserved (not just integer)
			fractionalPart := fractionalTick - float64(integerTick)
			if fractionalPart < 0.0 || fractionalPart >= 1.0 {
				t.Logf("Invalid fractional part: %.6f (should be in [0, 1))", fractionalPart)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Buffer size independence", func(t *testing.T) {
		property := func(totalSamples uint16, numBuffers uint8, tempoBPM uint8) bool {
			// Constrain inputs
			total := 1000 + int(totalSamples%9000) // 1000-9999 samples
			buffers := 1 + int(numBuffers%20)      // 1-20 buffers
			tempo := 60.0 + float64(tempoBPM%181)  // 60-240 BPM
			ppq := 480
			sampleRate := 44100

			// Create tempo map
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Test 1: Process all samples at once
			tg1, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}
			tg1.ProcessSamples(total)
			fractional1 := tg1.GetFractionalTick()

			// Test 2: Process samples in multiple buffers
			tg2, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			samplesPerBuffer := total / buffers
			remainder := total % buffers

			for i := 0; i < buffers; i++ {
				size := samplesPerBuffer
				if i == buffers-1 {
					size += remainder // Add remainder to last buffer
				}
				tg2.ProcessSamples(size)
			}
			fractional2 := tg2.GetFractionalTick()

			// Verify both approaches yield the same fractional tick
			error := math.Abs(fractional1 - fractional2)
			if error >= 0.01 {
				t.Logf("Buffer size affects result: error=%.6f ticks", error)
				t.Logf("Single buffer: %.6f, Multiple buffers: %.6f", fractional1, fractional2)
				t.Logf("Total samples: %d, Buffers: %d, Tempo: %.2f BPM", total, buffers, tempo)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// TestProperty1_TickCalculationFormulaAccuracy verifies that tick calculation
// matches the formula: ticks = (samples * tempo * PPQ) / (sample_rate * 60)
// **Validates: Requirements 1.1, 1.4**
// Feature: midi-timing-accuracy, Property 1: Tick Calculation Formula Accuracy
func TestProperty1_TickCalculationFormulaAccuracy(t *testing.T) {
	// Property: For any sample count, tempo (in BPM), PPQ value, and sample rate,
	// when calculating ticks from samples, the result should equal
	// (samples * tempo * PPQ) / (sample_rate * 60) within floating-point precision tolerance

	t.Run("Formula accuracy for single tempo", func(t *testing.T) {
		property := func(numSamples uint32, tempoBPM uint8, ppqValue uint16, sampleRateOffset uint16) bool {
			// Constrain inputs to reasonable ranges
			samples := int64(1 + (numSamples % 1000000))    // 1 to 1,000,000 samples
			tempo := 30.0 + float64(tempoBPM%200)           // 30-229 BPM (avoid overflow)
			ppq := 120 + int(ppqValue%841)                  // 120-960 PPQ
			sampleRate := 8000 + int(sampleRateOffset%1000) // 8000-8999 Hz (avoid overflow)

			// Create tempo map with single tempo
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				t.Logf("Failed to create TickGenerator: %v", err)
				return false
			}

			// Process samples
			tg.ProcessSamples(int(samples))

			// Get actual tick value
			actualTick := tg.GetFractionalTick()

			// Calculate expected tick using formula:
			// ticks = (samples * tempo * PPQ) / (sample_rate * 60)
			expectedTick := (float64(samples) * tempo * float64(ppq)) / (float64(sampleRate) * 60.0)

			// Verify within floating-point precision tolerance
			// Use relative error for large values, absolute error for small values
			tolerance := 0.001 // 0.1% relative error or 0.001 absolute error
			absoluteError := math.Abs(actualTick - expectedTick)
			relativeError := absoluteError / math.Max(expectedTick, 1.0)

			if absoluteError > tolerance && relativeError > tolerance {
				t.Logf("Formula mismatch: actual=%.6f, expected=%.6f, error=%.6f (%.4f%%)",
					actualTick, expectedTick, absoluteError, relativeError*100)
				t.Logf("Params: samples=%d, tempo=%.2f BPM, ppq=%d, sampleRate=%d",
					samples, tempo, ppq, sampleRate)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Formula accuracy across multiple buffers", func(t *testing.T) {
		property := func(bufferSizes []uint16, tempoBPM uint8, ppqValue uint16) bool {
			// Constrain inputs
			if len(bufferSizes) == 0 || len(bufferSizes) > 100 {
				return true // Skip invalid test cases
			}

			tempo := 30.0 + float64(tempoBPM%200) // 30-229 BPM (avoid overflow)
			ppq := 120 + int(ppqValue%841)        // 120-960 PPQ
			sampleRate := 44100                   // Standard sample rate

			// Create tempo map
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples in multiple buffers
			totalSamples := int64(0)
			for _, size := range bufferSizes {
				bufferSize := 64 + int(size%8129) // 64-8192 samples per buffer
				tg.ProcessSamples(bufferSize)
				totalSamples += int64(bufferSize)
			}

			// Get actual tick value
			actualTick := tg.GetFractionalTick()

			// Calculate expected tick using formula
			expectedTick := (float64(totalSamples) * tempo * float64(ppq)) / (float64(sampleRate) * 60.0)

			// Verify within tolerance
			tolerance := 0.001
			absoluteError := math.Abs(actualTick - expectedTick)
			relativeError := absoluteError / math.Max(expectedTick, 1.0)

			if absoluteError > tolerance && relativeError > tolerance {
				t.Logf("Formula mismatch after multiple buffers: actual=%.6f, expected=%.6f, error=%.6f",
					actualTick, expectedTick, absoluteError)
				t.Logf("Total samples: %d, Buffers: %d, Tempo: %.2f BPM, PPQ: %d",
					totalSamples, len(bufferSizes), tempo, ppq)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Formula accuracy with various sample rates", func(t *testing.T) {
		property := func(numSamples uint16, tempoBPM uint8, sampleRateChoice uint8) bool {
			// Constrain inputs
			samples := int64(1000 + (numSamples % 50000)) // 1000-50999 samples
			tempo := 60.0 + float64(tempoBPM%181)         // 60-240 BPM
			ppq := 480                                    // Standard PPQ

			// Choose from common sample rates
			sampleRates := []int{8000, 22050, 44100, 48000, 96000, 192000}
			sampleRate := sampleRates[int(sampleRateChoice)%len(sampleRates)]

			// Create tempo map
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples
			tg.ProcessSamples(int(samples))

			// Get actual tick value
			actualTick := tg.GetFractionalTick()

			// Calculate expected tick using formula
			expectedTick := (float64(samples) * tempo * float64(ppq)) / (float64(sampleRate) * 60.0)

			// Verify within tolerance
			tolerance := 0.001
			absoluteError := math.Abs(actualTick - expectedTick)
			relativeError := absoluteError / math.Max(expectedTick, 1.0)

			if absoluteError > tolerance && relativeError > tolerance {
				t.Logf("Formula mismatch: actual=%.6f, expected=%.6f, error=%.6f",
					actualTick, expectedTick, absoluteError)
				t.Logf("Samples: %d, Tempo: %.2f BPM, Sample rate: %d Hz",
					samples, tempo, sampleRate)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Formula accuracy with various PPQ values", func(t *testing.T) {
		property := func(numSamples uint16, tempoBPM uint8, ppqChoice uint8) bool {
			// Constrain inputs
			samples := int64(1000 + (numSamples % 50000)) // 1000-50999 samples
			tempo := 60.0 + float64(tempoBPM%181)         // 60-240 BPM
			sampleRate := 44100                           // Standard sample rate

			// Choose from common PPQ values
			ppqValues := []int{120, 240, 480, 960}
			ppq := ppqValues[int(ppqChoice)%len(ppqValues)]

			// Create tempo map
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples
			tg.ProcessSamples(int(samples))

			// Get actual tick value
			actualTick := tg.GetFractionalTick()

			// Calculate expected tick using formula
			expectedTick := (float64(samples) * tempo * float64(ppq)) / (float64(sampleRate) * 60.0)

			// Verify within tolerance
			tolerance := 0.001
			absoluteError := math.Abs(actualTick - expectedTick)
			relativeError := absoluteError / math.Max(expectedTick, 1.0)

			if absoluteError > tolerance && relativeError > tolerance {
				t.Logf("Formula mismatch: actual=%.6f, expected=%.6f, error=%.6f",
					actualTick, expectedTick, absoluteError)
				t.Logf("Samples: %d, Tempo: %.2f BPM, PPQ: %d",
					samples, tempo, ppq)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Formula accuracy with extreme values", func(t *testing.T) {
		property := func(sampleScale uint8, tempoChoice uint8, ppqChoice uint8) bool {
			// Test with extreme but valid values
			samples := int64(1 + int64(sampleScale%100)*1000) // 1 to 99,001 samples (avoid overflow)

			// Extreme tempos
			tempos := []float64{30.0, 60.0, 120.0, 180.0, 240.0, 300.0}
			tempo := tempos[int(tempoChoice)%len(tempos)]

			// Extreme PPQ values
			ppqValues := []int{120, 240, 480, 960}
			ppq := ppqValues[int(ppqChoice)%len(ppqValues)]

			sampleRate := 44100

			// Create tempo map
			microsPerBeat := int(60000000.0 / tempo)
			tempoMap := []TempoEvent{{Tick: 0, MicrosPerBeat: microsPerBeat}}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples
			tg.ProcessSamples(int(samples))

			// Get actual tick value
			actualTick := tg.GetFractionalTick()

			// Calculate expected tick using formula
			expectedTick := (float64(samples) * tempo * float64(ppq)) / (float64(sampleRate) * 60.0)

			// Verify within tolerance
			tolerance := 0.001
			absoluteError := math.Abs(actualTick - expectedTick)
			relativeError := absoluteError / math.Max(expectedTick, 1.0)

			if absoluteError > tolerance && relativeError > tolerance {
				t.Logf("Formula mismatch with extreme values: actual=%.6f, expected=%.6f, error=%.6f",
					actualTick, expectedTick, absoluteError)
				t.Logf("Samples: %d, Tempo: %.2f BPM, PPQ: %d",
					samples, tempo, ppq)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}

// TestProperty3_TempoChangeCorrectness verifies that tempo changes are handled correctly
// and tick calculations remain accurate across tempo boundaries
// **Validates: Requirements 1.2, 4.1, 4.2, 4.3, 4.4**
// Feature: midi-timing-accuracy, Property 3: Tempo Change Correctness
func TestProperty3_TempoChangeCorrectness(t *testing.T) {
	// Property: For any tempo map and sample position, when samples cross a tempo boundary,
	// subsequent tick calculations should use the new tempo value, and the tick value at
	// the boundary should be continuous (no jumps or gaps)

	t.Run("Monotonic progression across tempo changes", func(t *testing.T) {
		property := func(tempo1BPM uint8, tempo2BPM uint8, changeTickOffset uint16) bool {
			// Constrain inputs to reasonable ranges
			tempo1 := 60.0 + float64(tempo1BPM%181)       // 60-240 BPM
			tempo2 := 60.0 + float64(tempo2BPM%181)       // 60-240 BPM
			changeTick := 100 + int(changeTickOffset%900) // 100-999 ticks

			ppq := 480
			sampleRate := 44100

			// Create tempo map with two tempos
			microsPerBeat1 := int(60000000.0 / tempo1)
			microsPerBeat2 := int(60000000.0 / tempo2)
			tempoMap := []TempoEvent{
				{Tick: 0, MicrosPerBeat: microsPerBeat1},
				{Tick: changeTick, MicrosPerBeat: microsPerBeat2},
			}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				t.Logf("Failed to create TickGenerator: %v", err)
				return false
			}

			// Process samples in small increments to cross the tempo boundary
			bufferSize := 1000 // Small buffer
			previousTick := 0.0
			crossedBoundary := false

			for i := 0; i < 200; i++ { // Process enough buffers to cross the change
				tg.ProcessSamples(bufferSize)
				currentTick := tg.GetFractionalTick()

				// Verify monotonic progression (no backwards movement)
				if currentTick < previousTick {
					t.Logf("Tick moved backwards: previous=%.6f, current=%.6f",
						previousTick, currentTick)
					return false
				}

				// Check if we crossed the tempo boundary
				if previousTick < float64(changeTick) && currentTick >= float64(changeTick) {
					crossedBoundary = true

					// Verify no large jump at boundary (continuity)
					tickJump := currentTick - previousTick
					maxExpectedJump := float64(bufferSize) * math.Max(tempo1, tempo2) * float64(ppq) / (float64(sampleRate) * 60.0) * 1.5

					if tickJump > maxExpectedJump {
						t.Logf("Large tick jump at boundary: jump=%.6f, maxExpected=%.6f",
							tickJump, maxExpectedJump)
						return false
					}
				}

				previousTick = currentTick

				// Stop if we're well past the change point
				if currentTick > float64(changeTick)+100 {
					break
				}
			}

			// Verify we actually crossed the boundary
			if !crossedBoundary {
				t.Logf("Did not cross tempo boundary: finalTick=%.6f, changeTick=%d",
					previousTick, changeTick)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Multiple tempo changes maintain continuity", func(t *testing.T) {
		property := func(numChanges uint8, tempos []uint8, bufferSize uint16) bool {
			// Constrain inputs
			changes := 2 + int(numChanges%8) // 2-9 tempo changes
			if len(tempos) < changes {
				return true // Skip if not enough tempo values
			}

			bufSize := 1000 + int(bufferSize%9000) // 1000-9999 samples per buffer
			ppq := 480
			sampleRate := 44100

			// Create tempo map with multiple tempo changes
			tempoMap := make([]TempoEvent, changes)
			currentTick := 0

			for i := 0; i < changes; i++ {
				tempo := 60.0 + float64(tempos[i]%181) // 60-240 BPM
				microsPerBeat := int(60000000.0 / tempo)
				tempoMap[i] = TempoEvent{
					Tick:          currentTick,
					MicrosPerBeat: microsPerBeat,
				}
				// Space tempo changes 500 ticks apart
				currentTick += 500
			}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples through all tempo changes
			previousTick := 0.0
			for i := 0; i < changes*2; i++ { // Process enough buffers to cross all changes
				tg.ProcessSamples(bufSize)
				currentTick := tg.GetFractionalTick()

				// Verify monotonic progression (no backwards movement)
				if currentTick < previousTick {
					t.Logf("Tick moved backwards: previous=%.6f, current=%.6f",
						previousTick, currentTick)
					return false
				}

				// Verify no large jumps (continuity)
				tickDelta := currentTick - previousTick
				maxExpectedDelta := float64(bufSize) * 300.0 * float64(ppq) / (float64(sampleRate) * 60.0)

				if tickDelta > maxExpectedDelta*1.5 { // Allow 50% margin
					t.Logf("Large tick jump detected: delta=%.6f, maxExpected=%.6f",
						tickDelta, maxExpectedDelta)
					return false
				}

				previousTick = currentTick
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Tempo change does not cause backwards movement", func(t *testing.T) {
		property := func(tempo1BPM uint8, tempo2BPM uint8) bool {
			// Constrain inputs
			tempo1 := 60.0 + float64(tempo1BPM%181) // 60-240 BPM
			tempo2 := 60.0 + float64(tempo2BPM%181) // 60-240 BPM
			changeTick := 500                       // Fixed change point

			ppq := 480
			sampleRate := 44100

			// Create tempo map
			microsPerBeat1 := int(60000000.0 / tempo1)
			microsPerBeat2 := int(60000000.0 / tempo2)
			tempoMap := []TempoEvent{
				{Tick: 0, MicrosPerBeat: microsPerBeat1},
				{Tick: changeTick, MicrosPerBeat: microsPerBeat2},
			}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples and verify tick never decreases
			previousTick := 0.0
			bufferSize := 2048

			for i := 0; i < 50; i++ {
				tg.ProcessSamples(bufferSize)
				currentTick := tg.GetFractionalTick()

				if currentTick < previousTick {
					t.Logf("Tick decreased: previous=%.6f, current=%.6f",
						previousTick, currentTick)
					return false
				}

				previousTick = currentTick
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})

	t.Run("Rapid tempo changes", func(t *testing.T) {
		property := func(baseTempoOffset uint8, numChanges uint8) bool {
			// Constrain inputs
			baseTempo := 100.0 + float64(baseTempoOffset%100) // 100-199 BPM
			changes := 3 + int(numChanges%7)                  // 3-9 rapid changes

			ppq := 480
			sampleRate := 44100

			// Create tempo map with rapid changes (every 50 ticks)
			tempoMap := make([]TempoEvent, changes)
			for i := 0; i < changes; i++ {
				tempo := baseTempo + float64(i*10) // Gradually increase tempo
				microsPerBeat := int(60000000.0 / tempo)
				tempoMap[i] = TempoEvent{
					Tick:          i * 50, // Changes every 50 ticks
					MicrosPerBeat: microsPerBeat,
				}
			}

			// Create tick generator
			tg, err := NewTickGenerator(sampleRate, ppq, tempoMap)
			if err != nil {
				return false
			}

			// Process samples through all rapid changes
			previousTick := 0.0
			bufferSize := 500 // Small buffer to catch rapid changes

			for i := 0; i < changes*5; i++ { // Process enough buffers
				tg.ProcessSamples(bufferSize)
				currentTick := tg.GetFractionalTick()

				// Verify monotonic progression
				if currentTick < previousTick {
					t.Logf("Tick moved backwards during rapid changes: previous=%.6f, current=%.6f",
						previousTick, currentTick)
					return false
				}

				// Verify tick advanced (not stuck)
				if i > 0 && currentTick == previousTick {
					t.Logf("Tick did not advance: %.6f", currentTick)
					return false
				}

				previousTick = currentTick
			}

			// Verify we processed through all tempo changes
			finalTick := tg.GetFractionalTick()
			lastChangeTick := float64(tempoMap[changes-1].Tick)

			if finalTick < lastChangeTick {
				t.Logf("Did not process through all tempo changes: finalTick=%.6f, lastChange=%d",
					finalTick, tempoMap[changes-1].Tick)
				return false
			}

			return true
		}

		config := &quick.Config{MaxCount: 100}
		if err := quick.Check(property, config); err != nil {
			t.Error(err)
		}
	})
}
