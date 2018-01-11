package errors

import "sync"

// Adapter is an interface implemented by types that support adapting errors to
// be introspected by functions of the erorrs package.
type Adapter interface {
	// Adapt is called to adapt err, it either returnes err and false if it did
	// not recognize the error, or returns the adapted error and true.
	Adapt(err error) (error, bool)
}

// The AdapterFunc types is an implementation of the Adapter interface which
// makes it possible to use simple functions as error adapters.
type AdapterFunc func(error) (error, bool)

// Adapt satsifies the Adapter interface, calls f.
func (f AdapterFunc) Adapt(err error) (error, bool) { return f(err) }

// Adapt adapts err using the registered adapters.
//
// Programs usually do not need to call this function explicitly and can instead
// rely on the fact that functions like Wrap, WithMessage, WithStack... will
// automatically adapt the errors that they receive.
func Adapt(err error) error {
	switch err.(type) {
	case *baseError, *multiError, *errorWithMessage, *errorWithStack, *errorWithTypes, *errorWithTags, *errorTODO, *errorValue:
		// fast path: when the error is already one of the internal error types
		// of this package there is no need to go over the list of adapters.
		return err
	}
	return adapters.adapt(err, 1)
}

// Register registers a new error adapter.
func Register(a Adapter) { adapters.register(a) }

type adapterStore struct {
	mutex    sync.RWMutex
	adapters []Adapter
}

func (store *adapterStore) register(a Adapter) {
	if a != nil {
		store.mutex.Lock()
		store.adapters = append(store.adapters, a)
		store.mutex.Unlock()
	}
}

func (store *adapterStore) adapt(err error, depth int) error {
	if err != nil {
		store.mutex.RLock()
		defer store.mutex.RUnlock()

		for _, a := range store.adapters {
			if e, ok := a.Adapt(err); ok {
				err = e
				break
			}
		}
	}
	return err
}

// adapters is the global store of error adapters that the program has setup by
// calling Register.
var adapters adapterStore
