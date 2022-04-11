package tracing

import (
	"fmt"
	"net/http"
)

const (
	// Default RequestID header name.
	HeaderRequestID string = "X-Request-Id"
	// Default CausationID header name.
	HeaderCausationID string = "X-Causation-Id"
	// Default CorrelationID header name.
	HeaderCorrelationID string = "X-Correlation-Id"
)

// Tracing reader from Header.
type ReadHeader[T Metadata | RequestID] func(http.Header, string) (T, bool)

// Tracing writer to Header.
type WriteHeader[T Metadata | RequestID] func(http.Header, T)

// New Tracing for next event in execution chain.
type Next[T Metadata | RequestID] func(T, string) T

// Middleware options.
type Options[T Metadata | RequestID] func() (ReadHeader[T], WriteHeader[T], Next[T])

// Mapping function for getID argument.
func FromStringer(newStringer func() fmt.Stringer) func() string {
	return func() string {
		return newStringer().String()
	}
}

// Tracing middleware.
// Reads Tracing headers and writes next Tracing to Header and context.
func Middleware[T Metadata | RequestID, Opts Options[T]](
	opts Opts,
	getID func() string,
) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
			read, write, nextT := opts()
			id := getID()
			t, ok := read(req.Header, id)
			if ok {
				t = nextT(t, id)
			}

			ctx := req.Context()
			ctx = WithTracing(ctx, t)

			write(w.Header(), t)
			next.ServeHTTP(w, req.Clone(ctx))
		})
	}
}
