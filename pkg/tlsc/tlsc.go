package tlsc

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"os"
)

var ErrAppendCa = errors.New("failed to append ca to pool")

func LoadServerConfig(certFile, keyFile, clientCaFile string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(clientCaFile)
	if err != nil {
		return nil, err
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, ErrAppendCa
	}
	return &tls.Config{
		ClientAuth:   tls.RequireAndVerifyClientCert,
		Certificates: []tls.Certificate{cert},
		ClientCAs:    ca,
	}, nil
}

func LoadClientConfig(certFile, keyFile, serverCaFile, serverName string) (*tls.Config, error) {
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, err
	}
	ca := x509.NewCertPool()
	caBytes, err := os.ReadFile(serverCaFile)
	if err != nil {
		return nil, err
	}
	if ok := ca.AppendCertsFromPEM(caBytes); !ok {
		return nil, ErrAppendCa
	}
	return &tls.Config{
		ServerName:         serverName,
		Certificates:       []tls.Certificate{cert},
		RootCAs:            ca,
		InsecureSkipVerify: false,
	}, nil
}
