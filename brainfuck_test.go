package brainfuck

import (
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/go-cmp/cmp"
	"github.com/stretchr/testify/require"
)

//go:generate mockgen -source brainfuck_test.go -destination mock_test_writer.go -package brainfuck  TestOutputWriter
type TestOutputWriter interface {
	Write(int32) error
	Close() error
}

func TestBfInterpreter_Operations(t *testing.T) {
	t.Parallel()

	type Test struct {
		opFunc OpFunc[int32]

		srcData    []int32
		srcDataPtr int
		srcInput   []CmdType

		expErr     bool
		expData    []int32
		extDataPrt int
		expOutput  []int32
	}

	tests := map[string]Test{
		"ShiftRight": {
			opFunc: opShiftRight[int32],

			srcData:    make([]int32, 20),
			srcDataPtr: 0,
			srcInput:   nil,

			expErr:     false,
			expData:    make([]int32, 20),
			extDataPrt: 1,
			expOutput:  nil,
		},

		"ShiftLeft": {
			opFunc: opShiftLeft[int32],

			srcData:    make([]int32, 20),
			srcDataPtr: 10,
			srcInput:   nil,

			expErr:     false,
			expData:    make([]int32, 20),
			extDataPrt: 9,
			expOutput:  nil,
		},

		"Plus": {
			opFunc: opPlus[int32],

			srcData:    []int32{0, 0, 0, 0, 0},
			srcDataPtr: 0,
			srcInput:   nil,

			expErr:     false,
			expData:    []int32{1, 0, 0, 0, 0},
			extDataPrt: 0,
			expOutput:  nil,
		},

		"Minus": {
			opFunc: opMinus[int32],

			srcData:    []int32{0, 3, 0, 0, 0},
			srcDataPtr: 1,
			srcInput:   nil,

			expErr:     false,
			expData:    []int32{0, 2, 0, 0, 0},
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

			mockInputReader := NewMockInputReader(mockCtrl)
			mockInputReader.EXPECT().Close().AnyTimes().Return(nil)
			mockInputReader.EXPECT().Read(gomock.Any()).AnyTimes().
				DoAndReturn(func(string) (CmdType, error) {
					if cntInput >= len(test.srcInput) {
						return 0, errors.New("no more input")
					}
					cmd := test.srcInput[cntInput]
					cntInput++

					return cmd, nil
				})

			var output []int32

			mockOutputWriter := NewMockTestOutputWriter(mockCtrl)
			mockOutputWriter.EXPECT().Close().AnyTimes().Return(nil)
			mockOutputWriter.EXPECT().Write(gomock.Any()).AnyTimes().
				DoAndReturn(func(v int32) error {
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

		})
	}
}
