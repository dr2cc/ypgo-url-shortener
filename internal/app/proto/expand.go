package pb

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GrcpServer) Expand(ctx context.Context, r *ExpandRequest) (*ExpandResponse, error) {
	urlID := r.GetUrlId()
	if urlID == "" {
		return nil, status.Error(codes.InvalidArgument, "url_id is required")
	}
	shortURL, err := s.service.Expand(ctx, urlID)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	if shortURL.OriginalURL == "" {
		return nil, status.Error(codes.NotFound, "url id is not found")
	}

	if !shortURL.DeletedAt.IsZero() {
		return nil, status.Error(codes.NotFound, "url is deleted")
	}

	return &ExpandResponse{
		FullUrl: shortURL.OriginalURL,
	}, nil
}
