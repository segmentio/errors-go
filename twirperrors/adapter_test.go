package twirperrors

import (
	"testing"

	errors "github.com/segmentio/errors-go"
	"github.com/segmentio/errors-go/errorstest"
	"github.com/twitchtv/twirp"
)

func TestAdapt(t *testing.T) {
	errorstest.TestAdapter(t, errors.AdapterFunc(Adapt),
		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Canceled, ""),
			Types: []string{"Canceled", "Temporary", "Timeout"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Unknown, ""),
			Types: []string{"Unknown"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.InvalidArgument, ""),
			Types: []string{"InvalidArgument", "Validation"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.DeadlineExceeded, ""),
			Types: []string{"DeadlineExceeded", "Temporary", "Timeout"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.NotFound, ""),
			Types: []string{"NotFound"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.BadRoute, ""),
			Types: []string{"BadRoute", "Validation"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.AlreadyExists, ""),
			Types: []string{"AlreadyExists", "Conflict"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.PermissionDenied, ""),
			Types: []string{"PermissionDenied"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Unauthenticated, ""),
			Types: []string{"Unauthenticated"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.ResourceExhausted, ""),
			Types: []string{"ResourceExhausted", "Temporary", "Throttled"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.FailedPrecondition, ""),
			Types: []string{"FailedPrecondition"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Aborted, ""),
			Types: []string{"Aborted"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.OutOfRange, ""),
			Types: []string{"OutOfRange", "Validation"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Unimplemented, ""),
			Types: []string{"Temporary", "Unimplemented"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Internal, ""),
			Types: []string{"Internal", "Temporary"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.Unavailable, ""),
			Types: []string{"Temporary", "Unavailable"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.DataLoss, ""),
			Types: []string{"DataLoss"},
		},

		errorstest.AdapterTest{
			Error: twirp.NewError(twirp.NotFound, "").WithMeta("hello", "world").WithMeta("twitch", "tv"),
			Types: []string{"NotFound"},
			Tags: []errors.Tag{
				{Name: "hello", Value: "world"},
				{Name: "twitch", Value: "tv"},
			},
		},

		errorstest.AdapterTest{
			Error:   twirp.NewError(twirp.NotFound, "hello world!"),
			Message: "hello world!",
			Types:   []string{"NotFound"},
		},
	)
}
