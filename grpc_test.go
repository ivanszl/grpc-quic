package grpcquic

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"math/big"
	"testing"
	"time"

	"github.com/gfanton/grpc-quic/proto/hello"
	qgrpc "github.com/ivanszl/grpc-quic"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

type Hello struct{}

func (h *Hello) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	rep := new(hello.HelloReply)
	rep.Message = "Hello " + in.GetName()
	return rep, nil
}

func generateTLSConfig() (*tls.Config, error) {
	key, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		return nil, err
	}

	template := x509.Certificate{SerialNumber: big.NewInt(1)}
	certDER, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		return nil, err
	}

	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(key)})
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	tlsCert, err := tls.X509KeyPair(certPEM, keyPEM)
	if err != nil {
		return nil, err
	}

	return &tls.Config{Certificates: []tls.Certificate{tlsCert}}, nil
}

func testDial(t *testing.T, target string) {
	var (
		client *grpc.ClientConn
		server *grpc.Server
		err    error
	)

	defer func() {
		if client != nil {
			client.Close()
		}
		if server != nil {
			server.Stop()
		}
	}()

	Convey("Setup server", t, func(c C) {
		tlsConf, err := generateTLSConfig()
		So(err, ShouldBeNil)

		server, l, err := qgrpc.NewServer(target, opts.TLSConfig(tlsConf))
		So(err, ShouldBeNil)

		hello.RegisterGreeterServer(server, &Hello{})

		go func() {
			err := server.Serve(l)
			c.So(err, ShouldBeNil)
		}()
	})

	Convey("Setup client", t, func() {
		tlsConf := &tls.Config{InsecureSkipVerify: true}

		// Take a random port to listen from udp server
		client, err = qgrpc.Dial(target, opts.WithTLSConfig(tlsConf))
		So(err, ShouldBeNil)
	})

	Convey("Test basic dail", t, func() {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()

		greet := hello.NewGreeterClient(client)
		req := new(hello.HelloRequest)
		req.Name = "World"

		rep, err := greet.SayHello(ctx, req)
		So(err, ShouldBeNil)
		So(rep.GetMessage(), ShouldEqual, "Hello World")
	})
}

func TestDialUDP(t *testing.T) {
	testDial(t, "127.0.0.1:5847")
}