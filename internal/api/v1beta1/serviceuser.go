package v1beta1

import (
	"context"
	"encoding/json"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/shield/core/serviceuser"
	"github.com/raystack/shield/pkg/metadata"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

//go:generate mockery --name=ServiceUserService -r --case underscore --with-expecter --structname ServiceUserService --filename serviceuser_service.go --output=./mocks
type ServiceUserService interface {
	List(ctx context.Context, flt serviceuser.Filter) ([]serviceuser.ServiceUser, error)
	Create(ctx context.Context, serviceUser serviceuser.ServiceUser) (serviceuser.ServiceUser, error)
	Get(ctx context.Context, id string) (serviceuser.ServiceUser, error)
	Delete(ctx context.Context, id string) error
	ListKeys(ctx context.Context, serviceUserID string) ([]serviceuser.Credential, error)
	CreateKey(ctx context.Context, cred serviceuser.Credential) (serviceuser.Credential, error)
	GetKey(ctx context.Context, credID string) (serviceuser.Credential, error)
	DeleteKey(ctx context.Context, credID string) error
	CreateSecret(ctx context.Context, serviceUserID string) (serviceuser.Credential, error)
	GetSecret(ctx context.Context, serviceUserID string, credID string) (serviceuser.Credential, error)
	DeleteSecret(ctx context.Context, serviceUserID string, credID string) error
}

func (h Handler) ListServiceUsers(ctx context.Context, request *shieldv1beta1.ListServiceUsersRequest) (*shieldv1beta1.ListServiceUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*shieldv1beta1.ServiceUser
	usersList, err := h.serviceUserService.List(ctx, serviceuser.Filter{
		OrgID: request.GetOrgId(),
		State: serviceuser.State(request.GetState()),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, user := range usersList {
		userPB, err := transformServiceUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, userPB)
	}

	return &shieldv1beta1.ListServiceUsersResponse{
		Serviceusers: users,
	}, nil
}

func (h Handler) CreateServiceUser(ctx context.Context, request *shieldv1beta1.CreateServiceUserRequest) (*shieldv1beta1.CreateServiceUserResponse, error) {
	logger := grpczap.Extract(ctx)

	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap, err = metadata.Build(request.GetBody().GetMetadata().AsMap())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyError
		}
	}
	svUser, err := h.serviceUserService.Create(ctx, serviceuser.ServiceUser{
		Title:    request.GetBody().GetTitle(),
		OrgID:    request.GetOrgId(),
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1beta1.CreateServiceUserResponse{
		Serviceuser: svUserPb,
	}, nil
}

func (h Handler) GetServiceUser(ctx context.Context, request *shieldv1beta1.GetServiceUserRequest) (*shieldv1beta1.GetServiceUserResponse, error) {
	logger := grpczap.Extract(ctx)

	svUser, err := h.serviceUserService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1beta1.GetServiceUserResponse{
		Serviceuser: svUserPb,
	}, nil
}

func (h Handler) DeleteServiceUser(ctx context.Context, request *shieldv1beta1.DeleteServiceUserRequest) (*shieldv1beta1.DeleteServiceUserResponse, error) {
	logger := grpczap.Extract(ctx)
	err := h.serviceUserService.Delete(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.DeleteServiceUserResponse{}, nil
}

func (h Handler) CreateServiceUserKey(ctx context.Context, request *shieldv1beta1.CreateServiceUserKeyRequest) (*shieldv1beta1.CreateServiceUserKeyResponse, error) {
	logger := grpczap.Extract(ctx)

	svCred, err := h.serviceUserService.CreateKey(ctx, serviceuser.Credential{
		ServiceUserID: request.GetId(),
		Title:         request.GetTitle(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	svKey := &shieldv1beta1.KeyCredential{
		Type:        serviceuser.DefaultKeyType,
		Kid:         svCred.ID,
		PrincipalId: svCred.ServiceUserID,
		PrivateKey:  string(svCred.PrivateKey),
	}
	return &shieldv1beta1.CreateServiceUserKeyResponse{
		Key: svKey,
	}, nil
}

func (h Handler) ListServiceUserKeys(ctx context.Context, request *shieldv1beta1.ListServiceUserKeysRequest) (*shieldv1beta1.ListServiceUserKeysResponse, error) {
	logger := grpczap.Extract(ctx)
	var keys []*shieldv1beta1.ServiceUserKey
	credList, err := h.serviceUserService.ListKeys(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, svCred := range credList {
		jwkJson, err := json.Marshal(svCred.PublicKey)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		keys = append(keys, &shieldv1beta1.ServiceUserKey{
			Id:          svCred.ID,
			Title:       svCred.Title,
			PrincipalId: svCred.ServiceUserID,
			PublicKey:   string(jwkJson),
			CreatedAt:   timestamppb.New(svCred.CreatedAt),
		})
	}
	return &shieldv1beta1.ListServiceUserKeysResponse{
		Keys: keys,
	}, nil
}

func (h Handler) GetServiceUserKey(ctx context.Context, request *shieldv1beta1.GetServiceUserKeyRequest) (*shieldv1beta1.GetServiceUserKeyResponse, error) {
	logger := grpczap.Extract(ctx)
	svCred, err := h.serviceUserService.GetKey(ctx, request.GetKeyId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	jwks, err := toJSONWebKey(svCred.PublicKey)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1beta1.GetServiceUserKeyResponse{
		Keys: jwks.Keys,
	}, nil
}

func (h Handler) DeleteServiceUserKey(ctx context.Context, request *shieldv1beta1.DeleteServiceUserKeyRequest) (*shieldv1beta1.DeleteServiceUserKeyResponse, error) {
	logger := grpczap.Extract(ctx)
	err := h.serviceUserService.DeleteKey(ctx, request.GetKeyId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	return &shieldv1beta1.DeleteServiceUserKeyResponse{}, nil
}

func (h Handler) CreateServiceUserSecret(ctx context.Context, request *shieldv1beta1.CreateServiceUserSecretRequest) (*shieldv1beta1.CreateServiceUserSecretResponse, error) {
	return nil, grpcOperationUnsupported
}

func (h Handler) ListServiceUserSecrets(ctx context.Context, request *shieldv1beta1.ListServiceUserSecretsRequest) (*shieldv1beta1.ListServiceUserSecretsResponse, error) {
	return nil, grpcOperationUnsupported
}

func (h Handler) DeleteServiceUserSecret(ctx context.Context, request *shieldv1beta1.DeleteServiceUserSecretRequest) (*shieldv1beta1.DeleteServiceUserSecretResponse, error) {
	return nil, grpcOperationUnsupported
}

func transformServiceUserToPB(usr serviceuser.ServiceUser) (*shieldv1beta1.ServiceUser, error) {
	metaData, err := usr.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &shieldv1beta1.ServiceUser{
		Id:        usr.ID,
		Title:     usr.Title,
		State:     usr.State,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(usr.CreatedAt),
		UpdatedAt: timestamppb.New(usr.UpdatedAt),
	}, nil
}
