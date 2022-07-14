package bf_runner

import (
	"context"
	"errors"
	"fmt"
	"github.com/yurii-vyrovyi/brainfuck/stack"
	"io"
)

type BfRunner struct {
	data   []DataItem
	output io.Writer
	input  InputReader

	cmdPtr    int
	dataPtr   int
	commands  string
	loopStack stack.Stack[Loop]
}

type DataItem int

type InputReader interface {
	Read(string) (rune, error)
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
	CmdShiftRight = '>'
	CmdShiftLeft  = '<'
	CmdPlus       = '+'
	CmdMinus      = '-'
	CmdOut        = '.'
	CmdIn         = ','
	CmdStartLoop  = '['
	CmdEndLoop    = ']'
)

func New(dataSize int, w io.Writer, r InputReader) *BfRunner {

	if dataSize == 0 {
		dataSize = DefaultDataSize
	}

	return &BfRunner{
		data:   make([]DataItem, dataSize),
		output: w,
		input:  r,
	}
}

func (r *BfRunner) Run(ctx context.Context, commands string) ([]DataItem, error) {

	// if err := validate(commands); err != nil {
	// 	return nil, fmt.Errorf("bad commands: %w", err)
	// }

	r.cmdPtr = 0
	r.dataPtr = 0
	r.commands = commands

	// DEBUG
	fmt.Println()
	defer fmt.Println()

	for {
		if r.cmdPtr >= len(commands) {
			return r.data, nil
		}

		select {
		case <-ctx.Done():
			return nil, nil
		default:
		}

		if err := r.processCmd(r.cmdPtr); err != nil {
			return nil, fmt.Errorf("failed to process [#cmd: %d]: %w", r.cmdPtr, err)
		}

	}
}

func validate(commands string) error {

	loopCounter := 0

	for iCmd, cmd := range commands {
		switch cmd {
		case CmdShiftRight, CmdShiftLeft, CmdPlus, CmdMinus, CmdOut, CmdIn:

		case CmdStartLoop:
			loopCounter++

		case CmdEndLoop:
			loopCounter--
			if loopCounter < 0 {
				return fmt.Errorf(`number ']' is greater then number of '[' [#cmd: %d]`, iCmd)
			}

		default:
			return fmt.Errorf(`unknown [#cmd: %d]: '%c'`, iCmd, cmd)
		}
	}

	if loopCounter != 0 {
		return errors.New(`number '[' is not equal to number of ']'`)
	}

	return nil
}

func (r *BfRunner) processCmd(iCmd int) error {

	cmd := r.commands[iCmd]

	switch cmd {
	case CmdShiftRight:
		if r.dataPtr >= len(r.data)-1 {
			return fmt.Errorf("shift+ out of boundary")
		}
		r.dataPtr++

	case CmdShiftLeft:
		if r.dataPtr <= 0 {
			return fmt.Errorf("shift- out of boundary")
		}
		r.dataPtr--

	case CmdPlus:
		r.data[r.dataPtr] += 1

	case CmdMinus:
		r.data[r.dataPtr] -= 1

	case CmdOut:
		v := r.data[r.dataPtr]
		if _, err := r.output.Write([]byte(fmt.Sprintf("%d\r\n", v))); err != nil {
			return fmt.Errorf("failed to print value: %w", err)
		}

	case CmdIn:
		rn, err := r.input.Read(fmt.Sprintf("enter value [#cmd: %d]", r.cmdPtr))
		if err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}

		r.data[r.dataPtr] = DataItem(rn)

	case CmdStartLoop:

		// what's on top of the stack?
		loop := r.loopStack.Get()

		// is it a new loop?
		if loop == nil || loop.start != r.cmdPtr {
			loopEnd := findLoopEnd(r.cmdPtr, r.commands)
			if loopEnd == -1 {
				return fmt.Errorf("non-closed loop [#cmd: %d]", r.cmdPtr)
			}

			loop = &Loop{
				start: r.cmdPtr,
				end:   loopEnd,
			}

			r.loopStack.Push(*loop)
		}

		// should we exit this loop already?
		if r.data[r.dataPtr] == 0 {
			_ = r.loopStack.Pop()
			r.cmdPtr = loop.end // r.cmdPtr will be incremented at the end of func
		}

	case CmdEndLoop:
		loop := r.loopStack.Get()

		if loop == nil {
			return fmt.Errorf("stack is empty on closing loop [#cmd: %d]", r.cmdPtr)
		}

		r.cmdPtr = loop.start - 1 // r.cmdPtr will be incremented at the end of func
	}

	r.cmdPtr++

	return nil
}

func findLoopEnd(loopStart int, commands string) int {

	loopCnt := 0

	for i := loopStart; i < len(commands); i++ {

		switch commands[i] {
		case CmdStartLoop:
			loopCnt++
		case CmdEndLoop:
			loopCnt--
		}

		if loopCnt == 0 {
			return i
		}
	}

	return -1
}
