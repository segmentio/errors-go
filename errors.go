package errors

import (
	"fmt"
	"reflect"
	"strings"
)

// TODO is a non-nil error intended to act as a placeholder during development
// when writing the structure of the code but the implementation is still left
// to be written.
var TODO error

// New returns an error that formats as the given message. The returned error
// carries a capture of the stack trace.
//
//	err = errors.New("something went wrong")
//
func New(msg string) error {
	return &baseError{
		msg:   msg,
		stack: CaptureStackTrace(1),
	}
}

// Errorf returns an error that formats as fmt.Sprintf(msg, args...).
// The returned error carries a capture of the stack trace.
//
//	err = errors.Errorf("unexpected answer: %d", 42)
//
func Errorf(msg string, args ...interface{}) error {
	return &baseError{
		msg:   fmt.Sprintf(msg, args...),
		stack: CaptureStackTrace(1),
	}
}

// WithMessage returns an error that wraps err and prefix its original error
// error message with msg. If err is nil, WithMessage returns nil.
//
//	err = errors.WithMessage(err, "something went wrong")
//
func WithMessage(err error, msg string) error {
	if err == nil {
		return nil
	}
	return &errorWithMessage{
		cause: Adapt(err),
		msg:   msg,
	}
}

// WithStack returns an error that wraps err with a capture of the stack trace
// at the time the function is called. If err is nil, WithStack returns nil.
//
//	err = errors.WithStack(err)
//
// The error is adapted before the stack trace is added.
func WithStack(err error) error {
	return WithStackTrace(err, CaptureStackTrace(1))
}

// WithStackTrace returns an error that wraps err with the given stack trace.
// If err is nil, WithStackTrace returns nil.
//
//	err = errors.WithStackTrace(err, errors.CaptureStackTrace(1))
//
// The error is adapted before the stack trace is added.
func WithStackTrace(err error, stack StackTrace) error {
	if err == nil {
		return nil
	}
	return &errorWithStack{
		cause: Adapt(err),
		stack: stack,
	}
}

// WithTypes returns an error that wraps err and implements the given types so
// that calling errors.Is on the returned error with one of the given types will
// return true.
//
// The error is adapted before types are added.
func WithTypes(err error, types ...string) error {
	if err == nil {
		return nil
	}
	return &errorWithTypes{
		cause: Adapt(err),
		types: copyTypes(types),
	}
}

// WithTags returns an error that wraps err and tags it with the given key/value
// pairs. If err is nil the function returns nil.
//
// The error is adapted before tags are added.
func WithTags(err error, tags ...Tag) error {
	if err == nil {
		return nil
	}
	return &errorWithTags{
		cause: Adapt(err),
		tags:  makeTags(tags...),
	}
}

// Wrap returns an error that wraps err with msg as prefix to its original
// message and a capture of the stack trace at the time the function is called.
// If err is nil, Wrap returns nil.
//
//	err = errors.Wrap(err, "something went wrong")
//
// The error is adapted before being wrapped with a message and stack trace.
func Wrap(err error, msg string) error {
	return wrap(err, 1, msg)
}

// Wrapf returns an error that wraps err with fmt.Sprintf(msg, args...) as
// prefix to its original message and a capture of the stack trace at the time
// the function is called. If err is nil, Wrap returns nil.
//
//	err = errors.Wrapf(err, "unexpected answer: %d", 42)
//
// The error is adapted before being wrapped with a message and stack trace.
func Wrapf(err error, msg string, args ...interface{}) error {
	return wrap(err, 1, fmt.Sprintf(msg, args...))
}

func wrap(err error, depth int, msg string) error {
	if err == nil {
		return nil
	}
	return &errorWithMessage{
		cause: &errorWithStack{
			cause: Adapt(err),
			stack: CaptureStackTrace(depth + 1),
		},
		msg: msg,
	}
}

// Join composes an error from the list of errors passed as argument.
//
// The function strips all nil errors from the input argument list. The returned
// error has a Causes method which returns the list of non-nil errors that were
// given to the function.
//
//	err = errors.Join(err1, err2, err3)
//
// All errors passed to the function are adapted.
func Join(errs ...error) error {
	n := 0

	for _, e := range errs {
		if e != nil {
			n++
		}
	}

	if n == 0 {
		return nil
	}

	e := &multiError{
		errors: make([]error, 0, n),
	}

	for _, err := range errs {
		if err != nil {
			e.errors = append(e.errors, Adapt(err))
		}
	}

	return e
}

// Recv reads all errors from the given channel and returns one that combines
// them. All nil error are ignored.
//
//	ch := make(chan error)
//	wg := sync.WaitGroup{}
//
//	for _, t := range tasks {
//		wg.Add(1)
//		go func(t task) { ch <- t(); wg.Done() }
//	}
//
//	go func() { wg.Wait(); close(ch) }
//
//	err := errors.Recv(errch)
//
// All errors received on the channel are adapted.
func Recv(ch <-chan error) error {
	var errs []error

	for err := range ch {
		if err != nil {
			errs = append(errs, Adapt(err))
		}
	}

	if len(errs) == 0 {
		return nil
	}

	return &multiError{
		errors: errs,
	}
}

// Err constructs an error from a value of arbitrary type, using the following
// rules:
//
// - if v is nil, the function simply returns nil
//
// - if v is a string type, the function behaves like calling New
//
// - if v is already an error it is returned unchanged
//
// - if v has an Err() method returning an error, the function returns the
// result of calling it.
//
// - if v doesn't fall into any of those categories, the function behaves like
// calling Errorf("%+v", v).
//
// A common use case for this function is to implement internal error reporting
// based on raising panics (within a package), here is an example:
//
//	func F() (err error) {
//		defer func() { err = errors.Err(recover()) }()
//		// ...
//	}
//
func Err(v interface{}) error {
	switch value := v.(type) {
	case nil:
		return nil

	case string:
		return &baseError{
			msg:   value,
			stack: CaptureStackTrace(1),
		}

	case error:
		return value

	case interface {
		Err() error
	}:
		return value.Err()

	default:
		return &baseError{
			msg:   fmt.Sprintf("%+v", value),
			stack: CaptureStackTrace(1),
		}
	}
}

// Cause returns the cause of err, which may be err if it had no cause.
func Cause(err error) error {
	for {
		if e, ok := err.(errorCause); ok {
			if cause := e.Cause(); cause != nil {
				err = cause
				continue
			}
		}
		return err
	}
}

// Causes returns the list of causes of err, which may be an empty slice if err
// is nil or had no causes.
func Causes(err error) []error {
	original := err
	for {
		if e, ok := err.(errorCauses); ok {
			return e.Causes()
		}
		if e, ok := err.(errorCause); ok {
			if cause := e.Cause(); cause != nil {
				err = cause
				continue
			}
		}
		if err != original {
			return []error{err}
		}
		return nil
	}
}

// Is tests whether err is of type typ. Errors may implement types by defining
// methods that take no arguments and return a boolean value. Passing the name
// of those methods to Is tests for their existence and calls them to validate
// the type of the error.
//
// This model has been used in the standard library where some errors implement
// the Temporary and Timeout methods to give the program more details about the
// reason the error occured and the way it should be handled.
//
// Here is an example of using the Is function:
//
//	if errors.Is("Timeout", err) {
//		// ...
//	}
//	if errors.Is("Temporary", err) {
//		// ...
//	}
//
// The function walks through the graph of causes looking for an error which may
// implement the given type.
func Is(typ string, err error) bool {
	if err == nil {
		return false
	}

	if e, ok := err.(errorTypes); ok {
		for _, t := range e.Types() {
			if t == typ {
				return true
			}
		}
	}

	v := reflect.ValueOf(err)
	m := v.MethodByName(typ)

	if m.IsValid() {
		if f, ok := m.Interface().(func() bool); ok {
			return f()
		}
	}

	switch e := err.(type) {
	case errorCause:
		return Is(typ, e.Cause())

	case errorCauses:
		for _, cause := range e.Causes() {
			if ok := Is(typ, cause); ok {
				return true
			}
		}
	}

	return false
}

// Types returns a slice containing all the types implemented by err and its
// causes (if it had any).
func Types(err error) []string {
	return deepAppendTypes(nil, err)
}

// Tags returns a slice containing all the tags set on err and its causes
// (it if had any).
func Tags(err error) []Tag {
	return deepAppendTags(nil, err)
}

func inspect(err error) (msgs []string, types []string, tags []Tag, stacks []StackTrace, causes []error) {
	for err != nil {
		types = appendTypes(types, err)
		tags = appendTags(tags, err)

		if msg := message(err); len(msg) != 0 {
			msgs = append(msgs, msg)
		}

		if stack := stackTrace(err); len(stack) != 0 {
			stacks = append(stacks, stack)
		}

		switch e := err.(type) {
		case errorCauses:
			causes = e.Causes()
			err = nil

		case errorCause:
			err = e.Cause()

		case errorMessage:
			err = nil // prevent duplicating the message with the Error call

		default:
			msgs = append(msgs, e.Error())
			err = nil
		}
	}

	types = dedupeTypes(types)
	sortTags(tags)
	return
}

func walk(err error, do func(error)) {
	if err != nil {
		do(err)

		switch e := err.(type) {
		case errorCause:
			walk(e.Cause(), do)

		case errorCauses:
			for _, cause := range e.Causes() {
				walk(cause, do)
			}
		}
	}
}

type errorCause interface {
	Cause() error
}

type errorCauses interface {
	Causes() []error
}

type errorMessage interface {
	Message() string
}

type errorTypes interface {
	Types() []string
}

type errorTags interface {
	Tags() []Tag
}

type errorStackTrace interface {
	StackTrace() StackTrace
}

type baseError struct {
	msg   string
	stack StackTrace
}

func (e *baseError) Error() string {
	return e.msg
}

func (e *baseError) Message() string {
	return e.msg
}

func (e *baseError) StackTrace() StackTrace {
	return e.stack
}

func (e *baseError) Format(s fmt.State, v rune) {
	format(s, v, e)
}

type multiError struct {
	errors []error
}

func (e *multiError) Causes() []error {
	return e.errors
}

func (e *multiError) Error() string {
	s := make([]string, len(e.errors))
	for i, e := range e.errors {
		s[i] = e.Error()
	}
	return strings.Join(s, "; ")
}

func (e *multiError) Format(s fmt.State, v rune) {
	format(s, v, e)
}

type errorWithMessage struct {
	cause error
	msg   string
}

func (e *errorWithMessage) Cause() error {
	return e.cause
}

func (e *errorWithMessage) Error() string {
	return e.msg + ": " + e.cause.Error()
}

func (e *errorWithMessage) Message() string {
	return e.msg
}

func (e *errorWithMessage) Format(s fmt.State, v rune) {
	format(s, v, e)
}

type errorWithStack struct {
	cause error
	stack StackTrace
}

func (e *errorWithStack) Cause() error {
	return e.cause
}

func (e *errorWithStack) Error() string {
	return e.cause.Error()
}

func (e *errorWithStack) Format(s fmt.State, v rune) {
	format(s, v, e)
}

func (e *errorWithStack) StackTrace() StackTrace {
	return e.stack
}

type errorWithTypes struct {
	cause error
	types []string
}

func (e *errorWithTypes) Cause() error {
	return e.cause
}

func (e *errorWithTypes) Error() string {
	return e.cause.Error()
}

func (e *errorWithTypes) Format(s fmt.State, v rune) {
	format(s, v, e)
}

func (e *errorWithTypes) Types() []string {
	return e.types
}

type errorWithTags struct {
	cause error
	tags  []Tag
}

func (e *errorWithTags) Cause() error {
	return e.cause
}

func (e *errorWithTags) Error() string {
	return e.cause.Error()
}

func (e *errorWithTags) Format(s fmt.State, v rune) {
	format(s, v, e)
}

func (e *errorWithTags) Tags() []Tag {
	return e.tags
}

type errorTODO struct{}

func (*errorTODO) Error() string {
	return "TODO"
}

func init() {
	TODO = &errorTODO{}
}

func message(err error) string {
	if e, ok := err.(errorMessage); ok {
		return e.Message()
	}
	return ""
}

func stackTrace(err error) StackTrace {
	if e, ok := err.(errorStackTrace); ok {
		return e.StackTrace()
	}
	return nil
}
