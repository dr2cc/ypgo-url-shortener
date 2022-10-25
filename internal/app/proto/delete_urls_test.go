package pb

import (
	"context"
	"encoding/hex"
	"errors"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *ShortenTestSuite) TestDeleteUrlWithBadlyEncodedUserID() {
	request := &DeleteUrlsRequest{
		UserId: "badly encoded",
	}
	response, err := s.client.DeleteUrls(context.Background(), request)
	require.Error(s.T(), err)
	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestDeleteUrlsWithBadlyEncryptedUserID() {
	encoded := hex.EncodeToString([]byte("badly encrypted"))
	decoded, err := hex.DecodeString(encoded)
	require.NoError(s.T(), err)

	request := &DeleteUrlsRequest{
		UserId: encoded,
	}

	s.mockCrypto.EXPECT().Decrypt(decoded).Return([]byte{}, errors.New(""))

	response, err := s.client.DeleteUrls(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}

func (s *ShortenTestSuite) TestDeleteUrlsWithValidUserID() {
	userID := "encrypted id"
	encoded := hex.EncodeToString([]byte(userID))
	decoded, err := hex.DecodeString(encoded)
	require.NoError(s.T(), err)

	urlIds := []string{"id1", "id2"}
	request := &DeleteUrlsRequest{
		UrlIds: urlIds,
		UserId: encoded,
	}

	s.mockCrypto.EXPECT().Decrypt(decoded).Return([]byte(userID), nil)
	s.mockService.EXPECT().DeleteUrls(gomock.Any(), urlIds, userID)

	_, err = s.client.DeleteUrls(context.Background(), request)
	require.NoError(s.T(), err)
}

func (s *ShortenTestSuite) TestDeleteUrlsWithoutUserID() {
	urlIds := []string{"id1", "id2"}
	request := &DeleteUrlsRequest{
		UrlIds: urlIds,
	}

	response, err := s.client.DeleteUrls(context.Background(), request)
	require.Error(s.T(), err)

	assert.Nil(s.T(), response)

	grpcErr, ok := status.FromError(err)
	require.True(s.T(), ok)

	assert.Equal(s.T(), codes.InvalidArgument, grpcErr.Code())
}
