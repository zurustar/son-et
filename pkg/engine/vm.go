package engine

import (
	"fmt"
	"strings"

	"github.com/zurustar/son-et/pkg/compiler/interpreter"
)

// VM executes OpCode sequences within the engine.
type VM struct {
	state  *EngineState
	engine *Engine
	logger *Logger
}

// NewVM creates a new VM with the given state and logger.
func NewVM(state *EngineState, engine *Engine, logger *Logger) *VM {
	return &VM{
		state:  state,
		engine: engine,
		logger: logger,
	}
}

// ExecuteOp executes a single OpCode operation.
// This is the central dispatch point for all OpCode execution.
// Returns an error if the operation fails.
func (vm *VM) ExecuteOp(seq *Sequencer, op interpreter.OpCode) error {
	if vm.logger.GetLevel() >= DebugLevelDebug {
		vm.logger.LogDebug("ExecuteOp: %s (args: %d)", op.Cmd.String(), len(op.Args))
	}

	switch op.Cmd {
	case interpreter.OpAssign:
		return vm.executeAssign(seq, op)

	case interpreter.OpCall:
		return vm.executeCall(seq, op)

	case interpreter.OpIf:
		return vm.executeIf(seq, op)

	case interpreter.OpFor:
		return vm.executeFor(seq, op)

	case interpreter.OpWhile:
		return vm.executeWhile(seq, op)

	case interpreter.OpWait:
		return vm.executeWait(seq, op)

	case interpreter.OpSetStep:
		return vm.executeSetStep(seq, op)

	case interpreter.OpRegisterEventHandler:
		return vm.executeRegisterEventHandler(seq, op)

	case interpreter.OpBinaryOp:
		// Binary operations are evaluated as part of expressions
		// They should not appear at the statement level
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpBinaryOp cannot be executed as statement")

	case interpreter.OpRegisterSequence:
		// Sequence registration is handled by the engine, not the VM
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpRegisterSequence should be handled by engine")

	default:
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "Unknown OpCode: %s", op.Cmd.String())
	}
}

// executeAssign handles variable assignment: x = value
func (vm *VM) executeAssign(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpAssign requires 2 arguments, got %d", len(op.Args))
	}

	// First argument must be a Variable
	varName, ok := op.Args[0].(interpreter.Variable)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpAssign first argument must be Variable, got %T", op.Args[0])
	}

	// Second argument is the value (can be literal or expression)
	value, err := vm.evaluateValue(seq, op.Args[1])
	if err != nil {
		return err
	}

	// Set the variable
	seq.SetVariable(string(varName), value)
	vm.logger.LogDebug("Assign: %s = %v", varName, value)

	return nil
}

// executeCall handles function calls
func (vm *VM) executeCall(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) == 0 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpCall requires at least 1 argument (function name)")
	}

	// Get function name
	var funcName string
	switch fn := op.Args[0].(type) {
	case string:
		funcName = fn
	case interpreter.Variable:
		funcName = string(fn)
	default:
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpCall first argument must be string or Variable, got %T", op.Args[0])
	}

	// Handle special built-in functions
	switch funcName {
	case "define_function":
		return vm.executeDefineFunction(seq, op)
	case "return":
		// TODO: Implement return statement handling
		vm.logger.LogDebug("Return statement (not yet implemented)")
		return nil
	default:
		// Try to call user-defined function
		return vm.executeUserFunction(seq, funcName, op.Args[1:])
	}
}

// executeDefineFunction handles function definition registration
func (vm *VM) executeDefineFunction(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 4 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function requires 4 arguments (name, params, body), got %d", len(op.Args))
	}

	// Get function name
	funcName, ok := op.Args[1].(string)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function name must be string, got %T", op.Args[1])
	}

	// Get parameters
	params, ok := op.Args[2].([]any)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function params must be []any, got %T", op.Args[2])
	}

	// Convert parameters to strings
	paramNames := make([]string, len(params))
	for i, p := range params {
		paramNames[i] = fmt.Sprintf("%v", p)
	}

	// Get function body
	body, ok := op.Args[3].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "define_function body must be []OpCode, got %T", op.Args[3])
	}

	// Register function in engine state
	vm.state.RegisterFunction(funcName, paramNames, body)
	vm.logger.LogDebug("Defined function: %s with %d parameters", funcName, len(paramNames))

	return nil
}

// executeUserFunction handles user-defined function calls
func (vm *VM) executeUserFunction(seq *Sequencer, funcName string, args []any) error {
	// Look up function definition
	funcDef, ok := vm.state.GetFunction(funcName)
	if !ok {
		// Not a user-defined function - try built-in functions
		return vm.executeBuiltinFunction(seq, funcName, args)
	}

	// Evaluate arguments
	evaluatedArgs := make([]any, len(args))
	for i, arg := range args {
		val, err := vm.evaluateValue(seq, arg)
		if err != nil {
			return err
		}
		evaluatedArgs[i] = val
	}

	// Create new sequencer for function execution with current sequencer as parent
	funcSeq := NewSequencer(funcDef.Body, seq.GetMode(), seq)

	// Bind parameters to arguments
	for i, paramName := range funcDef.Parameters {
		if i < len(evaluatedArgs) {
			funcSeq.SetVariable(paramName, evaluatedArgs[i])
		} else {
			// Parameter not provided, use default value (0)
			funcSeq.SetVariable(paramName, 0)
		}
	}

	// Execute function body synchronously
	vm.logger.LogDebug("Calling user function: %s with %d arguments", funcName, len(evaluatedArgs))
	return vm.executeBlock(funcSeq, funcDef.Body)
}

// executeBuiltinFunction handles built-in function calls
func (vm *VM) executeBuiltinFunction(seq *Sequencer, funcName string, args []any) error {
	// Evaluate arguments
	evaluatedArgs := make([]any, len(args))
	for i, arg := range args {
		if vm.logger.GetLevel() >= DebugLevelDebug {
			vm.logger.LogDebug("  arg[%d] BEFORE eval: %v (type: %T)", i, arg, arg)
		}
		val, err := vm.evaluateValue(seq, arg)
		if err != nil {
			return err
		}
		evaluatedArgs[i] = val
		if vm.logger.GetLevel() >= DebugLevelDebug {
			vm.logger.LogDebug("  arg[%d] AFTER eval: %v (type: %T)", i, val, val)
		}
	}

	if vm.logger.GetLevel() >= DebugLevelDebug {
		vm.logger.LogDebug("Call: %s (built-in function)", funcName)
	}

	// Handle built-in functions
	switch strings.ToLower(funcName) {
	case "loadpic":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("LoadPic", fmt.Sprintf("%v", evaluatedArgs), "LoadPic requires 1 argument (filename)")
		}
		filename := fmt.Sprintf("%v", evaluatedArgs[0])
		picID := vm.engine.LoadPic(filename)
		// Store result in special return variable
		seq.SetVariable("__return__", int64(picID))
		return nil

	case "createpic":
		// CreatePic can be called with 1 or 2 arguments:
		// CreatePic(sourcePicID) - 1 arg: copy dimensions from source picture
		// CreatePic(width, height) - 2 args: create blank picture with specified size
		if len(evaluatedArgs) == 1 {
			// CreatePic(sourcePicID) - copy dimensions from source
			sourcePicID := int(vm.toInt(evaluatedArgs[0]))
			width := vm.engine.PicWidth(sourcePicID)
			height := vm.engine.PicHeight(sourcePicID)
			if width == 0 || height == 0 {
				// Source picture not found or has invalid dimensions
				vm.logger.LogError("CreatePic: source picture %d not found or has invalid dimensions", sourcePicID)
				seq.SetVariable("__return__", int64(0))
				return nil
			}
			picID := vm.engine.CreatePic(width, height)
			seq.SetVariable("__return__", int64(picID))
			return nil
		} else if len(evaluatedArgs) >= 2 {
			// CreatePic(width, height) - create blank picture
			width := int(vm.toInt(evaluatedArgs[0]))
			height := int(vm.toInt(evaluatedArgs[1]))
			picID := vm.engine.CreatePic(width, height)
			seq.SetVariable("__return__", int64(picID))
			return nil
		} else {
			return NewRuntimeError("CreatePic", fmt.Sprintf("%v", evaluatedArgs), "CreatePic requires 1 or 2 arguments")
		}

	case "delpic":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("DelPic", fmt.Sprintf("%v", evaluatedArgs), "DelPic requires 1 argument (picID)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.DelPic(picID)
		return nil

	case "picwidth":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PicWidth", fmt.Sprintf("%v", evaluatedArgs), "PicWidth requires 1 argument (picID)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		width := vm.engine.PicWidth(picID)
		seq.SetVariable("__return__", int64(width))
		return nil

	case "picheight":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PicHeight", fmt.Sprintf("%v", evaluatedArgs), "PicHeight requires 1 argument (picID)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		height := vm.engine.PicHeight(picID)
		seq.SetVariable("__return__", int64(height))
		return nil

	case "wininfo":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("WinInfo", fmt.Sprintf("%v", evaluatedArgs), "WinInfo requires 1 argument (index)")
		}
		index := int(vm.toInt(evaluatedArgs[0]))
		value := vm.engine.WinInfo(index)
		seq.SetVariable("__return__", int64(value))
		return nil

	case "movepic":
		// MovePic requires 9 arguments (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, mode)
		// Pad missing arguments with 0
		for len(evaluatedArgs) < 8 {
			evaluatedArgs = append(evaluatedArgs, int64(0))
		}
		srcID := int(vm.toInt(evaluatedArgs[0]))
		srcX := int(vm.toInt(evaluatedArgs[1]))
		srcY := int(vm.toInt(evaluatedArgs[2]))
		srcW := int(vm.toInt(evaluatedArgs[3]))
		srcH := int(vm.toInt(evaluatedArgs[4]))
		dstID := int(vm.toInt(evaluatedArgs[5]))
		dstX := int(vm.toInt(evaluatedArgs[6]))
		dstY := int(vm.toInt(evaluatedArgs[7]))
		vm.engine.MovePic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)
		return nil

	case "movespic":
		// MoveSPic supports 10 or 12 arguments
		// 10 args: (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)
		// 12 args: (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH, 0, trans_color)
		// Pad missing arguments with 0
		for len(evaluatedArgs) < 10 {
			evaluatedArgs = append(evaluatedArgs, int64(0))
		}
		srcID := int(vm.toInt(evaluatedArgs[0]))
		srcX := int(vm.toInt(evaluatedArgs[1]))
		srcY := int(vm.toInt(evaluatedArgs[2]))
		srcW := int(vm.toInt(evaluatedArgs[3]))
		srcH := int(vm.toInt(evaluatedArgs[4]))
		dstID := int(vm.toInt(evaluatedArgs[5]))
		dstX := int(vm.toInt(evaluatedArgs[6]))
		dstY := int(vm.toInt(evaluatedArgs[7]))
		dstW := int(vm.toInt(evaluatedArgs[8]))
		dstH := int(vm.toInt(evaluatedArgs[9]))
		// TODO: Handle transparent color if 12 arguments provided (evaluatedArgs[11])
		vm.engine.MoveSPic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY, dstW, dstH)
		return nil

	case "reversepic":
		if len(evaluatedArgs) < 8 {
			return NewRuntimeError("ReversePic", fmt.Sprintf("%v", evaluatedArgs), "ReversePic requires 8 arguments (srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)")
		}
		srcID := int(vm.toInt(evaluatedArgs[0]))
		srcX := int(vm.toInt(evaluatedArgs[1]))
		srcY := int(vm.toInt(evaluatedArgs[2]))
		srcW := int(vm.toInt(evaluatedArgs[3]))
		srcH := int(vm.toInt(evaluatedArgs[4]))
		dstID := int(vm.toInt(evaluatedArgs[5]))
		dstX := int(vm.toInt(evaluatedArgs[6]))
		dstY := int(vm.toInt(evaluatedArgs[7]))
		vm.engine.ReversePic(srcID, srcX, srcY, srcW, srcH, dstID, dstX, dstY)
		return nil

	case "openwin":
		// OpenWin can be called with 1, 5, or 8 arguments
		// OpenWin(pic) - 1 arg: x=0, y=0, w=0, h=0, picX=0, picY=0, col=0
		// OpenWin(pic, x, y, w, h) - 5 args: picX=0, picY=0, col=0
		// OpenWin(pic, x, y, w, h, picX, picY, col) - 8 args (full)
		// Pad missing arguments with 0
		for len(evaluatedArgs) < 8 {
			evaluatedArgs = append(evaluatedArgs, int64(0))
		}

		picID := int(vm.toInt(evaluatedArgs[0]))
		x := int(vm.toInt(evaluatedArgs[1]))
		y := int(vm.toInt(evaluatedArgs[2]))
		width := int(vm.toInt(evaluatedArgs[3]))
		height := int(vm.toInt(evaluatedArgs[4]))
		picX := int(vm.toInt(evaluatedArgs[5]))
		picY := int(vm.toInt(evaluatedArgs[6]))
		color := int(vm.toInt(evaluatedArgs[7]))
		winID := vm.engine.OpenWin(picID, x, y, width, height, picX, picY, color)
		seq.SetVariable("__return__", int64(winID))
		return nil

	case "movewin":
		// MoveWin can be called with 2 or 8 arguments
		// MoveWin(win, pic) - 2 args: short form (picture change only)
		// MoveWin(win, pic, x, y, width, height, pic_x, pic_y) - 8 args (full)
		if len(evaluatedArgs) < 2 {
			return NewRuntimeError("MoveWin", fmt.Sprintf("%v", evaluatedArgs), "MoveWin requires at least 2 arguments (winID, picID)")
		}

		// Pad missing arguments with 0
		for len(evaluatedArgs) < 8 {
			evaluatedArgs = append(evaluatedArgs, int64(0))
		}

		winID := int(vm.toInt(evaluatedArgs[0]))
		picID := int(vm.toInt(evaluatedArgs[1]))
		x := int(vm.toInt(evaluatedArgs[2]))
		y := int(vm.toInt(evaluatedArgs[3]))
		width := int(vm.toInt(evaluatedArgs[4]))
		height := int(vm.toInt(evaluatedArgs[5]))
		picX := int(vm.toInt(evaluatedArgs[6]))
		picY := int(vm.toInt(evaluatedArgs[7]))
		vm.engine.MoveWin(winID, picID, x, y, width, height, picX, picY)
		return nil

	case "closewin":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("CloseWin", fmt.Sprintf("%v", evaluatedArgs), "CloseWin requires 1 argument (winID)")
		}
		winID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.CloseWin(winID)
		return nil

	case "closewinall":
		vm.engine.CloseWinAll()
		return nil

	case "captitle":
		// CapTitle can take 1 or 2 arguments
		// 1 arg: CapTitle(caption) - applies to default window (ID=0)
		// 2 args: CapTitle(winID, caption)
		var winID int
		var caption string

		if len(evaluatedArgs) == 1 {
			winID = 0
			caption = fmt.Sprintf("%v", evaluatedArgs[0])
		} else if len(evaluatedArgs) == 2 {
			winID = int(vm.toInt(evaluatedArgs[0]))
			caption = fmt.Sprintf("%v", evaluatedArgs[1])
		} else {
			return NewRuntimeError("CapTitle", fmt.Sprintf("%v", evaluatedArgs), "CapTitle requires 1 or 2 arguments")
		}

		vm.engine.CapTitle(winID, caption)
		return nil

	case "getpicno":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("GetPicNo", fmt.Sprintf("%v", evaluatedArgs), "GetPicNo requires 1 argument (winID)")
		}
		winID := int(vm.toInt(evaluatedArgs[0]))
		picID := vm.engine.GetPicNo(winID)
		seq.SetVariable("__return__", int64(picID))
		return nil

	case "putcast":
		// PutCast can be called with 4, 5, 8, or 12 arguments:
		// PutCast(picID, destPicID, x, y) - 4 args: basic cast
		// PutCast(picID, destPicID, x, y, transparentColor) - 5 args: with transparency
		// PutCast(picID, destPicID, x, y, srcX, srcY, width, height) - 8 args: with clipping
		// PutCast(picID, destPicID, x, y, transparentColor, ?, ?, ?, width, height, srcX, srcY) - 12 args: full
		if len(evaluatedArgs) < 4 {
			return NewRuntimeError("PutCast", fmt.Sprintf("%v", evaluatedArgs), "PutCast requires at least 4 arguments (picID, destPicID, x, y)")
		}

		// Get basic arguments
		picID := int(vm.toInt(evaluatedArgs[0]))
		destPicID := int(vm.toInt(evaluatedArgs[1]))
		x := int(vm.toInt(evaluatedArgs[2]))
		y := int(vm.toInt(evaluatedArgs[3]))

		// Default values for optional arguments
		srcX := 0
		srcY := 0
		width := vm.engine.PicWidth(picID)
		height := vm.engine.PicHeight(picID)
		transparentColor := -1 // -1 means no transparency

		// Parse optional arguments based on count
		if len(evaluatedArgs) == 5 {
			// 5 args: includes transparent color
			transparentColor = int(vm.toInt(evaluatedArgs[4]))
		} else if len(evaluatedArgs) == 8 {
			// 8 args: includes clipping parameters
			srcX = int(vm.toInt(evaluatedArgs[4]))
			srcY = int(vm.toInt(evaluatedArgs[5]))
			width = int(vm.toInt(evaluatedArgs[6]))
			height = int(vm.toInt(evaluatedArgs[7]))
		} else if len(evaluatedArgs) >= 12 {
			// 12 args: transparentColor at [4], width/height at [8-9], srcX/srcY at [10-11]
			transparentColor = int(vm.toInt(evaluatedArgs[4]))
			// args[5-7] = unused
			width = int(vm.toInt(evaluatedArgs[8]))
			height = int(vm.toInt(evaluatedArgs[9]))
			srcX = int(vm.toInt(evaluatedArgs[10]))
			srcY = int(vm.toInt(evaluatedArgs[11]))
		}

		castID := vm.engine.PutCast(destPicID, picID, x, y, srcX, srcY, width, height, transparentColor)
		seq.SetVariable("__return__", int64(castID))
		return nil

	case "movecast":
		// MoveCast parameters (based on old implementation and y_saru sample):
		// args[0]: castID
		// args[1]: picID (source picture, often ignored for transparency preservation)
		// args[2]: x position
		// args[3]: y position
		// args[4]: unknown/transparent color (ignored for now)
		// args[5]: width
		// args[6]: height
		// args[7]: srcX (source clipping X)
		// args[8]: srcY (source clipping Y)
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("MoveCast", fmt.Sprintf("%v", evaluatedArgs), "MoveCast requires at least 3 arguments (castID, picID, x)")
		}

		castID := int(vm.toInt(evaluatedArgs[0]))
		// picID := int(vm.toInt(evaluatedArgs[1])) // Ignored for now to preserve transparency
		x := int(vm.toInt(evaluatedArgs[2]))
		y := 0
		if len(evaluatedArgs) > 3 {
			y = int(vm.toInt(evaluatedArgs[3]))
		}

		// Default to -1 (no change) for clipping parameters
		srcX := -1
		srcY := -1
		width := -1
		height := -1

		// Parse additional arguments for dimensions and source offset
		if len(evaluatedArgs) > 5 {
			width = int(vm.toInt(evaluatedArgs[5]))
		}
		if len(evaluatedArgs) > 6 {
			height = int(vm.toInt(evaluatedArgs[6]))
		}
		if len(evaluatedArgs) > 7 {
			srcX = int(vm.toInt(evaluatedArgs[7]))
		}
		if len(evaluatedArgs) > 8 {
			srcY = int(vm.toInt(evaluatedArgs[8]))
		}

		vm.engine.MoveCast(castID, x, y, srcX, srcY, width, height)
		return nil

	case "delcast":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("DelCast", fmt.Sprintf("%v", evaluatedArgs), "DelCast requires 1 argument (castID)")
		}
		castID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.DelCast(castID)
		return nil

	case "loadsoundfont":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("LoadSoundFont", fmt.Sprintf("%v", evaluatedArgs), "LoadSoundFont requires 1 argument (filename)")
		}
		filename := fmt.Sprintf("%v", evaluatedArgs[0])
		err := vm.engine.LoadSoundFont(filename)
		if err != nil {
			vm.logger.LogError("LoadSoundFont failed: %v", err)
		}
		return nil

	case "playmidi":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PlayMIDI", fmt.Sprintf("%v", evaluatedArgs), "PlayMIDI requires 1 argument (filename)")
		}
		filename := fmt.Sprintf("%v", evaluatedArgs[0])
		err := vm.engine.PlayMIDI(filename)
		if err != nil {
			vm.logger.LogError("PlayMIDI failed: %v", err)
		}
		return nil

	case "stopmidi":
		vm.engine.StopMIDI()
		return nil

	case "del_me":
		// Deactivate current sequence
		vm.engine.DeleteMe(seq.id)
		return nil

	case "del_us":
		// Deactivate all sequences in current group
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("del_us", fmt.Sprintf("%v", evaluatedArgs), "del_us requires 1 argument (groupID)")
		}
		groupID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.DeleteUs(groupID)
		return nil

	case "del_all":
		// Deactivate all sequences
		vm.engine.DeleteAll()
		return nil

	case "playwave":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PlayWAVE", fmt.Sprintf("%v", evaluatedArgs), "PlayWAVE requires 1 argument (filename)")
		}
		filename := fmt.Sprintf("%v", evaluatedArgs[0])
		err := vm.engine.PlayWAVE(filename)
		if err != nil {
			vm.logger.LogError("PlayWAVE failed: %v", err)
		}
		return nil

	case "loadrsc":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("LoadRsc", fmt.Sprintf("%v", evaluatedArgs), "LoadRsc requires 1 argument (filename)")
		}
		filename := fmt.Sprintf("%v", evaluatedArgs[0])
		resourceID, err := vm.engine.LoadRsc(filename)
		if err != nil {
			vm.logger.LogError("LoadRsc failed: %v", err)
			seq.SetVariable("__return__", int64(0))
			return nil
		}
		// Store resource ID in return variable
		seq.SetVariable("__return__", int64(resourceID))
		return nil

	case "playrsc":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("PlayRsc", fmt.Sprintf("%v", evaluatedArgs), "PlayRsc requires 1 argument (resourceID)")
		}
		resourceID := int(vm.toInt(evaluatedArgs[0]))
		err := vm.engine.PlayRsc(resourceID)
		if err != nil {
			vm.logger.LogError("PlayRsc failed: %v", err)
		}
		return nil

	case "delrsc":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("DelRsc", fmt.Sprintf("%v", evaluatedArgs), "DelRsc requires 1 argument (resourceID)")
		}
		resourceID := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.DelRsc(resourceID)
		return nil

	case "setfont":
		// SetFont(size, name, charset)
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("SetFont", fmt.Sprintf("%v", evaluatedArgs), "SetFont requires 3 arguments (size, name, charset)")
		}
		size := int(vm.toInt(evaluatedArgs[0]))
		name := fmt.Sprintf("%v", evaluatedArgs[1])
		charset := int(vm.toInt(evaluatedArgs[2]))
		vm.engine.SetFont(size, name, charset)
		return nil

	case "textcolor":
		// TextColor(r, g, b)
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("TextColor", fmt.Sprintf("%v", evaluatedArgs), "TextColor requires 3 arguments (r, g, b)")
		}
		r := int(vm.toInt(evaluatedArgs[0]))
		g := int(vm.toInt(evaluatedArgs[1]))
		b := int(vm.toInt(evaluatedArgs[2]))
		vm.engine.TextColor(r, g, b)
		return nil

	case "bgcolor":
		// BgColor(r, g, b)
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("BgColor", fmt.Sprintf("%v", evaluatedArgs), "BgColor requires 3 arguments (r, g, b)")
		}
		r := int(vm.toInt(evaluatedArgs[0]))
		g := int(vm.toInt(evaluatedArgs[1]))
		b := int(vm.toInt(evaluatedArgs[2]))
		vm.engine.BgColor(r, g, b)
		return nil

	case "backmode":
		// BackMode(mode)
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("BackMode", fmt.Sprintf("%v", evaluatedArgs), "BackMode requires 1 argument (mode)")
		}
		mode := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.BackMode(mode)
		return nil

	case "textwrite":
		// TextWrite(text, pic, x, y)
		if len(evaluatedArgs) < 4 {
			return NewRuntimeError("TextWrite", fmt.Sprintf("%v", evaluatedArgs), "TextWrite requires 4 arguments (text, pic, x, y)")
		}
		text := fmt.Sprintf("%v", evaluatedArgs[0])
		picID := int(vm.toInt(evaluatedArgs[1]))
		x := int(vm.toInt(evaluatedArgs[2]))
		y := int(vm.toInt(evaluatedArgs[3]))
		err := vm.engine.TextWrite(text, picID, x, y)
		if err != nil {
			vm.logger.LogError("TextWrite failed: %v", err)
		}
		return nil

	case "setlinesize":
		// SetLineSize(size)
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("SetLineSize", fmt.Sprintf("%v", evaluatedArgs), "SetLineSize requires 1 argument (size)")
		}
		size := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.SetLineSize(size)
		return nil

	case "setpaintcolor":
		// SetPaintColor(color)
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("SetPaintColor", fmt.Sprintf("%v", evaluatedArgs), "SetPaintColor requires 1 argument (color)")
		}
		colorValue := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.SetPaintColor(colorValue)
		return nil

	case "setrop":
		// SetROP(mode)
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("SetROP", fmt.Sprintf("%v", evaluatedArgs), "SetROP requires 1 argument (mode)")
		}
		mode := int(vm.toInt(evaluatedArgs[0]))
		vm.engine.SetROP(mode)
		return nil

	case "drawline":
		// DrawLine(pic, x1, y1, x2, y2)
		if len(evaluatedArgs) < 5 {
			return NewRuntimeError("DrawLine", fmt.Sprintf("%v", evaluatedArgs), "DrawLine requires 5 arguments (pic, x1, y1, x2, y2)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		x1 := int(vm.toInt(evaluatedArgs[1]))
		y1 := int(vm.toInt(evaluatedArgs[2]))
		x2 := int(vm.toInt(evaluatedArgs[3]))
		y2 := int(vm.toInt(evaluatedArgs[4]))
		err := vm.engine.DrawLine(picID, x1, y1, x2, y2)
		if err != nil {
			vm.logger.LogError("DrawLine failed: %v", err)
		}
		return nil

	case "drawcircle":
		// DrawCircle(pic, x, y, radius, fill_mode)
		if len(evaluatedArgs) < 5 {
			return NewRuntimeError("DrawCircle", fmt.Sprintf("%v", evaluatedArgs), "DrawCircle requires 5 arguments (pic, x, y, radius, fill_mode)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		x := int(vm.toInt(evaluatedArgs[1]))
		y := int(vm.toInt(evaluatedArgs[2]))
		radius := int(vm.toInt(evaluatedArgs[3]))
		fillMode := int(vm.toInt(evaluatedArgs[4]))
		err := vm.engine.DrawCircle(picID, x, y, radius, fillMode)
		if err != nil {
			vm.logger.LogError("DrawCircle failed: %v", err)
		}
		return nil

	case "drawrect":
		// DrawRect(pic, x1, y1, x2, y2, fill_mode)
		if len(evaluatedArgs) < 6 {
			return NewRuntimeError("DrawRect", fmt.Sprintf("%v", evaluatedArgs), "DrawRect requires 6 arguments (pic, x1, y1, x2, y2, fill_mode)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		x1 := int(vm.toInt(evaluatedArgs[1]))
		y1 := int(vm.toInt(evaluatedArgs[2]))
		x2 := int(vm.toInt(evaluatedArgs[3]))
		y2 := int(vm.toInt(evaluatedArgs[4]))
		fillMode := int(vm.toInt(evaluatedArgs[5]))
		err := vm.engine.DrawRect(picID, x1, y1, x2, y2, fillMode)
		if err != nil {
			vm.logger.LogError("DrawRect failed: %v", err)
		}
		return nil

	case "getcolor":
		// GetColor(pic, x, y)
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("GetColor", fmt.Sprintf("%v", evaluatedArgs), "GetColor requires 3 arguments (pic, x, y)")
		}
		picID := int(vm.toInt(evaluatedArgs[0]))
		x := int(vm.toInt(evaluatedArgs[1]))
		y := int(vm.toInt(evaluatedArgs[2]))
		colorValue, err := vm.engine.GetColor(picID, x, y)
		if err != nil {
			vm.logger.LogError("GetColor failed: %v", err)
			seq.SetVariable("__return__", int64(0))
			return nil
		}
		// Store result in return variable
		seq.SetVariable("__return__", int64(colorValue))
		return nil

	case "end_step":
		// end_step is used to break out of step(n) { ... } blocks
		// Return a special signal that executeFor will catch
		vm.logger.LogDebug("end_step: breaking out of step block")
		return &EndStepSignal{}

	// String operations
	case "strlen":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("StrLen", fmt.Sprintf("%v", evaluatedArgs), "StrLen requires 1 argument (string)")
		}
		str := fmt.Sprintf("%v", evaluatedArgs[0])
		length := len([]rune(str)) // Use rune count for proper Unicode support
		seq.SetVariable("__return__", int64(length))
		return nil

	case "substr":
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("SubStr", fmt.Sprintf("%v", evaluatedArgs), "SubStr requires 3 arguments (string, start, length)")
		}
		str := fmt.Sprintf("%v", evaluatedArgs[0])
		start := int(vm.toInt(evaluatedArgs[1]))
		length := int(vm.toInt(evaluatedArgs[2]))

		runes := []rune(str)
		if start < 0 || start >= len(runes) {
			seq.SetVariable("__return__", "")
			return nil
		}

		end := start + length
		if end > len(runes) {
			end = len(runes)
		}

		result := string(runes[start:end])
		seq.SetVariable("__return__", result)
		return nil

	case "strfind":
		if len(evaluatedArgs) < 2 {
			return NewRuntimeError("StrFind", fmt.Sprintf("%v", evaluatedArgs), "StrFind requires 2 arguments (string, search)")
		}
		str := fmt.Sprintf("%v", evaluatedArgs[0])
		search := fmt.Sprintf("%v", evaluatedArgs[1])

		// Find position (0-based rune index, -1 if not found)
		bytePos := strings.Index(str, search)
		if bytePos == -1 {
			seq.SetVariable("__return__", int64(-1))
			return nil
		}

		// Convert byte position to rune position
		runePos := len([]rune(str[:bytePos]))
		seq.SetVariable("__return__", int64(runePos))
		return nil

	case "strprint":
		// StrPrint(format, args...) - printf-style formatting
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("StrPrint", fmt.Sprintf("%v", evaluatedArgs), "StrPrint requires at least 1 argument (format)")
		}
		format := fmt.Sprintf("%v", evaluatedArgs[0])
		args := evaluatedArgs[1:]
		result := fmt.Sprintf(format, args...)
		seq.SetVariable("__return__", result)
		return nil

	case "strup":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("StrUp", fmt.Sprintf("%v", evaluatedArgs), "StrUp requires 1 argument (string)")
		}
		str := fmt.Sprintf("%v", evaluatedArgs[0])
		result := strings.ToUpper(str)
		seq.SetVariable("__return__", result)
		return nil

	case "strlow":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("StrLow", fmt.Sprintf("%v", evaluatedArgs), "StrLow requires 1 argument (string)")
		}
		str := fmt.Sprintf("%v", evaluatedArgs[0])
		result := strings.ToLower(str)
		seq.SetVariable("__return__", result)
		return nil

	case "charcode":
		// CharCode(char) - returns ASCII/Unicode code of first character
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("CharCode", fmt.Sprintf("%v", evaluatedArgs), "CharCode requires 1 argument (char)")
		}
		str := fmt.Sprintf("%v", evaluatedArgs[0])
		if len(str) == 0 {
			seq.SetVariable("__return__", int64(0))
			return nil
		}
		runes := []rune(str)
		code := int64(runes[0])
		seq.SetVariable("__return__", code)
		return nil

	case "strcode":
		// StrCode(code) - returns character from ASCII/Unicode code
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("StrCode", fmt.Sprintf("%v", evaluatedArgs), "StrCode requires 1 argument (code)")
		}
		code := int(vm.toInt(evaluatedArgs[0]))
		result := string(rune(code))
		seq.SetVariable("__return__", result)
		return nil

	// Array operations
	case "arraysize":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("ArraySize", fmt.Sprintf("%v", evaluatedArgs), "ArraySize requires 1 argument (array variable name)")
		}
		// Get array variable name
		varName := fmt.Sprintf("%v", evaluatedArgs[0])
		// Get array from sequencer
		value := seq.GetVariable(varName)
		if arr, ok := value.([]int64); ok {
			seq.SetVariable("__return__", int64(len(arr)))
		} else {
			seq.SetVariable("__return__", int64(0))
		}
		return nil

	case "delarrayall":
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("DelArrayAll", fmt.Sprintf("%v", evaluatedArgs), "DelArrayAll requires 1 argument (array variable name)")
		}
		// Get array variable name
		varName := fmt.Sprintf("%v", evaluatedArgs[0])
		// Clear array
		seq.SetVariable(varName, []int64{})
		return nil

	case "delarrayat":
		if len(evaluatedArgs) < 2 {
			return NewRuntimeError("DelArrayAt", fmt.Sprintf("%v", evaluatedArgs), "DelArrayAt requires 2 arguments (array variable name, index)")
		}
		// Get array variable name and index
		varName := fmt.Sprintf("%v", evaluatedArgs[0])
		index := int(vm.toInt(evaluatedArgs[1]))

		// Get array from sequencer
		value := seq.GetVariable(varName)
		if arr, ok := value.([]int64); ok {
			if index >= 0 && index < len(arr) {
				// Remove element at index
				newArr := append(arr[:index], arr[index+1:]...)
				seq.SetVariable(varName, newArr)
			}
		}
		return nil

	case "insarrayat":
		if len(evaluatedArgs) < 3 {
			return NewRuntimeError("InsArrayAt", fmt.Sprintf("%v", evaluatedArgs), "InsArrayAt requires 3 arguments (array variable name, index, value)")
		}
		// Get array variable name, index, and value
		varName := fmt.Sprintf("%v", evaluatedArgs[0])
		index := int(vm.toInt(evaluatedArgs[1]))
		value := vm.toInt(evaluatedArgs[2])

		// Get array from sequencer
		arrValue := seq.GetVariable(varName)
		arr, ok := arrValue.([]int64)
		if !ok {
			// Create new array if variable is not an array
			arr = []int64{}
		}

		// Insert element at index
		if index < 0 {
			index = 0
		}
		if index > len(arr) {
			index = len(arr)
		}

		// Create new array with inserted element
		newArr := make([]int64, len(arr)+1)
		copy(newArr[:index], arr[:index])
		newArr[index] = value
		copy(newArr[index+1:], arr[index:])

		seq.SetVariable(varName, newArr)
		return nil

	// System functions
	case "random":
		// Random(max) - returns random number from 0 to max-1
		// Random(min, max) - returns random number from min to max-1
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("Random", fmt.Sprintf("%v", evaluatedArgs), "Random requires at least 1 argument (max)")
		}

		var min, max int64
		if len(evaluatedArgs) == 1 {
			min = 0
			max = vm.toInt(evaluatedArgs[0])
		} else {
			min = vm.toInt(evaluatedArgs[0])
			max = vm.toInt(evaluatedArgs[1])
		}

		if max <= min {
			seq.SetVariable("__return__", min)
			return nil
		}

		// Generate random number in range [min, max)
		result := min + int64(vm.engine.random.IntN(int(max-min)))
		seq.SetVariable("__return__", result)
		return nil

	case "getsystime":
		// GetSysTime() - returns current Unix timestamp in seconds
		timestamp := vm.engine.GetSysTime()
		seq.SetVariable("__return__", timestamp)
		return nil

	case "whatday":
		// WhatDay() - returns day of month (1-31)
		day := vm.engine.WhatDay()
		seq.SetVariable("__return__", int64(day))
		return nil

	case "whattime":
		// WhatTime(mode) - returns time component
		// mode 0: hour (0-23)
		// mode 1: minute (0-59)
		// mode 2: second (0-59)
		if len(evaluatedArgs) < 1 {
			return NewRuntimeError("WhatTime", fmt.Sprintf("%v", evaluatedArgs), "WhatTime requires 1 argument (mode)")
		}
		mode := int(vm.toInt(evaluatedArgs[0]))
		value := vm.engine.WhatTime(mode)
		seq.SetVariable("__return__", int64(value))
		return nil

	// Legacy function stubs (Windows-specific functions no longer supported)
	case "shell":
		// Shell(program, args) - launches external programs (Windows-specific)
		// This is a legacy function that is no longer supported on cross-platform systems
		vm.logger.LogInfo("Shell() is a legacy Windows-specific function and is not supported")
		seq.SetVariable("__return__", int64(0))
		return nil

	case "getinistr":
		// GetIniStr(section, key, default, filename) - reads INI configuration files
		// This is a legacy function that is no longer supported
		vm.logger.LogInfo("GetIniStr() is a legacy function and is not supported")
		seq.SetVariable("__return__", "")
		return nil

	case "mci":
		// MCI(command) - Windows Media Control Interface commands
		// This is a legacy Windows-specific function that is no longer supported
		vm.logger.LogInfo("MCI() is a legacy Windows-specific function and is not supported")
		seq.SetVariable("__return__", int64(0))
		return nil

	case "strmci":
		// StrMCI(command) - String variant of MCI commands
		// This is a legacy Windows-specific function that is no longer supported
		vm.logger.LogInfo("StrMCI() is a legacy Windows-specific function and is not supported")
		seq.SetVariable("__return__", "")
		return nil

	default:
		// Unknown built-in function - just log and ignore
		vm.logger.LogDebug("Unknown built-in function: %s", funcName)
		return nil
	}
}

// executeIf handles if statements: if (condition) { thenBlock } else { elseBlock }
func (vm *VM) executeIf(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) < 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpIf requires at least 2 arguments, got %d", len(op.Args))
	}

	// Evaluate condition
	condition, err := vm.evaluateValue(seq, op.Args[0])
	if err != nil {
		return err
	}

	// Convert condition to boolean
	condBool := vm.toBool(condition)
	vm.logger.LogDebug("If: condition = %v", condBool)

	// Get then block
	thenBlock, ok := op.Args[1].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpIf second argument must be []OpCode, got %T", op.Args[1])
	}

	// Get else block (optional)
	var elseBlock []interpreter.OpCode
	if len(op.Args) >= 3 {
		elseBlock, ok = op.Args[2].([]interpreter.OpCode)
		if !ok {
			return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpIf third argument must be []OpCode, got %T", op.Args[2])
		}
	}

	// Execute appropriate block
	if condBool {
		return vm.executeBlock(seq, thenBlock)
	} else if len(elseBlock) > 0 {
		return vm.executeBlock(seq, elseBlock)
	}

	return nil
}

// executeFor handles for loops: for (init; condition; increment) { body }
func (vm *VM) executeFor(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 4 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpFor requires 4 arguments, got %d", len(op.Args))
	}

	// Execute init
	if op.Args[0] != nil {
		if initOp, ok := op.Args[0].(interpreter.OpCode); ok {
			if err := vm.ExecuteOp(seq, initOp); err != nil {
				return err
			}
		}
	}

	// Get condition, increment, and body
	condition := op.Args[1]
	increment := op.Args[2]
	body, ok := op.Args[3].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpFor fourth argument must be []OpCode, got %T", op.Args[3])
	}

	// Execute loop with termination checks
	iterations := 0
	maxIterations := 100000 // Safety limit
	for {
		// Check for termination every 100 iterations
		if iterations%100 == 0 && vm.engine.CheckTermination() {
			return ErrTerminated
		}

		// Evaluate condition
		if condition != nil {
			condValue, err := vm.evaluateValue(seq, condition)
			if err != nil {
				return err
			}
			if !vm.toBool(condValue) {
				break
			}
		}

		// Execute body
		if err := vm.executeBlock(seq, body); err != nil {
			// Check if it's an end_step signal (used to break out of step blocks)
			if IsEndStepSignal(err) {
				vm.logger.LogDebug("Caught end_step signal, breaking loop")
				break
			}
			return err
		}

		// Execute increment
		if increment != nil {
			if incOp, ok := increment.(interpreter.OpCode); ok {
				if err := vm.ExecuteOp(seq, incOp); err != nil {
					return err
				}
			}
		}

		iterations++
		if iterations >= maxIterations {
			vm.logger.LogError("For loop hit max iterations limit (%d)", maxIterations)
			return ErrTerminated
		}
	}

	return nil
}

// executeWhile handles while loops: while (condition) { body }
func (vm *VM) executeWhile(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpWhile requires 2 arguments, got %d", len(op.Args))
	}

	condition := op.Args[0]
	body, ok := op.Args[1].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpWhile second argument must be []OpCode, got %T", op.Args[1])
	}

	// Execute loop with termination checks
	iterations := 0
	maxIterations := 100000 // Safety limit
	for {
		// Check for termination every 100 iterations
		if iterations%100 == 0 && vm.engine.CheckTermination() {
			return ErrTerminated
		}

		// Evaluate condition
		condValue, err := vm.evaluateValue(seq, condition)
		if err != nil {
			return err
		}

		if !vm.toBool(condValue) {
			break
		}

		// Execute body
		if err := vm.executeBlock(seq, body); err != nil {
			return err
		}

		iterations++
		if iterations >= maxIterations {
			vm.logger.LogError("While loop hit max iterations limit (%d)", maxIterations)
			return ErrTerminated
		}
	}

	return nil
}

// executeWait handles wait operations: wait(n)
// The argument n represents the number of steps to wait.
// The actual tick duration is determined by ticksPerStep (set by SetStep).
//
// In step() blocks:
//
//	SetStep(65) sets ticksPerStep = 65 * 3 = 195 ticks (at 60 FPS)
//	Wait(1) waits for 1 * ticksPerStep = 195 ticks
//	Wait(2) waits for 2 * ticksPerStep = 390 ticks
//
// IMPORTANT: We do NOT decrement PC here. UpdateVM will increment PC after
// this function returns, so PC will point to the next instruction. While
// waiting, the sequence skips execution, so PC stays at the next instruction.
// When the wait completes, execution resumes at the next instruction.
func (vm *VM) executeWait(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 1 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpWait requires 1 argument, got %d", len(op.Args))
	}

	// Evaluate wait count (number of steps)
	waitValue, err := vm.evaluateValue(seq, op.Args[0])
	if err != nil {
		return err
	}

	// Convert to int (number of steps)
	steps := int(vm.toInt(waitValue))

	// Calculate ticks using ticksPerStep
	ticks := steps * seq.GetTicksPerStep()

	vm.logger.LogDebug("Wait: %d steps × %d ticks/step = %d ticks", steps, seq.GetTicksPerStep(), ticks)

	// Set wait counter
	seq.SetWait(ticks)

	return nil
}

// executeSetStep handles SetStep operations: SetStep(n)
// This sets the duration of each step for subsequent Wait operations.
// In TIME mode: step(n) means each step is n × 50ms = n × 3 ticks (at 60 FPS)
// In MIDI_TIME mode: step(n) means each step is n × (1/32nd note)
func (vm *VM) executeSetStep(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 1 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpSetStep requires 1 argument, got %d", len(op.Args))
	}

	// Evaluate step count
	stepValue, err := vm.evaluateValue(seq, op.Args[0])
	if err != nil {
		return err
	}

	// Convert to int
	stepCount := int(vm.toInt(stepValue))

	// Calculate ticksPerStep based on timing mode
	var ticksPerStep int
	if seq.GetMode() == TIME {
		// TIME mode: step(n) = n × 50ms = n × 3 ticks at 60 FPS
		ticksPerStep = stepCount * 3
	} else {
		// MIDI_TIME mode: step(n) = n × (1/32nd note)
		// In FILLY's tick system, 1 quarter note = 8 ticks (32nd note resolution)
		// Therefore, 1/32nd note = 1 tick
		// So step(n) = n ticks
		ticksPerStep = stepCount
	}

	vm.logger.LogInfo("SetStep: %d → %d ticks/step (mode: %d)", stepCount, ticksPerStep, seq.GetMode())

	// Set ticksPerStep in sequencer
	seq.SetTicksPerStep(ticksPerStep)

	return nil
}

// executeRegisterEventHandler handles mes() event handler registration
func (vm *VM) executeRegisterEventHandler(seq *Sequencer, op interpreter.OpCode) error {
	if len(op.Args) != 2 {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpRegisterEventHandler requires 2 arguments, got %d", len(op.Args))
	}

	// Get event type
	eventTypeStr, ok := op.Args[0].(string)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpRegisterEventHandler first argument must be string, got %T", op.Args[0])
	}

	// Get event body
	body, ok := op.Args[1].([]interpreter.OpCode)
	if !ok {
		return NewRuntimeError(op.Cmd.String(), fmt.Sprintf("%v", op.Args), "OpRegisterEventHandler second argument must be []OpCode, got %T", op.Args[1])
	}

	// Parse event type
	eventType := ParseEventType(eventTypeStr)

	// Register the event handler
	vm.engine.RegisterMesBlock(eventType, body, seq, 0)

	return nil
}

// executeBlock executes a block of OpCodes sequentially
func (vm *VM) executeBlock(seq *Sequencer, block []interpreter.OpCode) error {
	for _, op := range block {
		if err := vm.ExecuteOp(seq, op); err != nil {
			return err
		}
	}
	return nil
}

// evaluateValue evaluates a value (literal, variable, or expression)
func (vm *VM) evaluateValue(seq *Sequencer, value any) (any, error) {
	switch v := value.(type) {
	case int, int64, float64, string, bool:
		// Literal value
		return v, nil

	case interpreter.Variable:
		// Variable reference
		return seq.GetVariable(string(v)), nil

	case interpreter.OpCode:
		// Nested expression
		return vm.evaluateExpression(seq, v)

	default:
		return nil, fmt.Errorf("cannot evaluate value of type %T", value)
	}
}

// evaluateExpression evaluates an expression OpCode
func (vm *VM) evaluateExpression(seq *Sequencer, op interpreter.OpCode) (any, error) {
	switch op.Cmd {
	case interpreter.OpBinaryOp:
		return vm.evaluateBinaryOp(seq, op)

	case interpreter.OpCall:
		// Execute function call
		err := vm.executeCall(seq, op)
		if err != nil {
			return nil, err
		}
		// Read return value from special variable
		returnValue := seq.GetVariable("__return__")
		// Clear return variable for next call
		seq.SetVariable("__return__", int64(0))
		return returnValue, nil

	default:
		return nil, fmt.Errorf("cannot evaluate expression: %s", op.Cmd.String())
	}
}

// evaluateBinaryOp evaluates a binary operation
func (vm *VM) evaluateBinaryOp(seq *Sequencer, op interpreter.OpCode) (any, error) {
	if len(op.Args) != 3 {
		return nil, fmt.Errorf("OpBinaryOp requires 3 arguments, got %d", len(op.Args))
	}

	// Get operator
	operator, ok := op.Args[0].(string)
	if !ok {
		return nil, fmt.Errorf("OpBinaryOp first argument must be string, got %T", op.Args[0])
	}

	// Evaluate left and right operands
	left, err := vm.evaluateValue(seq, op.Args[1])
	if err != nil {
		return nil, err
	}

	right, err := vm.evaluateValue(seq, op.Args[2])
	if err != nil {
		return nil, err
	}

	// Perform operation
	return vm.applyBinaryOp(operator, left, right)
}

// applyBinaryOp applies a binary operator to two values
func (vm *VM) applyBinaryOp(op string, left, right any) (any, error) {
	// Convert to int64 for arithmetic
	leftInt := vm.toInt(left)
	rightInt := vm.toInt(right)

	switch op {
	case "+":
		return leftInt + rightInt, nil
	case "-":
		return leftInt - rightInt, nil
	case "*":
		return leftInt * rightInt, nil
	case "/":
		if rightInt == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return leftInt / rightInt, nil
	case "%":
		if rightInt == 0 {
			return nil, fmt.Errorf("modulo by zero")
		}
		return leftInt % rightInt, nil
	case "==":
		return leftInt == rightInt, nil
	case "!=":
		return leftInt != rightInt, nil
	case "<":
		return leftInt < rightInt, nil
	case ">":
		return leftInt > rightInt, nil
	case "<=":
		return leftInt <= rightInt, nil
	case ">=":
		return leftInt >= rightInt, nil
	case "&&":
		return vm.toBool(left) && vm.toBool(right), nil
	case "||":
		return vm.toBool(left) || vm.toBool(right), nil
	default:
		return nil, fmt.Errorf("unknown binary operator: %s", op)
	}
}

// toBool converts a value to boolean
func (vm *VM) toBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case int:
		return v != 0
	case int64:
		return v != 0
	case float64:
		return v != 0.0
	case string:
		return v != ""
	default:
		return false
	}
}

// toInt converts a value to int64
func (vm *VM) toInt(value any) int64 {
	switch v := value.(type) {
	case int:
		return int64(v)
	case int64:
		return v
	case float64:
		return int64(v)
	case bool:
		if v {
			return 1
		}
		return 0
	case string:
		// Try to parse string as int
		// For now, just return 0
		return 0
	default:
		return 0
	}
}
