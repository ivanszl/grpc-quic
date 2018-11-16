package net

import (
	"net"

	quic "github.com/lucas-clemente/quic-go"
)

type Listener struct {
	ql quic.Listener
}

func Listen(ql quic.Listener) net.Listener {
	return &Listener{ql}
}

func (l *Listener) Accept() (net.Conn, error) {
	sess, err := l.ql.Accept()
	if err != nil {
		return nil, err
	}

	s, err := sess.AcceptStream()
	if err != nil {
		return nil, err
	}

	return &QuicConn{sess, s}, nil
}

func (l *Listener) Close() error {
	return l.ql.Close()
}

func (l *Listener) Addr() net.Addr {
	return l.ql.Addr()
}
