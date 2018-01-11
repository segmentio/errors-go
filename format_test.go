package errors

import (
	"bytes"
	"fmt"
	"testing"
)

func TestIndent(t *testing.T) {
	b := &bytes.Buffer{}
	i := indent{}

	c0 := formatterContext{length: 1}
	c1 := formatterContext{}

	b.WriteString("A\n")
	i.push(c0)

	c0.index, c0.length = 0, 4
	i.nextNode(c0)
	i.writeTo(b)
	b.WriteString("B\n")

	c0.index++
	i.nextNode(c0)
	i.writeTo(b)
	b.WriteString("C\n")

	c0.index++
	i.nextNode(c0)
	i.writeTo(b)
	b.WriteString("D\n")
	i.push(c0)

	c1.index, c1.length = 0, 2
	i.nextNode(c1)
	i.writeTo(b)
	b.WriteString("E\n")

	i.nextLine(c1)
	i.writeTo(b)
	b.WriteString("F\n")

	i.nextLine(c1)
	i.writeTo(b)
	b.WriteString("G\n")

	c1.index++
	i.nextNode(c1)
	i.writeTo(b)
	b.WriteString("H\n")
	i.pop()

	c0.index++
	i.nextNode(c0)
	i.writeTo(b)
	b.WriteString("I\n")

	i.nextLine(c0)
	i.writeTo(b)
	b.WriteString("J\n")
	i.pop()

	s := b.String()
	r := `A
├── B
├── C
├── D
|   ├── E
|   |   F
|   |   G
|   └── H
└── I
    J
`

	if s != r {
		t.Error("bad indented tree representation:", s)
	}
}

func TestFormat(t *testing.T) {
	tests := []struct {
		error  error
		format string
		string string
	}{
		{
			error:  nil,
			format: "%s",
			string: "%!s(<nil>)",
		},

		{
			error:  nil,
			format: "%q",
			string: "%!q(<nil>)",
		},

		{
			error:  nil,
			format: "%T",
			string: "<nil>",
		},

		{
			error:  nil,
			format: "%v",
			string: "<nil>",
		},

		{
			error:  nil,
			format: "%+v",
			string: "<nil>",
		},

		{
			error:  nil,
			format: "%#v",
			string: "<nil>",
		},

		{
			error:  New("hello world"),
			format: "%d",
			string: "%!d(*errors.baseError)",
		},

		{
			error:  New("hello world"),
			format: "%s",
			string: "hello world",
		},

		{
			error:  New("hello world"),
			format: "%q",
			string: `"hello world"`,
		},

		{
			error:  New("hello world"),
			format: "%T",
			string: "*errors.baseError",
		},

		{
			error:  New("hello world"),
			format: "%v",
			string: "hello world",
		},

		{
			error:  WithMessage(New("hello world"), "answer 42"),
			format: "%v",
			string: "answer 42: hello world",
		},

		{
			error: WithMessage(
				Join(
					Wrap(TODO, "A\n\ttest multi-line messages"),
					New("B"),
					New("C"),
				),
				"answer 42",
			),
			format: "%v",
			string: `answer 42
├── A
|   	test multi-line messages: TODO
├── B
└── C`,
		},

		{
			error: WithMessage(
				Join(
					Join(
						Wrap(TODO, "A.1"),
						New("A.2"),
						New("A.3"),
					),
					New("B"),
					New("C"),
				),
				"answer 42",
			),
			format: "%v",
			string: `answer 42
├── .
|   ├── A.1: TODO
|   ├── A.2
|   └── A.3
├── B
└── C`,
		},

		{
			error: WithMessage(
				Join(
					Join(
						Wrap(TODO, "A.1"),
						New("A.2"),
						WithTypes(New("A.3"), "Timeout", "Temporary"),
					),
					New("B"),
					WithTags(New("C"), T("operation", "seek"), T("env", "production")),
				),
				"answer 42",
			),
			format: "%+v",
		},

		{
			error: WithMessage(
				Join(
					Wrap(TODO, "A"),
					New("B"),
					New("C"),
				),
				"answer 42",
			),
			format: "%#v",
			string: `*errors.errorWithMessage{msg:"answer 42" cause:*errors.multiError{msg:"" causes:[*errors.errorWithMessage{msg:"A" cause:*errors.errorWithStack{msg:"" cause:*errors.errorTODO{msg:""}}} *errors.baseError{msg:"B"} *errors.baseError{msg:"C"}]}}`,
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("format(%s)", test.format), func(t *testing.T) {
			s := fmt.Sprintf(test.format, test.error)

			switch {
			case len(test.string) == 0:
				// Stack traces going into the Go standard library may change
				// from version to version, for now let's just print it out to
				// see it work and exercise the code path (the code for printing
				// out stack traces is tested in stack_test.go already anyway).
				t.Log(s)

			case s != test.string:
				t.Error("bad string:")
				t.Logf("expected: %s", test.string)
				t.Logf("found:    %s", s)
			}
		})
	}
}
