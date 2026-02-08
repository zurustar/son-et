package vm

import "fmt"

// registerAudioBuiltins registers audio-related built-in functions.
func (vm *VM) registerAudioBuiltins() {
	// PlayMIDI: Play a MIDI file
	// Requirement 10.1: When PlayMIDI is called, system calls MIDI playback function.
	vm.RegisterBuiltinFunction("PlayMIDI", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("PlayMIDI requires filename argument")
		}
		filename, ok := args[0].(string)
		if !ok {
			v.log.Error("PlayMIDI filename must be string", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}
		if err := v.PlayMIDI(filename); err != nil {
			// Requirement 11.2: When file is not found, system logs error and continues execution.
			v.log.Error("PlayMIDI failed", "filename", filename, "error", err)
			return nil, nil
		}
		v.log.Debug("PlayMIDI called", "filename", filename)
		return nil, nil
	})

	// PlayWAVE: Play a WAV file
	// Requirement 10.2: When PlayWAVE is called, system calls WAV playback function.
	vm.RegisterBuiltinFunction("PlayWAVE", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("PlayWAVE requires filename argument")
		}
		filename, ok := args[0].(string)
		if !ok {
			v.log.Error("PlayWAVE filename must be string", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}
		if err := v.PlayWAVE(filename); err != nil {
			// Requirement 5.4: When WAV file is not found, system logs error and continues execution.
			// Requirement 5.5: When WAV file is corrupted, system logs error and continues execution.
			v.log.Error("PlayWAVE failed", "filename", filename, "error", err)
			return nil, nil
		}
		v.log.Debug("PlayWAVE called", "filename", filename)
		return nil, nil
	})
}
