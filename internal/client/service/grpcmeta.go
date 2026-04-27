package service

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// contextWithBearerToken добавляет JWT в исходящие gRPC метаданные.
func contextWithBearerToken(ctx context.Context, token string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
}
