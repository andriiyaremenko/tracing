package tracing

import "context"

type key int

var tracingKey key

// Adds tracing to context.
func WithTracing[T Metadata | RequestID](ctx context.Context, t T) context.Context {
	return context.WithValue(ctx, tracingKey, t)
}

// Reads tracing from context.
func GetTracing[T Metadata | RequestID](ctx context.Context) (T, bool) {
	v, ok := ctx.Value(tracingKey).(T)
	return v, ok
}
