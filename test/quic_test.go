package test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"math/big"
	"testing"
	"time"

	"github.com/gfanton/grpc-quic/proto/hello"
	qgrpc "github.com/ivanszl/grpc-quic"
	"github.com/ivanszl/grpc-quic/transports"
	quic "github.com/lucas-clemente/quic-go"
	. "github.com/smartystreets/goconvey/convey"
	"google.golang.org/grpc"
)

type Hello struct{}

func (h *Hello) SayHello(ctx context.Context, in *hello.HelloRequest) (*hello.HelloReply, error) {
	rep := new(hello.HelloReply)
	rep.Message = "Hello " + in.GetName()
	fmt.Println(in.GetName())
	return rep, nil
}

func generateTLSConfig() (*tls.Certificate, error) {
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

	return &tlsCert, nil
}

func TestDialUDP(t *testing.T) {
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

	conf := &quic.Config{
		KeepAlive: true,
	}
	target := "127.0.0.1:5847"

	cert, err := generateTLSConfig()

	Convey("Setup server", t, func(c C) {
		creds := transports.NewServerTLSFromCert(*cert)
		l, err := qgrpc.NewListener(target, conf, creds)
		So(err, ShouldBeNil)
		server = grpc.NewServer(grpc.Creds(creds))
		hello.RegisterGreeterServer(server, &Hello{})

		go func() {
			err := server.Serve(l)
			c.So(err, ShouldBeNil)
		}()
	})

	Convey("Setup client", t, func() {
		creds := transports.NewClientTLSSkipVerify()
		client, err = grpc.Dial(target,
			grpc.WithDialer(qgrpc.NewQuicDialer(conf, creds)),
			grpc.WithTransportCredentials(creds))

		So(err, ShouldBeNil)
	})

	Convey("Test basic dial", t, func() {
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
