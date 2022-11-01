package pb

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *GRPCServer) DeleteUrls(ctx context.Context, r *DeleteUrlsRequest) (*Empty, error) {
	if r.GetUserId() == "" {
		return nil, status.Error(codes.InvalidArgument, `user_id required`)
	}

	userID, err := s.decodeAndDecrypt(r.GetUserId())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, `invalid user_id`)
	}

	s.service.DeleteUrls(ctx, r.GetUrlIds(), userID)

	return &Empty{}, err
}
