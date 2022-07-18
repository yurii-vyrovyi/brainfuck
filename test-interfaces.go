package brainfuck

//go:generate mockgen -source test-interfaces.go -destination mock_brainfuck.go -package brainfuck

// These interfaces are necessary to generate InputReader and OutputWrite mocks.
// While original interfaces InputReader and OutputWrite use generics TestXXX ones use TestDataType as a data type.
// This type is used in all tests.
type (
	TestDataType int32

	TestInputReader interface {
		Read(string) (TestDataType, error)
		Close() error
	}

	TestOutputWriter interface {
		Write(TestDataType) error
		Close() error
	}
)
