// +build !go1.7

package errors

import "runtime"

func fileLineFunc(pc uintptr) (file string, line int, name string) {
	if fn := runtime.FuncForPC(pc); fn != nil {
		file, line = fn.FileLine(pc)
		name = fn.Name()
	}
	return
}
