package auth

import (
	"context"

	pb "github.com/avc-dev/gophkeeper/proto"
)

// service — локальный интерфейс; реализуется service/auth.Service.
type service interface {
	Register(ctx context.Context, email, password string) (token string, kdfSalt []byte, err error)
	Login(ctx context.Context, email, password string) (token string, kdfSalt []byte, err error)
}

// Handler реализует gRPC AuthService (регистрация и вход).
type Handler struct {
	pb.UnimplementedAuthServiceServer
	svc service
}

// New создаёт Handler с переданным сервисом аутентификации.
func New(svc service) *Handler {
	return &Handler{svc: svc}
}
