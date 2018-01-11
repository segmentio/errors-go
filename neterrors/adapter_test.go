package neterrors

import (
	"io"
	"net"
	"testing"

	errors "github.com/segmentio/errors-go"
	"github.com/segmentio/errors-go/errorstest"
)

func TestAdapt(t *testing.T) {
	errorstest.TestAdapter(t, errors.AdapterFunc(Adapt),
		errorstest.AdapterTest{
			Error: &net.AddrError{Err: "whatever"},
			Types: []string{},
		},

		errorstest.AdapterTest{
			Error: &net.AddrError{Err: "mismatched local address type"},
			Types: []string{"Validation"},
		},

		errorstest.AdapterTest{
			Error: &net.AddrError{Err: "unexpected address type"},
			Types: []string{"Validation"},
		},

		errorstest.AdapterTest{
			Error: &net.AddrError{Err: "unknown IP protocol specified"},
			Types: []string{"Validation"},
		},

		errorstest.AdapterTest{
			Error: &net.AddrError{Err: "invalid MAC address"},
			Types: []string{"Validation"},
		},

		errorstest.AdapterTest{
			Error: &net.AddrError{Err: "no suitable address found"},
			Types: []string{"Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.DNSError{},
			Types: []string{"Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.DNSError{IsTemporary: true},
			Types: []string{"Temporary", "Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.DNSError{IsTimeout: true},
			Types: []string{"Timeout", "Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.DNSError{IsTemporary: true, IsTimeout: true},
			Types: []string{"Temporary", "Timeout", "Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.ParseError{},
			Types: []string{"Validation"},
		},

		errorstest.AdapterTest{
			Error: &net.OpError{Err: io.ErrClosedPipe, Op: "read"},
			Types: []string{},
		},

		errorstest.AdapterTest{
			Error: &net.OpError{Err: io.ErrClosedPipe, Op: "dial"},
			Types: []string{"Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.OpError{Err: io.ErrClosedPipe, Op: "write"},
			Types: []string{"Unreachable"},
		},

		errorstest.AdapterTest{
			Error: &net.OpError{Err: &timeout{}, Op: "read"},
			Types: []string{"Temporary", "Timeout"},
		},

		errorstest.AdapterTest{
			Error: net.ErrWriteToConnected,
			Types: []string{"Validation"},
		},
	)
}

type timeout struct{}

func (*timeout) Error() string   { return "timeout" }
func (*timeout) Timeout() bool   { return true }
func (*timeout) Temporary() bool { return true }
