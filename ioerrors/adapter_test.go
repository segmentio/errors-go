package ioerrors

import (
	"io"
	"testing"

	errors "github.com/segmentio/errors-go"
	"github.com/segmentio/errors-go/errorstest"
)

func TestAdapt(t *testing.T) {
	errorstest.TestAdapter(t, errors.AdapterFunc(Adapt),
		errorstest.AdapterTest{
			Error: io.EOF,
			Types: []string{"EOF"},
		},

		errorstest.AdapterTest{
			Error: io.ErrClosedPipe,
			Types: []string{"ClosedPipe"},
		},

		errorstest.AdapterTest{
			Error: io.ErrNoProgress,
			Types: []string{"NoProgress"},
		},

		errorstest.AdapterTest{
			Error: io.ErrShortBuffer,
			Types: []string{"ShortBuffer"},
		},

		errorstest.AdapterTest{
			Error: io.ErrShortWrite,
			Types: []string{"ShortWrite"},
		},

		errorstest.AdapterTest{
			Error: io.ErrUnexpectedEOF,
			Types: []string{"UnexpectedEOF"},
		},
	)
}
