package ioerrors

import "io"

// Adapt checks the type of err and if it matches one of the error types or one
// of the error values of the standard io package, adapts it to make error types
// discoverable using the errors.Is function.
//
// This function is automatically installed as a global adapter when importing
// the neterrors package, a program likely should use errors.Adapt instead of
// calling this adapter directly.
func Adapt(err error) (error, bool) {
	switch err {
	case io.EOF:
		return &eof{err}, true

	case io.ErrClosedPipe:
		return &closedPipe{err}, true

	case io.ErrNoProgress:
		return &noProgress{err}, true

	case io.ErrShortBuffer:
		return &shortBuffer{err}, true

	case io.ErrShortWrite:
		return &shortWrite{err}, true

	case io.ErrUnexpectedEOF:
		return &unexpectedEOF{err}, true

	default:
		return err, false
	}
}

type eof struct{ cause error }

func (e *eof) Error() string { return e.cause.Error() }
func (e *eof) Cause() error  { return e.cause }
func (e *eof) EOF() bool     { return true }

type closedPipe struct{ cause error }

func (e *closedPipe) Error() string    { return e.cause.Error() }
func (e *closedPipe) Cause() error     { return e.cause }
func (e *closedPipe) ClosedPipe() bool { return true }

type noProgress struct{ cause error }

func (e *noProgress) Error() string    { return e.cause.Error() }
func (e *noProgress) Cause() error     { return e.cause }
func (e *noProgress) NoProgress() bool { return true }

type shortBuffer struct{ cause error }

func (e *shortBuffer) Error() string     { return e.cause.Error() }
func (e *shortBuffer) Cause() error      { return e.cause }
func (e *shortBuffer) ShortBuffer() bool { return true }

type shortWrite struct{ cause error }

func (e *shortWrite) Error() string    { return e.cause.Error() }
func (e *shortWrite) Cause() error     { return e.cause }
func (e *shortWrite) ShortWrite() bool { return true }

type unexpectedEOF struct{ cause error }

func (e *unexpectedEOF) Error() string       { return e.cause.Error() }
func (e *unexpectedEOF) Cause() error        { return e.cause }
func (e *unexpectedEOF) UnexpectedEOF() bool { return true }
