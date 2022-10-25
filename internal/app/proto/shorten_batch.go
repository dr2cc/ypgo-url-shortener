package pb

import (
	"context"

	"github.com/belamov/ypgo-url-shortener/internal/app/models"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GrcpServer) ShortenBatch(ctx context.Context, r *ShortenBatchRequest) (*ShortenBatchResponse, error) {
	userID, err := s.decodeAndDecrypt(r.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, `invalid user_id`)
	}

	if userID == "" {
		userID = s.service.GenerateNewUserID()
	}

	batch := make([]models.ShortURL, len(r.GetUrls()))

	for i, url := range r.GetUrls() {
		if url.OriginalUrl == "" {
			return nil, status.Error(codes.InvalidArgument, `full_url required`)
		}
		batch[i] = models.ShortURL{
			OriginalURL:   url.OriginalUrl,
			CorrelationID: url.CorrelationId,
		}
	}

	shortURLBatches, err := s.service.ShortenBatch(ctx, batch, userID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := make([]*ShortenBatchItemResponse, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = &ShortenBatchItemResponse{
			CorrelationId: shortURLBatch.CorrelationID,
			ResultUrl:     s.service.FormatShortURL(shortURLBatch.ID),
		}
	}

	return &ShortenBatchResponse{
		Urls: res,
	}, nil
}
