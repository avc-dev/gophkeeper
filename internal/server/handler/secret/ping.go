package secret

import (
	"context"
	"time"

	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *Handler) Ping(_ context.Context, _ *pb.PingRequest) (*pb.PingResponse, error) {
	return &pb.PingResponse{ServerTime: timestamppb.New(time.Now())}, nil
}
