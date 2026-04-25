package handler

import (
	"context"
	"strings"

	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type contextKey string

const ctxUserID contextKey = "user_id"

// публичные методы не требуют JWT.
var publicMethods = map[string]bool{
	"/gophkeeper.v1.AuthService/Register": true,
	"/gophkeeper.v1.AuthService/Login":    true,
	"/gophkeeper.v1.SecretsService/Ping":  true,
}

type tokenValidator interface {
	ValidateToken(token string) (uuid.UUID, error)
}

// AuthInterceptor проверяет JWT для всех методов кроме публичных.
func AuthInterceptor(v tokenValidator) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		if publicMethods[info.FullMethod] {
			return handler(ctx, req)
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "missing metadata")
		}

		values := md.Get("authorization")
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing token")
		}

		tokenStr := strings.TrimPrefix(values[0], "Bearer ")
		userID, err := v.ValidateToken(tokenStr)
		if err != nil {
			return nil, status.Error(codes.Unauthenticated, "invalid token")
		}

		return handler(context.WithValue(ctx, ctxUserID, userID), req)
	}
}

// UserIDFromContext извлекает user_id из контекста.
func UserIDFromContext(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(ctxUserID).(uuid.UUID)
	return id, ok
}
