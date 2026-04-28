package secret

import (
	"github.com/avc-dev/gophkeeper/internal/domain"
	"github.com/avc-dev/gophkeeper/internal/protoconv"
	pb "github.com/avc-dev/gophkeeper/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func secretToProto(s *domain.Secret) *pb.Secret {
	return &pb.Secret{
		Id:               s.ID.String(),
		Name:             s.Name,
		Type:             protoconv.TypeToProto(s.Type),
		EncryptedPayload: s.Payload,
		Metadata:         s.Metadata,
		Version:          s.Version,
		CreatedAt:        timestamppb.New(s.CreatedAt),
		UpdatedAt:        timestamppb.New(s.UpdatedAt),
	}
}
