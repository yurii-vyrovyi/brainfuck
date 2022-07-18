// Package brainfuck is brainfuck language interpreter.
// You can read more about the language itself here https://en.wikipedia.org/wiki/Brainfuck
//
// Despite the core idea of Brainfuck language its performance and shortness of interpreter were not the issue.
// This interpreter is dedicated rather to show Golang approaches.
//
// Here are points that explain an interpreter implementation:
//
// 1. Memory data type.
// We're using generics to define memory data type. This gives user freedom to pick any type that fits the best
//
// 2. Commands type.
// CmdType is byte. While we're not planning to give user a possibility to extend the language too much byte type is enough.
//
// 3. EOL
// Interpreter ignores unknown commands, hence '\r' and '\n' don't affect code execution. While user can write some comments.
// Again, while there's no some specific requirement it seems to be reasonable to keep implementation simple.
//
// 4. Commands and Data pointers
// CmdPtrType and DataPtrType are int. It could be changed if we have some requirement for the size of memory or
// an amount of Brainfuck commands that we're assuming to process.
//
// 5. Commands caching in loops
// One of requirements for this interpreter was reading commands in the flight and avoiding pre-loading commands.
// During the loop we're repeating the commands that were already read. It seems reasonable to keep it and to re-use it
// while iterating a loop.
// A map was picked as a cache container. It makes easier to manage it:
//
// - no need to extend memory moving along the loop body
//
// - no need to deal with commands order when caching it
//
// In case of performance requirements it may be implemented with slice that will make reading from cache faster.
//
// 6. Commands overloading
// Custom commands handlers have access to public members – Data, DataPtr, CmdPtr, Input and Output.
// So custom command handler can read and write data to memory, setup where other commands will read/write,
// manage what the next command will be and read/write data from/to user.
//
package brainfuck

import (
	"errors"
	"fmt"
	"io"

	"github.com/yurii-vyrovyi/brainfuck/stack"

	"golang.org/x/exp/constraints"
)

type BfInterpreter[DataType constraints.Signed] struct {

	// Data is a memory for brainfuck algorithm
	Data []DataType

	// CmdPtr is a current command pointer
	CmdPtr CmdPtrType

	// DataPtr is a data pointer – the cell that brainfuck commands affects
	DataPtr DataPtrType

	// Input is a reader for In brainfuck command
	Input InputReader[DataType]

	// Output is a writer where out brainfuck command writes to
	Output OutputWriter[DataType]

	// loopStack stores the addresses of a loops beginnings
	loopStack *stack.Stack[CmdPtrType]

	// cmdCache is a commands cache that is used while interpreter runs loops
	cmdCache CmdCache

	// opMap stores correspondence between commands and handlers
	opMap map[CmdType]OpFunc[DataType]

	// currentLoopEnd stores the command address of the end of the current loop
	currentLoopEnd CmdPtrType
}

type (
	// CmdType is a brainfuck commands type
	CmdType byte

	// CmdPtrType is a type of command pointer
	CmdPtrType int

	// DataPtrType is a type of data pointer
	DataPtrType int

	// CmdCache is a commands cache
	CmdCache map[CmdPtrType]CmdType

	// OpFunc is type for brainfuck commands handlers.
	OpFunc[DataType constraints.Signed] func(bf *BfInterpreter[DataType]) error
)

type (

	// InputReader is an interface of a reader that will be used for Input ( ',' In command).
	// Read() has a string parameter – a hint that might be used in case when user enters values.
	InputReader[DataType constraints.Signed] interface {
		Read(string) (DataType, error)
		Close() error
	}

	// OutputWriter is an interface that is used for Output ('.' Out command).
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

// New creates an instance of brainfuck interpreter
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
		Output:    output,
		Input:     input,
		opMap:     opMap,
		loopStack: stack.BuildStack[CmdPtrType](),
	}
}

// WithCmd allows to add or overload commands.
// Loop start and end commands ('[' and ']') can't be overloaded.
// This restriction is done because these commands change internal interpreter state aside of explicit
// Data, CmdPtr and DataPtr. Opening access to the rest of variables can make interpreter more vulnerable and behaviour undefined.
// If you need to implement other loop logics, pleases create a new commands (i.e. '{' and '}').
//
// cmd CmdType – is a command that may be used in the code
//
// opFunc OpFunc[DataType] – is a handler for the command
func (bf *BfInterpreter[DataType]) WithCmd(cmd CmdType, opFunc OpFunc[DataType]) *BfInterpreter[DataType] {

	// Overloading loop operators is forbidden because these commands change interpreter state aside of
	// data and pointers (loop stack, command cache etc.).
	// Overloading these commands may lead to memory leaks and undefined behaviour that will be hard to detect.
	if cmd != CmdStartLoop && cmd != CmdEndLoop {
		bf.opMap[cmd] = opFunc
	}

	return bf
}

// Run starts interpreting brainfuck code. It reads commands one by one from commands reader.
func (bf *BfInterpreter[DataType]) Run(commands io.Reader) ([]DataType, error) {

	bf.CmdPtr = 0
	bf.DataPtr = 0

	for {

		var cmd CmdType
		var ok bool

		// trying to read a command from cache
		if bf.cmdCache != nil {
			cmd, ok = bf.cmdCache[bf.CmdPtr]
		}

		// no cached command, let's get a new one from the reader
		if !ok {
			cmdBuffer := make([]byte, 1)

			_, err := commands.Read(cmdBuffer)
			if errors.Is(err, io.EOF) {
				return bf.Data, nil
			}

			if err != nil {
				return nil, fmt.Errorf("failed to read command: %w", err)
			}

			cmd = CmdType(cmdBuffer[0])
		}

		// ignoring commands without correspondent handler
		opFunc, ok := bf.opMap[cmd]
		if ok {

			// When the loop starts we're starting to cache commands
			if cmd == CmdStartLoop && bf.cmdCache == nil {
				bf.cmdCache = make(CmdCache)
				bf.cmdCache[bf.CmdPtr] = cmd
			}

			// If we're in loop we're caching every command
			if bf.loopStack.Len() > 0 && bf.cmdCache != nil {
				bf.cmdCache[bf.CmdPtr] = cmd
			}

			// processing command
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

// opShiftRight is default handler for ShiftRight ('>') command
func opShiftRight[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	if bf.DataPtr >= DataPtrType(len(bf.Data)-1) {
		return fmt.Errorf("shift+ moves out of boundary")
	}
	bf.DataPtr++

	return nil
}

// opShiftLeft is default handler for ShiftLeft ('<') command
func opShiftLeft[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	if bf.DataPtr <= 0 {
		return fmt.Errorf("shift- moves out of boundary")
	}
	bf.DataPtr--

	return nil
}

// opPlus is default handler for Plus ('+') command
func opPlus[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	bf.Data[bf.DataPtr] += 1

	return nil
}

// opMinus is default handler for Minus ('-') command
func opMinus[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	bf.Data[bf.DataPtr] -= 1
	return nil
}

// opOut is default handler for Out ('.') command
func opOut[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	v := bf.Data[bf.DataPtr]
	if err := bf.Output.Write(v); err != nil {
		return fmt.Errorf("failed to write value: %w", err)
	}

	return nil
}

// opIn is default handler for In (',') command
func opIn[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	rn, err := bf.Input.Read(fmt.Sprintf("enter value [#cmd: %d]", bf.CmdPtr))
	if err != nil {
		return fmt.Errorf("failed to read value: %w", err)
	}

	bf.Data[bf.DataPtr] = rn

	return nil
}

// opStartLoop is a handler for StartLoop ('[') command
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

// opEndLoop is a handler for EndLoop (']') command
func opEndLoop[DataType constraints.Signed](bf *BfInterpreter[DataType]) error {
	loop := bf.loopStack.Get()

	if loop == nil {
		return fmt.Errorf("stack is empty on closing loop [#cmd: %d]", bf.CmdPtr)
	}

	bf.currentLoopEnd = bf.CmdPtr
	bf.CmdPtr = *loop - 1 // bf.CmdPtr will be incremented

	return nil
}
