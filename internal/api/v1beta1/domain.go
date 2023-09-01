package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/pkg/errors"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	grpcDomainNotFoundErr      = status.Errorf(codes.NotFound, "domain whitelist request doesn't exist")
	grpcDomainAlreadyExistsErr = status.Errorf(codes.AlreadyExists, "domain name already exists for that organization")
	grpcInvalidHostErr         = status.Errorf(codes.NotFound, "invalid domain, no such host found")
	grpcTXTRecordNotFound      = status.Errorf(codes.NotFound, "required TXT record not found for domain verification")
	grpcDomainMisMatchErr      = status.Errorf(codes.InvalidArgument, "user and org's whitelisted domains doesn't match")
)

type DomainService interface {
	Get(ctx context.Context, id string) (domain.Domain, error)
	List(ctx context.Context, flt domain.Filter) ([]domain.Domain, error)
	ListJoinableOrgsByDomain(ctx context.Context, email string) ([]string, error)
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, toCreate domain.Domain) (domain.Domain, error)
	VerifyDomain(ctx context.Context, id string) (domain.Domain, error)
	Join(ctx context.Context, orgID string, userID string) error
}

func (h Handler) CreateOrganizationDomain(ctx context.Context, request *frontierv1beta1.CreateOrganizationDomainRequest) (*frontierv1beta1.CreateOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	if orgResp.State == organization.Disabled {
		return nil, grpcOrgDisabledErr
	}

	dmn, err := h.domainService.Create(ctx, domain.Domain{
		OrgID: orgResp.ID,
		Name:  request.GetDomain(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch err {
		case domain.ErrDuplicateKey:
			return nil, grpcDomainAlreadyExistsErr
		default:
			return nil, grpcInternalServerError
		}
	}

	domainPB := transformDomainToPB(dmn)
	return &frontierv1beta1.CreateOrganizationDomainResponse{Domain: &domainPB}, nil
}

func (h Handler) DeleteOrganizationDomain(ctx context.Context, request *frontierv1beta1.DeleteOrganizationDomainRequest) (*frontierv1beta1.DeleteOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	if orgResp.State == organization.Disabled {
		return nil, grpcOrgDisabledErr
	}

	if err := h.domainService.Delete(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		switch err {
		case domain.ErrNotExist:
			return nil, grpcDomainNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &frontierv1beta1.DeleteOrganizationDomainResponse{}, nil
}

func (h Handler) GetOrganizationDomain(ctx context.Context, request *frontierv1beta1.GetOrganizationDomainRequest) (*frontierv1beta1.GetOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	if orgResp.State == organization.Disabled {
		return nil, grpcOrgDisabledErr
	}

	domainResp, err := h.domainService.Get(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch err {
		case domain.ErrNotExist:
			return nil, grpcDomainNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	domainPB := transformDomainToPB(domainResp)
	return &frontierv1beta1.GetOrganizationDomainResponse{Domain: &domainPB}, nil
}

func (h Handler) JoinOrganization(ctx context.Context, request *frontierv1beta1.JoinOrganizationRequest) (*frontierv1beta1.JoinOrganizationResponse, error) {
	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	if orgResp.State == organization.Disabled {
		return nil, grpcOrgDisabledErr
	}

	// get current user
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	if err := h.domainService.Join(ctx, orgResp.ID, principal.ID); err != nil {
		logger.Error(err.Error())
		switch err {
		case domain.ErrDomainsMisMatch:
			return nil, grpcDomainMisMatchErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &frontierv1beta1.JoinOrganizationResponse{}, nil
}

func (h Handler) VerifyOrganizationDomain(ctx context.Context, request *frontierv1beta1.VerifyOrganizationDomainRequest) (*frontierv1beta1.VerifyOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	if orgResp.State == organization.Disabled {
		return nil, grpcOrgDisabledErr
	}

	domainResp, err := h.domainService.VerifyDomain(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch err {
		case domain.ErrInvalidDomain:
			return nil, grpcInvalidHostErr
		case domain.ErrNotExist:
			return nil, grpcDomainNotFoundErr
		case domain.ErrTXTrecordNotFound:
			return nil, grpcTXTRecordNotFound
		default:
			return nil, grpcInternalServerError
		}
	}

	return &frontierv1beta1.VerifyOrganizationDomainResponse{State: domainResp.State.String()}, nil
}

func (h Handler) ListOrganizationDomains(ctx context.Context, request *frontierv1beta1.ListOrganizationDomainsRequest) (*frontierv1beta1.ListOrganizationDomainsResponse, error) {
	logger := grpczap.Extract(ctx)
	orgResp, err := h.orgService.Get(ctx, request.GetOrgId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, organization.ErrNotExist):
			return nil, grpcOrgNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}
	if orgResp.State == organization.Disabled {
		return nil, grpcOrgDisabledErr
	}

	domains, err := h.domainService.List(ctx, domain.Filter{OrgID: orgResp.ID, State: domain.Status(request.GetState())})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
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

	return &frontierv1beta1.ListOrganizationDomainsResponse{Domains: domainPBs}, nil
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
