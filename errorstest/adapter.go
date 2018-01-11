package errorstest

import (
	"fmt"
	"testing"

	errors "github.com/segmentio/errors-go"
)

type AdapterTest struct {
	Error error
	Types []string
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
