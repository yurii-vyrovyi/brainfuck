package writer

import (
	"fmt"

	"golang.org/x/exp/constraints"
)

type StdOutWriter[DataType constraints.Signed] struct {
}

func BuildStdOutWriter[DataType constraints.Signed]() *StdOutWriter[DataType] {
	return &StdOutWriter[DataType]{}
}

func (w *StdOutWriter[DataType]) Write(v DataType) error {
	fmt.Println(fmt.Sprintf("%d\r", v))
	return nil
}

func (w *StdOutWriter[DataType]) Close() error {
	return nil
}
