package errors

import (
	"fmt"
	"io"
	"path"
	"runtime"
	"strings"
	"sync/atomic"
)

// Frame represents a program counter inside a stack frame.
type Frame uintptr

// pc returns the program counter for this frame; multiple frames may have the
// same PC value.
func (f Frame) pc() uintptr { return uintptr(f) }

func (f Frame) source() (string, int, string) {
	return sourceForPC(f.pc())
}

func (f Frame) file() string {
	file, _, _ := f.source()
	return file
}

func (f Frame) line() int {
	_, line, _ := f.source()
	return line
}

func (f Frame) name() string {
	_, _, name := f.source()
	return name
}

// Format formats the frame according to the fmt.Formatter interface.
//
//    %s    source file
//    %d    source line
//    %n    function name
//    %v    equivalent to %s:%d
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+s   path of source file relative to the compile time GOPATH
//    %#s   function name and path of source file
//    %+n   function name prefixed by its package name
//    %#n   function name prefixed by its full package path
//    %+v   equivalent to %+s:%d
//    %#v   equivalent to %#s:%d
func (f Frame) Format(s fmt.State, verb rune) {
	pc := f.pc()

	switch verb {
	case 's':
		switch {
		case s.Flag('#'):
			fn := runtime.FuncForPC(pc)
			if fn == nil {
				fmt.Fprintf(s, "(unknown)\n\t%#x", pc)
			} else {
				file, _ := fn.FileLine(pc)
				fmt.Fprintf(s, "%s\n\t%s", fn.Name(), file)
			}
		case s.Flag('+'):
			io.WriteString(s, f.file())
		default:
			io.WriteString(s, path.Base(f.file()))
		}

	case 'd':
		fmt.Fprintf(s, "%d", f.line())

	case 'n':
		funcName := f.name()
		switch {
		case s.Flag('#'):
		case s.Flag('+'):
			funcName = longFuncName(funcName)
		default:
			funcName = shortFuncName(funcName)
		}
		io.WriteString(s, funcName)

	case 'v':
		f.Format(s, 's')
		io.WriteString(s, ":")
		f.Format(s, 'd')

	case 'x':
		switch {
		case s.Flag('#'):
			fmt.Fprintf(s, "%#x", pc)
		default:
			fmt.Fprintf(s, "%x", pc)
		}

	case 'X':
		switch {
		case s.Flag('#'):
			fmt.Fprintf(s, "%#X", pc)
		default:
			fmt.Fprintf(s, "%X", pc)
		}

	default:
		fmt.Fprintf(s, "%%!%c(errors.Frame=%#x)", verb, pc)
	}
}

// StackTrace is stack of Frames from innermost (newest) to outermost (oldest).
type StackTrace []Frame

// Format formats the stack of Frames according to the fmt.Formatter interface.
//
//    %s	lists source files for each Frame in the stack
//    %v	lists the source file and line number for each Frame in the stack
//
// Format accepts flags that alter the printing of some verbs, as follows:
//
//    %+v   Prints filename, function, and line number for each Frame in the stack.
func (st StackTrace) Format(s fmt.State, verb rune) {
	switch verb {
	case 'v':
		switch {
		case s.Flag('#'):
			for _, f := range st {
				fmt.Fprintf(s, "\n%#v", f)
			}
		case s.Flag('+'):
			fmt.Fprintf(s, "%+v", []Frame(st))
		default:
			fmt.Fprintf(s, "%v", []Frame(st))
		}

	case 's':
		fmt.Fprintf(s, "%s", []Frame(st))

	default:
		fmt.Fprintf(s, "%%!%c(errors.StackFrame=%#x)", verb, []Frame(st))
	}
}

func shortFuncName(name string) string {
	name = longFuncName(name)
	if i := strings.Index(name, "."); i >= 0 {
		name = name[i+1:]
	}
	return name
}

func longFuncName(name string) string {
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// CaptureStackTrace walks the call stack that led to this function and records
// it as a StackTrace value. The skip argument is the number of stacks frames to
// skip, the frame for captureStackTrace is never included in the returned trace.
func CaptureStackTrace(skip int) StackTrace {
	frames := make([]uintptr, 100)
	length := runtime.Callers(skip+2, frames[:])

	if init := initializing(); init {
		for _, f := range frames {
			if init = strings.HasPrefix(shortFuncName(Frame(f).name()), "init."); init {
				break
			}
		}
		if init {
			return nil
		}
		completeInitialization()
	}

	return makeStackTrace(frames[:length])
}

func makeStackTrace(frames []uintptr) StackTrace {
	stackTrace := make(StackTrace, len(frames))
	for i, pc := range frames {
		stackTrace[i] = Frame(pc)
	}
	return stackTrace
}

// Atomic variable used to check if the initialization phase is complete, so
var initialized uint32

func initializing() bool {
	return atomic.LoadUint32(&initialized) == 0
}

func completeInitialization() {
	atomic.StoreUint32(&initialized, 1)
}
