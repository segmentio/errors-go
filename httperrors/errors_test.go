package httperrors

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"reflect"
	"testing"

	errors "github.com/segmentio/errors-go"
)

func TestNew(t *testing.T) {
	tests := []struct {
		code  int
		types []string
	}{
		{
			code:  http.StatusContinue,
			types: []string{"Continue"},
		},

		{
			code:  http.StatusSwitchingProtocols,
			types: []string{"SwitchingProtocols"},
		},

		{
			code:  http.StatusProcessing,
			types: []string{"Processing"},
		},

		{
			code:  http.StatusOK,
			types: []string{"OK"},
		},

		{
			code:  http.StatusCreated,
			types: []string{"Created"},
		},

		{
			code:  http.StatusAccepted,
			types: []string{"Accepted"},
		},

		{
			code:  http.StatusNonAuthoritativeInfo,
			types: []string{"NonAuthoritativeInfo"},
		},

		{
			code:  http.StatusNoContent,
			types: []string{"NoContent"},
		},

		{
			code:  http.StatusResetContent,
			types: []string{"ResetContent"},
		},

		{
			code:  http.StatusPartialContent,
			types: []string{"PartialContent"},
		},

		{
			code:  http.StatusMultiStatus,
			types: []string{"MultiStatus"},
		},

		{
			code:  http.StatusAlreadyReported,
			types: []string{"AlreadyReported"},
		},

		{
			code:  http.StatusIMUsed,
			types: []string{"IMUsed"},
		},

		{
			code:  http.StatusMultipleChoices,
			types: []string{"MultipleChoices"},
		},

		{
			code:  http.StatusMovedPermanently,
			types: []string{"MovedPermanently"},
		},

		{
			code:  http.StatusFound,
			types: []string{"Found", "Temporary"},
		},

		{
			code:  http.StatusSeeOther,
			types: []string{"SeeOther", "Temporary"},
		},

		{
			code:  http.StatusNotModified,
			types: []string{"NotModified", "Temporary"},
		},

		{
			code:  http.StatusUseProxy,
			types: []string{"UseProxy"},
		},

		{
			code:  http.StatusTemporaryRedirect,
			types: []string{"Temporary", "TemporaryRedirect"},
		},

		{
			code:  http.StatusPermanentRedirect,
			types: []string{"PermanentRedirect"},
		},

		{
			code:  http.StatusBadRequest,
			types: []string{"BadRequest"},
		},

		{
			code:  http.StatusUnauthorized,
			types: []string{"Unauthorized"},
		},

		{
			code:  http.StatusPaymentRequired,
			types: []string{"PaymentRequired", "Temporary"},
		},

		{
			code:  http.StatusForbidden,
			types: []string{"Forbidden"},
		},

		{
			code:  http.StatusNotFound,
			types: []string{"NotFound", "Temporary"},
		},

		{
			code:  http.StatusMethodNotAllowed,
			types: []string{"MethodNotAllowed"},
		},

		{
			code:  http.StatusNotAcceptable,
			types: []string{"NotAcceptable"},
		},

		{
			code:  http.StatusProxyAuthRequired,
			types: []string{"ProxyAuthRequired"},
		},

		{
			code:  http.StatusRequestTimeout,
			types: []string{"RequestTimeout", "Temporary", "Timeout"},
		},

		{
			code:  http.StatusConflict,
			types: []string{"Conflict"},
		},

		{
			code:  http.StatusGone,
			types: []string{"Gone"},
		},

		{
			code:  http.StatusLengthRequired,
			types: []string{"LengthRequired"},
		},

		{
			code:  http.StatusPreconditionFailed,
			types: []string{"PreconditionFailed"},
		},

		{
			code:  http.StatusRequestEntityTooLarge,
			types: []string{"RequestEntityTooLarge"},
		},

		{
			code:  http.StatusRequestURITooLong,
			types: []string{"RequestURITooLong"},
		},

		{
			code:  http.StatusUnsupportedMediaType,
			types: []string{"UnsupportedMediaType"},
		},

		{
			code:  http.StatusRequestedRangeNotSatisfiable,
			types: []string{"RequestedRangeNotSatisfiable"},
		},

		{
			code:  http.StatusExpectationFailed,
			types: []string{"ExpectationFailed"},
		},

		{
			code:  http.StatusTeapot,
			types: []string{"Teapot"},
		},

		{
			code:  http.StatusUnprocessableEntity,
			types: []string{"UnprocessableEntity"},
		},

		{
			code:  http.StatusLocked,
			types: []string{"Locked"},
		},

		{
			code:  http.StatusFailedDependency,
			types: []string{"FailedDependency"},
		},

		{
			code:  http.StatusUpgradeRequired,
			types: []string{"UpgradeRequired"},
		},

		{
			code:  http.StatusPreconditionRequired,
			types: []string{"PreconditionRequired"},
		},

		{
			code:  http.StatusTooManyRequests,
			types: []string{"Temporary", "Throttled", "TooManyRequests"},
		},

		{
			code:  http.StatusRequestHeaderFieldsTooLarge,
			types: []string{"RequestHeaderFieldsTooLarge"},
		},

		{
			code:  http.StatusUnavailableForLegalReasons,
			types: []string{"UnavailableForLegalReasons"},
		},

		{
			code:  http.StatusInternalServerError,
			types: []string{"InternalServerError", "Temporary"},
		},

		{
			code:  http.StatusNotImplemented,
			types: []string{"NotImplemented", "Temporary"},
		},

		{
			code:  http.StatusBadGateway,
			types: []string{"BadGateway", "Temporary"},
		},

		{
			code:  http.StatusServiceUnavailable,
			types: []string{"ServiceUnavailable", "Temporary"},
		},

		{
			code:  http.StatusGatewayTimeout,
			types: []string{"GatewayTimeout", "Temporary", "Timeout"},
		},

		{
			code:  http.StatusHTTPVersionNotSupported,
			types: []string{"HTTPVersionNotSupported"},
		},

		{
			code:  http.StatusVariantAlsoNegotiates,
			types: []string{"VariantAlsoNegotiates"},
		},

		{
			code:  http.StatusInsufficientStorage,
			types: []string{"InsufficientStorage", "Temporary"},
		},

		{
			code:  http.StatusLoopDetected,
			types: []string{"LoopDetected"},
		},

		{
			code:  http.StatusNotExtended,
			types: []string{"NotExtended"},
		},

		{
			code:  http.StatusNetworkAuthenticationRequired,
			types: []string{"NetworkAuthenticationRequired"},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.code), func(t *testing.T) {
			status := fmt.Sprintf("%d %s", test.code, http.StatusText(test.code))

			res := &http.Response{
				StatusCode: test.code,
				Status:     status,
				Request: &http.Request{
					Method: http.MethodGet,
					URL: &url.URL{
						Scheme: "https",
						Host:   "localhost:443",
						Path:   "/",
					},
					Header: http.Header{
						"Host": {"localhost"},
					},
				},
			}

			testTags := []errors.Tag{
				errors.T("host", "localhost"),
				errors.T("method", "GET"),
				errors.T("path", "/"),
				errors.T("scheme", "https"),
			}

			err := New(res)
			msg := "GET https://localhost/: " + status

			if errMsg := err.Error(); errMsg != msg {
				t.Error("bad error message:")
				t.Log("expected:", msg)
				t.Log("found:   ", errMsg)
			}

			if errTypes := errors.Types(err); !reflect.DeepEqual(errTypes, test.types) {
				t.Error("error types mismatch:")
				t.Log("expected:", test.types)
				t.Log("found:   ", errTypes)
			}

			if errTags := errors.Tags(err); !reflect.DeepEqual(errTags, testTags) {
				t.Error("error tags mismatch:")
				t.Log("expected:", testTags)
				t.Log("found:   ", errTags)
			}

			if e, ok := err.(errorStackTrace); !ok {
				t.Error("error stack trace is missing")
			} else if stack := e.StackTrace(); len(stack) == 0 {
				t.Error("error stack trace is empty")
			}
		})
	}
}

func TestWrap(t *testing.T) {
	t.Run("error", testWrapError)
	t.Run("200", testWrap200)
	t.Run("400", testWrap400)
}

func testWrapError(t *testing.T) {
	if _, err := Wrap(nil, &timeout{}); !errors.Is("Timeout", err) {
		t.Error("the wrapped error must be a timeout:", err)
	} else if msg := err.Error(); msg != "timeout" {
		t.Error("bad error message:", msg)
	} else if e, ok := err.(errorStackTrace); !ok {
		t.Error("error stack trace is missing")
	} else if stack := e.StackTrace(); len(stack) == 0 {
		t.Error("error stack trace is empty")
	}
}

func testWrap200(t *testing.T) {
	r, err := Wrap(&http.Response{
		StatusCode: http.StatusOK,
		Status:     "200 OK",
		Body:       &emptyBody{},
		Request: &http.Request{
			Method: http.MethodGet,
			URL: &url.URL{
				Scheme: "https",
				Host:   "localhost:443",
				Path:   "/",
			},
			Header: http.Header{
				"Host": {"localhost"},
			},
		},
	}, nil)
	if r == nil {
		t.Error("wrapping a successful response returned a nil response")
	}
	if err != nil {
		t.Error("wrapping a successful response returned a non-nil error:", err)
	}
}

func testWrap400(t *testing.T) {
	r, err := Wrap(&http.Response{
		StatusCode: http.StatusBadRequest,
		Status:     "400 Bad Request",
		Body:       &emptyBody{},
		Request: &http.Request{
			Method: http.MethodGet,
			URL: &url.URL{
				Scheme: "https",
				Host:   "localhost:443",
				Path:   "/",
			},
			Header: http.Header{
				"Host": {"localhost"},
			},
		},
	}, nil)
	if r != nil {
		t.Error("wrapping a non-successful response returned a non-nil response")
	}
	if !errors.Is("BadRequest", err) {
		t.Error("wrapping a non-successful response did not return an error of the correct type:", err)
	}
	if err != nil {
		msg := "GET https://localhost/: 400 Bad Request"

		if errMsg := err.Error(); errMsg != msg {
			t.Error("bad error message")
			t.Log("expected:", msg)
			t.Log("found:   ", errMsg)
		}

		if e, ok := err.(errorStackTrace); !ok {
			t.Error("error stack trace is missing")
		} else if stack := e.StackTrace(); len(stack) == 0 {
			t.Error("error stack trace is empty")
		}
	}
}

type timeout struct{}

func (*timeout) Error() string   { return "timeout" }
func (*timeout) Timeout() bool   { return true }
func (*timeout) Temporary() bool { return true }

type emptyBody struct{}

func (*emptyBody) Close() error               { return nil }
func (*emptyBody) Read(b []byte) (int, error) { return 0, io.EOF }

type errorStackTrace interface {
	StackTrace() errors.StackTrace
}
