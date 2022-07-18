package writer

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

// StdOutWriter implements brainfuck.OutputWriter interface.
// It's complimentary with reader.StdInReader and adds '\r' to the output.
type StdOutWriter[DataType constraints.Signed] struct{}

// Creates StdOutWriter instance.
func BuildStdOutWriter[DataType constraints.Signed]() *StdOutWriter[DataType] {
	return &StdOutWriter[DataType]{}
}

// Write writes a value to StdOut
func (w *StdOutWriter[DataType]) Write(v DataType) error {
	fmt.Println(fmt.Sprintf("%d\r", v))
	return nil
}

func (w *StdOutWriter[DataType]) Close() error {
	return nil
}
