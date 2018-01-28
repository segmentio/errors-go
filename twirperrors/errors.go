package twirperrors

import (
	"strings"

	errors "github.com/segmentio/errors-go"
	"github.com/twitchtv/twirp"
)

// New constructs a twirp error from another error. The error code is guessed by
// inspecting the types of err, and defaults to twirp.Unknown if the error had
// no types.
//
// If err is nil the function returns nil.
func New(err error) twirp.Error {
	if err == nil {
		return nil
	}

	msgs, types, tags, _, _ := errors.Inspect(err)

	for _, typ := range types {
		switch code := twirp.ErrorCode(typ); code {
		case "Canceled":
			return newError(twirp.Canceled, msgs, tags)

		case "Unknown":
			return newError(twirp.Unknown, msgs, tags)

		case "InvalidArgument":
			return newError(twirp.InvalidArgument, msgs, tags)

		case "DeadlineExceeded":
			return newError(twirp.DeadlineExceeded, msgs, tags)

		case "NotFound":
			return newError(twirp.NotFound, msgs, tags)

		case "BadRoute":
			return newError(twirp.BadRoute, msgs, tags)

		case "AlreadyExists":
			return newError(twirp.AlreadyExists, msgs, tags)

		case "PermissionDenied":
			return newError(twirp.PermissionDenied, msgs, tags)

		case "Unauthenticated":
			return newError(twirp.Unauthenticated, msgs, tags)

		case "ResourceExhausted":
			return newError(twirp.ResourceExhausted, msgs, tags)

		case "FailedPrecondition":
			return newError(twirp.FailedPrecondition, msgs, tags)

		case "Aborted":
			return newError(twirp.Aborted, msgs, tags)

		case "OutOfRange":
			return newError(twirp.OutOfRange, msgs, tags)

		case "Unimplemented":
			return newError(twirp.Unimplemented, msgs, tags)

		case "Internal":
			return newError(twirp.Internal, msgs, tags)

		case "Unavailable":
			return newError(twirp.Unavailable, msgs, tags)

		case "DataLoss":
			return newError(twirp.DataLoss, msgs, tags)
		}
	}

	return newError(twirp.Unknown, msgs, tags)
}

func newError(code twirp.ErrorCode, msgs []string, tags []errors.Tag) twirp.Error {
	twerr := twirp.NewError(code, strings.Join(msgs, ": "))
	for _, tag := range tags {
		twerr = twerr.WithMeta(tag.Name, tag.Value)
	}
	return twerr
}
