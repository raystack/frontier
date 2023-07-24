package v1beta1

import (
	"context"
	"encoding/json"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcServiceUserNotFound = status.Error(codes.NotFound, "service user not found")
var grpcSvcUserCredNotFound = status.Error(codes.NotFound, "service user credentials not found")

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
	CreateSecret(ctx context.Context, credential serviceuser.Credential) (serviceuser.Secret, error)
	ListSecret(ctx context.Context, serviceUserID string) ([]serviceuser.Credential, error)
	DeleteSecret(ctx context.Context, credID string) error
	ListByOrg(ctx context.Context, orgID string) ([]serviceuser.ServiceUser, error)
}

func (h Handler) ListServiceUsers(ctx context.Context, request *frontierv1beta1.ListServiceUsersRequest) (*frontierv1beta1.ListServiceUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*frontierv1beta1.ServiceUser
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

	return &frontierv1beta1.ListServiceUsersResponse{
		Serviceusers: users,
	}, nil
}

func (h Handler) CreateServiceUser(ctx context.Context, request *frontierv1beta1.CreateServiceUserRequest) (*frontierv1beta1.CreateServiceUserResponse, error) {
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
	audit.GetAuditor(ctx, request.GetOrgId()).
		Log(audit.ServiceUserCreatedEvent, audit.ServiceUserTarget(svUser.ID))
	return &frontierv1beta1.CreateServiceUserResponse{
		Serviceuser: svUserPb,
	}, nil
}

func (h Handler) GetServiceUser(ctx context.Context, request *frontierv1beta1.GetServiceUserRequest) (*frontierv1beta1.GetServiceUserResponse, error) {
	logger := grpczap.Extract(ctx)

	svUser, err := h.serviceUserService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.GetServiceUserResponse{
		Serviceuser: svUserPb,
	}, nil
}

func (h Handler) DeleteServiceUser(ctx context.Context, request *frontierv1beta1.DeleteServiceUserRequest) (*frontierv1beta1.DeleteServiceUserResponse, error) {
	logger := grpczap.Extract(ctx)
	err := h.serviceUserService.Delete(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	audit.GetAuditor(ctx, request.GetOrgId()).
		Log(audit.ServiceUserDeletedEvent, audit.ServiceUserTarget(request.GetId()))
	return &frontierv1beta1.DeleteServiceUserResponse{}, nil
}

func (h Handler) CreateServiceUserKey(ctx context.Context, request *frontierv1beta1.CreateServiceUserKeyRequest) (*frontierv1beta1.CreateServiceUserKeyResponse, error) {
	logger := grpczap.Extract(ctx)

	svCred, err := h.serviceUserService.CreateKey(ctx, serviceuser.Credential{
		ServiceUserID: request.GetId(),
		Title:         request.GetTitle(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	svKey := &frontierv1beta1.KeyCredential{
		Type:        serviceuser.DefaultKeyType,
		Kid:         svCred.ID,
		PrincipalId: svCred.ServiceUserID,
		PrivateKey:  string(svCred.PrivateKey),
	}
	return &frontierv1beta1.CreateServiceUserKeyResponse{
		Key: svKey,
	}, nil
}

func (h Handler) ListServiceUserKeys(ctx context.Context, request *frontierv1beta1.ListServiceUserKeysRequest) (*frontierv1beta1.ListServiceUserKeysResponse, error) {
	logger := grpczap.Extract(ctx)
	var keys []*frontierv1beta1.ServiceUserKey
	credList, err := h.serviceUserService.ListKeys(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	for _, svCred := range credList {
		jwkJson, err := json.Marshal(svCred.PublicKey)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		keys = append(keys, &frontierv1beta1.ServiceUserKey{
			Id:          svCred.ID,
			Title:       svCred.Title,
			PrincipalId: svCred.ServiceUserID,
			PublicKey:   string(jwkJson),
			CreatedAt:   timestamppb.New(svCred.CreatedAt),
		})
	}
	return &frontierv1beta1.ListServiceUserKeysResponse{
		Keys: keys,
	}, nil
}

func (h Handler) GetServiceUserKey(ctx context.Context, request *frontierv1beta1.GetServiceUserKeyRequest) (*frontierv1beta1.GetServiceUserKeyResponse, error) {
	logger := grpczap.Extract(ctx)
	svCred, err := h.serviceUserService.GetKey(ctx, request.GetKeyId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case err == serviceuser.ErrCredNotExist:
			return nil, grpcSvcUserCredNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	jwks, err := toJSONWebKey(svCred.PublicKey)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.GetServiceUserKeyResponse{
		Keys: jwks.Keys,
	}, nil
}

func (h Handler) DeleteServiceUserKey(ctx context.Context, request *frontierv1beta1.DeleteServiceUserKeyRequest) (*frontierv1beta1.DeleteServiceUserKeyResponse, error) {
	logger := grpczap.Extract(ctx)
	err := h.serviceUserService.DeleteKey(ctx, request.GetKeyId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case err == serviceuser.ErrCredNotExist:
			return nil, grpcSvcUserCredNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	return &frontierv1beta1.DeleteServiceUserKeyResponse{}, nil
}

func (h Handler) CreateServiceUserSecret(ctx context.Context, request *frontierv1beta1.CreateServiceUserSecretRequest) (*frontierv1beta1.CreateServiceUserSecretResponse, error) {
	logger := grpczap.Extract(ctx)
	secret, err := h.serviceUserService.CreateSecret(ctx, serviceuser.Credential{
		ServiceUserID: request.GetId(),
		Title:         request.GetTitle(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.CreateServiceUserSecretResponse{
		Secret: &frontierv1beta1.SecretCredential{
			Id:        secret.ID,
			Secret:    string(secret.Value),
			CreatedAt: timestamppb.New(secret.CreatedAt),
		},
	}, nil
}

func (h Handler) ListServiceUserSecrets(ctx context.Context, request *frontierv1beta1.ListServiceUserSecretsRequest) (*frontierv1beta1.ListServiceUserSecretsResponse, error) {
	logger := grpczap.Extract(ctx)

	credentials, err := h.serviceUserService.ListSecret(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	secretsPB := make([]*frontierv1beta1.SecretCredential, 0, len(credentials))
	for _, sec := range credentials {
		secretsPB = append(secretsPB, &frontierv1beta1.SecretCredential{
			Id:        sec.ID,
			Title:     sec.Title,
			CreatedAt: timestamppb.New(sec.CreatedAt),
		})
	}
	return &frontierv1beta1.ListServiceUserSecretsResponse{
		Secrets: secretsPB,
	}, nil
}

func (h Handler) DeleteServiceUserSecret(ctx context.Context, request *frontierv1beta1.DeleteServiceUserSecretRequest) (*frontierv1beta1.DeleteServiceUserSecretResponse, error) {
	logger := grpczap.Extract(ctx)
	err := h.serviceUserService.DeleteSecret(ctx, request.GetSecretId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.DeleteServiceUserSecretResponse{}, nil
}

func transformServiceUserToPB(usr serviceuser.ServiceUser) (*frontierv1beta1.ServiceUser, error) {
	metaData, err := usr.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.ServiceUser{
		Id:        usr.ID,
		OrgId:     usr.OrgID,
		Title:     usr.Title,
		State:     usr.State,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(usr.CreatedAt),
		UpdatedAt: timestamppb.New(usr.UpdatedAt),
	}, nil
}
