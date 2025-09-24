package v1beta1connect

import (
	"context"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (h *ConnectHandler) CreateOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationDomainResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	if err := h.domainService.Delete(ctx, request.Msg.GetId()); err != nil {
		switch err {
		case domain.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrDomainNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.DeleteOrganizationDomainResponse{}), nil
}

func (h *ConnectHandler) GetOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.GetOrganizationDomainResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	domainResp, err := h.domainService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch err {
		case domain.ErrNotExist:
			return nil, connect.NewError(connect.CodeNotFound, ErrDomainNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	domainPB := transformDomainToPB(domainResp)
	return connect.NewResponse(&frontierv1beta1.GetOrganizationDomainResponse{Domain: &domainPB}), nil
}

func (h *ConnectHandler) JoinOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.JoinOrganizationRequest]) (*connect.Response[frontierv1beta1.JoinOrganizationResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	// get current user
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
	}

	if err := h.domainService.Join(ctx, orgResp.ID, principal.ID); err != nil {
		switch err {
		case domain.ErrDomainsMisMatch:
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrDomainMismatch)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.JoinOrganizationResponse{}), nil
}

func (h *ConnectHandler) VerifyOrganizationDomain(ctx context.Context, request *connect.Request[frontierv1beta1.VerifyOrganizationDomainRequest]) (*connect.Response[frontierv1beta1.VerifyOrganizationDomainResponse], error) {
	_, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	return connect.NewResponse(&frontierv1beta1.VerifyOrganizationDomainResponse{State: domainResp.State.String()}), nil
}

func (h *ConnectHandler) ListOrganizationDomains(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationDomainsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationDomainsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetOrgId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeNotFound, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
		}
	}

	domains, err := h.domainService.List(ctx, domain.Filter{OrgID: orgResp.ID, State: domain.Status(request.Msg.GetState())})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, ErrInternalServerError)
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
