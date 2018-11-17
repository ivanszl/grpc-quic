package gquic

import (
	"context"
	"net"
	"time"

	"github.com/ivanszl/grpc-quic/transports"

	qnet "github.com/ivanszl/grpc-quic/net"
	quic "github.com/lucas-clemente/quic-go"
	"google.golang.org/grpc/credentials"
)

type Config quic.Config

func NewQuicDialer(conf *quic.Config, cred credentials.TransportCredentials) func(string, time.Duration) (net.Conn, error) {
	return func(target string, timeout time.Duration) (net.Conn, error) {
		creds, _ := cred.(*transports.Credentials)
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		sess, err := quic.DialAddrContext(ctx, target, creds.GetTLSConfig(), conf)
		if err != nil {
			return nil, err
		}

		return qnet.NewQuicConn(sess)
	}
}

func NewListener(addr string, conf *quic.Config, cred credentials.TransportCredentials) (net.Listener, error) {

	creds, _ := cred.(*transports.Credentials)

	ql, err := quic.ListenAddr(addr, creds.GetTLSConfig(), conf)
	if err != nil {
		return nil, err
	}

	return qnet.Listen(ql), nil
}
