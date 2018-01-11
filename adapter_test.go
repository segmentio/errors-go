package errors

import "testing"

func TestAdapter(t *testing.T) {
	adaptable := &adaptableError{}

	Register(AdapterFunc(func(err error) (error, bool) {
		if err != adaptable {
			return err, false
		}
		return &adapterError{cause: err}, true
	}))

	err0 := New("hello")
	err1 := Adapt(adaptable)
	err2 := Adapt(err0)
	err3 := Adapt(nil)

	if msg := err1.Error(); msg != "adapted: something went wrong" {
		t.Error("wrong message found on adapted error:", msg)
	}

	if err2 != err0 {
		t.Errorf("unadaptable errors must be preserved by a call to errors.Adapt: %#v %#v", err0, err2)
	}

	if err3 != nil {
		t.Error("adapting a nil error did not return nil:", err3)
	}

	if cause := Cause(err1); cause != adaptable {
		t.Error("wrong cause exposed on the adapted error:", cause)
	}
}

type adaptableError struct{}

func (*adaptableError) Error() string { return "something went wrong" }

type adapterError struct{ cause error }

func (e *adapterError) Error() string { return "adapted: " + e.cause.Error() }
func (e *adapterError) Cause() error  { return e.cause }
