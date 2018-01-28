package twirperrors

import (
	errors "github.com/segmentio/errors-go"
	"github.com/twitchtv/twirp"
)

func Adapt(err error) (error, bool) {
	if e, ok := err.(twirp.Error); ok {
		return &twirpError{cause: e}, true
	}
	return err, false
}

type twirpError struct {
	cause twirp.Error
}

func (e *twirpError) Cause() error { return e.cause }

func (e *twirpError) Error() string { return e.cause.Error() }

func (e *twirpError) Message() string { return e.cause.Msg() }

func (e *twirpError) Tags() []errors.Tag {
	meta := e.cause.MetaMap()
	tags := make([]errors.Tag, 0, len(meta))

	for name, value := range meta {
		tags = append(tags, errors.Tag{
			Name:  name,
			Value: value,
		})
	}

	return tags
}

// Twirp-specific error types

func (e *twirpError) Canceled() bool { return e.is(twirp.Canceled) }

func (e *twirpError) Unknown() bool { return e.is(twirp.Unknown) }

func (e *twirpError) InvalidArgument() bool { return e.is(twirp.InvalidArgument) }

func (e *twirpError) DeadlineExceeded() bool { return e.is(twirp.DeadlineExceeded) }

func (e *twirpError) NotFound() bool { return e.is(twirp.NotFound) }

func (e *twirpError) BadRoute() bool { return e.is(twirp.BadRoute) }

func (e *twirpError) AlreadyExists() bool { return e.is(twirp.AlreadyExists) }

func (e *twirpError) PermissionDenied() bool { return e.is(twirp.PermissionDenied) }

func (e *twirpError) Unauthenticated() bool { return e.is(twirp.Unauthenticated) }

func (e *twirpError) ResourceExhausted() bool { return e.is(twirp.ResourceExhausted) }

func (e *twirpError) FailedPrecondition() bool { return e.is(twirp.FailedPrecondition) }

func (e *twirpError) Aborted() bool { return e.is(twirp.Aborted) }

func (e *twirpError) OutOfRange() bool { return e.is(twirp.OutOfRange) }

func (e *twirpError) Unimplemented() bool { return e.is(twirp.Unimplemented) }

func (e *twirpError) Internal() bool { return e.is(twirp.Internal) }

func (e *twirpError) Unavailable() bool { return e.is(twirp.Unavailable) }

func (e *twirpError) DataLoss() bool { return e.is(twirp.DataLoss) }

func (e *twirpError) is(code twirp.ErrorCode) bool { return e.cause.Code() == code }

// Common error types

func (e *twirpError) Conflict() bool { return e.AlreadyExists() }

func (e *twirpError) Throttled() bool { return e.ResourceExhausted() }

func (e *twirpError) Timeout() bool { return e.Canceled() || e.DeadlineExceeded() }

func (e *twirpError) Validation() bool { return e.InvalidArgument() || e.OutOfRange() || e.BadRoute() }

func (e *twirpError) Temporary() bool {
	return e.Timeout() ||
		e.Throttled() ||
		e.Unimplemented() ||
		e.Internal() ||
		e.Unavailable()
}
