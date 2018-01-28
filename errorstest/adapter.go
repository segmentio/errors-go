package errorstest

import (
	"fmt"
	"testing"

	errors "github.com/segmentio/errors-go"
)

type AdapterTest struct {
	Error   error
	Message string
	Types   []string
	Tags    []errors.Tag
}

func TestAdapter(t *testing.T, a errors.Adapter, tests ...AdapterTest) {
	for _, test := range tests {
		t.Run(fmt.Sprintf("%T(%v)", test.Error, test.Error), func(t *testing.T) {
			err, ok := a.Adapt(test.Error)

			if !ok {
				t.Error("the error was not recognized")
				return
			}

			for _, typ := range test.Types {
				if !errors.Is(typ, err) {
					t.Errorf("%#v was expected to be a %q error", err, typ)
				}
			}

			if types := errors.Types(err); !typesEqual(types, test.Types) {
				t.Error("types mismatch")
				t.Log("expected:", test.Types)
				t.Log("found:   ", types)
			}

			if tags := errors.Tags(err); !tagsEqual(tags, test.Tags) {
				t.Error("tags mismatch")
				t.Log("expected:", test.Tags)
				t.Log("found:   ", tags)
			}

			if msg := message(err); msg != test.Message {
				t.Error("messages mismatch")
				t.Log("expected:", test.Message)
				t.Log("found:   ", msg)
			}

			if s := err.Error(); len(s) == 0 {
				t.Errorf("%#v has no error message", err)
			}

			if cause := errors.Cause(err); cause != test.Error {
				t.Error("invalid cause:", cause)
			}
		})
	}

	e1 := errors.New("non-adaptable")
	e2, ok := a.Adapt(e1)

	if ok {
		t.Error("errors.TODO is not a net error, it cannot be adapted by the neterrors adapters")
	}

	if e1 != e2 {
		t.Error("non-adapted errors must be returned unchanged by the neterrors adapter")
	}
}

func message(err error) string {
	if e, ok := err.(interface {
		Message() string
	}); ok {
		return e.Message()
	}
	return ""
}

func typesEqual(t1, t2 []string) bool {
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

func tagsEqual(t1, t2 []errors.Tag) bool {
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
