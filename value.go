package errors

import (
	"fmt"
	"strings"
)

// Value is a serializable error representation which carries all rich
// information of errors that can be constructed by this package.
//
// This type is useful to transmit errors between programs that communicate over
// some IPC mechanism. It may also be used as a form of reflection to discover
// the various components of an error.
type Value struct {
	Message string
	Tags    map[string]string
	Types   []string
	Stack   []string
	Causes  []Value
}

// ValueOf returns an error value representing err. If err is nil the function
// returns the zero-value of Value.
func ValueOf(err error) Value {
	if err == nil {
		return Value{}
	}

	msgs, types, tags, stacks, causes := inspect(err)

	v := Value{
		Message: strings.Join(msgs, ": "),
		Types:   types,
		Tags:    makeTagsMap(tags...),
	}

	if len(stacks) != 0 {
		v.Stack = make([]string, 0, len(stacks[0])*len(stacks))

		for i, stack := range stacks {
			if i != 0 {
				v.Stack = append(v.Stack, "")
			}
			for _, frame := range stack {
				v.Stack = append(v.Stack, fmt.Sprintf("%+v:%n", frame, frame))
			}
		}
	}

	if len(causes) != 0 {
		v.Causes = make([]Value, len(causes))

		for i, cause := range causes {
			v.Causes[i] = ValueOf(cause)
		}
	}

	return v
}

// Err constructs and returns an error from v, the error message, types, and
// causes are rebuilt and part of the returned error to match as closely as
// possible the information carried by the error that this value was built from
// in the first place.
//
// Note that the only information that isn't carried back from the Value into
// the returned error is the stack trace, because it may be coming from a
// different program there is no way to match it to the correct function
// pointers. Instead the stack trace information in the returned error is set
// to the call stack that led to this method call, which in general is way more
// relevant to the program which is calling this method.
//
// If v is the zero-value, the method returns a nil error.
func (v Value) Err() error {
	if v.IsNil() {
		return nil
	}

	e := &errorValue{
		msg:   v.Message,
		types: copyTypes(v.Types),
		tags:  makeTagsFromMap(v.Tags),
		stack: CaptureStackTrace(1),
	}

	if len(v.Causes) != 0 {
		e.causes = make([]error, len(v.Causes))

		for i := range v.Causes {
			e.causes[i] = v.Causes[i].Err()
		}
	}

	return e
}

// IsNil returns true if v represents a nil error (which means it is the
// zero-value).
func (v Value) IsNil() bool {
	return v.Message == "" && v.Tags == nil && v.Types == nil && v.Stack == nil && v.Causes == nil
}

type errorValue struct {
	msg    string
	causes []error
	types  []string
	tags   []Tag
	stack  StackTrace
}

func (e *errorValue) Error() string {
	return e.msg
}

func (e *errorValue) Message() string {
	return e.msg
}

func (e *errorValue) Types() []string {
	return e.types
}

func (e *errorValue) Tags() []Tag {
	return e.tags
}

func (e *errorValue) Causes() []error {
	return e.causes
}

func (e *errorValue) StackTrace() StackTrace {
	return e.stack
}

func (e *errorValue) Format(s fmt.State, v rune) {
	format(s, v, e)
}
