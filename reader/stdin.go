package reader

import (
	"bufio"
	"fmt"
	"os"

	"golang.org/x/exp/constraints"
	"golang.org/x/term"
)

type StdInReader[DataType constraints.Signed] struct {
	initialState *term.State
	in           *bufio.Reader
}

func BuildStdInReader[DataType constraints.Signed]() (*StdInReader[DataType], error) {
	state, err := term.MakeRaw(0)
	if err != nil {
		return nil, fmt.Errorf("failed to set stdin to raw: %w", err)
	}

	in := bufio.NewReader(os.Stdin)

	return &StdInReader[DataType]{
		initialState: state,
		in:           in,
	}, nil
}

func (r *StdInReader[DataType]) Close() error {
	if _, err := os.Stdin.Write([]byte{'\r'}); err != nil {
		return fmt.Errorf(`failed to print \r: %w`, err)
	}

	if err := term.Restore(0, r.initialState); err != nil {
		return fmt.Errorf("failed to restore terminal: %w", err)
	}

	return nil
}

func (r *StdInReader[DataType]) Read(msg string) (DataType, error) {
	_, err := os.Stdin.Write([]byte(msg + ": "))
	if err != nil {
		return 0, fmt.Errorf("failed to print message: %w", err)
	}

	b, _, err := r.in.ReadRune()
	if err != nil {
		return 0, err
	}

	_, err = os.Stdin.Write([]byte(fmt.Sprintf("%c\r\n", b)))
	if err != nil {
		return 0, fmt.Errorf("failed to print input value: %w", err)
	}

	return DataType(b), nil
}
