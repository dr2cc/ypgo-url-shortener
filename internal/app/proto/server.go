package pb

import (
	"encoding/hex"
	"log"
	"net"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/services/crypto"
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
)

type GrcpServer struct {
	UnimplementedShortenerServer
	ipChecker services.IPCheckerInterface
	service   services.ShortenerInterface
	server    *grpc.Server
	crypto    crypto.Cryptographer // interface that we'll use to encrypt and decrypt values
}

func (s *GrcpServer) Run() error {
	RegisterShortenerServer(s.server, s)

	listen, err := net.Listen("tcp", "localhost:3200")
	if err != nil {
		log.Fatal(err)
	}

	return s.server.Serve(listen)
}

func (s *GrcpServer) Shutdown() error {
	s.server.GracefulStop()
	return nil
}

func NewGRPCServer(
	_ *config.Config,
	ipChecker services.IPCheckerInterface,
	service services.ShortenerInterface,
	cryptographer crypto.Cryptographer,
) (*GrcpServer, error) {
	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_middleware.ChainStreamServer(
			grpc_recovery.StreamServerInterceptor(),
		)),
		grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(
			grpc_recovery.UnaryServerInterceptor(),
		)),
	)
	return &GrcpServer{
		server:    s,
		ipChecker: ipChecker,
		service:   service,
		crypto:    cryptographer,
	}, nil
}

// decodeAndDecrypt takes encrypted and encoded to hex string and returns decoded and decrypted string.
func (s *GrcpServer) decodeAndDecrypt(userID string) (string, error) {
	if userID == "" {
		return "", nil
	}
	decodedEncryptedUserID, err := hex.DecodeString(userID)
	if err != nil {
		return "", err
	}

	decryptedUserID, err := s.crypto.Decrypt(decodedEncryptedUserID)
	if err != nil {
		return "", err
	}

	return string(decryptedUserID), nil
}
