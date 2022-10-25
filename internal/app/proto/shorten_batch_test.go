package pb

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ShortenTestSuite) TestShortenBatchWithoutUrl() {
	userID := "id"
	encoded := hex.EncodeToString([]byte(userID))

	request := &ShortenBatchRequest{
		Urls: []*ShortenBatchItemRequest{
			{OriginalUrl: "url1", CorrelationId: "corId1"},
			{OriginalUrl: "", CorrelationId: "corId2"},
		},
		UserId: encoded,
	}

	s.mockCrypto.EXPECT().Decrypt([]byte(userID)).Return([]byte{123}, nil)

	response, err := s.client.ShortenBatch(context.Background(), request)
	require.Error(s.T(), err)
	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestShortenUnexpectedError() {
	userID := "id"
	encoded := hex.EncodeToString([]byte(userID))

	request := &ShortenBatchRequest{
		Urls: []*ShortenBatchItemRequest{
			{OriginalUrl: "url1", CorrelationId: "corId1"},
			{OriginalUrl: "url3", CorrelationId: "corId2"},
		},
		UserId: encoded,
	}

	s.mockCrypto.EXPECT().Decrypt([]byte(userID)).Return([]byte{123}, nil)
	s.mockService.EXPECT().ShortenBatch(gomock.Any(), gomock.Any(), gomock.Any()).Return(nil, errors.New(""))

	response, err := s.client.ShortenBatch(context.Background(), request)
	require.Error(s.T(), err)
	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.Internal, grpcErr.Code())
}

func (s *ShortenTestSuite) TestShortenBatchWithBadlyEncodedUserID() {
	request := &ShortenBatchRequest{
		Urls:   []*ShortenBatchItemRequest{},
		UserId: "badly encoded",
	}
	response, err := s.client.ShortenBatch(context.Background(), request)
	require.Error(s.T(), err)
	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestShortenBatchWithBadlyEncryptedUserID() {
	encoded := hex.EncodeToString([]byte("badly encrypted"))
	decoded, err := hex.DecodeString(encoded)
	require.NoError(s.T(), err)

	request := &ShortenBatchRequest{
		Urls:   []*ShortenBatchItemRequest{},
		UserId: encoded,
	}

	s.mockCrypto.EXPECT().Decrypt(decoded).Return([]byte{}, errors.New(""))

	response, err := s.client.ShortenBatch(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestShortenBatchWithValidUserID() {
	userID := "encrypted"
	encoded := hex.EncodeToString([]byte(userID))
	decoded, err := hex.DecodeString(encoded)
	require.NoError(s.T(), err)

	request := &ShortenBatchRequest{
		Urls: []*ShortenBatchItemRequest{
			{OriginalUrl: "url1", CorrelationId: "corId1"},
			{OriginalUrl: "url2", CorrelationId: "corId2"},
		},
		UserId: encoded,
	}

	expectedResult := []models.ShortURL{
		{OriginalURL: "url1", ID: "id1", CreatedByID: userID, CorrelationID: "corId1"},
		{OriginalURL: "url2", ID: "id2", CreatedByID: userID, CorrelationID: "corId1"},
	}
	s.mockCrypto.EXPECT().Decrypt(decoded).Return([]byte(userID), nil)
	s.mockService.EXPECT().FormatShortURL(expectedResult[0].ID).Return(expectedResult[0].ID)
	s.mockService.EXPECT().FormatShortURL(expectedResult[1].ID).Return(expectedResult[1].ID)
	s.mockService.EXPECT().ShortenBatch(gomock.Any(), gomock.Any(), userID).Return(expectedResult, nil)

	response, err := s.client.ShortenBatch(context.Background(), request)
	require.NoError(s.T(), err)

	assert.Len(s.T(), response.Urls, len(expectedResult))
	for i, expectedItem := range response.Urls {
		assert.Equal(s.T(), expectedItem.UserId, response.Urls[i].UserId)
		assert.Equal(s.T(), expectedItem.UrlId, response.Urls[i].UrlId)
		assert.Equal(s.T(), expectedItem.ResultUrl, response.Urls[i].ResultUrl)
		assert.Equal(s.T(), expectedItem.CorrelationId, response.Urls[i].CorrelationId)
	}
}

func (s *ShortenTestSuite) TestShortenBatchWithoutUserID() {
	request := &ShortenBatchRequest{
		Urls: []*ShortenBatchItemRequest{
			{OriginalUrl: "url1", CorrelationId: "corId1"},
			{OriginalUrl: "url2", CorrelationId: "corId2"},
		},
		UserId: "",
	}

	newUserId := "id"

	expectedResult := []models.ShortURL{
		{OriginalURL: "url1", ID: "id1", CreatedByID: newUserId, CorrelationID: "corId1"},
		{OriginalURL: "url2", ID: "id2", CreatedByID: newUserId, CorrelationID: "corId1"},
	}
	s.mockService.EXPECT().FormatShortURL(expectedResult[0].ID).Return(expectedResult[0].ID)
	s.mockService.EXPECT().FormatShortURL(expectedResult[1].ID).Return(expectedResult[1].ID)
	s.mockService.EXPECT().ShortenBatch(gomock.Any(), gomock.Any(), newUserId).Return(expectedResult, nil)
	s.mockService.EXPECT().GenerateNewUserID().Return(newUserId)

	response, err := s.client.ShortenBatch(context.Background(), request)
	require.NoError(s.T(), err)

	assert.Len(s.T(), response.Urls, len(expectedResult))
	for i, expectedItem := range response.Urls {
		assert.Equal(s.T(), expectedItem.UserId, response.Urls[i].UserId)
		assert.Equal(s.T(), expectedItem.UrlId, response.Urls[i].UrlId)
		assert.Equal(s.T(), expectedItem.ResultUrl, response.Urls[i].ResultUrl)
		assert.Equal(s.T(), expectedItem.CorrelationId, response.Urls[i].CorrelationId)
	}
}
