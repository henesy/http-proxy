package listeners

import (
	"net"
	"net/http"
	"time"

	"github.com/getlantern/idletiming"
)

// Wrapped idleConnListener that generates the wrapped idleConn
type idleConnListener struct {
	net.Listener
	idleTimeout time.Duration
}

func NewIdleConnListener(l net.Listener, timeout time.Duration) net.Listener {
	return &idleConnListener{
		Listener:    l,
		idleTimeout: timeout,
	}
}

func (l *idleConnListener) Accept() (c net.Conn, err error) {
	conn, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}

	return WrapIdleConn(conn, l.idleTimeout), nil
}

// WrapIdleConn wraps the given conn in an idletiming conn using the given
// idleTimeout.
func WrapIdleConn(conn net.Conn, idleTimeout time.Duration) net.Conn {
	iConn := idletiming.Conn(conn, idleTimeout, nil)

	sac, _ := conn.(WrapConnEmbeddable)
	return &idleConn{
		WrapConnEmbeddable: sac,
		Conn:               iConn,
	}
}

// Wrapped IdleTimingConn that supports OnState
type idleConn struct {
	WrapConnEmbeddable
	net.Conn
}

func (c *idleConn) OnState(s http.ConnState) {
	if c.WrapConnEmbeddable != nil {
		c.WrapConnEmbeddable.OnState(s)
	}
}

func (c *idleConn) ControlMessage(msgType string, data interface{}) {
	// Simply pass down the control message to the wrapped connection
	if c.WrapConnEmbeddable != nil {
		c.WrapConnEmbeddable.ControlMessage(msgType, data)
	}
}

func (c *idleConn) Wrapped() net.Conn {
	return c.Conn
}
