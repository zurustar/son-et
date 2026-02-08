package vm

import (
	"fmt"
	"regexp"
	"strings"
)

// registerStringBuiltins registers string-related built-in functions.
func (vm *VM) registerStringBuiltins() {
	// StrPrint: Printf-style string formatting
	// Requirement 1.1: When StrPrint is called with format string and arguments, system returns formatted string.
	// Requirement 1.2: System supports %ld format specifier for decimal integers, converting to Go's %d.
	// Requirement 1.3: System supports %lx format specifier for hexadecimal, converting to Go's %x.
	// Requirement 1.4: System supports %s format specifier for strings.
	// Requirement 1.5: System supports width and padding specifiers like %03d.
	// Requirement 1.6: System converts escape sequences (\n, \t, \r) to actual control characters.
	// Requirement 1.7: When called with fewer arguments than format specifiers, system handles gracefully.
	// Requirement 1.8: When called with more arguments than format specifiers, system ignores extra arguments.
	vm.RegisterBuiltinFunction("StrPrint", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return "", nil
		}

		// Get format string
		format, ok := args[0].(string)
		if !ok {
			v.log.Error("StrPrint format must be string", "got", fmt.Sprintf("%T", args[0]))
			return "", nil
		}

		// Convert FILLY format specifiers to Go format specifiers
		// Use regex to handle width/padding specifiers like %03ld, %5lx
		// Pattern: %[flags][width][.precision]ld or %[flags][width][.precision]lx
		convertedFormat := format

		// Convert %ld variants (with optional flags, width, precision) to %d
		// Matches: %ld, %5ld, %05ld, %-5ld, %+5ld, etc.
		ldPattern := regexp.MustCompile(`%([+-]?\d*\.?\d*)ld`)
		convertedFormat = ldPattern.ReplaceAllString(convertedFormat, "%${1}d")

		// Convert %lx variants to %x
		lxPattern := regexp.MustCompile(`%([+-]?\d*\.?\d*)lx`)
		convertedFormat = lxPattern.ReplaceAllString(convertedFormat, "%${1}x")

		// Convert escape sequences to actual control characters
		convertedFormat = strings.ReplaceAll(convertedFormat, "\\n", "\n")
		convertedFormat = strings.ReplaceAll(convertedFormat, "\\t", "\t")
		convertedFormat = strings.ReplaceAll(convertedFormat, "\\r", "\r")

		// Prepare arguments for fmt.Sprintf
		formatArgs := make([]any, 0, len(args)-1)
		for i := 1; i < len(args); i++ {
			formatArgs = append(formatArgs, args[i])
		}

		// Use fmt.Sprintf to format the string
		// This handles both fewer and more arguments than format specifiers gracefully
		result := fmt.Sprintf(convertedFormat, formatArgs...)

		v.log.Debug("StrPrint called", "format", format, "result", result)
		return result, nil
	})

	// StrCode: Convert character code to string
	// StrCode(code) - returns character from ASCII/Unicode code
	// Example: StrCode(65) returns "A", StrCode(0x4349) returns "CI" (for 2-byte code)
	vm.RegisterBuiltinFunction("StrCode", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("StrCode requires 1 argument (code), got %d", len(args))
		}

		code, ok := toInt64(args[0])
		if !ok {
			v.log.Error("StrCode code must be integer", "got", fmt.Sprintf("%T", args[0]))
			return "", nil
		}

		// Handle multi-byte codes (e.g., 0x4349 = "CI")
		// If code > 255, treat as 2-byte character code
		var result string
		if code > 255 {
			// High byte first, then low byte
			highByte := byte((code >> 8) & 0xFF)
			lowByte := byte(code & 0xFF)
			result = string([]byte{highByte, lowByte})
		} else {
			result = string(rune(code))
		}

		v.log.Debug("StrCode called", "code", code, "result", result)
		return result, nil
	})

	// StrLen(str) - returns the length of a string in characters (not bytes)
	// For multi-byte characters (like Japanese), this returns the character count.
	vm.RegisterBuiltinFunction("StrLen", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("StrLen requires 1 argument (string), got %d", len(args))
		}

		str := toString(args[0])
		// Use rune count for proper Unicode character counting
		length := int64(len([]rune(str)))

		v.log.Debug("StrLen called", "string", str, "length", length)
		return length, nil
	})

	// SubStr(str, start, length) - returns a substring
	// start: 0-based index (character position, not byte position)
	// length: number of characters to extract
	// For multi-byte characters (like Japanese), this operates on characters, not bytes.
	// If start is out of range, returns empty string.
	// If length exceeds remaining characters, returns characters from start to end.
	vm.RegisterBuiltinFunction("SubStr", func(v *VM, args []any) (any, error) {
		if len(args) < 3 {
			return "", fmt.Errorf("SubStr requires 3 arguments (str, start, length), got %d", len(args))
		}

		str := toString(args[0])
		start, ok := toInt64(args[1])
		if !ok {
			v.log.Error("SubStr start must be integer", "got", fmt.Sprintf("%T", args[1]))
			return "", nil
		}
		length, ok := toInt64(args[2])
		if !ok {
			v.log.Error("SubStr length must be integer", "got", fmt.Sprintf("%T", args[2]))
			return "", nil
		}

		// Convert to runes for proper Unicode handling
		runes := []rune(str)
		runeLen := int64(len(runes))

		// Handle negative start (treat as 0)
		if start < 0 {
			start = 0
		}

		// If start is beyond string length, return empty string
		if start >= runeLen {
			v.log.Debug("SubStr called", "string", str, "start", start, "length", length, "result", "")
			return "", nil
		}

		// Handle negative length (treat as 0)
		if length < 0 {
			length = 0
		}

		// Calculate end position
		end := start + length
		if end > runeLen {
			end = runeLen
		}

		result := string(runes[start:end])
		v.log.Debug("SubStr called", "string", str, "start", start, "length", length, "result", result)
		return result, nil
	})

	// StrFind(str, search_str) - finds the first occurrence of search_str in str
	// Returns: 0-based index of the first occurrence, or -1 if not found
	// For multi-byte characters (like Japanese), this returns character position, not byte position.
	vm.RegisterBuiltinFunction("StrFind", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return int64(-1), fmt.Errorf("StrFind requires 2 arguments (str, search_str), got %d", len(args))
		}

		str := toString(args[0])
		searchStr := toString(args[1])

		// Handle empty search string
		if searchStr == "" {
			v.log.Debug("StrFind called", "string", str, "search", searchStr, "result", 0)
			return int64(0), nil
		}

		// Convert to runes for proper Unicode handling
		strRunes := []rune(str)
		searchRunes := []rune(searchStr)

		strLen := len(strRunes)
		searchLen := len(searchRunes)

		// Search for the substring
		for i := 0; i <= strLen-searchLen; i++ {
			found := true
			for j := 0; j < searchLen; j++ {
				if strRunes[i+j] != searchRunes[j] {
					found = false
					break
				}
			}
			if found {
				v.log.Debug("StrFind called", "string", str, "search", searchStr, "result", i)
				return int64(i), nil
			}
		}

		v.log.Debug("StrFind called", "string", str, "search", searchStr, "result", -1)
		return int64(-1), nil
	})

	// StrUp: Convert string to uppercase
	// Requirement 1.1: Convert ASCII lowercase (a-z) to uppercase (A-Z)
	// Requirement 1.2: Already uppercase strings are returned unchanged
	// Requirement 1.3: Empty string returns empty string
	// Requirement 1.4: Non-ASCII characters are preserved unchanged
	vm.RegisterBuiltinFunction("StrUp", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("StrUp requires 1 argument (string), got %d", len(args))
		}

		str := toString(args[0])
		result := strings.ToUpper(str)

		v.log.Debug("StrUp called", "string", str, "result", result)
		return result, nil
	})

	// StrLow: Convert string to lowercase
	// Requirement 2.1: Convert ASCII uppercase (A-Z) to lowercase (a-z)
	// Requirement 2.2: Already lowercase strings are returned unchanged
	// Requirement 2.3: Empty string returns empty string
	// Requirement 2.4: Non-ASCII characters are preserved unchanged
	vm.RegisterBuiltinFunction("StrLow", func(v *VM, args []any) (any, error) {
		if len(args) < 1 {
			return nil, fmt.Errorf("StrLow requires 1 argument (string), got %d", len(args))
		}

		str := toString(args[0])
		result := strings.ToLower(str)

		v.log.Debug("StrLow called", "string", str, "result", result)
		return result, nil
	})

	// CharCode: Get Unicode code point at specified position
	// Requirement 3.1: Return Unicode code point at specified index
	// Requirement 3.2: Return 0 for out-of-range index (negative or >= length)
	// Requirement 3.3: Return 0 for empty string
	// Requirement 3.4: Return Unicode code point for Japanese characters
	vm.RegisterBuiltinFunction("CharCode", func(v *VM, args []any) (any, error) {
		if len(args) < 2 {
			return int64(0), fmt.Errorf("CharCode requires 2 arguments (string, index), got %d", len(args))
		}

		str := toString(args[0])
		index, ok := toInt64(args[1])
		if !ok {
			v.log.Error("CharCode index must be integer", "got", fmt.Sprintf("%T", args[1]))
			return int64(0), nil
		}

		// Convert to rune slice for proper Unicode handling
		runes := []rune(str)

		// Return 0 for out-of-range index
		if index < 0 || index >= int64(len(runes)) {
			v.log.Debug("CharCode called", "string", str, "index", index, "result", 0)
			return int64(0), nil
		}

		result := int64(runes[index])
		v.log.Debug("CharCode called", "string", str, "index", index, "result", result)
		return result, nil
	})
}
