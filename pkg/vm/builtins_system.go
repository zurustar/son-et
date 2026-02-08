package vm

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// registerSystemBuiltins registers system-related built-in functions.
// This includes event handler management, execution control, INI file operations,
// legacy Windows functions, messaging, and other system utilities.
func (vm *VM) registerSystemBuiltins() {
	// del_me: Remove the currently executing handler
	// Requirement 10.3: When del_me is called, system removes current event handler.
	vm.RegisterBuiltinFunction("del_me", func(v *VM, args []any) (any, error) {
		if v.currentHandler != nil {
			v.currentHandler.Remove()
			v.log.Debug("del_me called", "handler", v.currentHandler.ID)
		}
		return nil, nil
	})

	// del_us: Same as del_me
	// Requirement 10.5: When del_us is called, system removes current event handler.
	vm.RegisterBuiltinFunction("del_us", func(v *VM, args []any) (any, error) {
		if v.currentHandler != nil {
			v.currentHandler.Remove()
			v.log.Debug("del_us called", "handler", v.currentHandler.ID)
		}
		return nil, nil
	})

	// del_all: Remove all registered handlers
	// Requirement 10.4: When del_all is called, system removes all event handlers.
	vm.RegisterBuiltinFunction("del_all", func(v *VM, args []any) (any, error) {
		v.handlerRegistry.UnregisterAll()
		v.log.Debug("del_all called")
		return nil, nil
	})

	// end_step: End the current step block
	// Requirement 6.7: When end_step is called, system terminates step block execution.
	// Requirement 10.6: When end_step is called, system terminates current step block.
	vm.RegisterBuiltinFunction("end_step", func(v *VM, args []any) (any, error) {
		if v.currentHandler != nil {
			// Reset the handler's step counter and wait counter
			v.currentHandler.StepCounter = 0
			v.currentHandler.WaitCounter = 0
			// Reset PC to end of handler to stop execution
			v.currentHandler.CurrentPC = len(v.currentHandler.OpCodes)
			v.log.Debug("end_step called", "handler", v.currentHandler.ID)
		} else {
			// Reset VM's step counter
			v.stepCounter = 0
			v.log.Debug("end_step called (no handler)")
		}
		return nil, nil
	})

	// Wait: Wait for specified number of events
	// Requirement 17.1: When Wait(n) is called, system pauses execution for n events.
	// Requirement 17.2: When Wait(n) is called in mes(TIME) handler, system waits for n TIME events.
	// Requirement 17.3: When Wait(n) is called in mes(MIDI_TIME) handler, system waits for n MIDI_TIME events.
	// Requirement 17.4: When Wait(0) is called, system continues execution immediately.
	// Requirement 17.5: When Wait(n) is called with n<0, system treats it as Wait(0).
	// Requirement 17.6: System maintains separate wait counter for each handler.
	vm.RegisterBuiltinFunction("Wait", func(v *VM, args []any) (any, error) {
		// Get wait count from arguments
		waitCount := 1 // Default to 1 if not specified
		if len(args) >= 1 {
			if wc, ok := toInt64(args[0]); ok {
				waitCount = int(wc)
			} else if f, fok := toFloat64(args[0]); fok {
				waitCount = int(f)
			}
		}

		// Requirement 17.4: When Wait(0) is called, system continues execution immediately.
		// Requirement 17.5: When Wait(n) is called with n<0, system treats it as Wait(0).
		if waitCount <= 0 {
			v.log.Debug("Wait: wait count <= 0, continuing immediately", "waitCount", waitCount)
			return nil, nil
		}

		// If a handler is currently executing, set its WaitCounter
		// Requirement 17.6: System maintains separate wait counter for each handler.
		if v.currentHandler != nil {
			v.currentHandler.WaitCounter = waitCount
			// ログは削除（頻繁すぎるため）
			// Return a wait marker to signal the handler should pause
			return &waitMarker{WaitCount: waitCount}, nil
		}

		// If no handler is executing (e.g., in main code), we can't really wait
		// Log a warning and continue
		v.log.Warn("Wait called outside of event handler, ignoring", "waitCount", waitCount)
		return nil, nil
	})

	// ExitTitle: Terminate the program
	// Requirement 10.7: When ExitTitle is called, system terminates program.
	// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
	// Requirement 15.2: When ExitTitle is called, system closes all windows.
	// Requirement 15.3: When ExitTitle is called, system cleans up all resources.
	// Requirement 15.4: When ExitTitle is called, system terminates event loop.
	// Requirement 15.7: System provides graceful shutdown mechanism.
	vm.RegisterBuiltinFunction("ExitTitle", func(v *VM, args []any) (any, error) {
		v.log.Info("ExitTitle called, initiating graceful shutdown")

		// Stop all audio playback
		// Requirement 15.1: When ExitTitle is called, system stops all audio playback.
		if v.audioSystem != nil {
			v.audioSystem.StopTimer()
			v.audioSystem.Shutdown()
		}

		// Remove all event handlers
		v.handlerRegistry.UnregisterAll()

		// Stop the VM
		// Requirement 15.4: When ExitTitle is called, system terminates event loop.
		v.Stop()

		return nil, nil
	})

	// GetMesNo(seqID) - returns the sequence ID if the mes() block exists, 0 otherwise
	// This is used to check if a specific mes() block (event handler) is registered.
	// In FILLY, mes() blocks are assigned numeric IDs in registration order.
	vm.RegisterBuiltinFunction("GetMesNo", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			v.log.Warn("GetMesNo requires 1 argument (seqID)")
			return int64(0), nil
		}

		seqID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("GetMesNo seqID must be integer", "got", fmt.Sprintf("%T", args[0]))
			return int64(0), nil
		}

		// Check if handler with this ID exists
		_, exists := v.handlerRegistry.GetHandlerByNumber(int(seqID))
		if exists {
			v.log.Debug("GetMesNo called", "seqID", seqID, "exists", true)
			return seqID, nil
		}

		v.log.Debug("GetMesNo called", "seqID", seqID, "exists", false)
		return int64(0), nil
	})

	// DelMes(seqID) - deletes a mes() block by its sequence ID
	// This removes the event handler with the specified ID.
	vm.RegisterBuiltinFunction("DelMes", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			v.log.Warn("DelMes requires 1 argument (seqID)")
			return nil, nil
		}

		seqID, ok := toInt64(args[0])
		if !ok {
			v.log.Error("DelMes seqID must be integer", "got", fmt.Sprintf("%T", args[0]))
			return nil, nil
		}

		// Get handler by number and unregister it
		handler, exists := v.handlerRegistry.GetHandlerByNumber(int(seqID))
		if exists {
			v.handlerRegistry.Unregister(handler.ID)
			v.log.Debug("DelMes called", "seqID", seqID, "deleted", true)
		} else {
			v.log.Debug("DelMes called", "seqID", seqID, "deleted", false, "reason", "not found")
		}

		return nil, nil
	})

	// GetIniStr(section, key, default, filename) - reads a string value from an INI file
	// INI file format:
	//   [Section]
	//   key=value
	// Returns the value if found, or the default value if not found.
	vm.RegisterBuiltinFunction("GetIniStr", func(v *VM, args []any) (any, error) {
		if len(args) < 4 {
			v.log.Warn("GetIniStr requires 4 arguments (section, key, default, filename)")
			if len(args) >= 3 {
				if defaultVal, ok := args[2].(string); ok {
					return defaultVal, nil
				}
			}
			return "", nil
		}

		section, _ := args[0].(string)
		key, _ := args[1].(string)
		defaultVal, _ := args[2].(string)
		filename, _ := args[3].(string)

		// Resolve file path
		fullPath := v.resolveFilePath(filename)

		// Read INI file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			v.log.Debug("GetIniStr: file not found, returning default", "filename", filename, "default", defaultVal)
			return defaultVal, nil
		}

		// Parse INI file
		lines := strings.Split(string(content), "\n")
		currentSection := ""
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Handle Windows line endings
			line = strings.TrimSuffix(line, "\r")
			if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
				continue
			}

			// Check for section header [Section]
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				currentSection = strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
				continue
			}

			// Check for key=value in matching section
			if strings.EqualFold(currentSection, section) {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), key) {
					value := strings.TrimSpace(parts[1])
					v.log.Debug("GetIniStr called", "section", section, "key", key, "value", value)
					return value, nil
				}
			}
		}

		v.log.Debug("GetIniStr: key not found, returning default", "section", section, "key", key, "default", defaultVal)
		return defaultVal, nil
	})

	// GetIniInt(section, key, default, filename) - reads an integer value from an INI file
	// INI file format:
	//   [Section]
	//   key=value
	// Returns the value as integer if found, or the default value if not found.
	vm.RegisterBuiltinFunction("GetIniInt", func(v *VM, args []any) (any, error) {
		if len(args) < 4 {
			v.log.Warn("GetIniInt requires 4 arguments (section, key, default, filename)")
			if len(args) >= 3 {
				if defaultVal, ok := toInt64(args[2]); ok {
					return defaultVal, nil
				}
			}
			return int64(0), nil
		}

		section := toString(args[0])
		key := toString(args[1])
		defaultVal, _ := toInt64(args[2])
		filename := toString(args[3])

		// Resolve file path
		fullPath := v.resolveFilePath(filename)

		// Read INI file
		content, err := os.ReadFile(fullPath)
		if err != nil {
			v.log.Debug("GetIniInt: file not found, returning default", "filename", filename, "default", defaultVal)
			return defaultVal, nil
		}

		// Parse INI file
		lines := strings.Split(string(content), "\n")
		currentSection := ""
		for _, line := range lines {
			line = strings.TrimSpace(line)
			// Handle Windows line endings
			line = strings.TrimSuffix(line, "\r")
			if line == "" || strings.HasPrefix(line, ";") || strings.HasPrefix(line, "#") {
				continue
			}

			// Check for section header [Section]
			if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
				currentSection = strings.TrimPrefix(strings.TrimSuffix(line, "]"), "[")
				continue
			}

			// Check for key=value in matching section
			if strings.EqualFold(currentSection, section) {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), key) {
					valueStr := strings.TrimSpace(parts[1])
					value, ok := toInt64(valueStr)
					if !ok {
						v.log.Debug("GetIniInt: value not integer, returning default", "value", valueStr, "default", defaultVal)
						return defaultVal, nil
					}
					v.log.Debug("GetIniInt called", "section", section, "key", key, "value", value)
					return value, nil
				}
			}
		}

		v.log.Debug("GetIniInt: key not found, returning default", "section", section, "key", key, "default", defaultVal)
		return defaultVal, nil
	})

	// WriteIniInt(section, key, value, filename) - writes an integer value to an INI file
	// Creates the file if it doesn't exist, creates the section if it doesn't exist.
	vm.RegisterBuiltinFunction("WriteIniInt", func(v *VM, args []any) (any, error) {
		if len(args) < 4 {
			v.log.Warn("WriteIniInt requires 4 arguments (section, key, value, filename)")
			return nil, nil
		}

		section := toString(args[0])
		key := toString(args[1])
		value, _ := toInt64(args[2])
		filename := toString(args[3])

		// Resolve file path
		fullPath := v.resolveFilePath(filename)

		// Read existing INI file or create empty content
		var lines []string
		content, err := os.ReadFile(fullPath)
		if err == nil {
			lines = strings.Split(string(content), "\n")
		}

		// Find and update or add the key
		sectionFound := false
		keyFound := false
		currentSection := ""
		var result []string

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			trimmedLine = strings.TrimSuffix(trimmedLine, "\r")

			// Check for section header
			if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
				// If we were in the target section and didn't find the key, add it before this section
				if sectionFound && !keyFound {
					result = append(result, fmt.Sprintf("%s=%d", key, value))
					keyFound = true
				}
				currentSection = strings.TrimPrefix(strings.TrimSuffix(trimmedLine, "]"), "[")
				if strings.EqualFold(currentSection, section) {
					sectionFound = true
				}
				result = append(result, lines[i])
				continue
			}

			// Check for key=value in matching section
			if sectionFound && !keyFound && strings.EqualFold(currentSection, section) {
				parts := strings.SplitN(trimmedLine, "=", 2)
				if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), key) {
					result = append(result, fmt.Sprintf("%s=%d", key, value))
					keyFound = true
					continue
				}
			}

			result = append(result, lines[i])
		}

		// If section was found but key wasn't, add the key at the end of the section
		if sectionFound && !keyFound {
			result = append(result, fmt.Sprintf("%s=%d", key, value))
			keyFound = true
		}

		// If section wasn't found, add it with the key
		if !sectionFound {
			result = append(result, fmt.Sprintf("[%s]", section))
			result = append(result, fmt.Sprintf("%s=%d", key, value))
		}

		// Write back to file
		err = os.WriteFile(fullPath, []byte(strings.Join(result, "\n")), 0644)
		if err != nil {
			v.log.Error("WriteIniInt: failed to write file", "filename", filename, "error", err)
			return nil, nil
		}

		v.log.Debug("WriteIniInt called", "section", section, "key", key, "value", value, "filename", filename)
		return nil, nil
	})

	// WriteIniStr(section, key, value, filename) - writes a string value to an INI file
	// Creates the file if it doesn't exist, creates the section if it doesn't exist.
	vm.RegisterBuiltinFunction("WriteIniStr", func(v *VM, args []any) (any, error) {
		if len(args) < 4 {
			v.log.Warn("WriteIniStr requires 4 arguments (section, key, value, filename)")
			return nil, nil
		}

		section := toString(args[0])
		key := toString(args[1])
		value := toString(args[2])
		filename := toString(args[3])

		// Resolve file path
		fullPath := v.resolveFilePath(filename)

		// Read existing INI file or create empty content
		var lines []string
		content, err := os.ReadFile(fullPath)
		if err == nil {
			lines = strings.Split(string(content), "\n")
		}

		// Find and update or add the key
		sectionFound := false
		keyFound := false
		currentSection := ""
		var result []string

		for i, line := range lines {
			trimmedLine := strings.TrimSpace(line)
			trimmedLine = strings.TrimSuffix(trimmedLine, "\r")

			// Check for section header
			if strings.HasPrefix(trimmedLine, "[") && strings.HasSuffix(trimmedLine, "]") {
				// If we were in the target section and didn't find the key, add it before this section
				if sectionFound && !keyFound {
					result = append(result, fmt.Sprintf("%s=%s", key, value))
					keyFound = true
				}
				currentSection = strings.TrimPrefix(strings.TrimSuffix(trimmedLine, "]"), "[")
				if strings.EqualFold(currentSection, section) {
					sectionFound = true
				}
				result = append(result, lines[i])
				continue
			}

			// Check for key=value in matching section
			if sectionFound && !keyFound && strings.EqualFold(currentSection, section) {
				parts := strings.SplitN(trimmedLine, "=", 2)
				if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), key) {
					result = append(result, fmt.Sprintf("%s=%s", key, value))
					keyFound = true
					continue
				}
			}

			result = append(result, lines[i])
		}

		// If section was found but key wasn't, add the key at the end of the section
		if sectionFound && !keyFound {
			result = append(result, fmt.Sprintf("%s=%s", key, value))
			keyFound = true
		}

		// If section wasn't found, add it with the key
		if !sectionFound {
			result = append(result, fmt.Sprintf("[%s]", section))
			result = append(result, fmt.Sprintf("%s=%s", key, value))
		}

		// Write back to file
		err = os.WriteFile(fullPath, []byte(strings.Join(result, "\n")), 0644)
		if err != nil {
			v.log.Error("WriteIniStr: failed to write file", "filename", filename, "error", err)
			return nil, nil
		}

		v.log.Debug("WriteIniStr called", "section", section, "key", key, "value", value, "filename", filename)
		return nil, nil
	})

	// Debug: Set debug level (placeholder - does nothing for now)
	vm.RegisterBuiltinFunction("Debug", func(v *VM, args []any) (any, error) {
		if len(args) >= 1 {
			level, _ := toInt64(args[0])
			v.log.Debug("Debug called", "level", level)
		}
		return nil, nil
	})

	// Shell(command) - executes a shell command
	// This is a legacy Windows-specific function that is not supported on cross-platform systems.
	vm.RegisterBuiltinFunction("Shell", func(v *VM, args []any) (any, error) {
		v.log.Info("Shell() is a legacy Windows-specific function and is not supported")
		return int64(0), nil
	})

	// MCI(command) - Windows Media Control Interface commands
	// This is a legacy Windows-specific function that is not supported on cross-platform systems.
	vm.RegisterBuiltinFunction("MCI", func(v *VM, args []any) (any, error) {
		v.log.Info("MCI() is a legacy Windows-specific function and is not supported")
		return int64(0), nil
	})

	// StrMCI(command) - String variant of MCI commands
	// This is a legacy Windows-specific function that is not supported on cross-platform systems.
	vm.RegisterBuiltinFunction("StrMCI", func(v *VM, args []any) (any, error) {
		v.log.Info("StrMCI() is a legacy Windows-specific function and is not supported")
		return "", nil
	})

	// PostMes: Send a custom message to mes() blocks
	// PostMes(messageType, p1, p2, p3, p4)
	// This triggers event handlers registered with mes(USER, userID) or other event types.
	// Parameters:
	//   - messageType: event type (0=TIME, 1=MIDI_TIME, 2=MIDI_END, 3=KEY, 4=CLICK, 5=RBDOWN, 6=RBDBLCLK, or USER ID)
	//   - p1, p2, p3, p4: message parameters (stored in MesP1-MesP4)
	vm.RegisterBuiltinFunction("PostMes", func(v *VM, args []any) (any, error) {
		if len(args) < 5 {
			return nil, fmt.Errorf("PostMes requires 5 arguments (messageType, p1, p2, p3, p4), got %d", len(args))
		}

		messageType, _ := toInt64(args[0])
		p1, _ := toInt64(args[1])
		p2, _ := toInt64(args[2])
		p3, _ := toInt64(args[3])
		p4, _ := toInt64(args[4])

		v.log.Debug("PostMes called", "messageType", messageType, "p1", p1, "p2", p2, "p3", p3, "p4", p4)

		// Create event with parameters
		var eventType EventType
		switch messageType {
		case 0:
			eventType = EventTIME
		case 1:
			eventType = EventMIDI_TIME
		case 2:
			eventType = EventMIDI_END
		case 3:
			eventType = EventKEY
		case 4:
			eventType = EventCLICK
		case 5:
			eventType = EventRBDOWN
		case 6:
			eventType = EventRBDBLCLK
		default:
			// Treat as USER event with custom ID
			eventType = EventUSER
		}

		event := &Event{
			Type:      eventType,
			Timestamp: time.Now(),
			Params: map[string]any{
				"MesP1":       int(p1),
				"MesP2":       int(p2),
				"MesP3":       int(p3),
				"MesP4":       int(p4),
				"MessageType": int(messageType),
			},
		}

		// Push event to queue
		v.eventQueue.Push(event)
		return nil, nil
	})

	// MsgBox: Display a message box
	// MsgBox(message, flags)
	// Flags (Windows MessageBox style):
	//   Button types:
	//     0x00 = OK button only
	//     0x01 = OK/Cancel buttons
	//     0x03 = Abort/Retry/Ignore buttons
	//     0x04 = Yes/No buttons
	//     0x05 = Retry/Cancel buttons
	//     0x06 = Cancel/Try Again/Continue buttons
	//   Icon types:
	//     0x10 = Error icon
	//     0x20 = Question icon
	//     0x30 = Warning icon
	//     0x40 = Information icon
	// Returns (Windows MessageBox return values):
	//   IDOK = 1, IDCANCEL = 2, IDABORT = 3, IDRETRY = 4, IDIGNORE = 5,
	//   IDYES = 6, IDNO = 7, IDCLOSE = 8, IDHELP = 9
	// In headless mode, logs the message and returns appropriate default:
	//   - OK button types return 1 (IDOK)
	//   - Yes/No button types return 6 (IDYES)
	vm.RegisterBuiltinFunction("MsgBox", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			v.log.Warn("MsgBox requires at least 1 argument (message)")
			return int64(1), nil
		}

		message := toString(args[0])
		flags := int64(0)
		if len(args) >= 2 {
			flags, _ = toInt64(args[1])
		}

		// Extract button type (lower 4 bits)
		buttonType := flags & 0x0F

		// Determine icon type from flags
		iconType := "info"
		if flags&0x10 != 0 {
			iconType = "error"
		} else if flags&0x20 != 0 {
			iconType = "question"
		} else if flags&0x30 != 0 {
			iconType = "warning"
		} else if flags&0x40 != 0 {
			iconType = "info"
		}

		// Determine button description and default return value
		var buttonDesc string
		var defaultReturn int64

		switch buttonType {
		case MB_OK:
			buttonDesc = "OK"
			defaultReturn = IDOK
		case MB_OKCANCEL:
			buttonDesc = "OK/Cancel"
			defaultReturn = IDOK
		case MB_ABORTRETRYIGNORE:
			buttonDesc = "Abort/Retry/Ignore"
			defaultReturn = IDABORT
		case MB_YESNO:
			buttonDesc = "Yes/No"
			defaultReturn = IDYES
		case MB_RETRYCANCEL:
			buttonDesc = "Retry/Cancel"
			defaultReturn = IDRETRY
		case MB_CANCELTRYCONTINUE:
			buttonDesc = "Cancel/Try Again/Continue"
			defaultReturn = IDCANCEL
		default:
			buttonDesc = "OK"
			defaultReturn = IDOK
		}

		v.log.Info("MsgBox", "message", message, "icon", iconType, "buttons", buttonDesc, "defaultReturn", defaultReturn)

		// In headless mode, return the default button value
		return defaultReturn, nil
	})

	// GetSysTime: Get current system time in seconds
	// GetSysTime() - returns current Unix timestamp in seconds since Unix epoch
	// This is commonly used for timing and performance measurement.
	vm.RegisterBuiltinFunction("GetSysTime", func(v *VM, args []any) (any, error) {
		// Return current time in seconds (not milliseconds)
		seconds := time.Now().Unix()
		v.log.Debug("GetSysTime called", "seconds", seconds)
		return seconds, nil
	})
}
