package graphics

import (
	"testing"
)

func TestNewCommandQueue(t *testing.T) {
	queue := NewCommandQueue()
	if queue == nil {
		t.Fatal("NewCommandQueue returned nil")
	}
	if queue.Len() != 0 {
		t.Errorf("Expected empty queue, got length %d", queue.Len())
	}
}

func TestCommandQueuePushAndPopAll(t *testing.T) {
	queue := NewCommandQueue()

	// Push some commands
	cmd1 := Command{Type: CmdMovePic, Args: []any{1, 2, 3}}
	cmd2 := Command{Type: CmdOpenWin, Args: []any{"test"}}
	cmd3 := Command{Type: CmdDrawLine, Args: []any{10, 20, 30, 40}}

	queue.Push(cmd1)
	queue.Push(cmd2)
	queue.Push(cmd3)

	if queue.Len() != 3 {
		t.Errorf("Expected queue length 3, got %d", queue.Len())
	}

	// PopAll should return all commands in FIFO order
	commands := queue.PopAll()
	if len(commands) != 3 {
		t.Fatalf("Expected 3 commands, got %d", len(commands))
	}

	if commands[0].Type != CmdMovePic {
		t.Errorf("Expected first command type %v, got %v", CmdMovePic, commands[0].Type)
	}
	if commands[1].Type != CmdOpenWin {
		t.Errorf("Expected second command type %v, got %v", CmdOpenWin, commands[1].Type)
	}
	if commands[2].Type != CmdDrawLine {
		t.Errorf("Expected third command type %v, got %v", CmdDrawLine, commands[2].Type)
	}

	// Queue should be empty after PopAll
	if queue.Len() != 0 {
		t.Errorf("Expected empty queue after PopAll, got length %d", queue.Len())
	}
}

func TestCommandQueuePopAllEmpty(t *testing.T) {
	queue := NewCommandQueue()

	// PopAll on empty queue should return nil
	commands := queue.PopAll()
	if commands != nil {
		t.Errorf("Expected nil from PopAll on empty queue, got %v", commands)
	}
}

func TestCommandQueueMultiplePopAll(t *testing.T) {
	queue := NewCommandQueue()

	// Push commands
	for i := 0; i < 5; i++ {
		queue.Push(Command{Type: CmdMovePic, Args: []any{i}})
	}

	// First PopAll
	commands1 := queue.PopAll()
	if len(commands1) != 5 {
		t.Errorf("Expected 5 commands in first PopAll, got %d", len(commands1))
	}

	// Second PopAll should return nil
	commands2 := queue.PopAll()
	if commands2 != nil {
		t.Errorf("Expected nil from second PopAll, got %v", commands2)
	}

	// Push more commands
	for i := 0; i < 3; i++ {
		queue.Push(Command{Type: CmdOpenWin, Args: []any{i}})
	}

	// Third PopAll should return new commands
	commands3 := queue.PopAll()
	if len(commands3) != 3 {
		t.Errorf("Expected 3 commands in third PopAll, got %d", len(commands3))
	}
}

func TestCommandTypes(t *testing.T) {
	// Verify all command types are defined
	types := []CommandType{
		CmdMovePic,
		CmdMoveSPic,
		CmdTransPic,
		CmdReversePic,
		CmdOpenWin,
		CmdMoveWin,
		CmdCloseWin,
		CmdPutCast,
		CmdMoveCast,
		CmdDelCast,
		CmdTextWrite,
		CmdDrawLine,
		CmdDrawRect,
		CmdFillRect,
		CmdDrawCircle,
	}

	// Check that types are sequential
	for i, cmdType := range types {
		if int(cmdType) != i {
			t.Errorf("Expected command type %d to have value %d, got %d", i, i, int(cmdType))
		}
	}
}
