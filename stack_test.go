package errors

import (
	"fmt"
	"os"
	"testing"
)

var cwd = getwd()

func TestFormatStackTrace(t *testing.T) {
	f := func() StackTrace { return CaptureStackTrace(0) } // line 9
	g := func() StackTrace { return f() }

	stack := g()[:3]

	tests := []struct {
		format string
		result string
	}{
		{
			format: "%s",
			result: `[stack_test.go stack_test.go stack_test.go]`,
		},
		{
			format: "%v",
			result: `[stack_test.go:12 stack_test.go:13 stack_test.go:15]`,
		},
		{
			format: "%+v",
			result: `[github.com/segmentio/errors-go/stack_test.go:12 github.com/segmentio/errors-go/stack_test.go:13 github.com/segmentio/errors-go/stack_test.go:15]`,
		},
		{
			format: "%#v",
			result: `
github.com/segmentio/errors-go.TestFormatStackTrace.func1
	` + cwd + `/stack_test.go:12
github.com/segmentio/errors-go.TestFormatStackTrace.func2
	` + cwd + `/stack_test.go:13
github.com/segmentio/errors-go.TestFormatStackTrace
	` + cwd + `/stack_test.go:15`,
		},
	}

	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			if s := fmt.Sprintf(test.format, stack); s != test.result {
				t.Error("bad result:")
				t.Log("expected:\n", test.result)
				t.Log("found:\n", s)
			}
		})
	}
}

func TestFormatStackFrame(t *testing.T) {
	f := func() StackTrace { return CaptureStackTrace(0) } // line 57
	g := func() StackTrace { return f() }

	stack := g()[:3]

	tests := []struct {
		args   []interface{}
		format string
		result string
	}{
		{
			args:   []interface{}{stack[0]},
			format: "%s",
			result: `stack_test.go`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%d",
			result: `57`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%n",
			result: `TestFormatStackFrame.func1`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%+n",
			result: `errors-go.TestFormatStackFrame.func1`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%#n",
			result: `github.com/segmentio/errors-go.TestFormatStackFrame.func1`,
		},
		{
			args:   []interface{}{stack[0], stack[0], stack[0]},
			format: "%s:%d:%n",
			result: `stack_test.go:57:TestFormatStackFrame.func1`,
		},
		{
			args:   []interface{}{stack[1], stack[1], stack[1]},
			format: "%s:%d:%n",
			result: `stack_test.go:58:TestFormatStackFrame.func2`,
		},
		{
			args:   []interface{}{stack[2], stack[2], stack[2]},
			format: "%s:%d:%n",
			result: `stack_test.go:60:TestFormatStackFrame`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%+s",
			result: `github.com/segmentio/errors-go/stack_test.go`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%#s",
			result: `github.com/segmentio/errors-go.TestFormatStackFrame.func1
	` + cwd + `/stack_test.go`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%v",
			result: `stack_test.go:57`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%+v",
			result: `github.com/segmentio/errors-go/stack_test.go:57`,
		},
		{
			args:   []interface{}{stack[0]},
			format: "%#v",
			result: `github.com/segmentio/errors-go.TestFormatStackFrame.func1
	` + cwd + `/stack_test.go:57`,
		},

		{
			args: []interface{}{StackTrace{
				Frame(1),
				Frame(2),
				Frame(3),
			}},
			format: "%a",
			result: "%!a(errors.StackFrame=[0x1 0x2 0x3])",
		},
	}

	for _, test := range tests {
		t.Run(test.format, func(t *testing.T) {
			if s := fmt.Sprintf(test.format, test.args...); s != test.result {
				t.Error("bad result:")
				t.Log("expected:\n", test.result)
				t.Log("found:\n", s)
			}
		})
	}
}

func TestInvalidFrame(t *testing.T) {
	f := Frame(0)

	file, line, fn := f.source()

	if file != "" {
		t.Errorf("source file of an invalid frame must be \"\", got %q", file)
	}

	if line != 0 {
		t.Error("source line of an invalid frame must be zero, got", line)
	}

	if fn != "" {
		t.Errorf("source function of an invalid frame must be \"\", got %q", fn)
	}

	if s := fmt.Sprintf("%#s", f); s != "(unknown)\n\t0x0" {
		t.Error("bad string representation of invalid frame:", s)
	}
}

func TestFormatFrameAddress(t *testing.T) {
	f := Frame(0x123456789A)

	tests := []struct {
		f string
		s string
	}{
		{f: "%x", s: "123456789a"},
		{f: "%#x", s: "0x123456789a"},
		{f: "%X", s: "123456789A"},
		{f: "%#X", s: "0X123456789A"},
		{f: "%a", s: "%!a(errors.Frame=0x123456789a)"},
	}

	for _, test := range tests {
		t.Run(test.f, func(t *testing.T) {
			if s := fmt.Sprintf(test.f, f); s != test.s {
				t.Error("bad string representation")
				t.Log("expected:", test.s)
				t.Log("found:   ", s)
			}
		})
	}
}

func getwd() string {
	path, _ := os.Getwd()
	return path
}
