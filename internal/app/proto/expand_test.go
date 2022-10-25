package pb

import (
	"context"
	"errors"
	"time"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ShortenTestSuite) TestExpandUrl() {
	urlID := "id"
	request := &ExpandRequest{UrlId: urlID}

	expectedURL := models.ShortURL{
		OriginalURL: "url",
		ID:          urlID,
	}
	s.mockService.EXPECT().Expand(gomock.Any(), request.UrlId).Return(expectedURL, nil)

	response, err := s.client.Expand(context.Background(), request)
	require.NoError(s.T(), err)

	assert.Equal(s.T(), response.FullUrl, expectedURL.OriginalURL)
}

func (s *ShortenTestSuite) TestExpandWithoutUrlID() {
	request := &ExpandRequest{}

	response, err := s.client.Expand(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestExpandUnexpectedError() {
	request := &ExpandRequest{UrlId: "id"}

	s.mockService.EXPECT().Expand(gomock.Any(), request.UrlId).Return(models.ShortURL{}, errors.New(""))

	response, err := s.client.Expand(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.Internal, grpcErr.Code())
}

func (s *ShortenTestSuite) TestExpandNotFound() {
	request := &ExpandRequest{UrlId: "id"}

	s.mockService.EXPECT().Expand(gomock.Any(), request.UrlId).Return(models.ShortURL{}, nil)

	response, err := s.client.Expand(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.NotFound, grpcErr.Code())
}

func (s *ShortenTestSuite) TestExpandDeleted() {
	request := &ExpandRequest{UrlId: "id"}

	s.mockService.EXPECT().Expand(gomock.Any(), request.UrlId).Return(models.ShortURL{DeletedAt: time.Now(), OriginalURL: "url"}, nil)

	response, err := s.client.Expand(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.NotFound, grpcErr.Code())
}
