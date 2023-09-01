package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testDomainID1 = uuid.New().String()
	testDomainID2 = uuid.New().String()
	testDomainMap = map[string]domain.Domain{
		testDomainID1: {
			ID:        testDomainID1,
			Name:      "raystack.org",
			OrgID:     testOrgID,
			Token:     "_frontier-domain-verification=1234567890",
			State:     domain.Verified,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		testDomainID2: {
			ID:        testDomainID2,
			Name:      "raystack.com",
			OrgID:     testOrgID,
			Token:     "_frontier-domain-verification=1234567890",
			State:     domain.Pending,
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
	dm1PB = &frontierv1beta1.Domain{
		Id:        testDomainID1,
		Name:      "raystack.org",
		OrgId:     testOrgID,
		Token:     "_frontier-domain-verification=1234567890",
		State:     domain.Verified.String(),
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_CreateOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *frontierv1beta1.CreateOrganizationDomainRequest
		want    *frontierv1beta1.CreateOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should return error when org doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error when org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			request: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
		{
			name: "should return error when domain already exists",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: testOrgID,
					Name:  "raystack.org",
				}).Return(domain.Domain{}, domain.ErrDuplicateKey)
			},
			request: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			},
			want:    nil,
			wantErr: grpcDomainAlreadyExistsErr,
		},
		{
			name: "should create org domain with valid request",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: testOrgID,
					Name:  "raystack.org",
				}).Return(testDomainMap[testDomainID1], nil)
			},
			request: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			},
			want: &frontierv1beta1.CreateOrganizationDomainResponse{
				Domain: dm1PB,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &mocks.OrganizationService{}
			ds := &mocks.DomainService{}
			if tt.setup != nil {
				tt.setup(os, ds)
			}
			h := Handler{
				orgService:    os,
				domainService: ds,
			}
			got, err := h.CreateOrganizationDomain(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_DeleteOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *frontierv1beta1.DeleteOrganizationDomainRequest
		want    *frontierv1beta1.DeleteOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should return error when org doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error when org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			request: &frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
		{
			name: "should return error when domain doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(domain.ErrNotExist)
			},
			request: &frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcDomainNotFoundErr,
		},
		{
			name: "should delete org domain with valid request",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(nil)
			},
			request: &frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want: &frontierv1beta1.DeleteOrganizationDomainResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &mocks.OrganizationService{}
			ds := &mocks.DomainService{}
			if tt.setup != nil {
				tt.setup(os, ds)
			}
			h := Handler{
				orgService:    os,
				domainService: ds,
			}
			got, err := h.DeleteOrganizationDomain(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_GetOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *frontierv1beta1.GetOrganizationDomainRequest
		want    *frontierv1beta1.GetOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should return error when org doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error when org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			request: &frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
		{
			name: "should return error when domain doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(domain.Domain{}, domain.ErrNotExist)
			},
			request: &frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcDomainNotFoundErr,
		},
		{
			name: "should get org domain with valid request",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(testDomainMap[testDomainID1], nil)
			},
			request: &frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want: &frontierv1beta1.GetOrganizationDomainResponse{
				Domain: dm1PB,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &mocks.OrganizationService{}
			ds := &mocks.DomainService{}
			if tt.setup != nil {
				tt.setup(os, ds)
			}
			h := Handler{
				orgService:    os,
				domainService: ds,
			}
			got, err := h.GetOrganizationDomain(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_JoinOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService, us *mocks.AuthnService)
		request *frontierv1beta1.JoinOrganizationRequest
		want    *frontierv1beta1.JoinOrganizationResponse
		wantErr error
	}{
		{
			name: "should return error when org doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, us *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error when org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, us *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			request: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
		{
			name: "should return error when unable domain mismatch",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, us *mocks.AuthnService) {
				usr := testUserMap[testUserID]
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(
					authenticate.Principal{
						ID:   testUserID,
						User: &usr,
					}, nil)
				ds.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), testOrgID, testUserID).Return(domain.ErrDomainsMisMatch)
			},
			request: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcDomainMisMatchErr,
		},
		{
			name: "should join org with valid request",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, us *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				us.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(
					authenticate.Principal{
						ID: testUserID,
						User: &user.User{
							ID:    testUserID,
							Email: "test@notraystack.org",
						},
					}, nil)
				ds.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), testOrgID, testUserID).Return(nil)
			},
			request: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			},
			want: &frontierv1beta1.JoinOrganizationResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &mocks.OrganizationService{}
			ds := &mocks.DomainService{}
			us := &mocks.AuthnService{}
			if tt.setup != nil {
				tt.setup(os, ds, us)
			}
			h := Handler{
				orgService:    os,
				domainService: ds,
				authnService:  us,
			}
			got, err := h.JoinOrganization(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_ListOrganizationDomains(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *frontierv1beta1.ListOrganizationDomainsRequest
		want    *frontierv1beta1.ListOrganizationDomainsResponse
		wantErr error
	}{
		{
			name: "should return error when org doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error when org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			request: &frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
		{
			name: "should list org domains with valid request",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), domain.Filter{
					OrgID: testOrgID,
				}).Return([]domain.Domain{testDomainMap[testDomainID1]}, nil)
			},
			request: &frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			},
			want: &frontierv1beta1.ListOrganizationDomainsResponse{
				Domains: []*frontierv1beta1.Domain{dm1PB},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &mocks.OrganizationService{}
			ds := &mocks.DomainService{}
			if tt.setup != nil {
				tt.setup(os, ds)
			}
			h := Handler{
				orgService:    os,
				domainService: ds,
			}
			got, err := h.ListOrganizationDomains(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}

func TestHandler_VerifyOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *frontierv1beta1.VerifyOrganizationDomainRequest
		want    *frontierv1beta1.VerifyOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should return error when org doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error when org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(organization.Organization{State: organization.Disabled}, nil)
			},
			request: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
		{
			name: "should return error when domain doesn't exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(domain.Domain{}, domain.ErrNotExist)
			},
			request: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcDomainNotFoundErr,
		},
		{
			name: "should return error when TXT record not found",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(testDomainMap[testDomainID1], domain.ErrTXTrecordNotFound)
			},
			request: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want:    nil,
			wantErr: grpcTXTRecordNotFound,
		},
		{
			name: "should verify org domain with valid request",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), testDomainID1).Return(testDomainMap[testDomainID1], nil)
			},
			request: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			},
			want: &frontierv1beta1.VerifyOrganizationDomainResponse{
				State: domain.Verified.String(),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os := &mocks.OrganizationService{}
			ds := &mocks.DomainService{}
			if tt.setup != nil {
				tt.setup(os, ds)
			}
			h := Handler{
				orgService:    os,
				domainService: ds,
			}
			got, err := h.VerifyOrganizationDomain(context.Background(), tt.request)
			assert.EqualValues(t, tt.wantErr, err)
			assert.EqualValues(t, tt.want, got)
		})
	}
}
