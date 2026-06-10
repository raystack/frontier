package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationDomainResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateOrganizationDomain.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	dmn, err := h.domainService.Create(ctx, domain.Domain{
		OrgID: orgResp.ID,
		Name:  request.Msg.GetDomain(),
	})
	if err != nil {
		switch err {
		case domain.ErrDuplicateKey:
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrDomainAlreadyExists)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateOrganizationDomain.Create: org_id=%s domain_name=%s: %w",
				request.Msg.GetOrgId(), request.Msg.GetDomain(), err))
		}
	}

	domainPB := transformDomainToPB(dmn)
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationDomainResponse{Domain: &domainPB}), nil
}

func (h *ConnectHandler) DeleteOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.DeleteOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.DeleteOrganizationDomainResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteOrganizationDomain.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	if err := h.domainService.Delete(ctx, request.Msg.GetId()); err != nil {
		switch err {
		case domain.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrDomainNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DeleteOrganizationDomain.Delete: org_id=%s domain_id=%s: %w",
				request.Msg.GetOrgId(), request.Msg.GetId(), err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationDomainResponse{}), nil
}

func (h *ConnectHandler) GetOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.GetOrganizationDomainResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetOrganizationDomain.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	domainResp, err := h.domainService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch err {
		case domain.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrDomainNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetOrganizationDomain.Get: org_id=%s domain_id=%s: %w",
				request.Msg.GetOrgId(), request.Msg.GetId(), err))
		}
	}

	domainPB := transformDomainToPB(domainResp)
	return connect.NewResponse(&frontierv1beta1.GetOrganizationDomainResponse{Domain: &domainPB}), nil
}

func (h *ConnectHandler) JoinOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.JoinOrganizationRequest]) (*connect.Response[frontierv1beta1.JoinOrganizationResponse], error) {
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	if err := h.domainService.Join(ctx, request.Msg.GetOrgId(), principal.ID); err != nil {
		switch {
		case errors.Is(err, domain.ErrDomainsMisMatch):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrDomainMismatch)
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, membership.ErrInvalidOrgRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, membership.ErrAlreadyMember):
			return nil, connect.NewError(connect.CodeAlreadyExists, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("JoinOrganization.Join: org_id=%s principal_id=%s: %w",
				request.Msg.GetOrgId(), principal.ID, err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.JoinOrganizationResponse{}), nil
}

func (h *ConnectHandler) VerifyOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.VerifyOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.VerifyOrganizationDomainResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("VerifyOrganizationDomain.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	domainResp, err := h.domainService.VerifyDomain(ctx, request.Msg.GetId())
	if err != nil {
		switch err {
		case domain.ErrInvalidDomain:
			return nil, connect.NewError(connect.CodeNotFound, ErrInvalidHost)
		case domain.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrDomainNotFound)
		case domain.ErrTXTrecordNotFound:
			return nil, connect.NewError(connect.CodeNotFound, ErrTXTRecordNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("VerifyOrganizationDomain.VerifyDomain: org_id=%s domain_id=%s: %w",
				request.Msg.GetOrgId(), request.Msg.GetId(), err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.VerifyOrganizationDomainResponse{State: domainResp.State.String()}), nil
}

func (h *ConnectHandler) ListOrganizationDomains(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationDomainsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationDomainsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationDomains.Get: org_id=%s: %w", request.Msg.GetOrgId(), err))
		}
	}

	domains, err := h.domainService.List(ctx, domain.Filter{OrgID: orgResp.ID, State: domain.Status(request.Msg.GetState())})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationDomains.List: org_id=%s state=%s: %w",
			request.Msg.GetOrgId(), request.Msg.GetState(), err))
	}

	var domainPBs []*frontierv1beta1.Domain
	for _, d := range domains {
		domainPBs = append(domainPBs, &frontierv1beta1.Domain{
			Id:        d.ID,
			Name:      d.Name,
			OrgId:     d.OrgID,
			Token:     d.Token,
			State:     d.State.String(),
			CreatedAt: timestamppb.New(d.CreatedAt),
			UpdatedAt: timestamppb.New(d.UpdatedAt),
		})
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationDomainsResponse{Domains: domainPBs}), nil
}

func transformDomainToPB(from domain.Domain) frontierv1beta1.Domain {
	return frontierv1beta1.Domain{
		Id:        from.ID,
		Name:      from.Name,
		OrgId:     from.OrgID,
		Token:     from.Token,
		State:     from.State.String(),
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}
}
