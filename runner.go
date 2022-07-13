package bf_runner

import (
	"context"
	"fmt"
	"io"
)

type BfRunner struct {
	data   []DataItem
	output io.Writer
	input  InputReader

	cmdPtr  int
	dataPtr int
}

type DataItem int

type InputReader interface {
	Read(string) (rune, error)
	Close() error
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

func (r *BfRunner) Run(ctx context.Context, commands string) error {

	if err := validate(commands); err != nil {
		return fmt.Errorf("bad commands: %w", err)
	}

	r.cmdPtr = 0
	r.dataPtr = 0

	// DEBUG
	fmt.Println()
	defer fmt.Println()

	for {
		if r.cmdPtr >= len(commands) {
			return nil
		}

		select {
		case <-ctx.Done():
			return nil
		default:
		}

		if err := r.processCmd(commands[r.cmdPtr]); err != nil {
			return fmt.Errorf("failed to process cmd [# %d]: %w", r.cmdPtr, err)
		}

		r.cmdPtr++
	}
}

func validate(commands string) error {

	for iCmd, cmd := range commands {
		switch cmd {
		case CmdShiftRight, CmdShiftLeft, CmdPlus, CmdMinus, CmdOut, CmdIn, CmdStartLoop, CmdEndLoop:

		default:
			return fmt.Errorf(`unknown cmd [%d]: '%c'`, iCmd, cmd)
		}
	}

	return nil
}

func (r *BfRunner) processCmd(cmd byte) error {

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

	case CmdEndLoop:

	default:
		return fmt.Errorf(`unknown cmd: '%c'`, cmd)
	}

	return nil
}
