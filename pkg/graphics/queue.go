package graphics

import (
	"sync"
)

// CommandType は描画コマンドの種類を表す
type CommandType int

const (
	CmdMovePic CommandType = iota
	CmdMoveSPic
	CmdTransPic
	CmdReversePic
	CmdOpenWin
	CmdMoveWin
	CmdCloseWin
	CmdPutCast
	CmdMoveCast
	CmdDelCast
	CmdTextWrite
	CmdDrawLine
	CmdDrawRect
	CmdFillRect
	CmdDrawCircle
)

// Command は描画コマンドを表す
type Command struct {
	Type CommandType
	Args []any
}

// CommandQueue はスレッドセーフな描画コマンドキュー
type CommandQueue struct {
	commands []Command
	mu       sync.Mutex
}

// NewCommandQueue は新しいCommandQueueを作成する
func NewCommandQueue() *CommandQueue {
	return &CommandQueue{
		commands: make([]Command, 0),
	}
}

// Push はコマンドをキューに追加する（スレッドセーフ）
func (cq *CommandQueue) Push(cmd Command) {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	cq.commands = append(cq.commands, cmd)
}

// PopAll はキュー内のすべてのコマンドを取り出して返す
// キューは空になる
func (cq *CommandQueue) PopAll() []Command {
	cq.mu.Lock()
	defer cq.mu.Unlock()

	if len(cq.commands) == 0 {
		return nil
	}

	result := make([]Command, len(cq.commands))
	copy(result, cq.commands)
	cq.commands = cq.commands[:0] // スライスをクリア

	return result
}

// Len はキュー内のコマンド数を返す
func (cq *CommandQueue) Len() int {
	cq.mu.Lock()
	defer cq.mu.Unlock()
	return len(cq.commands)
}
