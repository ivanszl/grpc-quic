package net

import (
	"net"
	"time"

	quic "github.com/lucas-clemente/quic-go"
)

type QuicConn struct {
	sess   quic.Session
	stream quic.Stream
}

func NewConn(sess quic.Session) (net.Conn, error) {
	stream, err := sess.OpenStreamSync()
	if err != nil {
		return nil, err
	}

	return &QuicConn{sess, stream}, nil
}

func (c *QuicConn) Read(b []byte) (int, error) {
	return c.stream.Read(b)
}

func (c *QuicConn) Write(b []byte) (int, error) {
	return c.stream.Write(b)
}

func (c *QuicConn) Close() error {
	c.stream.Close()

	return c.sess.Close()
}

func (c *QuicConn) LocalAddr() net.Addr {
	return c.sess.LocalAddr()
}

func (c *QuicConn) RemoteAddr() net.Addr {
	return c.sess.RemoteAddr()
}

func (c *QuicConn) SetDeadLine(t time.Time) error {
	return c.stream.SetDeadLine(t)
}

func (c *QuicConn) SetReadDeadLine(t time.Time) error {
	return c.stream.SetReadDeadLine(t)
}

func (c *QuicConn) SetWriteDeadLine(t time.Time) error {
	return c.stream.SetWriteDeadLine(t)
}
