package reader

import (
	"bufio"
	"golang.org/x/exp/constraints"
	"os"
)

// FileReader implements brainfuck.InputReader.
type FileReader[DataType constraints.Signed] struct {
	f  *os.File
	in *bufio.Reader
}

// BuildFileReader creates and instance of FileReader and opens the file that it will read from.
func BuildFileReader[DataType constraints.Signed](fileName string) (*FileReader[DataType], error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	in := bufio.NewReader(f)

	return &FileReader[DataType]{
		f:  f,
		in: in,
	}, nil
}

// Close closes the file that FileReader read data from
func (r *FileReader[DataType]) Close() error {
	return r.f.Close()
}

// Read reads data from file byte per byte
func (r *FileReader[DataType]) Read(_ string) (DataType, error) {
	b, err := r.in.ReadByte()
	if err != nil {
		return 0, err
	}

	return DataType(b), nil
}
