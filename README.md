# errors-go [![CircleCI](https://circleci.com/gh/segmentio/errors-go.svg?style=shield)](https://circleci.com/gh/segmentio/errors-go) [![Go Report Card](https://goreportcard.com/badge/github.com/segmentio/errors-go)](https://goreportcard.com/report/github.com/segmentio/errors-go) [![GoDoc](https://godoc.org/github.com/segmentio/errors-go?status.svg)](https://godoc.org/github.com/segmentio/errors-go)

## Motivations

Error management in Go is very flexible, the builtin `error` interface only
require the error value to expose an `Error() string` method which gives a
human-readable representation of what went wrong. A typical way to report
errors is through the use of *constant* values that a program can compare
error values returned by functions against to determine the issue (this is
the model used by `io.EOF` for example).

However, this is not very flexible and creates a hard dependency between the
programs and the packages they use. If a package's implementation changes the
error value returned under some circumstances, or adds new errors, the programs
depending on that package may need to be updated to adapt to the package's new
behavior.

One approach that spread across the standard library to loosen the dependency
between the package producing the errors and the program handling them has been
to establish convention on interfaces that the actual error type implements.
This allows the program to *ask questions* about the error value it received,
like whether or not it's a temporary error, or if it happened because of a
timeout (because the error type exposes `Temporary() bool` and `Timeout() bool`
methods).
This approach offers strong decoupling between components of a pgroam and is one
of the pillar concepts that this package is built uppon.

Another limitation of carrying a single error value is that it doesn't allow
layers of abstractions to add and carry context about the consequences of an
error. When `io.ErrUnexpectedEOF` is returned by a low-level I/O routine, the
next abstraction layer can either propagate the error, or return a different
error value, the former gives no information about the consequence of the error,
the latter looses information of what the original error was.

Packages like [github.com/pkg/errors](https://github.com/pkg/errors) attempt to
provide more powerful tools to compose errors and aggregate context about the
cause and consequences of those errors, but they are by choice of their authors
narrowed to solving a single aspect of error management.

This is where the `errors-go` package comes into play. It is built to be a
drop-in replacement for `pkg/errors` while offering a wider set of error
mamangement tools that we wished we had countless times in order to build
software that is more robust, expressive, and maintainable.

## Types

In the `errors-go` package, errors can carry types, and a program can
dynamically test what types an error value has. Error types are methods that
take no arguments and return a boolean value. This is the same mechanism used
in the standard library to test whether errors are temporary, or if they
happened because of a timeout.

For example, this error may be of type `Temporary`, `Timeout`, and `Unreachable`
```go
type myError struct {
    unreachable bool
    timeout     bool
}

func (e *myError) Error()       string { return "..." }
func (e *myError) Temporary()   bool   { return true }
func (e *myError) Timeout()     bool   { return e.timeout }
func (e *myError) Unreachable() bool   { return e.unreachable }
```

and a program may use the `errors.Is(typ string, err error) bool` function to
test what types the error has
```go
switch {
case errors.Is("Timeout", err):
    ...
case errors.Is("Unreachable", err):
    ...
default:
    ...
}
```

`errors.Is` dynamically discovers whether the error has a method matching the
type name, and calls it to test the error type.

Resources:
- [`errors.Is`](https://godoc.org/github.com/segmentio/errors-go#Is)
- [`errors.Types`](https://godoc.org/github.com/segmentio/errors-go#Types)
- [`errors.WithTypes`](https://godoc.org/github.com/segmentio/errors-go#WithTypes)

## Tagging

Tags are a list of arbitrary key/value pairs that can be attached to any errors,
it provides a way to aggregate errors based on tag values.
They are useful when errors are logged, traced, or measured, since the tags can
be extracted from the error and injected into the logging, tracing, or metric
collection systems.

To add tags to an error, a program may either define a `Tags() []errors.Tag`
method on an error type, or use the `errors.WithTags` function to add tags to
an error value, for example:
```go
operation := "HelloWorld"

if err := rpc.Call(operation); err != nil {
    return errors.WithTags(err,
        errors.T("component", "rpc-client"),
        errors.T("operation", operation),
    )
}
```

Resources:
- [`errors.T`](https://godoc.org/github.com/segmentio/errors-go#T)
- [`errors.Tag`](https://godoc.org/github.com/segmentio/errors-go#Tag)
- [`errors.Tags`](https://godoc.org/github.com/segmentio/errors-go#Tags)
- [`errors.WithTags`](https://godoc.org/github.com/segmentio/errors-go#WithTags)

## Causes

When wrapping an error value to add context to it (may it be types, tags, or
other attributes), a program can retrieve the original error by using the
`errors.Cause(error) error` function. This mechanism is identical to what is
done in the [`github.com/pkg/errors`](https://github.com/pkg/errors) package
and is useful when a program needs to compare the original error value against
some *contant* like `io.EOF` for example.

However, it is common in Go to end up with more than one cause for an error.
When a program spawns multiple goroutine to do I/O operations in parallel, some
may succeed and some may fail. Instead of having to discard all errors but the
first one, or re-invent yet another multi-error type, the `errors-go` package
offers two methods to construct error values from a set of errors:
```go
func Join(errs ...error) error
```
```go
func Recv(errs <-chan error) error
```

A program then cannot count on retrieving the single cause and instead can
use the `errors.Causes(error) []error` function to extract all causes that
lead to generating this error.

This means that the `errors-go` package allows programs to build error trees,
where each node is an error value (including wrappers) with a list of causes,
that may as well have been wrapped, and may also have causes themselves.

Resources:
- [`errors.Cause`](https://godoc.org/github.com/segmentio/errors-go#Cause)
- [`errors.Causes`](https://godoc.org/github.com/segmentio/errors-go#Causes)
- [`errors.Join`](https://godoc.org/github.com/segmentio/errors-go#Recv)
- [`errors.Recv`](https://godoc.org/github.com/segmentio/errors-go#Recv)
- [`errors.Wrap`](https://godoc.org/github.com/segmentio/errors-go#Wrap)
- [`errors.Wrapf`](https://godoc.org/github.com/segmentio/errors-go#Wrapf)

## Adapters

Due to Go's very flexible error handling model, packages have adopted different
approaches, which are not always easy to plug together. This is true even
within the standard library itself, where some packages use specific error
types, exported or unexported error values, or a combination of those. Those
differences result in heterogenous error handling code within a single program.

To work around this issue the `errors-go` package uses the *Adapter* concept.
An Adapter is meant to convert errors generated by a package into errors that
can be manipulated using the `errors-go` functions.

Adapters are intended to be registered during the initialization phase of a
package (using its `init` function) in order to be available globally whenever
an error needs to be adapted.

Errors are adapted automatically by calls to any of the wrapper functions of the
`errors-go` package (like `Wrap`, `WithMessage`, `WithStack`, etc...). It means
that all a program needs to do is import adapter packages and its error wrapping
functions will be enhanced to do proper error decoration of errors coming from
those packages.

Resources:
- [`errors.Adapt`](https://godoc.org/github.com/segmentio/errors-go#Adapt)
- [`errors.Register`](https://godoc.org/github.com/segmentio/errors-go#Register)

## Formatting

Errors are eventually meant to be consumed by developers and operators of a
program, which means formatting of the error values into human-readable forms
is a key feature of an error package.

Go has understood this well by making the `error` interface's single method one
that returns an error message. However, a single error message often isn't
enough to communicate all the context of what went wrong, and using a format
that allows the context to be fully exposed is highly valuable, both during
development and operations.

Errors wrapped by, or produced by the `errors-go` package all use a text format
which exposes the message, types, tags, stack traces, and the tree of causes
that resulted in the error. Those information can be enabled or disabled based
on the format string being used (e.g. `%v`, `%+v`, `%#v`), here's an example of
the full format:
```
error message (type ...) [tag=value ...]
stack traces
...
├── sub error message (type ...) [tag=value ...]
|   stack traces
|   ...
└── last error message (type ...) [tag=value ...]
    stack traces
```

- The first line is the error message, the types and tags of the error.
- The stack traces are printed in the same format as the panic stack traces.
- The tree-like format provides a representation of the tree of error causes.

This format ensures that all information carried by the error are available in
a way that is both familiar with current practices (Go debug traces, format of
errors from the `github.com/pkg/errors` package, ...) while still exposing other
properties of errors of the `errors-go` package.

## Serializing

Serializability is not a property that's very easy to obtain from Go errors.
Because the standard `error` interface only gives access to an error message
one can simply serialize this message to pass errors across services over the
network, but this process would lose other information like types, tags, or
causes.

To address this limitation the `errors-go` package uses an intermediary,
serializable type to represent errors, which can also be used as a form of
reflection to explore the inner structure of an error value.

The `errors.ValueOf(error)` function returns a `errors.Value` which snapshots
a representation of the error passed as argument. Values can then be manipulated
and serialized and deserialized by an program, and an error carrying the same
properties as the original can be reconstructed from the value.

*Note: The only property of an error which is not equivalent in errors built
from values and their original is the stack trace. This design choice was made
because programs exchanging error information over the network rarely need to
carry a stack trace of a different code than theirs. So instead, the stack trace
in errors that are reconstructed from values is a new capture of the call stack
within the current program.*

Resources:
- [`errors.Value`](https://godoc.org/github.com/segmentio/errors-go#Value)
- [`errors.ValueOf`](https://godoc.org/github.com/segmentio/errors-go#ValueOf)
