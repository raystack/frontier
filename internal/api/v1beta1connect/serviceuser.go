package v1beta1connect

import (
	"context"
	"encoding/json"

	"connectrpc.com/connect"
	"github.com/lestrrat-go/jwx/v2/jwk"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type JsonWebKeySet struct {
	Keys []*frontierv1beta1.JSONWebKey `json:"keys"`
}

func toJSONWebKey(keySet jwk.Set) (*JsonWebKeySet, error) {
	jwks := &JsonWebKeySet{
		Keys: []*frontierv1beta1.JSONWebKey{},
	}
	keySetJson, err := json.Marshal(keySet)
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(keySetJson, &jwks); err != nil {
		return nil, err
	}
	return jwks, nil
}

func (h *ConnectHandler) ListServiceUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListServiceUsersResponse], error) {
	var users []*frontierv1beta1.ServiceUser
	usersList, err := h.serviceUserService.List(ctx, serviceuser.Filter{
		OrgID: request.Msg.GetOrgId(),
		State: serviceuser.State(request.Msg.GetState()),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, user := range usersList {
		userPB, err := transformServiceUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		users = append(users, userPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
		Serviceusers: users,
	}), nil
}

func (h *ConnectHandler) ListAllServiceUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListAllServiceUsersResponse], error) {
	var serviceUsers []*frontierv1beta1.ServiceUser
	serviceUsersList, err := h.serviceUserService.ListAll(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	for _, su := range serviceUsersList {
		serviceUserPB, err := transformServiceUserToPB(su)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		serviceUsers = append(serviceUsers, serviceUserPB)
	}

	return connect.NewResponse(&frontierv1beta1.ListAllServiceUsersResponse{
		ServiceUsers: serviceUsers,
	}), nil
}

func (h *ConnectHandler) GetServiceUser(ctx context.Context, request *connect.Request[frontierv1beta1.GetServiceUserRequest]) (*connect.Response[frontierv1beta1.GetServiceUserResponse], error) {
	svUser, err := h.serviceUserService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, err)
	}
	return connect.NewResponse(&frontierv1beta1.GetServiceUserResponse{
		Serviceuser: svUserPb,
	}), nil
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

func (h *ConnectHandler) CreateServiceUser(ctx context.Context, request *connect.Request[frontierv1beta1.CreateServiceUserRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserResponse], error) {
	var metaDataMap metadata.Metadata
	if request.Msg.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())
	}

	svUser, err := h.serviceUserService.Create(ctx, serviceuser.ServiceUser{
		Title:    request.Msg.GetBody().GetTitle(),
		OrgID:    request.Msg.GetOrgId(),
		Metadata: metaDataMap,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	svUserPb, err := transformServiceUserToPB(svUser)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).
		LogWithAttrs(audit.ServiceUserCreatedEvent, audit.ServiceUserTarget(svUser.ID), map[string]string{
			"title": svUser.Title,
		})

	return connect.NewResponse(&frontierv1beta1.CreateServiceUserResponse{
		Serviceuser: svUserPb,
	}), nil
}

func (h *ConnectHandler) DeleteServiceUser(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteServiceUserRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserResponse], error) {
	err := h.serviceUserService.Delete(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	audit.GetAuditor(ctx, request.Msg.GetOrgId()).
		Log(audit.ServiceUserDeletedEvent, audit.ServiceUserTarget(request.Msg.GetId()))

	return connect.NewResponse(&frontierv1beta1.DeleteServiceUserResponse{}), nil
}

func (h *ConnectHandler) CreateServiceUserJWK(ctx context.Context, request *connect.Request[frontierv1beta1.CreateServiceUserJWKRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserJWKResponse], error) {
	svCred, err := h.serviceUserService.CreateKey(ctx, serviceuser.Credential{
		ServiceUserID: request.Msg.GetId(),
		Title:         request.Msg.GetTitle(),
	})
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserCredNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	svKey := &frontierv1beta1.KeyCredential{
		Type:        serviceuser.DefaultKeyType,
		Kid:         svCred.ID,
		PrincipalId: svCred.ServiceUserID,
		PrivateKey:  string(svCred.PrivateKey),
	}
	return connect.NewResponse(&frontierv1beta1.CreateServiceUserJWKResponse{
		Key: svKey,
	}), nil
}

func (h *ConnectHandler) ListServiceUserJWKs(ctx context.Context, request *connect.Request[frontierv1beta1.ListServiceUserJWKsRequest]) (*connect.Response[frontierv1beta1.ListServiceUserJWKsResponse], error) {
	var keys []*frontierv1beta1.ServiceUserJWK
	credList, err := h.serviceUserService.ListKeys(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case err == serviceuser.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserCredNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	for _, svCred := range credList {
		jwkJson, err := json.Marshal(svCred.PublicKey)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		keys = append(keys, &frontierv1beta1.ServiceUserJWK{
			Id:          svCred.ID,
			Title:       svCred.Title,
			PrincipalId: svCred.ServiceUserID,
			PublicKey:   string(jwkJson),
			CreatedAt:   timestamppb.New(svCred.CreatedAt),
		})
	}
	return connect.NewResponse(&frontierv1beta1.ListServiceUserJWKsResponse{
		Keys: keys,
	}), nil
}

func (h *ConnectHandler) GetServiceUserJWK(ctx context.Context, request *connect.Request[frontierv1beta1.GetServiceUserJWKRequest]) (*connect.Response[frontierv1beta1.GetServiceUserJWKResponse], error) {
	svCred, err := h.serviceUserService.GetKey(ctx, request.Msg.GetKeyId())
	if err != nil {
		switch {
		case err == serviceuser.ErrCredNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserCredNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	jwks, err := toJSONWebKey(svCred.PublicKey)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.GetServiceUserJWKResponse{
		Keys: jwks.Keys,
	}), nil
}

func (h *ConnectHandler) DeleteServiceUserJWK(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteServiceUserJWKRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserJWKResponse], error) {
	err := h.serviceUserService.DeleteKey(ctx, request.Msg.GetKeyId())
	if err != nil {
		switch {
		case err == serviceuser.ErrCredNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrServiceUserCredNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteServiceUserJWKResponse{}), nil
}

func (h *ConnectHandler) CreateServiceUserCredential(ctx context.Context, request *connect.Request[frontierv1beta1.CreateServiceUserCredentialRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserCredentialResponse], error) {
	secret, err := h.serviceUserService.CreateSecret(ctx, serviceuser.Credential{
		ServiceUserID: request.Msg.GetId(),
		Title:         request.Msg.GetTitle(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	return connect.NewResponse(&frontierv1beta1.CreateServiceUserCredentialResponse{
		Secret: &frontierv1beta1.SecretCredential{
			Id:        secret.ID,
			Title:     secret.Title,
			Secret:    secret.Value,
			CreatedAt: timestamppb.New(secret.CreatedAt),
		},
	}), nil
}

func (h *ConnectHandler) ListServiceUserCredentials(ctx context.Context, request *connect.Request[frontierv1beta1.ListServiceUserCredentialsRequest]) (*connect.Response[frontierv1beta1.ListServiceUserCredentialsResponse], error) {
	credentials, err := h.serviceUserService.ListSecret(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	secretsPB := make([]*frontierv1beta1.SecretCredential, 0, len(credentials))
	for _, sec := range credentials {
		secretsPB = append(secretsPB, &frontierv1beta1.SecretCredential{
			Id:        sec.ID,
			Title:     sec.Title,
			CreatedAt: timestamppb.New(sec.CreatedAt),
		})
	}
	return connect.NewResponse(&frontierv1beta1.ListServiceUserCredentialsResponse{
		Secrets: secretsPB,
	}), nil
}

func (h *ConnectHandler) DeleteServiceUserCredential(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteServiceUserCredentialRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserCredentialResponse], error) {
	err := h.serviceUserService.DeleteSecret(ctx, request.Msg.GetSecretId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.DeleteServiceUserCredentialResponse{}), nil
}

func (h *ConnectHandler) CreateServiceUserToken(ctx context.Context, request *connect.Request[frontierv1beta1.CreateServiceUserTokenRequest]) (*connect.Response[frontierv1beta1.CreateServiceUserTokenResponse], error) {
	secret, err := h.serviceUserService.CreateToken(ctx, serviceuser.Credential{
		ServiceUserID: request.Msg.GetId(),
		Title:         request.Msg.GetTitle(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.CreateServiceUserTokenResponse{
		Token: &frontierv1beta1.ServiceUserToken{
			Id:        secret.ID,
			Title:     secret.Title,
			Token:     secret.Value,
			CreatedAt: timestamppb.New(secret.CreatedAt),
		},
	}), nil
}

func (h *ConnectHandler) ListServiceUserTokens(ctx context.Context, request *connect.Request[frontierv1beta1.ListServiceUserTokensRequest]) (*connect.Response[frontierv1beta1.ListServiceUserTokensResponse], error) {
	credentials, err := h.serviceUserService.ListToken(ctx, request.Msg.GetId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	secretsPB := make([]*frontierv1beta1.ServiceUserToken, 0, len(credentials))
	for _, sec := range credentials {
		secretsPB = append(secretsPB, &frontierv1beta1.ServiceUserToken{
			Id:        sec.ID,
			Title:     sec.Title,
			CreatedAt: timestamppb.New(sec.CreatedAt),
		})
	}
	return connect.NewResponse(&frontierv1beta1.ListServiceUserTokensResponse{
		Tokens: secretsPB,
	}), nil
}

func (h *ConnectHandler) DeleteServiceUserToken(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteServiceUserTokenRequest]) (*connect.Response[frontierv1beta1.DeleteServiceUserTokenResponse], error) {
	err := h.serviceUserService.DeleteToken(ctx, request.Msg.GetTokenId())
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}
	return connect.NewResponse(&frontierv1beta1.DeleteServiceUserTokenResponse{}), nil
}

func (h *ConnectHandler) ListServiceUserProjects(ctx context.Context, request *connect.Request[frontierv1beta1.ListServiceUserProjectsRequest]) (*connect.Response[frontierv1beta1.ListServiceUserProjectsResponse], error) {
	projList, err := h.projectService.ListByUser(ctx, request.Msg.GetId(), schema.ServiceUserPrincipal, project.Filter{
		OrgID: request.Msg.GetOrgId(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
	}

	var projects []*frontierv1beta1.Project
	var accessPairsPb []*frontierv1beta1.ListServiceUserProjectsResponse_AccessPair
	for _, v := range projList {
		projPB, err := transformProjectToPB(v)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
		projects = append(projects, projPB)
	}

	if len(request.Msg.GetWithPermissions()) > 0 {
		resourceIds := utils.Map(projList, func(res project.Project) string {
			return res.ID
		})
		successCheckPairs, err := h.fetchAccessPairsOnResource(ctx, schema.ProjectNamespace, resourceIds, request.Msg.GetWithPermissions())
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

	return connect.NewResponse(&frontierv1beta1.ListServiceUserProjectsResponse{
		Projects:    projects,
		AccessPairs: accessPairsPb,
	}), nil
}
