package writer

import (
	"fmt"
	"os"

	"golang.org/x/exp/constraints"
)

type FileWriter[DataType constraints.Signed] struct {
	f *os.File
}

func BuildFileWriter[DataType constraints.Signed](fileName string) (*FileWriter[DataType], error) {

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &FileWriter[DataType]{
		f: f,
	}, nil
}

func (w *FileWriter[DataType]) Write(v DataType) error {
	if _, err := w.f.Write([]byte(fmt.Sprintf("%d ", v))); err != nil {
		return err
	}

	return nil
}

func (w *FileWriter[DataType]) Close() error {
	return w.f.Close()
}
