package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/serviceuser"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"

	"google.golang.org/protobuf/types/known/timestamppb"
)

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
