package neterrors

import (
	"net"
	"strings"
)

// Adapt checks the type of err and if it matches one of the error types or one
// of the error values of the standard net package, adapts it to make error
// types discoverable using the errors.Is function.
//
// This function is automatically installed as a global adapter when importing
// the neterrors package, a program likely should use errors.Adapt instead of
// calling this adapter directly.
func Adapt(err error) (error, bool) {
	switch e := err.(type) {
	case *net.AddrError:
		return &addrError{e}, true

	case *net.DNSError:
		return &dnsError{e}, true

	case *net.ParseError:
		return &parseError{e}, true

	case *net.OpError:
		return &opError{e}, true
	}

	switch err {
	case net.ErrWriteToConnected:
		return &validation{err}, true
	}

	return err, false
}

type addrError struct{ cause error }

func (e *addrError) Cause() error  { return e.cause }
func (e *addrError) Error() string { return e.cause.Error() }

func (e *addrError) Validation() bool {
	s := e.cause.Error()
	// according to https://golang.org/search?q=AddrError%7B, those are common
	// prefixes of address errors.
	return strings.HasPrefix(s, "mismatched ") ||
		strings.HasPrefix(s, "unexpected ") ||
		strings.HasPrefix(s, "invalid ") ||
		strings.HasPrefix(s, "unknown ")
}

func (e *addrError) Unreachable() bool {
	// /src/net/net.go: errNoSuitableAddress
	return e.cause.Error() == "no suitable address found"
}

type dnsError struct{ cause *net.DNSError }

func (e *dnsError) Cause() error      { return e.cause }
func (e *dnsError) Error() string     { return e.cause.Error() }
func (e *dnsError) Temporary() bool   { return e.cause.Temporary() }
func (e *dnsError) Timeout() bool     { return e.cause.Timeout() }
func (e *dnsError) Unreachable() bool { return true }

type parseError struct{ cause *net.ParseError }

func (e *parseError) Cause() error     { return e.cause }
func (e *parseError) Error() string    { return e.cause.Error() }
func (e *parseError) Validation() bool { return true }

type opError struct{ cause *net.OpError }

func (e *opError) Cause() error      { return e.cause }
func (e *opError) Error() string     { return e.cause.Error() }
func (e *opError) Temporary() bool   { return e.cause.Temporary() }
func (e *opError) Timeout() bool     { return e.cause.Timeout() }
func (e *opError) Unreachable() bool { return e.cause.Op == "dial" || e.cause.Op == "write" }

type validation struct{ cause error }

func (e *validation) Cause() error     { return e.cause }
func (e *validation) Error() string    { return e.cause.Error() }
func (e *validation) Validation() bool { return true }
