package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
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
	dm2PB = &frontierv1beta1.Domain{
		Id:        testDomainID2,
		Name:      "raystack.com",
		OrgId:     testOrgID,
		Token:     "_frontier-domain-verification=1234567890",
		State:     domain.Pending.String(),
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_CreateOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *connect.Request[frontierv1beta1.CreateOrganizationDomainRequest]
		want    *connect.Response[frontierv1beta1.CreateOrganizationDomainResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return already exists error if domain already exists",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), domain.Domain{
					OrgID: testOrgID,
					Name:  "raystack.org",
				}).Return(domain.Domain{}, domain.ErrDuplicateKey)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, ErrDomainAlreadyExists),
		},
		{
			name: "should return internal error if domain service fails",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), domain.Domain{
					OrgID: testOrgID,
					Name:  "raystack.org",
				}).Return(domain.Domain{}, errors.New("domain service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should create domain successfully",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), domain.Domain{
					OrgID: testOrgID,
					Name:  "raystack.org",
				}).Return(testDomainMap[testDomainID1], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  testOrgID,
				Domain: "raystack.org",
			}),
			want:    connect.NewResponse(&frontierv1beta1.CreateOrganizationDomainResponse{Domain: dm1PB}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockDomainService := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockDomainService)
			}
			mockDep := &ConnectHandler{
				orgService:    mockOrgService,
				domainService: mockDomainService,
			}
			resp, err := mockDep.CreateOrganizationDomain(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *connect.Request[frontierv1beta1.DeleteOrganizationDomainRequest]
		want    *connect.Response[frontierv1beta1.DeleteOrganizationDomainResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if domain does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrDomainNotFound),
		},
		{
			name: "should return internal error if domain service fails",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(errors.New("domain service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should delete domain successfully",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteOrganizationDomainResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockDomainService := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockDomainService)
			}
			mockDep := &ConnectHandler{
				orgService:    mockOrgService,
				domainService: mockDomainService,
			}
			resp, err := mockDep.DeleteOrganizationDomain(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *connect.Request[frontierv1beta1.GetOrganizationDomainRequest]
		want    *connect.Response[frontierv1beta1.GetOrganizationDomainResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if domain does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.Domain{}, domain.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrDomainNotFound),
		},
		{
			name: "should return internal error if domain service fails",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.Domain{}, errors.New("domain service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should get domain successfully",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(testDomainMap[testDomainID1], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    connect.NewResponse(&frontierv1beta1.GetOrganizationDomainResponse{Domain: dm1PB}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockDomainService := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockDomainService)
			}
			mockDep := &ConnectHandler{
				orgService:    mockOrgService,
				domainService: mockDomainService,
			}
			resp, err := mockDep.GetOrganizationDomain(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_JoinOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService)
		request *connect.Request[frontierv1beta1.JoinOrganizationRequest]
		want    *connect.Response[frontierv1beta1.JoinOrganizationResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: testUserID}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: testUserID}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: testUserID}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return invalid argument error if domains mismatch",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: testUserID}, nil)
				ds.EXPECT().Join(mock.AnythingOfType("context.backgroundCtx"), testOrgID, testUserID).Return(domain.ErrDomainsMisMatch)
			},
			request: connect.NewRequest(&frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrDomainMismatch),
		},
		{
			name: "should return internal error if domain service fails",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: testUserID}, nil)
				ds.EXPECT().Join(mock.AnythingOfType("context.backgroundCtx"), testOrgID, testUserID).Return(errors.New("domain service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should join organization successfully",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService, as *mocks.AuthnService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{ID: testUserID}, nil)
				ds.EXPECT().Join(mock.AnythingOfType("context.backgroundCtx"), testOrgID, testUserID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.JoinOrganizationRequest{
				OrgId: testOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.JoinOrganizationResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockDomainService := new(mocks.DomainService)
			mockAuthnService := new(mocks.AuthnService)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockDomainService, mockAuthnService)
			}
			mockDep := &ConnectHandler{
				orgService:    mockOrgService,
				domainService: mockDomainService,
				authnService:  mockAuthnService,
			}
			resp, err := mockDep.JoinOrganization(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_ListOrganizationDomains(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *connect.Request[frontierv1beta1.ListOrganizationDomainsRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationDomainsResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return internal error if domain service fails",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), domain.Filter{OrgID: testOrgID, State: domain.Status("")}).Return([]domain.Domain{}, errors.New("domain service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should list domains successfully",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				domains := []domain.Domain{testDomainMap[testDomainID1], testDomainMap[testDomainID2]}
				ds.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), domain.Filter{OrgID: testOrgID, State: domain.Status("")}).Return(domains, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationDomainsResponse{
				Domains: []*frontierv1beta1.Domain{dm1PB, dm2PB},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockDomainService := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockDomainService)
			}
			mockDep := &ConnectHandler{
				orgService:    mockOrgService,
				domainService: mockDomainService,
			}
			resp, err := mockDep.ListOrganizationDomains(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_VerifyOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(os *mocks.OrganizationService, ds *mocks.DomainService)
		request *connect.Request[frontierv1beta1.VerifyOrganizationDomainRequest]
		want    *connect.Response[frontierv1beta1.VerifyOrganizationDomainResponse]
		wantErr error
	}{
		{
			name: "should return internal error if org service return some error",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if org is disabled",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrOrgDisabled),
		},
		{
			name: "should return not found error if org does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if domain is invalid",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.Domain{}, domain.ErrInvalidDomain)
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrInvalidHost),
		},
		{
			name: "should return not found error if domain does not exist",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.Domain{}, domain.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrDomainNotFound),
		},
		{
			name: "should return not found error if TXT record not found",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.Domain{}, domain.ErrTXTrecordNotFound)
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrTXTRecordNotFound),
		},
		{
			name: "should return internal error if domain service fails",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(domain.Domain{}, errors.New("domain service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should verify domain successfully",
			setup: func(os *mocks.OrganizationService, ds *mocks.DomainService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				ds.EXPECT().VerifyDomain(mock.AnythingOfType("context.backgroundCtx"), testDomainID1).Return(testDomainMap[testDomainID1], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: testOrgID,
				Id:    testDomainID1,
			}),
			want: connect.NewResponse(&frontierv1beta1.VerifyOrganizationDomainResponse{
				State: domain.Verified.String(),
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockDomainService := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockOrgService, mockDomainService)
			}
			mockDep := &ConnectHandler{
				orgService:    mockOrgService,
				domainService: mockDomainService,
			}
			resp, err := mockDep.VerifyOrganizationDomain(context.Background(), tt.request)
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
