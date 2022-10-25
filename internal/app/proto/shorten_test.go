package pb

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/services"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ShortenTestSuite) TestRequestWithoutUrl() {
	request := &ShortenRequest{}
	response, err := s.client.Shorten(context.Background(), request)
	require.Error(s.T(), err)
	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestRequestWithBadlyEncodedUserID() {
	request := &ShortenRequest{
		Url:    "url",
		UserId: "badly encoded",
	}
	response, err := s.client.Shorten(context.Background(), request)
	require.Error(s.T(), err)
	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestRequestWithBadlyEncryptedUserID() {
	encoded := hex.EncodeToString([]byte("badly encrypted"))
	decoded, err := hex.DecodeString(encoded)
	require.NoError(s.T(), err)

	request := &ShortenRequest{
		Url:    "url",
		UserId: encoded,
	}

	s.mockCrypto.EXPECT().Decrypt(decoded).Return([]byte{}, errors.New(""))

	response, err := s.client.Shorten(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestRequestWithValidUserID() {
	userID := "encrypted"
	encoded := hex.EncodeToString([]byte(userID))
	decoded, err := hex.DecodeString(encoded)
	require.NoError(s.T(), err)

	request := &ShortenRequest{
		Url:    "url",
		UserId: encoded,
	}

	expectedResult := models.ShortURL{
		OriginalURL: "url",
		ID:          "id",
		CreatedByID: userID,
	}
	s.mockCrypto.EXPECT().Decrypt(decoded).Return([]byte(userID), nil)
	s.mockService.EXPECT().Shorten(gomock.Any(), request.Url, userID).Return(expectedResult, nil)
	s.mockService.EXPECT().FormatShortURL(expectedResult.ID).Return(expectedResult.ID)

	response, err := s.client.Shorten(context.Background(), request)
	require.NoError(s.T(), err)

	expectedResponse := ShorteningResponse{
		ResultUrl: expectedResult.ID,
		UserId:    userID,
		UrlId:     expectedResult.ID,
	}
	assert.Equal(s.T(), expectedResponse.ResultUrl, response.ResultUrl)
	assert.Equal(s.T(), expectedResponse.UserId, response.UserId)
	assert.Equal(s.T(), expectedResponse.UrlId, response.UrlId)
}

func (s *ShortenTestSuite) TestRequestWithoutUserID() {
	request := &ShortenRequest{
		Url: "url",
	}

	userID := "id"
	expectedResult := models.ShortURL{
		OriginalURL: "url",
		ID:          "",
		CreatedByID: userID,
	}
	s.mockService.EXPECT().GenerateNewUserID().Return(userID)
	s.mockService.EXPECT().Shorten(gomock.Any(), request.Url, userID).Return(expectedResult, nil)
	s.mockService.EXPECT().FormatShortURL(expectedResult.ID).Return(expectedResult.ID)

	response, err := s.client.Shorten(context.Background(), request)
	require.NoError(s.T(), err)

	expectedResponse := ShorteningResponse{
		ResultUrl: expectedResult.ID,
		UserId:    userID,
		UrlId:     expectedResult.ID,
	}
	assert.Equal(s.T(), expectedResponse.ResultUrl, response.ResultUrl)
	assert.Equal(s.T(), expectedResponse.UserId, response.UserId)
	assert.Equal(s.T(), expectedResponse.UrlId, response.UrlId)
}

func (s *ShortenTestSuite) TestRequestAlreadyShorten() {
	request := &ShortenRequest{
		Url: "url",
	}

	userID := "id"
	expectedResult := models.ShortURL{
		OriginalURL: "url",
		ID:          "",
		CreatedByID: userID,
	}
	s.mockService.EXPECT().GenerateNewUserID().Return(userID)
	s.mockService.EXPECT().Shorten(gomock.Any(), request.Url, userID).Return(
		expectedResult,
		services.NewShorteningError(expectedResult, storage.NewNotUniqueURLError(expectedResult, errors.New(""))),
	)
	s.mockService.EXPECT().FormatShortURL(expectedResult.ID).Return(expectedResult.ID)

	response, err := s.client.Shorten(context.Background(), request)
	require.NoError(s.T(), err)

	expectedResponse := ShorteningResponse{
		ResultUrl: expectedResult.ID,
		UrlId:     expectedResult.ID,
	}
	assert.Equal(s.T(), expectedResponse.ResultUrl, response.ResultUrl)
	assert.Equal(s.T(), "", response.UserId)
	assert.Equal(s.T(), expectedResponse.UrlId, response.UrlId)
}

func (s *ShortenTestSuite) TestUnknownServiceError() {
	request := &ShortenRequest{
		Url: "url",
	}

	userID := "id"
	expectedResult := models.ShortURL{
		OriginalURL: "url",
		ID:          "",
		CreatedByID: userID,
	}
	s.mockService.EXPECT().GenerateNewUserID().Return(userID)
	s.mockService.EXPECT().Shorten(gomock.Any(), request.Url, userID).Return(
		expectedResult,
		services.NewShorteningError(models.ShortURL{}, errors.New("unexpected error")),
	)

	response, err := s.client.Shorten(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.Internal, grpcErr.Code())
}
