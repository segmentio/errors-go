package errors

import (
	"runtime"
	"testing"
)

func TestSourceForPC(t *testing.T) {
	pc := [1]uintptr{}
	runtime.Callers(1, pc[:])

	file, line, name := sourceForPC(pc[0])

	if file != "github.com/segmentio/errors-go/source_test.go" {
		t.Error("bad file:", file)
	}

	if line != 10 {
		t.Error("bad line:", line)
	}

	if name != "github.com/segmentio/errors-go.TestSourceForPC" {
		t.Error("bad name:", name)
	}
}
