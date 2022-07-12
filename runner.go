package bf_runner

import (
	"context"
	"fmt"
	"io"
)

type BfRunner struct {
	data   []DataItem
	output io.Writer
	input  io.Reader

	cmdPtr  int
	dataPtr int
}

type DataItem int

// type Error string
//
// func (e Error) Error() string {
// 	return string(e)
// }
//
// const ErrInterrupted = Error("interrupted")

const DefaultDataSize = 4096

func New(dataSize int, w io.Writer, r io.Reader) *BfRunner {

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
		case '>', '<', '+', '-', '.', ',', '[', ']':

		default:
			return fmt.Errorf(`unknown cmd [%d]: '%c'`, iCmd, cmd)
		}
	}

	return nil
}

func (r *BfRunner) processCmd(cmd byte) error {

	switch cmd {
	case '>':
		if r.dataPtr >= len(r.data)-1 {
			return fmt.Errorf("shift+ out of boundary")
		}
		r.dataPtr++

	case '<':
		if r.dataPtr <= 0 {
			return fmt.Errorf("shift- out of boundary")
		}
		r.dataPtr--

	case '+':
		r.data[r.dataPtr] += 1

	case '-':
		r.data[r.dataPtr] -= 1

	case '.':
		v := r.data[r.dataPtr]
		if _, err := r.output.Write([]byte(fmt.Sprintf("%d", v))); err != nil {
			return fmt.Errorf("failed to print value: %w", err)
		}

	case ',':
		b := make([]byte, 3)
		if _, err := r.input.Read(b); err != nil {
			return fmt.Errorf("failed to read value: %w", err)
		}

		// TODO: unicode char to unm value

	case '[':

	case ']':

	default:
		return fmt.Errorf(`unknown cmd: '%c'`, cmd)
	}

	return nil
}
