package bf_runner

import (
	"context"
	"fmt"
)

type BfRunner struct {
	data []DataItem

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

func New(dataSize int) *BfRunner {

	if dataSize == 0 {
		dataSize = DefaultDataSize
	}

	return &BfRunner{
		data: make([]DataItem, dataSize),
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

		fmt.Printf("%c", commands[r.cmdPtr])

		r.cmdPtr++
	}
}

func validate(commands string) error {
	// TODO: implement

	return nil
}
