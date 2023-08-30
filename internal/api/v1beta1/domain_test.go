package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/domain"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CreateOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.DomainService)
		req     *frontierv1beta1.CreateOrganizationDomainRequest
		want    *frontierv1beta1.CreateOrganizationDomainResponse
		wantErr error
	}{
		{name: "should create Organization Domain on success",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: "org_id",
					Name:  "example.com",
				}).Return(domain.Domain{
					ID:    "some_id",
					Name:  "example.com",
					OrgID: "org_id",
					Token: "some_token",
					State: domain.Pending,
				}, nil)
			},
			req: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  "org_id",
				Domain: "example.com",
			},
			want: &frontierv1beta1.CreateOrganizationDomainResponse{Domain: &frontierv1beta1.Domain{
				Id:        "some_id",
				Name:      "example.com",
				OrgId:     "org_id",
				Token:     "some_token",
				State:     domain.Pending.String(),
				UpdatedAt: timestamppb.New(time.Time{}),
				CreatedAt: timestamppb.New(time.Time{}),
			}},
			wantErr: nil,
		},
		{
			name: "should return error if Org Id or Name is empty",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: "",
					Name:  "",
				}).Return(domain.Domain{}, grpcBadBodyError)
			},
			req: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  "",
				Domain: "example.com",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error if Organization does not exist",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: "some-id",
					Name:  "example.com",
				}).Return(domain.Domain{}, organization.ErrNotExist)
			},
			req: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  "some-id",
				Domain: "example.com",
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error if Domain arleady exist",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: "some-id",
					Name:  "example.com",
				}).Return(domain.Domain{}, domain.ErrDuplicateKey)
			},
			req: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  "some-id",
				Domain: "example.com",
			},
			want:    nil,
			wantErr: grpcDomainAlreadyExistsErr,
		},
		{
			name: "should return internal error if domain service return some error",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), domain.Domain{
					OrgID: "some-id",
					Name:  "example.com",
				}).Return(domain.Domain{}, errors.New("some error"))
			},
			req: &frontierv1beta1.CreateOrganizationDomainRequest{
				OrgId:  "some-id",
				Domain: "example.com",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomain := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockDomain)
			}
			mockDom := Handler{domainService: mockDomain}
			resp, err := mockDom.CreateOrganizationDomain(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.DomainService)
		req     *frontierv1beta1.DeleteOrganizationDomainRequest
		want    *frontierv1beta1.DeleteOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should delete domain on success",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(nil)
			},
			req: &frontierv1beta1.DeleteOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want:    &frontierv1beta1.DeleteOrganizationDomainResponse{},
			wantErr: nil,
		},
		{
			name: "should return error if Org Id or domain Id is empty",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "").Return(grpcBadBodyError)
			},
			req: &frontierv1beta1.DeleteOrganizationDomainRequest{
				Id:    "",
				OrgId: "",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error if organization does not exist",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(organization.ErrNotExist)
			},
			req: &frontierv1beta1.DeleteOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error is domain does not exist",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.ErrNotExist)
			},
			req: &frontierv1beta1.DeleteOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcDomainNotFoundErr,
		},
		{
			name: "should return an internal server error if domain service fails to delete the domain",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(errors.New("some_error"))
			},
			req: &frontierv1beta1.DeleteOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomain := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockDomain)
			}
			mockDom := Handler{domainService: mockDomain}
			resp, err := mockDom.DeleteOrganizationDomain(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_GetOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.DomainService)
		req     *frontierv1beta1.GetOrganizationDomainRequest
		want    *frontierv1beta1.GetOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should get the Org Domain on success",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{
					ID:    "some_id",
					Name:  "example.com",
					OrgID: "org_id",
					Token: "some_token",
					State: domain.Pending,
				}, nil)
			},
			req: &frontierv1beta1.GetOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want: &frontierv1beta1.GetOrganizationDomainResponse{Domain: &frontierv1beta1.Domain{
				Id:        "some_id",
				Name:      "example.com",
				OrgId:     "org_id",
				Token:     "some_token",
				State:     domain.Pending.String(),
				UpdatedAt: timestamppb.New(time.Time{}),
				CreatedAt: timestamppb.New(time.Time{}),
			}},
			wantErr: nil,
		},
		{
			name: "should return error if Org Id or domain Id is empty",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(domain.Domain{
					ID:    "",
					Name:  "",
					OrgID: "",
					Token: "",
					State: "",
				}, grpcBadBodyError)
			},
			req: &frontierv1beta1.GetOrganizationDomainRequest{
				Id:    "",
				OrgId: "",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error is domain does not exist",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{
					ID:    "some_id",
					Name:  "example.com",
					OrgID: "org_id",
					Token: "some_tokenpending",
					State: "",
				}, domain.ErrNotExist)
			},
			req: &frontierv1beta1.GetOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcDomainNotFoundErr,
		},
		{
			name: "should return an internal server error if domain service fails to get ",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{
					ID:    "some_id",
					Name:  "example.com",
					OrgID: "org_id",
					Token: "some_tokenpending",
					State: "",
				}, errors.New("some_error"))
			},
			req: &frontierv1beta1.GetOrganizationDomainRequest{
				Id:    "some_id",
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomainServ := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockDomainServ)
			}
			mockDom := Handler{domainService: mockDomainServ}
			resp, err := mockDom.GetOrganizationDomain(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_JoinOrganization(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.DomainService, a *mocks.AuthnService)
		req     *frontierv1beta1.JoinOrganizationRequest
		want    *frontierv1beta1.JoinOrganizationResponse
		wantErr error
	}{
		{
			name: "should join the org on success",
			setup: func(m *mocks.DomainService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "user_id",
				}, nil)
				m.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), "org_id", "user_id").Return(nil)
			},
			req: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: "org_id",
			},
			want:    &frontierv1beta1.JoinOrganizationResponse{},
			wantErr: nil,
		},
		{
			name: "should return error if Org Id is empty",
			setup: func(m *mocks.DomainService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "user_id",
				}, nil)
				m.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), "", "user_id").Return(grpcBadBodyError)
			},
			req: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: "",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return an  error if Authn service returns some error",
			setup: func(m *mocks.DomainService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{}, errors.New("some_err"))
				m.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), "org_id", "").Return(grpcInternalServerError)
			},
			req: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return error if Org Id does not exist",
			setup: func(m *mocks.DomainService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "user_id",
				}, nil)
				m.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), "org_id", "user_id").Return(organization.ErrNotExist)
			},
			req: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcOrgNotFoundErr,
		},
		{
			name: "should return error is domain miss match with org id",
			setup: func(m *mocks.DomainService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "user_id",
				}, nil)
				m.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), "org_id", "user_id").Return(domain.ErrDomainsMisMatch)
			},
			req: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcDomainMisMatchErr,
		},
		{
			name: "should return an  error if domain service return some error",
			setup: func(m *mocks.DomainService, a *mocks.AuthnService) {
				a.EXPECT().GetPrincipal(mock.AnythingOfType("*context.emptyCtx")).Return(authenticate.Principal{
					ID: "user_id",
				}, nil)
				m.EXPECT().Join(mock.AnythingOfType("*context.emptyCtx"), "org_id", "user_id").Return(errors.New("some_err"))
			},
			req: &frontierv1beta1.JoinOrganizationRequest{
				OrgId: "org_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomainServ := new(mocks.DomainService)
			mockAuthSerc := new(mocks.AuthnService)
			if tt.setup != nil {
				tt.setup(mockDomainServ, mockAuthSerc)
			}
			mockDom := Handler{
				domainService: mockDomainServ,
				authnService:  mockAuthSerc,
			}
			resp, err := mockDom.JoinOrganization(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_VerifyOrganizationDomain(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.DomainService)
		req     *frontierv1beta1.VerifyOrganizationDomainRequest
		want    *frontierv1beta1.VerifyOrganizationDomainResponse
		wantErr error
	}{
		{
			name: "should Verify the Org Domain on success",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{
					State: domain.Verified,
				}, nil)
			},
			req: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: "org_id",
				Id:    "some_id",
			},
			want: &frontierv1beta1.VerifyOrganizationDomainResponse{
				State: domain.Verified.String(),
			},
			wantErr: nil,
		},
		{
			name: "should return error if Org Id or Id is empty",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), "").Return(domain.Domain{}, grpcBadBodyError)
			},
			req: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: "",
				Id:    "",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error if domain is invalid",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{}, domain.ErrInvalidDomain)
			},
			req: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: "org_id",
				Id:    "some_id",
			},
			want:    nil,
			wantErr: grpcInvalidHostErr,
		},
		{
			name: "should return error if domain does not exist",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{}, domain.ErrNotExist)
			},
			req: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: "org_id",
				Id:    "some_id",
			},
			want:    nil,
			wantErr: grpcDomainNotFoundErr,
		},
		{
			name: "should return error if TXT record not found",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{}, domain.ErrTXTrecordNotFound)
			},
			req: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: "org_id",
				Id:    "some_id",
			},
			want:    nil,
			wantErr: grpcTXTRecordNotFound,
		},
		{
			name: "should return an  error if domain service return some error",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().VerifyDomain(mock.AnythingOfType("*context.emptyCtx"), "some_id").Return(domain.Domain{}, errors.New("some_error"))
			},
			req: &frontierv1beta1.VerifyOrganizationDomainRequest{
				OrgId: "org_id",
				Id:    "some_id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomSer := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockDomSer)
			}
			mockDom := Handler{domainService: mockDomSer}
			resp, err := mockDom.VerifyOrganizationDomain(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func Test_ListOrganizationDomains(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *mocks.DomainService)
		req     *frontierv1beta1.ListOrganizationDomainsRequest
		want    *frontierv1beta1.ListOrganizationDomainsResponse
		wantErr error
	}{
		{
			name: "should list domains on success",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), domain.Filter{
					OrgID: "org_id",
					State: domain.Verified,
				}).Return([]domain.Domain{
					{
						ID:    "some_id",
						Name:  "example.com",
						OrgID: "org_id",
						Token: "some_token",
						State: domain.Pending,
					},
				}, nil)
			},

			req: &frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: "org_id",
				State: domain.Verified.String(),
			},
			want: &frontierv1beta1.ListOrganizationDomainsResponse{
				Domains: []*frontierv1beta1.Domain{
					{
						Id:        "some_id",
						Name:      "example.com",
						OrgId:     "org_id",
						Token:     "some_token",
						State:     domain.Pending.String(),
						UpdatedAt: timestamppb.New(time.Time{}),
						CreatedAt: timestamppb.New(time.Time{}),
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should return error if Org Id or State is empty",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), domain.Filter{
					OrgID: "",
					State: "",
				}).Return([]domain.Domain{}, grpcBadBodyError)
			},

			req: &frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: "",
				State: "",
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return an  error if domain service return some error",
			setup: func(m *mocks.DomainService) {
				m.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), domain.Filter{
					OrgID: "org_id",
					State: domain.Verified,
				}).Return([]domain.Domain{}, errors.New("some_error"))
			},

			req: &frontierv1beta1.ListOrganizationDomainsRequest{
				OrgId: "org_id",
				State: domain.Verified.String(),
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockDomSer := new(mocks.DomainService)
			if tt.setup != nil {
				tt.setup(mockDomSer)
			}
			mockDom := Handler{domainService: mockDomSer}
			resp, err := mockDom.ListOrganizationDomains(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
