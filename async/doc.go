// Package async shares one value and one error per key between concurrent
// goroutines.
//
// A [Container] is safe for concurrent use and its zero value is ready to
// use: producers store results with [Container.Set] or [Container.TrySet], and
// consumers read them back with [Container.Get]. Use TrySet when the key comes
// from untrusted input, since Set panics on a nil or non-comparable key.
package async
