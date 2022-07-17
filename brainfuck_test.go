package brainfuck

import (
	"errors"
	"github.com/yurii-vyrovyi/brainfuck/stack"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -source brainfuck_test.go -destination mock_brainfuck.go -package brainfuck
type (
	TestInputReader interface {
		Read(string) (int32, error)
		Close() error
	}

	TestOutputWriter interface {
		Write(int32) error
		Close() error
	}
)

func TestBfInterpreter_Operations(t *testing.T) {
	t.Parallel()

	type Test struct {
		opFunc OpFunc[int32]

		srcData     []int32
		srcDataPtr  DataPtrType
		srcInput    []int32
		srcInError  error
		srcOutError error

		expErr     bool
		expData    []int32
		extDataPrt DataPtrType
		expOutput  []int32
	}

	tests := map[string]Test{
		"ShiftRight": {
			opFunc: opShiftRight[int32],

			srcData:     make([]int32, 20),
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    make([]int32, 20),
			extDataPrt: 1,
			expOutput:  nil,
		},

		"ShiftLeft": {
			opFunc: opShiftLeft[int32],

			srcData:     make([]int32, 20),
			srcDataPtr:  10,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    make([]int32, 20),
			extDataPrt: 9,
			expOutput:  nil,
		},

		"Plus": {
			opFunc: opPlus[int32],

			srcData:     []int32{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []int32{1, 0, 0, 0, 0},
			extDataPrt: 0,
			expOutput:  nil,
		},

		"Minus": {
			opFunc: opMinus[int32],

			srcData:     []int32{0, 3, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []int32{0, 2, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"Out OK": {
			opFunc: opOut[int32],

			srcData:     []int32{0, 15, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []int32{0, 15, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  []int32{15},
		},

		"Out Error": {
			opFunc: opOut[int32],

			srcData:     []int32{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: errors.New("output error"),

			expErr:     true,
			expData:    []int32{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"In OK": {
			opFunc: opIn[int32],

			srcData:     []int32{0, 0, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    []int32{12},
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []int32{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"Input Error": {
			opFunc: opIn[int32],

			srcData:     []int32{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  errors.New("input error"),
			srcOutError: nil,

			expErr:     true,
			expData:    []int32{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"OpError": {
			opFunc: opShiftLeft[int32],

			srcData:     []int32{0, 0, 0, 0, 0},
			srcDataPtr:  0,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     true,
			expData:    []int32{0, 12, 0, 0, 0},
			extDataPrt: 1,
			expOutput:  nil,
		},

		"CustomOp": {
			opFunc: func(bf *BfInterpreter[int32]) error {
				bf.Data[bf.DataPtr] = bf.Data[bf.DataPtr] * bf.Data[bf.DataPtr]
				return nil
			},

			srcData:     []int32{0, 3, 0, 0, 0},
			srcDataPtr:  1,
			srcInput:    nil,
			srcInError:  nil,
			srcOutError: nil,

			expErr:     false,
			expData:    []int32{0, 9, 0, 0, 0},
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
				DoAndReturn(func(string) (int32, error) {

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

			var output []int32

			mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
			mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
			mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().
				DoAndReturn(func(v int32) error {

					if test.srcOutError != nil {
						return test.srcOutError
					}

					output = append(output, v)
					return nil
				})

			bf := New[int32](len(test.srcData), mockInputReader, mockOutputWriter)

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
		srcData      []int32
		srcDataPtr   DataPtrType
		srcCmdPtr    CmdPtrType
		srcLoopStack *stack.Stack[CmdPtrType]
		srLoopEnd    CmdPtrType

		expData      []int32
		extDataPrt   DataPtrType
		expCmdPtr    CmdPtrType
		expLoopStack *stack.Stack[CmdPtrType]
	}

	tests := map[string]Test{
		"Start a new loop": {
			srcData:    []int32{0, 1, 0, 0},
			srcDataPtr: 1,
			srcCmdPtr:  2,
			srLoopEnd:  4,

			expData:      []int32{0, 1, 0, 0},
			extDataPrt:   1,
			expCmdPtr:    2,
			expLoopStack: stack.BuildStack[CmdPtrType](2),
		},

		"Start nesting loop": {
			srcData:      []int32{0, 1, 0, 0},
			srcDataPtr:   1,
			srcCmdPtr:    2,
			srcLoopStack: stack.BuildStack[CmdPtrType](1),
			srLoopEnd:    4,

			expData:      []int32{0, 1, 0, 0},
			extDataPrt:   1,
			expCmdPtr:    2,
			expLoopStack: stack.BuildStack[CmdPtrType](2, 1),
		},

		"Exit nesting loop": {
			srcData:      []int32{0, 1, 0, 0},
			srcDataPtr:   2,
			srcCmdPtr:    2,
			srcLoopStack: stack.BuildStack[CmdPtrType](2, 1),
			srLoopEnd:    4,

			expData:      []int32{0, 1, 0, 0},
			extDataPrt:   2,
			expCmdPtr:    4, // cmdPtr is moved to the END of the loop
			expLoopStack: stack.BuildStack[CmdPtrType](1),
		},

		"Exit topmost loop": {
			srcData:      []int32{0, 1, 0, 0},
			srcDataPtr:   2,
			srcCmdPtr:    1,
			srcLoopStack: stack.BuildStack[CmdPtrType](1),
			srLoopEnd:    4,

			expData:      []int32{0, 1, 0, 0},
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
			mockInputReader.EXPECT().Read(gomock.Any()).AnyTimes().Return(int32(0), nil)

			mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
			mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
			mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().Return(nil)

			bf := New[int32](len(test.srcData), mockInputReader, mockOutputWriter)

			bf.Data = test.srcData
			bf.DataPtr = test.srcDataPtr
			bf.CmdPtr = test.srcCmdPtr
			bf.currentLoopEnd = test.srLoopEnd

			if test.srcLoopStack != nil {
				bf.loopStack = test.srcLoopStack
			}

			err := opStartLoop[int32](bf)

			require.NoError(t, err)
			require.True(t, cmp.Equal(test.expData, bf.Data))
			require.Equal(t, test.extDataPrt, bf.DataPtr)
			require.Equal(t, test.expCmdPtr, bf.CmdPtr)

			stacksCmp := test.expLoopStack.Equals(bf.loopStack, func(a, b *CmdPtrType) bool { return *a == *b })
			require.True(t, stacksCmp)
		})
	}
}
