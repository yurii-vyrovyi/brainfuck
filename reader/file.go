package reader

import (
	"bufio"
	"golang.org/x/exp/constraints"
	"os"
)

type FileReader[DataType constraints.Signed] struct {
	f  *os.File
	in *bufio.Reader
}

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

func (r *FileReader[DataType]) Close() error {
	return r.f.Close()
}

func (r *FileReader[DataType]) Read(_ string) (DataType, error) {
	b, err := r.in.ReadByte()
	if err != nil {
		return 0, err
	}

	return DataType(b), nil
}
