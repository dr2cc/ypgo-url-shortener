package pb

import (
	"context"
	"errors"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"github.com/belamov/ypgo-url-shortener/internal/app/storage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) Shorten(ctx context.Context, r *ShortenRequest) (*ShorteningResponse, error) {
	if r.Url == "" {
		return nil, status.Error(codes.InvalidArgument, `full_url required`)
	}

	userID, err := s.decodeAndDecrypt(r.UserId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, `invalid user_id`)
	}

	if userID == "" {
		userID = s.service.GenerateNewUserID()
	}

	shortURL, err := s.service.Shorten(ctx, r.Url, userID)
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		// we cannot return "conflict" status with response, response becomes nil for client
		return s.newShorteningResponse(shortURL, ""), nil
	}
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return s.newShorteningResponse(shortURL, userID), nil
}

func (s *GRPCServer) newShorteningResponse(shortURL models.ShortURL, userID string) *ShorteningResponse {
	return &ShorteningResponse{
		ResultUrl: s.service.FormatShortURL(shortURL.ID),
		UserId:    userID,
		UrlId:     shortURL.ID,
	}
}
