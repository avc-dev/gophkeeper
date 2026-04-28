package service

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// ContextWithBearerToken adds a JWT Bearer token to outgoing gRPC metadata.
func ContextWithBearerToken(ctx context.Context, token string) context.Context {
	return metadata.NewOutgoingContext(ctx, metadata.Pairs("authorization", "Bearer "+token))
}
