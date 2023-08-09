package v1beta1

import (
	"context"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/organization"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	grpcDomainNotFoundErr = status.Errorf(codes.NotFound, "domain whitelist request doesn't exist")
	grpcInvalidHostErr    = status.Errorf(codes.NotFound, "invalid domain. No such host found")
	grpcTXTRecordNotFound = status.Errorf(codes.NotFound, "required TXT record not found for domain verification")
	grpcDomainMisMatchErr = status.Errorf(codes.InvalidArgument, "user and org's whitelisted domains doesn't match")
)

type DomainService interface {
	Get(ctx context.Context, id string) (domain.Domain, error)
	List(ctx context.Context, flt domain.Filter) ([]domain.Domain, error)
	Delete(ctx context.Context, id string) error
	Create(ctx context.Context, toCreate domain.Domain) (domain.Domain, error)
	VerifyDomain(ctx context.Context, id string) (domain.Domain, error)
	Join(ctx context.Context, orgID string, userID string) error
}

func (h Handler) AddOrganizationDomain(ctx context.Context, request *frontierv1beta1.AddOrganizationDomainRequest) (*frontierv1beta1.AddOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetOrgId() == "" || request.GetDomain() == "" {
		return nil, grpcBadBodyError
	}

	dmn, err := h.domainService.Create(ctx, domain.Domain{
		OrgID: request.GetOrgId(),
		Name:  request.GetDomain(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	domainPB := transformDomainToPB(dmn)
	return &frontierv1beta1.AddOrganizationDomainResponse{Domain: &domainPB}, nil
}

func (h Handler) RemoveOrganizationDomain(ctx context.Context, request *frontierv1beta1.RemoveOrganizationDomainRequest) (*frontierv1beta1.RemoveOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetId() == "" || request.GetOrgId() == "" {
		return nil, grpcBadBodyError
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

	return &frontierv1beta1.RemoveOrganizationDomainResponse{}, nil
}

func (h Handler) GetOrganizationDomain(ctx context.Context, request *frontierv1beta1.GetOrganizationDomainRequest) (*frontierv1beta1.GetOrganizationDomainResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetId() == "" || request.GetOrgId() == "" {
		return nil, grpcBadBodyError
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
	orgId := request.GetOrgId()
	if orgId == "" {
		return nil, grpcBadBodyError
	}

	// get current user
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	if err := h.domainService.Join(ctx, orgId, principal.ID); err != nil {
		logger.Error(err.Error())
		switch err {
		case organization.ErrNotExist:
			return nil, grpcOrgNotFoundErr
		case domain.ErrDomainsMisMatch:
			return nil, grpcDomainMisMatchErr
		default:
			return nil, grpcInternalServerError
		}
	}

	return &frontierv1beta1.JoinOrganizationResponse{}, nil
}

func (h Handler) VerifyOrgDomain(ctx context.Context, request *frontierv1beta1.VerifyOrgDomainRequest) (*frontierv1beta1.VerifyOrgDomainResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetId() == "" || request.GetOrgId() == "" {
		return nil, grpcBadBodyError
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

	return &frontierv1beta1.VerifyOrgDomainResponse{State: domainResp.State}, nil
}

func (h Handler) ListOrganizationDomains(ctx context.Context, request *frontierv1beta1.ListOrganizationDomainsRequest) (*frontierv1beta1.ListOrganizationDomainsResponse, error) {
	logger := grpczap.Extract(ctx)

	if request.GetOrgId() == "" {
		return nil, grpcBadBodyError
	}

	domains, err := h.domainService.List(ctx, domain.Filter{OrgID: request.GetOrgId(), State: request.GetState()})
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
			State:     d.State,
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
		State:     from.State,
		CreatedAt: timestamppb.New(from.CreatedAt),
		UpdatedAt: timestamppb.New(from.UpdatedAt),
	}
}
