package timeoutconn

import (
	"net"
	"time"
)

// timeoutListener wraps a net.Listener
type timeoutListener struct {
	net.Listener
	timeout time.Duration
}

// Accept waits for a connection and starts the timer.
func (l *timeoutListener) Accept() (net.Conn, error) {
	nc, err := l.Listener.Accept()
	if err != nil {
		return nil, err
	}
	c := &timeoutConn{
		timeout: l.timeout,
		Conn:    nc,
	}
	c.startTimer()
	return c, err
}

// timeoutConn wraps a net.Conn and will close the connection when the timer fires
type timeoutConn struct {
	net.Conn
	timeout    time.Duration
	closeTimer *time.Timer
}

// Close closes the connection
func (c *timeoutConn) Close() error {
	return c.Conn.Close()
}

// Read reads from the connection and resets the timer.
func (c *timeoutConn) Read(b []byte) (int, error) {
	c.resetTimer()
	return c.Conn.Read(b)
}

// Write writes to the connection.
func (c *timeoutConn) Write(b []byte) (int, error) {
	c.resetTimer()
	return c.Conn.Write(b)
}

// startTimer starts the close timer
func (c *timeoutConn) startTimer() {
	c.closeTimer = time.AfterFunc(c.timeout, func() { c.Close() })
}

// resetTimer will reset the close
func (c *timeoutConn) resetTimer() {
	if c.closeTimer != nil {
		c.closeTimer.Reset(c.timeout)
	}
}

// Listen listens on an address
func Listen(network, address string, timeout time.Duration) (net.Listener, error) {
	ln, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewListener(ln, timeout), nil
}

// NewListener wraps a net.Listener
func NewListener(l net.Listener, timeout time.Duration) net.Listener {
	return &timeoutListener{
		Listener: l,
		timeout:  timeout,
	}
}
