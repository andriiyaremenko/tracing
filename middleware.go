package tracing

import (
	"fmt"
	"net/http"
)

const (
	HeaderRequestID     string = "X-Request-Id"
	HeaderCausationID   string = "X-Causation-Id"
	HeaderCorrelationID string = "X-Correlation-Id"
)

type ReadHeader[T Metadata | RequestID] func(http.Header, string) (T, bool)
type WriteHeader[T Metadata | RequestID] func(http.Header, T)
type Next[T Metadata | RequestID] func(T, string) T

type Options[T Metadata | RequestID] func() (ReadHeader[T], WriteHeader[T], Next[T])

func FromStringer(newStringer func() fmt.Stringer) func() string {
	return func() string {
		return newStringer().String()
	}
}

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
