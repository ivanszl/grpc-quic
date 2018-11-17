package transports

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io/ioutil"
	"net"

	qnet "github.com/ivanszl/grpc-quic/net"
	"google.golang.org/grpc/credentials"
)

type QuicInfo struct {
	conn *qnet.QuicConn
}

type Credentials struct {
	tlsConfig  *tls.Config
	isQuic     bool
	serverName string
	grpcCreds  credentials.TransportCredentials
}

var _ credentials.TransportCredentials = (*Credentials)(nil)

func NewQuicInfo(c *qnet.QuicConn) *QuicInfo {
	return &QuicInfo{c}
}

func (qi *QuicInfo) AuthType() string {
	return "quic-tls"
}

func (qi *QuicInfo) Conn() net.Conn {
	return qi.conn
}

func NewServerTLSFromFile(certFile, keyFile string) (*Credentials, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	grpcCreds := credentials.NewTLS(tlsConfig)
	return &Credentials{
		grpcCreds: grpcCreds,
		tlsConfig: tlsConfig,
	}, nil
}

func NewServerTLSFromCert(cert tls.Certificate) credentials.TransportCredentials {

	tlsConfig := &tls.Config{Certificates: []tls.Certificate{cert}}
	grpcCreds := credentials.NewTLS(tlsConfig)
	return &Credentials{
		grpcCreds: grpcCreds,
		tlsConfig: tlsConfig,
	}
}

func NewClientTLSFromFile(certFile, serverNameOverride string) (credentials.TransportCredentials, error) {
	b, err := ioutil.ReadFile(certFile)
	if err != nil {
		return nil, err
	}
	cp := x509.NewCertPool()
	if !cp.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("credentials: failed to append certificates")
	}

	tlsConfig := &tls.Config{ServerName: serverNameOverride, RootCAs: cp}
	grpcCreds := credentials.NewTLS(tlsConfig)
	return &Credentials{
		grpcCreds: grpcCreds,
		tlsConfig: tlsConfig,
	}, nil
}

func NewClientTLSFromCert(cp *x509.CertPool, serverNameOverride string) credentials.TransportCredentials {
	tlsConfig := &tls.Config{ServerName: serverNameOverride, RootCAs: cp}
	grpcCreds := credentials.NewTLS(tlsConfig)
	return &Credentials{
		grpcCreds: grpcCreds,
		tlsConfig: tlsConfig,
	}
}

func NewClientTLSSkipVerify() credentials.TransportCredentials {
	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	grpcCreds := credentials.NewTLS(tlsConfig)
	return &Credentials{
		grpcCreds: grpcCreds,
		tlsConfig: tlsConfig,
	}
}

func (cred *Credentials) ClientHandshake(ctx context.Context, authority string, conn net.Conn) (net.Conn, credentials.AuthInfo, error) {
	if c, ok := conn.(*qnet.QuicConn); ok {
		cred.isQuic = true
		return conn, NewQuicInfo(c), nil
	}

	return cred.grpcCreds.ClientHandshake(ctx, authority, conn)
}

func (cred *Credentials) ServerHandshake(conn net.Conn) (net.Conn, credentials.AuthInfo, error) {

	if c, ok := conn.(*qnet.QuicConn); ok {
		cred.isQuic = true
		return conn, NewQuicInfo(c), nil
	}
	return cred.grpcCreds.ServerHandshake(conn)
}

func (cred *Credentials) Info() credentials.ProtocolInfo {
	if cred.isQuic {
		return credentials.ProtocolInfo{
			ProtocolVersion:  "/quic/1.0.0",
			SecurityProtocol: "quic-tls",
			SecurityVersion:  "1.2.0",
			ServerName:       cred.serverName,
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

func (cred *Credentials) GetTLSConfig() *tls.Config {
	return cred.tlsConfig
}
