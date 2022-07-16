package bf_runner

import (
	"errors"
	"fmt"
	"io"

	"github.com/yurii-vyrovyi/brainfuck/stack"

	"golang.org/x/exp/constraints"
)

type BfInterpreter[DataType constraints.Signed] struct {
	Data    []DataType
	CmdPtr  int
	DataPtr int

	commands io.Reader
	output   io.Writer
	input    InputReader

	loopStack stack.Stack[int]
	cmdCache  CmdCache
	opMap     map[CmdType]OpFunc[DataType]

	currentLoopEnd int
}

type (
	CmdType                             byte
	CmdCache                            map[int]CmdType
	OpFunc[DataType constraints.Signed] func(bf *BfInterpreter[DataType], cmd CmdType) error
)

type InputReader interface {
	Read(string) (CmdType, error)
	Close() error
}

const (
	DefaultDataSize = 4096
)

const (
	CmdShiftRight = CmdType('>')
	CmdShiftLeft  = CmdType('<')
	CmdPlus       = CmdType('+')
	CmdMinus      = CmdType('-')
	CmdOut        = CmdType('.')
	CmdIn         = CmdType(',')
	CmdStartLoop  = CmdType('[')
	CmdEndLoop    = CmdType(']')
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

	opMap := map[CmdType]OpFunc[DataType]{
		CmdShiftRight: opShiftRight[DataType],
		CmdShiftLeft:  opShiftLeft[DataType],
		CmdPlus:       opPlus[DataType],
		CmdMinus:      opMinus[DataType],
		CmdOut:        opOut[DataType],
		CmdIn:         opIn[DataType],
		CmdStartLoop:  opStartLoop[DataType],
		CmdEndLoop:    opEndLoop[DataType],
	}

	return &BfInterpreter[DataType]{
		Data:     make([]DataType, dataSize),
		commands: commands,
		output:   output,
		input:    input,
		opMap:    opMap,
	}
}

func (bf *BfInterpreter[DataType]) WithCmd(cmd CmdType, opFunc OpFunc[DataType]) *BfInterpreter[DataType] {
	bf.opMap[cmd] = opFunc
	return bf
}

func (bf *BfInterpreter[DataType]) Run() ([]DataType, error) {

	bf.CmdPtr = 0
	bf.DataPtr = 0

	cmdBuffer := make([]byte, 1)

	for {

		var cmd CmdType
		var ok bool

		if bf.cmdCache != nil {
			cmd, ok = bf.cmdCache[bf.CmdPtr]
		}

		if !ok {
			_, err := bf.commands.Read(cmdBuffer)
			if errors.Is(err, io.EOF) {
				return bf.Data, nil
			}

			if err != nil {
				return nil, fmt.Errorf("failed to read command: %w", err)
			}

			cmd = CmdType(cmdBuffer[0])
		}

		opFunc, ok := bf.opMap[cmd]
		if ok {

			if bf.loopStack.Len() > 0 && bf.cmdCache != nil {
				bf.cmdCache[bf.CmdPtr] = cmd
			}

			if err := opFunc(bf, cmd); err != nil {
				return nil, fmt.Errorf("failed to process [#cmd: %d]: %w", bf.CmdPtr, err)
			}
		}

		bf.CmdPtr++
	}
}

func opShiftRight[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	if bf.DataPtr >= len(bf.Data)-1 {
		return fmt.Errorf("shift+ out of boundary")
	}
	bf.DataPtr++

	return nil
}

func opShiftLeft[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	if bf.DataPtr <= 0 {
		return fmt.Errorf("shift- out of boundary")
	}
	bf.DataPtr--

	return nil
}

func opPlus[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	bf.Data[bf.DataPtr] += 1

	return nil
}

func opMinus[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	bf.Data[bf.DataPtr] -= 1
	return nil
}

func opOut[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	v := bf.Data[bf.DataPtr]
	if _, err := bf.output.Write([]byte(fmt.Sprintf("%d\r\n", v))); err != nil {
		return fmt.Errorf("failed to print value: %w", err)
	}

	return nil
}

func opIn[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	rn, err := bf.input.Read(fmt.Sprintf("enter value [#cmd: %d]", bf.CmdPtr))
	if err != nil {
		return fmt.Errorf("failed to read value: %w", err)
	}

	bf.Data[bf.DataPtr] = DataType(rn)

	return nil
}

func opStartLoop[DataType constraints.Signed](bf *BfInterpreter[DataType], cmd CmdType) error {
	// what's on top of the stack?
	loop := bf.loopStack.Get()

	// is it a new loop?
	if loop == nil || *loop != bf.CmdPtr {
		bf.loopStack.Push(bf.CmdPtr)
	}

	if bf.cmdCache == nil {
		bf.cmdCache = make(CmdCache)
		bf.cmdCache[bf.CmdPtr] = cmd
	}

	// should we stay in loop?
	if bf.Data[bf.DataPtr] != 0 {
		return nil
	}

	_ = bf.loopStack.Pop()
	bf.CmdPtr = bf.currentLoopEnd // bf.CmdPtr will be incremented

	if bf.loopStack.Len() == 0 {
		bf.cmdCache = nil
	}

	return nil
}

func opEndLoop[DataType constraints.Signed](bf *BfInterpreter[DataType], _ CmdType) error {
	loop := bf.loopStack.Get()

	if loop == nil {
		return fmt.Errorf("stack is empty on closing loop [#cmd: %d]", bf.CmdPtr)
	}

	bf.currentLoopEnd = bf.CmdPtr
	bf.CmdPtr = *loop - 1 // bf.CmdPtr will be incremented

	return nil
}
