package awserrors

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	errors "github.com/segmentio/errors-go"
)

func TestAdapt(t *testing.T) {
	tests := []struct {
		scenario string
		function func(*testing.T)
	}{
		{
			scenario: "adapting errors that don't come from AWS simply returns the original error",
			function: testAdaptNonAwsError,
		},

		{
			scenario: "adapting errors exposes the correct code, message, and causes",
			function: testAdaptError,
		},

		{
			scenario: "adapting batch errors exposes the correct code, message, and causes",
			function: testAdaptBatchError,
		},

		{
			scenario: "adapting request errors exposes the correct code, message, and causes",
			function: testAdaptRequestError,
		},
	}

	for _, test := range tests {
		t.Run(test.scenario, test.function)
	}
}

func testAdaptNonAwsError(t *testing.T) {
	e1 := errors.New("some error")
	e2, ok := Adapt(e1)

	if ok {
		t.Error("adapting non-AWS errors must return false to indicate that the error was not recognized")
	}

	if e1 != e2 {
		t.Error("adapting non-AWS errors must return the original error unchanged")
	}
}

func testAdaptError(t *testing.T) {
	e0 := errors.New("base error")
	e1 := &singleError{code: "ThrottlingException", msg: "too many requests", orig: e0}
	e2, ok := Adapt(e1)

	if !ok {
		t.Error("adapting AWS errors must return true to indicate that the error was recognized")
	}

	s1 := e1.Error()
	s2 := e2.Error()

	if s1 != s2 {
		t.Error("the adapted AWS error must preserve the original error string")
		t.Log("expected:", s1)
		t.Log("found:   ", s2)
	}

	c1 := code(e1)
	c2 := code(e2)

	if c1 != c2 {
		t.Error("the adapted AWS error must preserve the original error code")
		t.Log("expected:", c1)
		t.Log("found:   ", c2)
	}

	m1 := message(e1)
	m2 := message(e2)

	if m1 != m2 {
		t.Error("the adapted AWS error must preserve the original error message")
		t.Log("expected:", m1)
		t.Log("found:   ", m2)
	}

	if cause := errors.Cause(e2); cause != e0 {
		t.Error("the adapted AWS error must expose the original error as a cause")
		t.Log("expected:", e0)
		t.Log("found:   ", cause)
	}

	if causes := errors.Causes(e2); !equalErrors(causes, []error{e0}) {
		t.Error("the adapted AWS error must expose the original error as causes")
		t.Log("expected:", []error{e0})
		t.Log("found:   ", causes)
	}

	if id := requestID(e2); id != "" {
		t.Error("unexpected request ID on AWS error:", id)
	}

	if sc := statusCode(e2); sc != 0 {
		t.Error("unexpected status code on AWS error:", sc)
	}
}

func testAdaptBatchError(t *testing.T) {
	e0 := []error{
		errors.New("base error 1"),
		errors.New("base error 2"),
		errors.New("base error 3"),
	}
	e1 := &batchError{code: "ThrottlingException", msg: "too many requests", orig: e0}
	e2, ok := Adapt(e1)

	if !ok {
		t.Error("adapting AWS errors must return true to indicate that the error was recognized")
	}

	s1 := e1.Error()
	s2 := e2.Error()

	if s1 != s2 {
		t.Error("the adapted AWS error must preserve the original error string")
		t.Log("expected:", s1)
		t.Log("found:   ", s2)
	}

	c1 := code(e1)
	c2 := code(e2)

	if c1 != c2 {
		t.Error("the adapted AWS error must preserve the original error code")
		t.Log("expected:", c1)
		t.Log("found:   ", c2)
	}

	m1 := message(e1)
	m2 := message(e2)

	if m1 != m2 {
		t.Error("the adapted AWS error must preserve the original error message")
		t.Log("expected:", m1)
		t.Log("found:   ", m2)
	}

	if causes := errors.Causes(e2); !equalErrors(causes, e0) {
		t.Error("the adapted AWS error must expose the original error as causes")
		t.Log("expected:", e0)
		t.Log("found:   ", causes)
	}
}

func testAdaptRequestError(t *testing.T) {
	e1 := &requestError{code: "ThrottlingException", msg: "too many requests", id: "1234567890", status: 429}
	e2, ok := Adapt(e1)

	if !ok {
		t.Error("adapting AWS errors must return true to indicate that the error was recognized")
	}

	s1 := e1.Error()
	s2 := e2.Error()

	if s1 != s2 {
		t.Error("the adapted AWS error must preserve the original error string")
		t.Log("expected:", s1)
		t.Log("found:   ", s2)
	}

	c1 := code(e1)
	c2 := code(e2)

	if c1 != c2 {
		t.Error("the adapted AWS error must preserve the original error code")
		t.Log("expected:", c1)
		t.Log("found:   ", c2)
	}

	m1 := message(e1)
	m2 := message(e2)

	if m1 != m2 {
		t.Error("the adapted AWS error must preserve the original error message")
		t.Log("expected:", m1)
		t.Log("found:   ", m2)
	}

	id1 := requestID(e1)
	id2 := requestID(e2)

	if id1 != id2 {
		t.Error("the adapted AWS error must preserve the original request ID")
		t.Log("expected:", id1)
		t.Log("found:   ", id2)
	}

	sc1 := statusCode(e1)
	sc2 := statusCode(e2)

	if sc1 != sc2 {
		t.Error("the adapted AWS error must preserve the original status code")
		t.Log("expected:", sc1)
		t.Log("found:   ", sc2)
	}
}

func code(err error) string {
	e, ok := err.(interface {
		Code() string
	})
	if ok {
		return e.Code()
	}
	return ""
}

func message(err error) string {
	e, ok := err.(interface {
		Message() string
	})
	if ok {
		return e.Message()
	}
	return ""
}

func requestID(err error) string {
	e, ok := err.(interface {
		RequestID() string
	})
	if ok {
		return e.RequestID()
	}
	return ""
}

func statusCode(err error) int {
	e, ok := err.(interface {
		StatusCode() int
	})
	if ok {
		return e.StatusCode()
	}
	return 0
}

func equalErrors(errs1 []error, errs2 []error) bool {
	if len(errs1) != len(errs2) {
		return false
	}

	for i := range errs1 {
		if errs1[i] != errs2[i] {
			return false
		}
	}

	return true
}

type singleError struct {
	code string
	msg  string
	orig error
}

func (e *singleError) Error() string   { return awserr.SprintError(e.code, e.msg, "", e.orig) }
func (e *singleError) Code() string    { return e.code }
func (e *singleError) Message() string { return e.msg }
func (e *singleError) OrigErr() error  { return e.orig }

type batchError struct {
	code string
	msg  string
	orig []error
}

func (e *batchError) Error() string {
	return awserr.SprintError(e.code, e.msg, "", errors.Join(e.orig...))
}
func (e *batchError) Code() string      { return e.code }
func (e *batchError) Message() string   { return e.msg }
func (e *batchError) OrigErrs() []error { return e.orig }

type requestError struct {
	code   string
	msg    string
	id     string
	status int
}

func (e *requestError) Error() string {
	return awserr.SprintError(e.code, e.msg, fmt.Sprintf("id=%s, status=%d", e.id, e.status), nil)
}
func (e *requestError) Code() string      { return e.code }
func (e *requestError) Message() string   { return e.msg }
func (e *requestError) OrigErr() error    { return nil }
func (e *requestError) StatusCode() int   { return e.status }
func (e *requestError) RequestID() string { return e.id }
