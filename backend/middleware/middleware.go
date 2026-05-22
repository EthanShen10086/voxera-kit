// Package middleware provides reusable HTTP middleware for the voxera-kit
// ecosystem. Each middleware is expressed as a [Func] that wraps an
// [http.Handler], and [Chain] composes them in left-to-right order.
package middleware

import "net/http"

// Func is the signature shared by every middleware in this package.
// It takes a handler and returns a new handler that adds behavior before or
// after the inner handler executes.
type Func func(http.Handler) http.Handler

// Chain applies the given middlewares to handler in the order they are
// provided. The first middleware in the list is the outermost wrapper.
func Chain(handler http.Handler, mws ...Func) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		handler = mws[i](handler)
	}
	return handler
}
