// +build go1.7

package errors

import "runtime"

func fileLineFunc(pc uintptr) (file string, line int, name string) {
	caller := [1]uintptr{pc}
	frames := runtime.CallersFrames(caller[:])

	if frames != nil {
		f, _ := frames.Next()
		file = f.File
		line = f.Line
		name = f.Func.Name()
	}

	return
}
