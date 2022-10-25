package pb

import (
	"context"
	"fmt"
	"log"
	"net"
	"testing"

	"github.com/belamov/ypgo-url-shortener/internal/app/config"
	"github.com/belamov/ypgo-url-shortener/internal/app/mocks"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

type Reporter struct {
	T *testing.T
}

// ensure Reporter implements gomock.TestReporter.
var _ gomock.TestReporter = Reporter{}

// Errorf is equivalent testing.T.Errorf.
func (r Reporter) Errorf(format string, args ...interface{}) {
	r.T.Errorf(format, args...)
}

// Fatalf crashes the program with a panic to allow users to diagnose
// missing expects.
func (r Reporter) Fatalf(format string, args ...interface{}) {
	panic(fmt.Sprintf(format, args...))
}

type ShortenTestSuite struct {
	suite.Suite
	client      ShortenerClient
	conn        *grpc.ClientConn
	mockCtrl    *gomock.Controller
	grpcServer  *grpc.Server
	mockService *mocks.MockShortenerInterface
	mockCrypto  *mocks.MockCryptographer
}

func (s *ShortenTestSuite) SetupTest() {
	grpcServer := grpc.NewServer()

	ctrl := gomock.NewController(Reporter{s.T()})

	mockService := mocks.NewMockShortenerInterface(ctrl)
	mockIPChecker := mocks.NewMockIPCheckerInterface(ctrl)
	mockCrypto := mocks.NewMockCryptographer(ctrl)

	appServer, err := NewGRPCServer(&config.Config{}, mockIPChecker, mockService, mockCrypto)
	require.NoError(s.T(), err)

	RegisterShortenerServer(grpcServer, appServer)

	s.mockCtrl = ctrl
	s.grpcServer = grpcServer

	listener := bufconn.Listen(1024 * 1024)

	dialer := func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}

	conn, err := grpc.DialContext(
		context.Background(),
		"",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(dialer),
	)
	if err != nil {
		log.Fatal(err)
	}

	go func(s *ShortenTestSuite) {
		errServe := s.grpcServer.Serve(listener)

		require.NoError(s.T(), errServe)
	}(s)

	s.client = NewShortenerClient(conn)
	s.conn = conn
	s.mockService = mockService
	s.mockCrypto = mockCrypto
}

func (s *ShortenTestSuite) TearDownTest() {
	s.mockCtrl.Finish()
	s.grpcServer.Stop()
}

func (s *ShortenTestSuite) TearDownSuite() {
	err := s.conn.Close()
	require.NoError(s.T(), err)
}

func TestShortenTestSuite(t *testing.T) {
	suite.Run(t, new(ShortenTestSuite))
}
