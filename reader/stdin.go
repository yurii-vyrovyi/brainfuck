package reader

import (
	"bufio"
	"fmt"
	"os"

	bf "github.com/yurii-vyrovyi/brainfuck"

	"golang.org/x/term"
)

type StdInReader struct {
	initialState *term.State
	in           *bufio.Reader
}

func BuildStdInReader() (*StdInReader, error) {
	state, err := term.MakeRaw(0)
	if err != nil {
		return nil, fmt.Errorf("failed to set stdin to raw: %w", err)
	}

	in := bufio.NewReader(os.Stdin)

	return &StdInReader{
		initialState: state,
		in:           in,
	}, nil
}

func (r *StdInReader) Close() error {
	if _, err := os.Stdin.Write([]byte("\r")); err != nil {
		return fmt.Errorf(`failed to print \r: %w`, err)
	}

	if err := term.Restore(0, r.initialState); err != nil {
		return fmt.Errorf("failed to restore terminal: %w", err)
	}

	return nil
}

func (r *StdInReader) Read(msg string) (bf.CmdType, error) {
	_, err := os.Stdin.Write([]byte(msg + ": "))
	if err != nil {
		return 0, fmt.Errorf("failed to print message: %w", err)
	}

	b, err := r.in.ReadByte()
	if err != nil {
		return 0, err
	}

	_, err = os.Stdin.Write([]byte(fmt.Sprintf("%c\r\n", b)))
	if err != nil {
		return 0, fmt.Errorf("failed to print input value: %w", err)
	}

	return bf.CmdType(b), nil
}
