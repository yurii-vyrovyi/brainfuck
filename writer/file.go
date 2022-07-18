package writer

import (
	"fmt"
	"os"

	"golang.org/x/exp/constraints"
)

// FileWriter implements brainfuck.OutputWriter interface.
// File writer stores output to a file.
type FileWriter[DataType constraints.Signed] struct {
	f *os.File
}

// BuildFileWriter creates FileWriter instance and creates/opens a file that data will be written to.
func BuildFileWriter[DataType constraints.Signed](fileName string) (*FileWriter[DataType], error) {

	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}

	return &FileWriter[DataType]{
		f: f,
	}, nil
}

// Write writes value to the file
func (w *FileWriter[DataType]) Write(v DataType) error {
	if _, err := w.f.Write([]byte(fmt.Sprintf("%d ", v))); err != nil {
		return err
	}

	return nil
}

// Close closes the underlying file.
func (w *FileWriter[DataType]) Close() error {
	return w.f.Close()
}
