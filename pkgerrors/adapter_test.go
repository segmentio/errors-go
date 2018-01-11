package pkgerrors

import (
	stderrors "errors"
	"fmt"
	"testing"

	pkgerrors "github.com/pkg/errors"
	"github.com/segmentio/errors-go"
)

func TestAdapt(t *testing.T) {
	e0 := stderrors.New("base error")
	e1 := pkgerrors.WithStack(e0) // line 14
	e2, ok2 := Adapt(e0)
	e3, ok3 := Adapt(e1)

	if ok2 {
		t.Error("errors that are not from the github.com/pkg/errors package must not be adapted")
	}

	if !ok3 {
		t.Error("errors from the github.com/pkg/errors package that have a stack must be adapted")
	}

	if e2 != e0 {
		t.Error("non-adapted errors must be preserved by a call to Adapt")
	}

	s1 := e1.Error()
	s3 := e3.Error()

	if s1 != s3 {
		t.Error("bad error string")
		t.Log("expected:", s1)
		t.Log("found:   ", s3)
	}

	c1 := errors.Cause(e1)
	c3 := errors.Cause(e3)

	if c1 != c3 {
		t.Error("bad error cause")
		t.Log("expected:", c1)
		t.Log("found:   ", c3)
	}

	stack := stackTrace(e3)
	stack = stack[:1] // just capture the frame within this function

	if s := fmt.Sprint(stack); s != "[adapter_test.go:14]" {
		t.Error("bad stack trace representation:", s)
	}
}

func stackTrace(err error) errors.StackTrace {
	e, ok := err.(interface {
		StackTrace() errors.StackTrace
	})
	if ok {
		return e.StackTrace()
	}
	return nil
}
