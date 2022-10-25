// Package server is wrapper around built in http server.
package server

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	handlers "github.com/belamov/ypgo-url-shortener/internal/app/http_handlers"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
)

type HTTPS struct {
	server    *http.Server
	tlsConfig *tls.Config
}

func (s *HTTPS) Run() error {
	conn, err := net.Listen("tcp", s.server.Addr)
	if err != nil {
		log.Fatal(err)
	}

	tlsListener := tls.NewListener(conn, s.tlsConfig)
	return s.server.Serve(tlsListener)
}

func (s *HTTPS) Shutdown() error {
	return s.server.Shutdown(context.Background())
}

func NewHTTPS(config *config.Config, ipChecker services.IPCheckerInterface, service *services.Shortener) (Server, error) {
	server := &http.Server{
		Addr:              config.ServerAddress,
		Handler:           handlers.NewRouter(service, ipChecker, config),
		ReadHeaderTimeout: 1 * time.Second,
	}

	certificate, err := getCertificate()
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certificate},
		MinVersion:   tls.VersionTLS12,
	}

	httpsServer := &HTTPS{
		server:    server,
		tlsConfig: tlsConfig,
	}
	return httpsServer, nil
}

func getCertificate() (tls.Certificate, error) {
	// nolint TODO: save and load from file
	// tls.LoadX509KeyPair("server.crt", "server.key")
	return generateCertificate()
}

func generateCertificate() (tls.Certificate, error) {
	serialNumber, err := getRandomCertificateSerialNumber()
	if err != nil {
		return tls.Certificate{}, err
	}

	cert := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"shortener"},
			Country:      []string{"RU"},
		},
		IPAddresses:  []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}, //nolint:gomnd
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(10, 0, 0), //nolint:gomnd
		SubjectKeyId: []byte{1, 2, 3, 4, 6},        //nolint:gomnd
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	privateKey, err := rsa.GenerateKey(rand.Reader, 4096) //nolint:gomnd
	if err != nil {
		return tls.Certificate{}, err
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &privateKey.PublicKey, privateKey)
	if err != nil {
		return tls.Certificate{}, err
	}

	var certPEM bytes.Buffer
	err = pem.Encode(&certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})
	if err != nil {
		return tls.Certificate{}, err
	}

	var privateKeyPEM bytes.Buffer
	err = pem.Encode(&privateKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return tls.Certificate{}, err
	}

	return tls.X509KeyPair(certPEM.Bytes(), privateKeyPEM.Bytes())
}

func getRandomCertificateSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128) //nolint:gomnd
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	return serialNumber, err
}
