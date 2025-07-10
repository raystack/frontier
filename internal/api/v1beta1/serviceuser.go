package v1beta1

import (
	"context"
	"encoding/json"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var grpcServiceUserNotFound = status.Error(codes.NotFound, "service user not found")
var grpcSvcUserCredNotFound = status.Error(codes.NotFound, "service user credentials not found")

type ServiceUserService interface {
	List(ctx context.Context, flt serviceuser.Filter) ([]serviceuser.ServiceUser, error)
	ListAll(ctx context.Context) ([]serviceuser.ServiceUser, error)
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
	CreateToken(ctx context.Context, credential serviceuser.Credential) (serviceuser.Token, error)
	ListToken(ctx context.Context, serviceUserID string) ([]serviceuser.Credential, error)
	DeleteToken(ctx context.Context, credID string) error
	ListByOrg(ctx context.Context, orgID string) ([]serviceuser.ServiceUser, error)
	IsSudo(ctx context.Context, id string, permissionName string) (bool, error)
	Sudo(ctx context.Context, id string, relationName string) error
	UnSudo(ctx context.Context, id string) error
	GetByIDs(ctx context.Context, ids []string) ([]serviceuser.ServiceUser, error)
}

func (h Handler) ListServiceUsers(ctx context.Context, request *frontierv1beta1.ListServiceUsersRequest) (*frontierv1beta1.ListServiceUsersResponse, error) {
	var users []*frontierv1beta1.ServiceUser
	usersList, err := h.serviceUserService.List(ctx, serviceuser.Filter{
		OrgID: request.GetOrgId(),
		State: serviceuser.State(request.GetState()),
	})
	if err != nil {
		return nil, err
	}

	for _, user := range usersList {
		userPB, err := transformServiceUserToPB(user)
		if err != nil {
			return nil, err
		}
		users = append(users, userPB)
	}

	return &frontierv1beta1.ListServiceUsersResponse{
		Serviceusers: users,
	}, nil
}

func (h Handler) ListAllServiceUsers(ctx context.Context, request *frontierv1beta1.ListAllServiceUsersRequest) (*frontierv1beta1.ListAllServiceUsersResponse, error) {
	var serviceUsers []*frontierv1beta1.ServiceUser

	// ListAll returns all service users across all organizations without any filtering
	serviceUsersList, err := h.serviceUserService.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	for _, su := range serviceUsersList {
		serviceUserPB, err := transformServiceUserToPB(su)
		if err != nil {
			return nil, err
		}
		serviceUsers = append(serviceUsers, serviceUserPB)
	}

	return &frontierv1beta1.ListAllServiceUsersResponse{
		ServiceUsers: serviceUsers,
	}, nil
}

func (h Handler) CreateServiceUser(ctx context.Context, request *frontierv1beta1.CreateServiceUserRequest) (*frontierv1beta1.CreateServiceUserResponse, error) {
	var metaDataMap metadata.Metadata
	var err error
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())
	}
	svUser, err := h.serviceUserService.Create(ctx, serviceuser.ServiceUser{
		Title:    request.GetBody().GetTitle(),
		OrgID:    request.GetOrgId(),
		Metadata: metaDataMap,
	})
	if err != nil {
		return nil, err
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		return nil, err
	}

	audit.GetAuditor(ctx, request.GetOrgId()).
		LogWithAttrs(audit.ServiceUserCreatedEvent, audit.ServiceUserTarget(svUser.ID), map[string]string{
			"title": svUser.Title,
		})
	return &frontierv1beta1.CreateServiceUserResponse{
		Serviceuser: svUserPb,
	}, nil
}

func (h Handler) GetServiceUser(ctx context.Context, request *frontierv1beta1.GetServiceUserRequest) (*frontierv1beta1.GetServiceUserResponse, error) {
	svUser, err := h.serviceUserService.Get(ctx, request.GetId())
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, err
		}
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetServiceUserResponse{
		Serviceuser: svUserPb,
	}, nil
}

func (h Handler) DeleteServiceUser(ctx context.Context, request *frontierv1beta1.DeleteServiceUserRequest) (*frontierv1beta1.DeleteServiceUserResponse, error) {
	err := h.serviceUserService.Delete(ctx, request.GetId())
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, err
		}
	}

	audit.GetAuditor(ctx, request.GetOrgId()).
		Log(audit.ServiceUserDeletedEvent, audit.ServiceUserTarget(request.GetId()))
	return &frontierv1beta1.DeleteServiceUserResponse{}, nil
}

func (h Handler) CreateServiceUserJWK(ctx context.Context, request *frontierv1beta1.CreateServiceUserJWKRequest) (*frontierv1beta1.CreateServiceUserJWKResponse, error) {
	svCred, err := h.serviceUserService.CreateKey(ctx, serviceuser.Credential{
		ServiceUserID: request.GetId(),
		Title:         request.GetTitle(),
	})
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, err
		}
	}

	svKey := &frontierv1beta1.KeyCredential{
		Type:        serviceuser.DefaultKeyType,
		Kid:         svCred.ID,
		PrincipalId: svCred.ServiceUserID,
		PrivateKey:  string(svCred.PrivateKey),
	}
	return &frontierv1beta1.CreateServiceUserJWKResponse{
		Key: svKey,
	}, nil
}

func (h Handler) ListServiceUserJWKs(ctx context.Context, request *frontierv1beta1.ListServiceUserJWKsRequest) (*frontierv1beta1.ListServiceUserJWKsResponse, error) {
	var keys []*frontierv1beta1.ServiceUserJWK
	credList, err := h.serviceUserService.ListKeys(ctx, request.GetId())
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, grpcServiceUserNotFound
		default:
			return nil, err
		}
	}

	for _, svCred := range credList {
		jwkJson, err := json.Marshal(svCred.PublicKey)
		if err != nil {
			return nil, err
		}
		keys = append(keys, &frontierv1beta1.ServiceUserJWK{
			Id:          svCred.ID,
			Title:       svCred.Title,
			PrincipalId: svCred.ServiceUserID,
			PublicKey:   string(jwkJson),
			CreatedAt:   timestamppb.New(svCred.CreatedAt),
		})
	}
	return &frontierv1beta1.ListServiceUserJWKsResponse{
		Keys: keys,
	}, nil
}

func (h Handler) GetServiceUserJWK(ctx context.Context, request *frontierv1beta1.GetServiceUserJWKRequest) (*frontierv1beta1.GetServiceUserJWKResponse, error) {
	svCred, err := h.serviceUserService.GetKey(ctx, request.GetKeyId())
	if err != nil {
		switch {
		case err == serviceuser.ErrCredNotExist:
			return nil, grpcSvcUserCredNotFound
		default:
			return nil, err
		}
	}

	jwks, err := toJSONWebKey(svCred.PublicKey)
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.GetServiceUserJWKResponse{
		Keys: jwks.Keys,
	}, nil
}

func (h Handler) DeleteServiceUserJWK(ctx context.Context, request *frontierv1beta1.DeleteServiceUserJWKRequest) (*frontierv1beta1.DeleteServiceUserJWKResponse, error) {
	err := h.serviceUserService.DeleteKey(ctx, request.GetKeyId())
	if err != nil {
		switch {
		case err == serviceuser.ErrCredNotExist:
			return nil, grpcSvcUserCredNotFound
		default:
			return nil, err
		}
	}

	return &frontierv1beta1.DeleteServiceUserJWKResponse{}, nil
}

func (h Handler) CreateServiceUserCredential(ctx context.Context, request *frontierv1beta1.CreateServiceUserCredentialRequest) (*frontierv1beta1.CreateServiceUserCredentialResponse, error) {
	secret, err := h.serviceUserService.CreateSecret(ctx, serviceuser.Credential{
		ServiceUserID: request.GetId(),
		Title:         request.GetTitle(),
	})
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateServiceUserCredentialResponse{
		Secret: &frontierv1beta1.SecretCredential{
			Id:        secret.ID,
			Title:     secret.Title,
			Secret:    secret.Value,
			CreatedAt: timestamppb.New(secret.CreatedAt),
		},
	}, nil
}

func (h Handler) ListServiceUserCredentials(ctx context.Context, request *frontierv1beta1.ListServiceUserCredentialsRequest) (*frontierv1beta1.ListServiceUserCredentialsResponse, error) {
	credentials, err := h.serviceUserService.ListSecret(ctx, request.GetId())
	if err != nil {
		return nil, err
	}
	secretsPB := make([]*frontierv1beta1.SecretCredential, 0, len(credentials))
	for _, sec := range credentials {
		secretsPB = append(secretsPB, &frontierv1beta1.SecretCredential{
			Id:        sec.ID,
			Title:     sec.Title,
			CreatedAt: timestamppb.New(sec.CreatedAt),
		})
	}
	return &frontierv1beta1.ListServiceUserCredentialsResponse{
		Secrets: secretsPB,
	}, nil
}

func (h Handler) DeleteServiceUserCredential(ctx context.Context, request *frontierv1beta1.DeleteServiceUserCredentialRequest) (*frontierv1beta1.DeleteServiceUserCredentialResponse, error) {
	err := h.serviceUserService.DeleteSecret(ctx, request.GetSecretId())
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.DeleteServiceUserCredentialResponse{}, nil
}

func (h Handler) CreateServiceUserToken(ctx context.Context, request *frontierv1beta1.CreateServiceUserTokenRequest) (*frontierv1beta1.CreateServiceUserTokenResponse, error) {
	secret, err := h.serviceUserService.CreateToken(ctx, serviceuser.Credential{
		ServiceUserID: request.GetId(),
		Title:         request.GetTitle(),
	})
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.CreateServiceUserTokenResponse{
		Token: &frontierv1beta1.ServiceUserToken{
			Id:        secret.ID,
			Title:     secret.Title,
			Token:     secret.Value,
			CreatedAt: timestamppb.New(secret.CreatedAt),
		},
	}, nil
}

func (h Handler) ListServiceUserTokens(ctx context.Context, request *frontierv1beta1.ListServiceUserTokensRequest) (*frontierv1beta1.ListServiceUserTokensResponse, error) {
	credentials, err := h.serviceUserService.ListToken(ctx, request.GetId())
	if err != nil {
		return nil, err
	}
	secretsPB := make([]*frontierv1beta1.ServiceUserToken, 0, len(credentials))
	for _, sec := range credentials {
		secretsPB = append(secretsPB, &frontierv1beta1.ServiceUserToken{
			Id:        sec.ID,
			Title:     sec.Title,
			CreatedAt: timestamppb.New(sec.CreatedAt),
		})
	}
	return &frontierv1beta1.ListServiceUserTokensResponse{
		Tokens: secretsPB,
	}, nil
}

func (h Handler) DeleteServiceUserToken(ctx context.Context, request *frontierv1beta1.DeleteServiceUserTokenRequest) (*frontierv1beta1.DeleteServiceUserTokenResponse, error) {
	err := h.serviceUserService.DeleteToken(ctx, request.GetTokenId())
	if err != nil {
		return nil, err
	}
	return &frontierv1beta1.DeleteServiceUserTokenResponse{}, nil
}

func (h Handler) ListServiceUserProjects(ctx context.Context, request *frontierv1beta1.ListServiceUserProjectsRequest) (*frontierv1beta1.ListServiceUserProjectsResponse, error) {
	projList, err := h.projectService.ListByUser(ctx, request.GetId(), schema.ServiceUserPrincipal, project.Filter{
		OrgID: request.GetOrgId(),
	})
	if err != nil {
		return nil, err
	}

	var projects []*frontierv1beta1.Project
	var accessPairsPb []*frontierv1beta1.ListServiceUserProjectsResponse_AccessPair
	for _, v := range projList {
		projPB, err := transformProjectToPB(v)
		if err != nil {
			return nil, err
		}
		projects = append(projects, projPB)
	}

	if len(request.GetWithPermissions()) > 0 {
		resourceIds := utils.Map(projList, func(res project.Project) string {
			return res.ID
		})
		successCheckPairs, err := h.fetchAccessPairsOnResource(ctx, schema.ProjectNamespace, resourceIds, request.GetWithPermissions())
		if err != nil {
			return nil, err
		}
		for _, successCheck := range successCheckPairs {
			resID := successCheck.Relation.Object.ID

			// find all permission checks on same resource
			pairsForCurrentGroup := utils.Filter(successCheckPairs, func(pair relation.CheckPair) bool {
				return pair.Relation.Object.ID == resID
			})
			// fetch permissions
			permissions := utils.Map(pairsForCurrentGroup, func(pair relation.CheckPair) string {
				return pair.Relation.RelationName
			})
			accessPairsPb = append(accessPairsPb, &frontierv1beta1.ListServiceUserProjectsResponse_AccessPair{
				ProjectId:   resID,
				Permissions: permissions,
			})
		}
	}

	return &frontierv1beta1.ListServiceUserProjectsResponse{
		Projects:    projects,
		AccessPairs: accessPairsPb,
	}, nil
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
