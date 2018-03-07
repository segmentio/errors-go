package errors

import (
	"errors"
	"fmt"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		err   error
		types []string
		tags  []Tag
	}{
		{
			err: nil,
		},

		{
			err: New("hello world"),
		},

		{
			err: Errorf("hello world"),
		},

		{
			err:   &timeout{},
			types: []string{"Temporary", "Timeout"},
		},

		{
			err: Join(),
		},

		{
			err:   Join(nil, New(""), &timeout{}),
			types: []string{"Temporary", "Timeout"},
		},

		{
			err: Recv(errorChan()),
		},

		{
			err:   Recv(errorChan(nil, New(""), &timeout{})),
			types: []string{"Temporary", "Timeout"},
		},

		{
			err: WithTypes(nil),
		},

		{
			err: WithTags(nil),
		},

		{
			err: WithTypes(New("hello")),
		},

		{
			err:   WithTypes(New("hello"), "Temporary", "Timeout"),
			types: []string{"Temporary", "Timeout"},
		},

		{
			err:  WithTags(New("hello"), T("A", "1"), T("B", "2"), T("C", "3")),
			tags: []Tag{{"A", "1"}, {"B", "2"}, {"C", "3"}},
		},

		{
			err: WithTags(
				Join(
					WithTags(New("hello"), T("A", "1"), T("B", "2"), T("C", "3")),
					New("world"),
					WithTags(New("!!!"), T("D", "4")),
				),
				T("A", "a"),
			),
			tags: []Tag{{"A", "1"}, {"A", "a"}, {"B", "2"}, {"C", "3"}, {"D", "4"}},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.err), func(t *testing.T) {
			for _, subtest := range []struct {
				info string
				err  error
			}{
				{"base", test.err},
				{"with stack", WithStack(test.err)},
				{"with message", WithMessage(test.err, "hello world")},
				{"wrap", Wrap(test.err, "hello world")},
				{"wrapf", Wrapf(test.err, "hello %s", "world")},
			} {
				t.Run(subtest.info, func(t *testing.T) {
					testError(t, subtest.err, test.types, test.tags)
				})
			}
		})
	}
}

func TestCauses(t *testing.T) {
	join := Join

	recv := func(errs ...error) error {
		return Recv(errorChan(errs...))
	}

	withStack := func(errs ...error) error {
		return WithStack(Join(errs...))
	}

	withMessage := func(errs ...error) error {
		return WithMessage(Join(errs...), "hello world")
	}

	wrap := func(errs ...error) error {
		return Wrap(Join(errs...), "hello world")
	}

	wrapf := func(errs ...error) error {
		return Wrapf(Join(errs...), "hello %s", "world")
	}

	newErrors := []struct {
		info     string
		newError func(...error) error
	}{
		{"join", join},
		{"recv", recv},
		{"withStack", withStack},
		{"withmessage", withMessage},
		{"wrap", wrap},
		{"wrapf", wrapf},
	}

	errorCauses := [][]error{
		nil,
		{New("")},
		{&timeout{}},
		{New(""), &timeout{}},
	}

	for _, test := range newErrors {
		t.Run(test.info, func(t *testing.T) {
			for _, causes := range errorCauses {
				t.Run(fmt.Sprintf("%#v", causes), func(t *testing.T) {
					err := test.newError(causes...)
					found := Causes(err)

					if oneCause := Cause(err); oneCause != nil {
						found = Causes(oneCause)
					}

					if !reflect.DeepEqual(causes, found) {
						t.Error("bad error causes:")
						t.Logf("expected: %#v", causes)
						t.Logf("found:    %#v", found)
					}
				})
			}
		})
	}
}

func TestErr(t *testing.T) {
	tests := []struct {
		s string
		v interface{}
	}{
		{
			s: "",
			v: nil,
		},

		{
			s: "hello world",
			v: New("hello world"),
		},

		{
			s: "hello world",
			v: errors.New("hello world"),
		},

		{
			s: "42",
			v: "42",
		},

		{
			s: "42",
			v: 42,
		},

		{
			s: "hello world",
			v: ValueOf(New("hello world")),
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.v), func(t *testing.T) {
			e := Err(test.v)

			if e == nil {
				if len(test.s) != 0 {
					t.Error("unexpected nil error")
				}
				return
			}

			if s := e.Error(); s != test.s {
				t.Errorf("bad error message: %q !+ %q", s, test.s)
			}
		})
	}
}

func testError(t *testing.T, err error, types []string, tags []Tag) {
	if errTypes := Types(err); !equalTypes(errTypes, types) {
		t.Error("bad error types:", errTypes, "!=", types)
	}

	if errTags := Tags(err); !equalTags(errTags, tags) {
		t.Error("bad error tags:", errTags, "!=", tags)
	}

	if Is("whatever", err) {
		t.Errorf("%#v was expected to not be a %q error", err, "whatever")
	}

	if err != nil {
		if s := err.Error(); len(s) == 0 {
			t.Errorf("%#v has no error message", err)
		}

		// Not very useful tests, but it exercises the code
		// paths and ensures there are no invalid pointer
		// dereferences.
		fmt.Fprintf(ioutil.Discard, "%s", err)
		fmt.Fprintf(ioutil.Discard, "%q", err)
		fmt.Fprintf(ioutil.Discard, "%v", err)
		fmt.Fprintf(ioutil.Discard, "%+v", err)
	}

	for _, typ := range types {
		if !Is(typ, err) {
			t.Errorf("%#v was expected to be a %q error", err, typ)
		}
	}

	if st, ok := err.(interface {
		StackTrace() StackTrace
	}); ok {
		if len(st.StackTrace()) == 0 {
			t.Errorf("%#v has an empty stack trace", err)
		}
	}
}

func TestCauseWithNilCause(t *testing.T) {
	e1 := &errorWithNilCause{}
	e2 := Wrap(e1, "")
	e3 := Cause(e2)
	e4 := Causes(e2)
	e5 := Causes(TODO)

	if e3 != e1 {
		t.Error("when an error along the chain of causes returns a nil cause it must be considered the cause")
	}

	if len(e4) != 1 || e4[0] != e1 {
		t.Error("when an error along the chain of causes returns a nil cause it must be considered the cause")
	}

	if e5 != nil {
		t.Error("errors that have no causes must return no causes")
	}
}

type timeout struct{}

func (*timeout) Error() string   { return "timeout" }
func (*timeout) Timeout() bool   { return true }
func (*timeout) Temporary() bool { return true }

func errorChan(errs ...error) <-chan error {
	ch := make(chan error, len(errs))
	for _, err := range errs {
		ch <- err
	}
	close(ch)
	return ch
}

func equalTypes(t1, t2 []string) bool {
	if len(t1) != len(t2) {
		return false
	}

	for i := range t1 {
		if t1[i] != t2[i] {
			return false
		}
	}

	return true
}

func equalTags(t1, t2 []Tag) bool {
	if len(t1) != len(t2) {
		return false
	}

	for i := range t1 {
		if t1[i] != t2[i] {
			return false
		}
	}

	return true
}

type errorWithNilCause struct{}

func (*errorWithNilCause) Error() string { return "" }
func (*errorWithNilCause) Cause() error  { return nil }

func TestLookupTag(t *testing.T) {
	taggedErr := WithTags(errors.New("tagged"), T("key", "value1"))

	tests := []struct {
		err    error
		result string
	}{
		{
			err: nil,
		},
		{
			err: errors.New("foreign"),
		},
		{
			err:    taggedErr,
			result: "value1",
		},
		{
			err:    WithTags(Wrap(taggedErr, "double"), T("key", "value2")),
			result: "value2",
		},
	}
	for _, test := range tests {
		var subtestName string
		if test.err == nil {
			subtestName = "<nil>"
		} else {
			subtestName = test.err.Error()
		}
		t.Run(subtestName, func(t *testing.T) {
			if actual := LookupTag(test.err, "key"); actual != test.result {
				t.Error("bad result:")
				t.Logf("expected: %#v", test.result)
				t.Logf("found:    %#v", actual)
			}
		})
	}
}
