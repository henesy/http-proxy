package filters

import (
	"context"
	"net"
	"time"
)

type contextKey string

const (
	ctxKeyDownstream    = contextKey("downstream")
	ctxKeyRequestNumber = contextKey("requestNumber")
)

// Context is a wrapper for Context that exposes some additional
// information specific to its use in proxies.
type Context interface {
	context.Context

	// DownstreamConn retrieves the downstream connection from the given Context.
	DownstreamConn() net.Conn

	// RequestNumber indicates how many requests have been received on the current
	// connection. The RequestNumber for the first request is 1, for the second is 2
	// and so forth.
	RequestNumber() int

	// IncrementRequestNumber increments the request number by 1 and returns a new
	// context.
	IncrementRequestNumber() Context

	// WithCancel mimics the method on context.Context
	WithCancel() (Context, context.CancelFunc)

	// WithDeadline mimics the method on context.Context
	WithDeadline(deadline time.Time) (Context, context.CancelFunc)

	// WithTimeout mimics the method on context.Context
	WithTimeout(timeout time.Duration) (Context, context.CancelFunc)

	// WithValue mimics the method on context.Context
	WithValue(key, val interface{}) Context
}

// WrapContext wraps the given context.Context into a Context containing the
// given downstream net.Conn.
func WrapContext(ctx context.Context, downstream net.Conn) Context {
	return (&ctext{ctx}).
		WithValue(ctxKeyRequestNumber, 1).
		WithValue(ctxKeyDownstream, downstream)
}

// BackgroundContext creates a background Context without an associated
// connection.
func BackgroundContext() Context {
	return WrapContext(context.Background(), nil)
}

// ctext implements Context
type ctext struct {
	context.Context
}

func (ctx *ctext) DownstreamConn() net.Conn {
	return ctx.Value(ctxKeyDownstream).(net.Conn)
}

func (ctx *ctext) RequestNumber() int {
	return ctx.Value(ctxKeyRequestNumber).(int)
}

func (ctx *ctext) IncrementRequestNumber() Context {
	return ctx.WithValue(ctxKeyRequestNumber, ctx.RequestNumber()+1)
}

func (ctx *ctext) WithCancel() (Context, context.CancelFunc) {
	result, cancel := context.WithCancel(ctx)
	return &ctext{result}, cancel
}

func (ctx *ctext) WithDeadline(deadline time.Time) (Context, context.CancelFunc) {
	result, cancel := context.WithDeadline(ctx, deadline)
	return &ctext{result}, cancel
}

func (ctx *ctext) WithTimeout(timeout time.Duration) (Context, context.CancelFunc) {
	result, cancel := context.WithTimeout(ctx, timeout)
	return &ctext{result}, cancel
}

func (ctx *ctext) WithValue(key, val interface{}) Context {
	return &ctext{context.WithValue(ctx, key, val)}
}
