package pkgerrors

import errors "github.com/segmentio/errors-go"

func init() {
	errors.Register(errors.AdapterFunc(Adapt))
}
