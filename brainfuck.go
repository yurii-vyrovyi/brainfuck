package brainfuck

import (
	"errors"
	"fmt"
	"io"

	"github.com/yurii-vyrovyi/brainfuck/stack"

	"golang.org/x/exp/constraints"
)

type BfInterpreter[DataType constraints.Signed] struct {
	Data    []DataType
	CmdPtr  CmdPtrType
	DataPtr DataPtrType

	commands io.Reader
	input    InputReader[DataType]
	output   OutputWriter[DataType]

	loopStack *stack.Stack[CmdPtrType]
	cmdCache  CmdCache
	opMap     map[CmdType]OpFunc[DataType]

	currentLoopEnd CmdPtrType
}

type (
	CmdType     byte
	CmdPtrType  int
	DataPtrType int

	CmdCache                            map[CmdPtrType]CmdType
	OpFunc[DataType constraints.Signed] func(bf *BfInterpreter[DataType]) error
)

type (
	InputReader[DataType constraints.Signed] interface {
		Read(string) (DataType, error)
		Close() error
	}

	OutputWriter[DataType constraints.Signed] interface {
		Write(DataType) error
		Close() error
	}
)

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
	input InputReader[DataType],
	output OutputWriter[DataType],
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
		Data:      make([]DataType, dataSize),
		output:    output,
		input:     input,
		opMap:     opMap,
		loopStack: stack.BuildStack[CmdPtrType](),
	}
}

func (bf *BfInterpreter[DataType]) WithCmd(cmd CmdType, opFunc OpFunc[DataType]) *BfInterpreter[DataType] {

	// Overloading loop operators is forbidden because these commands change interpreter state aside of
	// data and pointers (loop stack, command cache etc.).
	// Overloading these commands may lead to memory leaks and undefined behaviour that will be hard to detect.
	if cmd != CmdStartLoop && cmd != CmdEndLoop {
		bf.opMap[cmd] = opFunc
	}

	return bf
}

func (bf *BfInterpreter[DataType]) Run(commands io.Reader) ([]DataType, error) {

	bf.commands = commands
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

			// When the loop starts we're starting to cache commands
			// If the loop is nested we're keeping caching.
			if bf.cmdCache == nil {
				bf.cmdCache = make(CmdCache)
			}

			bf.cmdCache[bf.CmdPtr] = cmd

			if err := opFunc(bf); err != nil {
				return nil, fmt.Errorf("failed to process [#cmd: %d]: %w", bf.CmdPtr, err)
			}

			// Cache is not necessary anymore when we finish the topmost loop
			if bf.loopStack.Len() == 0 {
				bf.cmdCache = nil
			}
		}

		bf.CmdPtr++
	}
}

func opShiftRight[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	if bf.DataPtr >= DataPtrType(len(bf.Data)-1) {
		return fmt.Errorf("shift+ moves out of boundary")
	}
	bf.DataPtr++

	return nil
}

func opShiftLeft[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	if bf.DataPtr <= 0 {
		return fmt.Errorf("shift- moves out of boundary")
	}
	bf.DataPtr--

	return nil
}

func opPlus[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	bf.Data[bf.DataPtr] += 1

	return nil
}

func opMinus[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	bf.Data[bf.DataPtr] -= 1
	return nil
}

func opOut[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	v := bf.Data[bf.DataPtr]
	if err := bf.output.Write(v); err != nil {
		return fmt.Errorf("failed to write value: %w", err)
	}

	return nil
}

func opIn[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	rn, err := bf.input.Read(fmt.Sprintf("enter value [#cmd: %d]", bf.CmdPtr))
	if err != nil {
		return fmt.Errorf("failed to read value: %w", err)
	}

	bf.Data[bf.DataPtr] = rn

	return nil
}

func opStartLoop[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {

	// what's on top of the stack?
	loop := bf.loopStack.Get()

	// is it a new loop?
	if loop == nil || *loop != bf.CmdPtr {
		bf.loopStack.Push(bf.CmdPtr)
	}

	// should we stay in loop?
	if bf.Data[bf.DataPtr] != 0 {
		return nil
	}

	_ = bf.loopStack.Pop()
	bf.CmdPtr = bf.currentLoopEnd // bf.CmdPtr will be incremented

	return nil
}

func opEndLoop[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	loop := bf.loopStack.Get()

	if loop == nil {
		return fmt.Errorf("stack is empty on closing loop [#cmd: %d]", bf.CmdPtr)
	}

	bf.currentLoopEnd = bf.CmdPtr
	bf.CmdPtr = *loop - 1 // bf.CmdPtr will be incremented

	return nil
}
