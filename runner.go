package bf_runner

import (
	"errors"
	"fmt"
	"io"

	"github.com/yurii-vyrovyi/brainfuck/stack"

	"golang.org/x/exp/constraints"
)

type BfInterpreter[DataType constraints.Signed] struct {
	data     []DataType
	commands io.Reader
	output   io.Writer
	input    InputReader

	cmdPtr    int
	dataPtr   int
	loopStack stack.Stack[int]
	cmdCache  CmdCache

	currentLoopEnd int
}

type CmdCache map[int]byte

type InputReader interface {
	Read(string) (byte, error)
	Close() error
}

type Loop struct {
	start int
	end   int
}

const (
	DefaultDataSize = 4096
)

const (
	CmdShiftRight = byte('>')
	CmdShiftLeft  = byte('<')
	CmdPlus       = byte('+')
	CmdMinus      = byte('-')
	CmdOut        = byte('.')
	CmdIn         = byte(',')
	CmdStartLoop  = byte('[')
	CmdEndLoop    = byte(']')
)

func New[DataType constraints.Signed](
	dataSize int,
	commands io.Reader,
	output io.Writer,
	input InputReader,
) *BfInterpreter[DataType] {

	if dataSize == 0 {
		dataSize = DefaultDataSize
	}

	return &BfInterpreter[DataType]{
		data:     make([]DataType, dataSize),
		commands: commands,
		output:   output,
		input:    input,
	}
}

func (bf *BfInterpreter[DataType]) Run() ([]DataType, error) {

	bf.cmdPtr = 0
	bf.dataPtr = 0

	cmdBuffer := make([]byte, 1)

	for {

		var cmd byte
		var ok bool

		if bf.cmdCache != nil {
			cmd, ok = bf.cmdCache[bf.cmdPtr]
		}

		if !ok {
			_, err := bf.commands.Read(cmdBuffer)
			if errors.Is(err, io.EOF) {
				return bf.data, nil
			}

			if err != nil {
				return nil, fmt.Errorf("failed to read command: %w", err)
			}

			cmd = cmdBuffer[0]
		}

		if err := bf.execCmd(cmd); err != nil {
			return nil, fmt.Errorf("failed to process [#cmd: %d]: %w", bf.cmdPtr, err)
		}

		bf.cmdPtr++
	}
}

func (bf *BfInterpreter[DataType]) execCmd(cmd byte) error {

	if bf.loopStack.Len() > 0 && bf.cmdCache != nil {
		bf.cmdCache[bf.cmdPtr] = cmd
	}

	switch cmd {
	case CmdShiftRight:
		if bf.dataPtr >= len(bf.data)-1 {
			return fmt.Errorf("shift+ out of boundary")
		}
		bf.dataPtr++

	case CmdShiftLeft:
		if bf.dataPtr <= 0 {
			return fmt.Errorf("shift- out of boundary")
		}
		bf.dataPtr--

	case CmdPlus:
		bf.data[bf.dataPtr] += 1

	case CmdMinus:
		bf.data[bf.dataPtr] -= 1

	case CmdOut:
		v := bf.data[bf.dataPtr]
		if _, err := bf.output.Write([]byte(fmt.Sprintf("%d\r\n", v))); err != nil {
			return fmt.Errorf("failed to print value: %w", err)
		}

	case CmdIn:
		rn, err := bf.input.Read(fmt.Sprintf("enter value [#cmd: %d]", bf.cmdPtr))
		if err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}

		bf.data[bf.dataPtr] = DataType(rn)

	case CmdStartLoop:

		// what's on top of the stack?
		loop := bf.loopStack.Get()

		// is it a new loop?
		if loop == nil || *loop != bf.cmdPtr {
			bf.loopStack.Push(bf.cmdPtr)
		}

		if bf.cmdCache == nil {
			bf.cmdCache = make(CmdCache)
			bf.cmdCache[bf.cmdPtr] = cmd
		}

		// should we stay in loop?
		if bf.data[bf.dataPtr] != 0 {
			break
		}

		_ = bf.loopStack.Pop()
		bf.cmdPtr = bf.currentLoopEnd // bf.cmdPtr will be incremented

		if bf.loopStack.Len() == 0 {
			bf.cmdCache = nil
		}

	case CmdEndLoop:
		loop := bf.loopStack.Get()

		if loop == nil {
			return fmt.Errorf("stack is empty on closing loop [#cmd: %d]", bf.cmdPtr)
		}

		bf.currentLoopEnd = bf.cmdPtr
		bf.cmdPtr = *loop - 1 // bf.cmdPtr will be incremented
	}

	return nil
}
