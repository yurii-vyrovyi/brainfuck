package brainfuck

import (
	"bytes"
	"errors"
	"testing"

	"github.com/yurii-vyrovyi/brainfuck/stack"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -source brainfuck_test.go -destination mock_brainfuck.go -package brainfuck

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

func TestBfInterpreter_Operations(t *testing.T) {
	t.Parallel()

	type Test struct {
		opFunc OpFunc[TestDataType]

		srcData     []TestDataType
		srcDataPtr  DataPtrType
		srcInput    []TestDataType
		srcInError  error
		srcOutError error

		expErr     bool
		expData    []TestDataType
		extDataPrt DataPtrType
		expOutput  []TestDataType
	}

	tests := map[string]Test{
		"ShiftRight": {
			opFunc: opShiftRight[TestDataType],

			srcData:     make([]TestDataType, 20),
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    make([]TestDataType, 20),
			extDataPrt: 1,
			expOutput:  nil,
		},

		"ShiftLeft": {
			opFunc: opShiftLeft[TestDataType],

			srcData:     make([]TestDataType, 20),
			srcDataPtr:  10,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    make([]TestDataType, 20),
			extDataPrt: 9,
			expOutput:  nil,
		},

		"Plus": {
			opFunc: opPlus[TestDataType],

			srcData:     []TestDataType{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []TestDataType{1, 0, 0, 0, 0},
			extDataPrt: 0,
			expOutput:  nil,
		},

		"Minus": {
			opFunc: opMinus[TestDataType],

			srcData:     []TestDataType{0, 3, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []TestDataType{0, 2, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"Out OK": {
			opFunc: opOut[TestDataType],

			srcData:     []TestDataType{0, 15, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []TestDataType{0, 15, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  []TestDataType{15},
		},

		"Out Error": {
			opFunc: opOut[TestDataType],

			srcData:     []TestDataType{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: errors.New("output error"),

			expErr:     true,
			expData:    []TestDataType{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"In OK": {
			opFunc: opIn[TestDataType],

			srcData:     []TestDataType{0, 0, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    []TestDataType{12},
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []TestDataType{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"Input Error": {
			opFunc: opIn[TestDataType],

			srcData:     []TestDataType{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  errors.New("input error"),
			srcOutError: nil,

			expErr:     true,
			expData:    []TestDataType{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"OpError": {
			opFunc: opShiftLeft[TestDataType],

			srcData:     []TestDataType{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     true,
			expData:    []TestDataType{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"CustomOp": {
			opFunc: func(bf *BfInterpreter[TestDataType]) error {
				bf.Data[bf.DataPtr] = bf.Data[bf.DataPtr] * bf.Data[bf.DataPtr]
				return nil
			},

			srcData:     []TestDataType{0, 3, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []TestDataType{0, 9, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			cntInput := 0

			mockCtrl := gomock.NewController(t)

			mockInputReader := NewMockTestInputReader(mockCtrl)
			mockInputReader.EXPECT().Close().AnyTimes().Return(nil)
			mockInputReader.EXPECT().Read(gomock.Any()).AnyTimes().
				DoAndReturn(func(string) (TestDataType, error) {

					if test.srcInError != nil {
						return 0, test.srcInError
					}

					if cntInput >= len(test.srcInput) {
						return 0, errors.New("no more input")
					}
					v := test.srcInput[cntInput]
					cntInput++

					return v, nil
				})

			var output []TestDataType

			mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
			mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
			mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().
				DoAndReturn(func(v TestDataType) error {

					if test.srcOutError != nil {
						return test.srcOutError
					}

					output = append(output, v)
					return nil
				})

			bf := New[TestDataType](len(test.srcData), mockInputReader, mockOutputWriter)

			bf.Data = test.srcData
			bf.DataPtr = test.srcDataPtr

			err := test.opFunc(bf)

			if test.expErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.True(t, cmp.Equal(test.expData, bf.Data))
			require.Equal(t, test.extDataPrt, bf.DataPtr)
			require.True(t, cmp.Equal(test.expOutput, output))
		})
	}
}

func TestBfInterpreter_StartLoop(t *testing.T) {
	t.Parallel()

	type Test struct {
		srcData      []TestDataType
		srcDataPtr   DataPtrType
		srcCmdPtr    CmdPtrType
		srcLoopStack *stack.Stack[CmdPtrType]
		srLoopEnd    CmdPtrType

		expData      []TestDataType
		extDataPrt   DataPtrType
		expCmdPtr    CmdPtrType
		expLoopStack *stack.Stack[CmdPtrType]
	}

	tests := map[string]Test{
		"Start a new loop": {
			srcData:    []TestDataType{0, 1, 0, 0},
			srcDataPtr: 1,
			srcCmdPtr:  2,
			srLoopEnd:  4,

			expData:      []TestDataType{0, 1, 0, 0},
			extDataPrt:   1,
			expCmdPtr:    2,
			expLoopStack: stack.BuildStack[CmdPtrType](2),
		},

		"Start nesting loop": {
			srcData:      []TestDataType{0, 1, 0, 0},
			srcDataPtr:   1,
			srcCmdPtr:    2,
			srcLoopStack: stack.BuildStack[CmdPtrType](1),
			srLoopEnd:    4,

			expData:      []TestDataType{0, 1, 0, 0},
			extDataPrt:   1,
			expCmdPtr:    2,
			expLoopStack: stack.BuildStack[CmdPtrType](2, 1),
		},

		"Exit nesting loop": {
			srcData:      []TestDataType{0, 1, 0, 0},
			srcDataPtr:   2,
			srcCmdPtr:    2,
			srcLoopStack: stack.BuildStack[CmdPtrType](2, 1),
			srLoopEnd:    4,

			expData:      []TestDataType{0, 1, 0, 0},
			extDataPrt:   2,
			expCmdPtr:    4, // cmdPtr is moved to the END of the loop
			expLoopStack: stack.BuildStack[CmdPtrType](1),
		},

		"Exit topmost loop": {
			srcData:      []TestDataType{0, 1, 0, 0},
			srcDataPtr:   2,
			srcCmdPtr:    1,
			srcLoopStack: stack.BuildStack[CmdPtrType](1),
			srLoopEnd:    4,

			expData:      []TestDataType{0, 1, 0, 0},
			extDataPrt:   2,
			expCmdPtr:    4, // cmdPtr is moved to the END of the loop
			expLoopStack: stack.BuildStack[CmdPtrType](),
		},
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			mockCtrl := gomock.NewController(t)

			mockInputReader := NewMockTestInputReader(mockCtrl)
			mockInputReader.EXPECT().Close().AnyTimes().Return(nil)
			mockInputReader.EXPECT().Read(gomock.Any()).AnyTimes().Return(TestDataType(0), nil)

			mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
			mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
			mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().Return(nil)

			bf := New[TestDataType](len(test.srcData), mockInputReader, mockOutputWriter)

			bf.Data = test.srcData
			bf.DataPtr = test.srcDataPtr
			bf.CmdPtr = test.srcCmdPtr
			bf.currentLoopEnd = test.srLoopEnd

			if test.srcLoopStack != nil {
				bf.loopStack = test.srcLoopStack
			}

			err := opStartLoop[TestDataType](bf)

			require.NoError(t, err)
			require.True(t, cmp.Equal(test.expData, bf.Data))
			require.Equal(t, test.extDataPrt, bf.DataPtr)
			require.Equal(t, test.expCmdPtr, bf.CmdPtr)

			stacksCmp := test.expLoopStack.Equals(bf.loopStack, func(a, b *CmdPtrType) bool { return *a == *b })
			require.True(t, stacksCmp)
		})
	}
}

func TestBfInterpreter_EndLoop(t *testing.T) {
	t.Parallel()

	mockCtrl := gomock.NewController(t)

	mockInputReader := NewMockTestInputReader(mockCtrl)
	mockInputReader.EXPECT().Close().AnyTimes().Return(nil)
	mockInputReader.EXPECT().Read(gomock.Any()).AnyTimes().Return(TestDataType(0), nil)

	mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
	mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
	mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().Return(nil)

	bf := New[TestDataType](10, mockInputReader, mockOutputWriter)
	bf.CmdPtr = 6

	bf.loopStack = stack.BuildStack[CmdPtrType](3)

	err := opEndLoop[TestDataType](bf)
	require.NoError(t, err)

	require.Equal(t, CmdPtrType(2), bf.CmdPtr) // moves cmdPtr to a command BEFORE the loop beginning
	require.Equal(t, CmdPtrType(6), bf.currentLoopEnd)
}

func TestBfInterpreter_Run(t *testing.T) {
	t.Parallel()

	type Test struct {
		srcCommands []byte
		srcInput    []TestDataType

		expOutput []TestDataType
		expData   []TestDataType
	}

	tests := map[string]Test{
		"no loop": {
			srcCommands: []byte(`>>+++.>>,+++.`),
			srcInput:    []TestDataType{12},
			expOutput:   []TestDataType{3, 15},
			expData:     []TestDataType{0, 0, 3, 0, 15},
		},

		"one loop": {
			srcCommands: []byte(`>>+++.>>+++[>+++<-]>.`),
			srcInput:    []TestDataType{12},
			expOutput:   []TestDataType{3, 9},
			expData:     []TestDataType{0, 0, 3, 0, 0, 9},
		},

		"nested loop": {
			srcCommands: []byte(`>>+++.>+++[>++[>++.<-]<-]>.`),
			srcInput:    []TestDataType{12},
			expOutput:   []TestDataType{3, 2, 4, 6, 8, 10, 12, 0},
			expData:     []TestDataType{0, 0, 3, 0, 0, 12, 0},
		},
	}

	//nolint:paralleltest
	for description, test := range tests {
		test := test

		t.Run(description, func(t *testing.T) {
			t.Parallel()

			cntInput := 0

			mockCtrl := gomock.NewController(t)

			mockInputReader := NewMockTestInputReader(mockCtrl)
			mockInputReader.EXPECT().Close().AnyTimes().Return(nil)
			mockInputReader.EXPECT().Read(gomock.Any()).AnyTimes().
				DoAndReturn(func(string) (TestDataType, error) {
					if cntInput >= len(test.srcInput) {
						return 0, errors.New("no more input")
					}
					v := test.srcInput[cntInput]
					cntInput++

					return v, nil
				})

			var output []TestDataType

			mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
			mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
			mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().
				DoAndReturn(func(v TestDataType) error {
					output = append(output, v)
					return nil
				})

			bf := New[TestDataType](len(test.expData), mockInputReader, mockOutputWriter)

			resData, err := bf.Run(bytes.NewReader(test.srcCommands))

			require.NoError(t, err)

			require.True(t, cmp.Equal(test.expData, resData))
			require.True(t, cmp.Equal(test.expOutput, output))

		})
	}
}
