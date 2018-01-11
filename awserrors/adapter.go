package awserrors

import (
	"github.com/aws/aws-sdk-go/aws/awserr"
	errors "github.com/segmentio/errors-go"
)

// Adapt checks the type of err and if it matches one of the error types or one
// of the error values of the standard net package, adapts it to make error
// types discoverable using the errors.Is function.
//
// This function is automatically installed as a global adapter when importing
// the neterrors package, a program likely should use errors.Adapt instead of
// calling this adapter directly.
func Adapt(err error) (error, bool) {
	switch e := err.(type) {
	case awserr.Error:
		return &awsError{e}, true

	case awserr.BatchError:
		return &awsBatchError{e}, true

	default:
		return err, false
	}
}

func adaptErrors(errs []error) []error {
	adapted := make([]error, 0, len(errs))

	for _, e := range errs {
		if e != nil {
			adapted = append(adapted, errors.Adapt(e))
		}
	}

	return adapted
}

type awsError struct {
	cause awserr.Error
}

func (e *awsError) Error() string {
	return e.cause.Error()
}

func (e *awsError) Code() string {
	return e.cause.Code()
}

func (e *awsError) Message() string {
	return e.cause.Message()
}

func (e *awsError) Cause() error {
	return errors.Adapt(e.cause.OrigErr())
}

func (e *awsError) Causes() []error {
	if b, ok := e.cause.(awserr.BatchedErrors); ok {
		return adaptErrors(b.OrigErrs())
	}
	if cause := e.Cause(); cause != nil {
		return []error{cause}
	}
	return nil
}

func (e *awsError) StatusCode() int {
	if r, ok := e.cause.(awserr.RequestFailure); ok {
		return r.StatusCode()
	}
	return 0
}

func (e *awsError) RequestID() string {
	if r, ok := e.cause.(awserr.RequestFailure); ok {
		return r.RequestID()
	}
	return ""
}

type awsBatchError struct {
	cause awserr.BatchError
}

func (e *awsBatchError) Error() string {
	return e.cause.Error()
}

func (e *awsBatchError) Code() string {
	return e.cause.Code()
}

func (e *awsBatchError) Message() string {
	return e.cause.Message()
}

func (e *awsBatchError) Causes() []error {
	return adaptErrors(e.cause.OrigErrs())
}
