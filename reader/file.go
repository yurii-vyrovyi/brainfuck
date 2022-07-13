package reader

import (
	"bufio"
	"os"
)

type FileReader struct {
	f  *os.File
	in *bufio.Reader
}

func BuildFileReader(fileName string) (*FileReader, error) {
	f, err := os.Open(fileName)
	if err != nil {
		return nil, err
	}

	in := bufio.NewReader(f)

	return &FileReader{
		f:  f,
		in: in,
	}, nil
}

func (r *FileReader) Close() error {
	return r.f.Close()
}

func (r *FileReader) Read(_ string) (rune, error) {
	c, _, err := r.in.ReadRune()
	if err != nil {
		return 0, err
	}

	return c, nil
}
