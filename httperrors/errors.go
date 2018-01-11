package httperrors

import (
	"net/http"

	errors "github.com/segmentio/errors-go"
)

// New constructs an error from a HTTP response.
//
// The error returned by this function capture the stack trace, is tagged with
// the method, scheme, host, and path of the original request that this response
// was obtained from (taken from the Request field of the given response).
//
// The type of the error is set to the status of the response, for example a 404
// Not Found error will return an error of type "NotFound".
//
// It also may carry three other high-level types, "Temporary", "Timeout", and
// "Throttled" which are deducted from the status of the response, for example
// a 408 Request Timeout status will construct an error of type "Timeout".
//
// Note that the response body is left untouched, so the program still has to
// close it at some point.
//
// Here's an example of common use of the function:
//
//	r, err := http.Get("http://localhost:1234/")
//	if err != nil {
//		return errors.Adapt(err)
//	}
//	defer r.Body.Close()
//	if r.StatusCode >= 300 {
//		return httperrors.New(r)
//	}
//	// ...
//
func New(res *http.Response) error {
	return newHTTPError(res, errors.CaptureStackTrace(1))
}

// Wrap is similar to New but takes an extra error argument so it can be used
// on the returned values from functions like http.Get or http.(*Client).Do.
//
// The function considers all responses with status codes equal or greater than
// 300 to be errors and in this case it closes the response body and returns an
// error.
//
// If err is not nil, the function returns the result of calling errors.Adapt
// with err as argument.
//
// Here's an example of common use of the function:
//
//	r, err := httperrors.Wrap(http.Get("http://localhost:1234/"))
//  if err != nil {
//		return err
//	}
//
func Wrap(res *http.Response, err error) (*http.Response, error) {
	if err != nil {
		return nil, errors.WithStackTrace(err, errors.CaptureStackTrace(1))
	}

	if res.StatusCode >= 300 {
		res.Body.Close()
		return nil, newHTTPError(res, errors.CaptureStackTrace(1))
	}

	return res, nil
}

type httpError struct {
	code   int
	status string
	method string
	scheme string
	host   string
	path   string
	tags   []errors.Tag
	stack  errors.StackTrace
}

func newHTTPError(res *http.Response, stack errors.StackTrace) *httpError {
	e := &httpError{
		code:   res.StatusCode,
		status: res.Status,
		stack:  stack,
	}

	if req := res.Request; req != nil {
		e.method = req.Method
		e.scheme = req.URL.Scheme
		e.host = req.URL.Host
		e.path = req.URL.Path

		if host := req.Header.Get("Host"); len(host) != 0 {
			e.host = host
		}

		e.tags = []errors.Tag{
			{"method", e.method},
			{"scheme", e.scheme},
			{"host", e.host},
			{"path", e.path},
		}
	}

	return e
}

func (e *httpError) Error() string {
	b := make([]byte, 0, len(e.method)+len(e.scheme)+len(e.host)+len(e.path)+len(e.status)+6)

	if len(e.method) != 0 {
		b = append(b, e.method...)
		b = append(b, ' ')
	}

	if len(e.scheme) != 0 {
		b = append(b, e.scheme...)
		b = append(b, ':', '/', '/')
	}

	b = append(b, e.host...)
	b = append(b, e.path...)

	if len(b) != 0 {
		b = append(b, ':', ' ')
	}

	b = append(b, e.status...)
	return string(b)
}

func (e *httpError) Tags() []errors.Tag {
	return e.tags
}

func (e *httpError) StackTrace() errors.StackTrace {
	return e.stack
}

func (e *httpError) Temporary() bool {
	return e.Timeout() ||
		e.Throttled() ||
		e.PaymentRequired() ||
		e.Found() ||
		e.SeeOther() ||
		e.NotModified() ||
		e.TemporaryRedirect() ||
		e.NotFound() ||
		e.InternalServerError() ||
		e.NotImplemented() ||
		e.BadGateway() ||
		e.ServiceUnavailable() ||
		e.InsufficientStorage()
}

func (e *httpError) Timeout() bool {
	return e.RequestTimeout() || e.GatewayTimeout()
}

func (e *httpError) Throttled() bool {
	return e.TooManyRequests()
}

// 1xx
func (e *httpError) Continue() bool           { return e.is(http.StatusContinue) }
func (e *httpError) SwitchingProtocols() bool { return e.is(http.StatusSwitchingProtocols) }
func (e *httpError) Processing() bool         { return e.is(http.StatusProcessing) }

// 2xx
func (e *httpError) OK() bool                   { return e.is(http.StatusOK) }
func (e *httpError) Created() bool              { return e.is(http.StatusCreated) }
func (e *httpError) Accepted() bool             { return e.is(http.StatusAccepted) }
func (e *httpError) NonAuthoritativeInfo() bool { return e.is(http.StatusNonAuthoritativeInfo) }
func (e *httpError) NoContent() bool            { return e.is(http.StatusNoContent) }
func (e *httpError) ResetContent() bool         { return e.is(http.StatusResetContent) }
func (e *httpError) PartialContent() bool       { return e.is(http.StatusPartialContent) }
func (e *httpError) MultiStatus() bool          { return e.is(http.StatusMultiStatus) }
func (e *httpError) AlreadyReported() bool      { return e.is(http.StatusAlreadyReported) }
func (e *httpError) IMUsed() bool               { return e.is(http.StatusIMUsed) }

// 3xx
func (e *httpError) MultipleChoices() bool   { return e.is(http.StatusMultipleChoices) }
func (e *httpError) MovedPermanently() bool  { return e.is(http.StatusMovedPermanently) }
func (e *httpError) Found() bool             { return e.is(http.StatusFound) }
func (e *httpError) SeeOther() bool          { return e.is(http.StatusSeeOther) }
func (e *httpError) NotModified() bool       { return e.is(http.StatusNotModified) }
func (e *httpError) UseProxy() bool          { return e.is(http.StatusUseProxy) }
func (e *httpError) TemporaryRedirect() bool { return e.is(http.StatusTemporaryRedirect) }
func (e *httpError) PermanentRedirect() bool { return e.is(http.StatusPermanentRedirect) }

// 4xx
func (e *httpError) BadRequest() bool            { return e.is(http.StatusBadRequest) }
func (e *httpError) Unauthorized() bool          { return e.is(http.StatusUnauthorized) }
func (e *httpError) PaymentRequired() bool       { return e.is(http.StatusPaymentRequired) }
func (e *httpError) Forbidden() bool             { return e.is(http.StatusForbidden) }
func (e *httpError) NotFound() bool              { return e.is(http.StatusNotFound) }
func (e *httpError) MethodNotAllowed() bool      { return e.is(http.StatusMethodNotAllowed) }
func (e *httpError) NotAcceptable() bool         { return e.is(http.StatusNotAcceptable) }
func (e *httpError) ProxyAuthRequired() bool     { return e.is(http.StatusProxyAuthRequired) }
func (e *httpError) RequestTimeout() bool        { return e.is(http.StatusRequestTimeout) }
func (e *httpError) Conflict() bool              { return e.is(http.StatusConflict) }
func (e *httpError) Gone() bool                  { return e.is(http.StatusGone) }
func (e *httpError) LengthRequired() bool        { return e.is(http.StatusLengthRequired) }
func (e *httpError) PreconditionFailed() bool    { return e.is(http.StatusPreconditionFailed) }
func (e *httpError) RequestEntityTooLarge() bool { return e.is(http.StatusRequestEntityTooLarge) }
func (e *httpError) RequestURITooLong() bool     { return e.is(http.StatusRequestURITooLong) }
func (e *httpError) UnsupportedMediaType() bool  { return e.is(http.StatusUnsupportedMediaType) }
func (e *httpError) RequestedRangeNotSatisfiable() bool {
	return e.is(http.StatusRequestedRangeNotSatisfiable)
}
func (e *httpError) ExpectationFailed() bool    { return e.is(http.StatusExpectationFailed) }
func (e *httpError) Teapot() bool               { return e.is(http.StatusTeapot) }
func (e *httpError) UnprocessableEntity() bool  { return e.is(http.StatusUnprocessableEntity) }
func (e *httpError) Locked() bool               { return e.is(http.StatusLocked) }
func (e *httpError) FailedDependency() bool     { return e.is(http.StatusFailedDependency) }
func (e *httpError) UpgradeRequired() bool      { return e.is(http.StatusUpgradeRequired) }
func (e *httpError) PreconditionRequired() bool { return e.is(http.StatusPreconditionRequired) }
func (e *httpError) TooManyRequests() bool      { return e.is(http.StatusTooManyRequests) }
func (e *httpError) RequestHeaderFieldsTooLarge() bool {
	return e.is(http.StatusRequestHeaderFieldsTooLarge)
}
func (e *httpError) UnavailableForLegalReasons() bool {
	return e.is(http.StatusUnavailableForLegalReasons)
}

// 5xx
func (e *httpError) InternalServerError() bool     { return e.is(http.StatusInternalServerError) }
func (e *httpError) NotImplemented() bool          { return e.is(http.StatusNotImplemented) }
func (e *httpError) BadGateway() bool              { return e.is(http.StatusBadGateway) }
func (e *httpError) ServiceUnavailable() bool      { return e.is(http.StatusServiceUnavailable) }
func (e *httpError) GatewayTimeout() bool          { return e.is(http.StatusGatewayTimeout) }
func (e *httpError) HTTPVersionNotSupported() bool { return e.is(http.StatusHTTPVersionNotSupported) }
func (e *httpError) VariantAlsoNegotiates() bool   { return e.is(http.StatusVariantAlsoNegotiates) }
func (e *httpError) InsufficientStorage() bool     { return e.is(http.StatusInsufficientStorage) }
func (e *httpError) LoopDetected() bool            { return e.is(http.StatusLoopDetected) }
func (e *httpError) NotExtended() bool             { return e.is(http.StatusNotExtended) }
func (e *httpError) NetworkAuthenticationRequired() bool {
	return e.is(http.StatusNetworkAuthenticationRequired)
}

func (e *httpError) is(code int) bool { return e.code == code }
