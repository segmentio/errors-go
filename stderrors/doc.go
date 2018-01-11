// Package stderrors exposes no APIs and is used for the sole purpose of setting
// up global adapters for all packages of the standard library supported by the
// errors-go project.
//
// For example, instead of importing each error package individually to set up
// adapters,
//
//	import (
//		_ "github.com/segmentio/errors-go/ioerrors"
//		_ "github.com/segmentio/errors-go/neterrors"
//		...
//	)
//
// a program can import this package instead with
//
//	import (
//		_ "github.com/segmentio/errors-go/stderrors"
//	)
//
// Note that using this package may result in a larger compiled binary if it
// causes the import of unused packages, this is usually not an issue as large
// programs often make use of most packages of the standard library.
//
// Importing this package doesn't install adapters for packages like awserrors
// because they don't provide adapters for errors of the standard library.
package stderrors
