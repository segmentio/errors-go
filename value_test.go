package errors

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestValueOf(t *testing.T) {
	tests := []struct {
		err error
		val Value
	}{
		{
			err: nil,
			val: Value{},
		},

		{
			err: New("hello world!"),
			val: Value{
				Message: "hello world!",
				Stack: []string{
					"github.com/segmentio/errors-go/value_test.go:21:TestValueOf",
				},
			},
		},

		{
			err: Join(
				New("A"),
				New("B"),
				WithTypes(New("C"), "type1", "type2", "type3"),
			),
			val: Value{
				Causes: []Value{
					{
						Message: "A",
						Stack: []string{
							"github.com/segmentio/errors-go/value_test.go:32:TestValueOf",
						},
					},
					{
						Message: "B",
						Stack: []string{
							"github.com/segmentio/errors-go/value_test.go:33:TestValueOf",
						},
					},
					{
						Types:   []string{"type1", "type2", "type3"},
						Message: "C",
						Stack: []string{
							"github.com/segmentio/errors-go/value_test.go:34:TestValueOf",
						},
					},
				},
			},
		},

		{
			err: WithTags(New("hello world!"), T("A", "1"), T("B", "2"), T("C", "3")),
			val: Value{
				Message: "hello world!",
				Tags: map[string]string{
					"A": "1",
					"B": "2",
					"C": "3",
				},
				Stack: []string{
					"github.com/segmentio/errors-go/value_test.go:62:TestValueOf",
				},
			},
		},

		{
			err: WithStack(
				WithStack(
					New("multiple stacks"),
				),
			),
			val: Value{
				Message: "multiple stacks",
				Stack: []string{
					"github.com/segmentio/errors-go/value_test.go:77:TestValueOf",
					"",
					"github.com/segmentio/errors-go/value_test.go:78:TestValueOf",
					"",
					"github.com/segmentio/errors-go/value_test.go:79:TestValueOf",
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%T", test.err), func(t *testing.T) {
			val := ValueOf(test.err)
			stripRuntimeStackFrames(&val)

			if !reflect.DeepEqual(val, test.val) {
				t.Error("bad error value:")
				t.Logf("expected: %#v", test.val)
				t.Logf("found:    %#v", val)
			}

			err := val.Err()

			if t1, t2 := Tags(test.err), Tags(err); !equalTags(t1, t2) {
				t.Error("tags mismatch on the error constructed from a value:")
				t.Log("expected:", t1)
				t.Log("found:   ", t2)
			}

			if t1, t2 := Types(test.err), Types(err); !equalTypes(t1, t2) {
				t.Error("types mismatch on the error constructed from a value:")
				t.Log("expected:", t1)
				t.Log("found:   ", t2)
			}

			if err != nil {
				if stack := stackTrace(err); len(stack) == 0 {
					t.Error("missing stack trace on error constructed from a value")
				}
			}

			s1 := fmt.Sprint(test.err)
			s2 := fmt.Sprint(err)

			if s1 != s2 {
				t.Error("the formatted representation of the error constructed from a value is incorrect")
				t.Log("expected:", s1)
				t.Log("found:   ", s2)
			}
		})
	}
}

func stripRuntimeStackFrames(v *Value) {
	if len(v.Stack) != 0 {
		i := 0

		for _, s := range v.Stack {
			if s == "" || strings.HasPrefix(s, "github.com/segmentio/errors-go") {
				v.Stack[i] = s
				i++
			}
		}

		v.Stack = v.Stack[:i]
	}
	for i := range v.Causes {
		stripRuntimeStackFrames(&v.Causes[i])
	}
}
