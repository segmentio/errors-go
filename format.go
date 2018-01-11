package errors

import (
	"fmt"
	"io"
	"strings"
)

// format is the implementation of a generic error formatting functions which
// supports the following verbs:
//
//    %s    writes the value returned by calling .Error on the the error
//    %q    same as %s but quotes and escapes the string
//    %v    prints the error messages in a tree-like format
//
// format accepts flags that alter the printing of some verbs, as follows:
//
//    %+v   similar to %v but prints the stack traces below the messages
//    %#v   prints errors with their messages and causes in  Go-like syntax
func format(s fmt.State, v rune, err error) {
	switch v {
	case 's':
		io.WriteString(s, err.Error())

	case 'q':
		fmt.Fprintf(s, "%q", err.Error())

	case 'v':
		if s.Flag('#') {
			f := goformatter{state: s}
			f.format(err)
		} else {
			f := formatter{state: s}
			f.format(formatterContext{length: 1}, err)
		}

	default:
		fmt.Fprintf(s, "%%!%c(%T)", v, err)
	}
}

type formatterContext struct {
	index       int  // index in the parent list of causes
	length      int  // length of the parent list of causes
	needNewLine bool // whether a new line must be printed
}

func (fctx *formatterContext) last() bool {
	i := fctx.index
	n := fctx.length - 1
	return !(i < n)
}

// formatter provides the implementation of an error formatter which prints
// errors and the graph of potential causes in a style similar to the tree(1)
// command. It is used when writing errors with the "%v" and "%+v" formats.
type formatter struct {
	state  fmt.State
	indent indent
}

func (f *formatter) format(fctx formatterContext, err error) {
	msgs, types, tags, stacks, causes := inspect(err)

	if len(msgs) == 0 {
		msgs = []string{"."}
	}

	f.writeNode(fctx, msgs, types, tags, stacks)
	f.indent.push(fctx)
	defer f.indent.pop()

	fctx.length = len(causes)
	fctx.needNewLine = true

	for i, cause := range causes {
		fctx.index = i
		f.format(fctx, cause)
	}
}

func (f *formatter) writeNewLine(fctx formatterContext) {
	f.writeString("\n")
	f.indent.nextLine(fctx)
}

func (f *formatter) writeString(s string) {
	io.WriteString(f.state, s)
}

func (f *formatter) writeIndent() {
	f.indent.writeTo(f.state)
}

func (f *formatter) writeNode(fctx formatterContext, msgs []string, types []string, tags []Tag, stacks []StackTrace) {
	if fctx.needNewLine {
		f.writeNewLine(fctx)
	}

	f.indent.nextNode(fctx)
	lines := strings.Split(strings.Join(msgs, ": "), "\n")

	for i, line := range lines {
		if i != 0 {
			f.writeNewLine(fctx)
		}
		f.writeIndent()
		f.writeString(line)
	}

	f.writeTypes(types)
	f.writeTags(tags)

	if f.state.Flag('+') {
		f.writeStacks(fctx, stacks)
	}
}

func (f *formatter) writeTypes(types []string) {
	if len(types) != 0 {
		f.writeString(" (")

		for i, t := range types {
			if i != 0 {
				f.writeString(" ")
			}
			f.writeString(t)
		}

		f.writeString(")")
	}
}

func (f *formatter) writeTags(tags []Tag) {
	if len(tags) != 0 {
		f.writeString(" [")

		for i, t := range tags {
			if i != 0 {
				f.writeString(" ")
			}
			fmt.Fprintf(f.state, "%s:%q", t.Name, t.Value)
		}

		f.writeString("]")
	}
}

func (f *formatter) writeStacks(fctx formatterContext, stacks []StackTrace) {
	for i, stack := range stacks {
		if i != 0 {
			f.writeNewLine(fctx)
			f.writeIndent()
		}
		for _, frame := range stack {
			f.writeFrame(fctx, frame)
		}
		f.writeNewLine(fctx)
		f.writeIndent()
	}
}

func (f *formatter) writeFrame(fctx formatterContext, frame Frame) {
	f.writeFrameFunc(fctx, frame)
	f.writeFrameFile(fctx, frame)
}

func (f *formatter) writeFrameFunc(fctx formatterContext, frame Frame) {
	f.writeNewLine(fctx)
	f.writeIndent()
	fmt.Fprintf(f.state, "%#n", frame)
}

func (f *formatter) writeFrameFile(fctx formatterContext, frame Frame) {
	f.writeNewLine(fctx)
	f.writeIndent()
	fmt.Fprintf(f.state, "\t%+s:%d", frame, frame)
}

// goformatter is an error formatter which prints errors with their message and
// causes in a Go-like syntax. It is used when writing errors with the "%#v"
// format.
type goformatter struct {
	state fmt.State
}

func (f *goformatter) format(err error) {
	var msg string
	if e, ok := err.(errorMessage); ok {
		msg = e.Message()
	}

	f.print("%T{msg:%q", err, msg)

	switch e := err.(type) {
	case errorCauses:
		f.print(" causes:[")

		for i, cause := range e.Causes() {
			if i != 0 {
				f.print(" ")
			}
			f.format(cause)
		}

		f.print("]")

	case errorCause:
		f.print(" cause:")
		f.format(e.Cause())
	}

	f.print("}")
}

func (f *goformatter) print(s string, a ...interface{}) {
	fmt.Fprintf(f.state, s, a...)
}

// indent is a helper type used to format a tree-like representation that
// supports multi-line nodes.
type indent struct {
	symbols []string
}

func (i *indent) push(fctx formatterContext) {
	i.nextLine(fctx)
	i.symbols = append(i.symbols, "")
}

func (i *indent) pop() {
	i.symbols = i.symbols[:i.lastIndex()]
}

func (i *indent) nextNode(fctx formatterContext) {
	if fctx.last() {
		i.set("└── ")
	} else {
		i.set("├── ")
	}
}

func (i *indent) nextLine(fctx formatterContext) {
	if fctx.last() {
		i.set("    ")
	} else {
		i.set("|   ")
	}
}

func (i *indent) writeTo(w io.Writer) {
	for _, symbol := range i.symbols {
		io.WriteString(w, symbol)
	}
}

func (i *indent) lastIndex() int {
	return len(i.symbols) - 1
}

func (i *indent) empty() bool {
	return len(i.symbols) == 0
}

func (i *indent) set(s string) {
	if !i.empty() {
		i.symbols[i.lastIndex()] = s
	}
}
