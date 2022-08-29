package timeoutconn

import (
	"io"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestConnRead(t *testing.T) {
	t.Parallel()
	ln, err := Listen("tcp", ":0", 50*time.Millisecond)
	assert.Nil(t, err)
	quit := make(chan bool)

	go func() {
		clientConn, err := net.Dial("tcp", ln.Addr().String())
		assert.Nil(t, err)
		defer clientConn.Close()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		count := 0
		for range ticker.C {
			_, err = clientConn.Write([]byte("xyzzy"))
			assert.Nil(t, err)
			count++
			if count > 10 {
				quit <- true
			}
		}
	}()

	srvConn, err := ln.Accept()
	assert.Nil(t, err)
	buf := make([]byte, 32)

loop:
	for {
		select {
		case <-quit:
			break loop
		default:
			_, err = srvConn.Read(buf)
			assert.Nil(t, err)
		}
	}
	srvConn.Close()
}

func TestSlowReadTimesOut(t *testing.T) {
	t.Parallel()
	ln, err := Listen("tcp", ":0", 50*time.Millisecond)
	assert.Nil(t, err)

	go func() {
		clientConn, err := net.Dial("tcp", ln.Addr().String())
		assert.Nil(t, err)
		_, err = clientConn.Write([]byte("xyzzy\n"))
		assert.Nil(t, err)
	}()

	srvConn, err := ln.Accept()
	assert.Nil(t, err)
	buf := make([]byte, 32)
	// wait until the timeout is passed
	<-time.After(100 * time.Millisecond)
	_, err = srvConn.Read(buf)
	assert.ErrorIs(t, err, net.ErrClosed)
}

func TestWriteKeepsConnAlive(t *testing.T) {
	t.Parallel()
	ln, err := Listen("tcp", ":0", 50*time.Millisecond)
	assert.Nil(t, err)
	quit := make(chan bool)

	go func() {
		clientConn, err := net.Dial("tcp", ln.Addr().String())
		assert.Nil(t, err)
		defer clientConn.Close()
		ticker := time.NewTicker(20 * time.Millisecond)
		defer ticker.Stop()
		count := 0
		buf := make([]byte, 32)
		for range ticker.C {
			_, err = clientConn.Read(buf)
			assert.Nil(t, err)
			count++
			if count > 10 {
				quit <- true
			}
		}
	}()

	srvConn, err := ln.Accept()
	assert.Nil(t, err)
	defer srvConn.Close()

loop:
	for {
		select {
		case <-quit:
			break loop
		default:
			_, err = srvConn.Write([]byte("hello!"))
			assert.Nil(t, err)
		}
	}
}

func TestSlowWriteTimesOut(t *testing.T) {
	t.Parallel()
	ln, err := Listen("tcp", ":0", 50*time.Millisecond)
	assert.Nil(t, err)

	go func() {
		buf := make([]byte, 32)
		clientConn, err := net.Dial("tcp", ln.Addr().String())
		assert.Nil(t, err)
		_, err = clientConn.Read(buf)
		assert.ErrorIs(t, err, io.EOF)
	}()

	srvConn, err := ln.Accept()
	assert.Nil(t, err)
	// wait until the timeout is passed
	<-time.After(100 * time.Millisecond)
	_, err = srvConn.Write([]byte("hello?"))
	assert.ErrorIs(t, err, net.ErrClosed)
}
