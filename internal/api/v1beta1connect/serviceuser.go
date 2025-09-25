package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

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
			return nil, connect.NewError(connect.CodeNotFound, serviceuser.ErrNotExist)
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
			return nil, connect.NewError(connect.CodeNotFound, serviceuser.ErrNotExist)
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
			return nil, connect.NewError(connect.CodeNotFound, serviceuser.ErrNotExist)
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
