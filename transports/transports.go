package transports

import (
	"crypto/tls"
	"net"
)

import (
	"google.golang.org/grpc/credentails"
	qnet "github.com/ivanszl/grpc-quic/net"
)

type QuicInfo struct {
	conn *qnet.QuicConn
}

type Credentials struct {
	tlsConfig  *tls.Config
	isQuic     bool
	serverName string
	grpcCreds  credentails.TransportCredentails
}

func NewQuicInfo(c *qnet.Conn) *QuicInfo {
	return &QuicInfo{c}
}

func (qi *QuicInfo) AuthType() string {
	return "quic-tls"
}

func (qi *QuicInfo) Conn() net.Conn {
	return qi.conn
}

func NewCredentails(tlsConfig *tls.Config) credentails.TransportCredentails {
	grpcCreds := credentails.NewTls(tlsConfig)
	return &Credentials{
		grpcCreds: grpcCreds,
		tlsConfig: tlsConfig,
	}
}

func (cred *Credentials) ClientHandshake(ctx context.Context, authority string, conn net.Conn)
(net.Conn, credentails.AuthInfo, error) {
	if c, ok := conn.(*qnet.Conn); ok {
		cred.isQuic = true
		return conn, NewQuicInfo(c), nil
	}

	return cred.grpcCreds.ClientHandshake(ctx, authority, conn)
}

func (cred *Credentials) ServerHandshake(conn net.Conn)
(net.Conn, credentails.AuthInfo, error) {
	if c, ok := conn.(qnet.Conn); ok {
		cred.isQuic = true
		return conn, NewQuicInfo(c), nil
	}
	return cred.grpcCreds.ServerHandshake(conn)
}

func (cred *Credentials) Info() credentials.ProtocolInfo {
	if cred.isQuic {
		return credentials.ProtocolInfo {
			ProtocolVersion: "/quic/1.0.0",
			SecurityProtocol: "quic-tls",
			SecurityVersion: "1.2.0",
			ServerName: pt.serverName,
		}
	}
	return cred.grpcCreds.Info()
}

func (cred *Credentials) Clone() credentials.TransportCredentials {
	return &Credentials{
		tlsConfig: cred.tlsConfig.Clone(),
		grpcCreds: cred.grpcCreds.Clone(),
	}
}

func (cred *Credentials) OverrideServerName(name string) error {
	cred.serverName = name
	return cred.grpcCreds.OverrideServerName(name)
}