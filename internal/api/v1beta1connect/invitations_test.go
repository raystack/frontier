package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testInvitation1ID = uuid.New()
	testInvitation2ID = uuid.New()
	testInvitation3ID = uuid.New()
	testOrg2ID        = uuid.New().String()
	testUserEmail     = "test@raystack.org"
	testUser2Email    = "user2@raystack.org"
	testUser3Email    = "tu3@raystack.org"
	testInvitationMap = map[string]invitation.Invitation{
		testInvitation1ID.String(): {
			ID:          testInvitation1ID,
			UserEmailID: testUserEmail,
			OrgID:       testOrgID,
			GroupIDs: []string{
				testGroupID,
			},
			Metadata: metadata.Metadata{
				"group_ids": testGroupID,
			},
			CreatedAt: time.Time{},
			ExpiresAt: time.Time{},
		},
		testInvitation2ID.String(): {
			ID:          testInvitation2ID,
			UserEmailID: testUser2Email,
			OrgID:       testOrg2ID,
			GroupIDs:    []string{},
			Metadata: metadata.Metadata{
				"group_ids": "",
			},
			CreatedAt: time.Time{},
			ExpiresAt: time.Time{},
		},
		testInvitation3ID.String(): {
			ID:          testInvitation3ID,
			UserEmailID: testUser3Email,
			OrgID:       testOrg2ID,
			GroupIDs:    []string{},
			Metadata: metadata.Metadata{
				"group_ids": "",
			},
			CreatedAt: time.Time{}.AddDate(0, 0, -8),
			ExpiresAt: time.Time{}.AddDate(0, 0, -1),
		},
	}
)

func TestHandler_ListOrganizationInvitations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.ListOrganizationInvitationsRequest]
		want    *connect.Response[frontierv1beta1.ListOrganizationInvitationsResponse]
		wantErr error
	}{
		{
			name: "should return an error if listing invitation returns an error",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), invitation.Filter{
					OrgID: testOrgID,
				}).Return(nil, errors.New("new-error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationInvitationsRequest{
				OrgId: testOrgID,
			}),
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
			want:    nil,
		},
		{
			name: "should return the list of invitations belonging to an org on success",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				var testInvitationList []invitation.Invitation
				for _, u := range testInvitationMap {
					if u.OrgID == testOrgID {
						testInvitationList = append(testInvitationList, u)
					}
				}
				is.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), invitation.Filter{
					OrgID: testOrgID,
				}).Return(testInvitationList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationInvitationsRequest{
				OrgId: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListOrganizationInvitationsResponse{Invitations: []*frontierv1beta1.Invitation{
				{
					Id:       testInvitation1ID.String(),
					OrgId:    testOrgID,
					UserId:   testUserEmail,
					GroupIds: []string{testGroupID},
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"group_ids": structpb.NewStringValue(testGroupID),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					ExpiresAt: timestamppb.New(time.Time{}),
				},
			}}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvitationSvc := new(mocks.InvitationService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockInvitationSvc, mockOrgSvc)
			}
			h := ConnectHandler{
				invitationService: mockInvitationSvc,
				orgService:        mockOrgSvc,
			}
			got, err := h.ListOrganizationInvitations(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_ListCurrentUserInvitations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService, as *mocks.AuthnService)
		request *connect.Request[frontierv1beta1.ListCurrentUserInvitationsRequest]
		want    *connect.Response[frontierv1beta1.ListCurrentUserInvitationsResponse]
		wantErr error
	}{
		{
			name: "should return an error if listing current user invitations returns an error",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{
					User: &user.User{Email: testUserEmail},
				}, nil)
				is.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), testUserEmail).Return(nil, errors.New("new-error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListCurrentUserInvitationsRequest{}),
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
			want:    nil,
		},
		{
			name: "should return the list of invitations and orgs belonging to current user on success",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.AnythingOfType("context.backgroundCtx")).Return(authenticate.Principal{
					User: &user.User{Email: testUserEmail},
				}, nil)
				var testInvitationList []invitation.Invitation
				for _, u := range testInvitationMap {
					if u.UserEmailID == testUserEmail {
						testInvitationList = append(testInvitationList, u)
					}
				}
				is.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), testUserEmail).Return(testInvitationList, nil)
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListCurrentUserInvitationsRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListCurrentUserInvitationsResponse{
				Invitations: []*frontierv1beta1.Invitation{
					{
						Id:       testInvitation1ID.String(),
						OrgId:    testOrgID,
						UserId:   testUserEmail,
						GroupIds: []string{testGroupID},
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"group_ids": structpb.NewStringValue(testGroupID),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						ExpiresAt: timestamppb.New(time.Time{}),
					},
				},
				Orgs: []*frontierv1beta1.Organization{
					{
						Id:    testOrgID,
						Name:  "org-1",
						Title: "",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"email":  structpb.NewStringValue("org1@org1.com"),
								"age":    structpb.NewNumberValue(21),
								"intern": structpb.NewBoolValue(true),
							},
						},
						State:     "enabled",
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvitationSvc := new(mocks.InvitationService)
			mockOrgSvc := new(mocks.OrganizationService)
			mockAuthnSvc := new(mocks.AuthnService)
			if tt.setup != nil {
				tt.setup(mockInvitationSvc, mockOrgSvc, mockAuthnSvc)
			}
			h := ConnectHandler{
				invitationService: mockInvitationSvc,
				orgService:        mockOrgSvc,
				authnService:      mockAuthnSvc,
			}
			got, err := h.ListCurrentUserInvitations(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_ListUserInvitations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService)
		request *connect.Request[frontierv1beta1.ListUserInvitationsRequest]
		want    *connect.Response[frontierv1beta1.ListUserInvitationsResponse]
		wantErr error
	}{
		{
			name: "should return an error if listing user invitation returns an error",
			setup: func(is *mocks.InvitationService) {
				is.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), testUserEmail).Return(nil, errors.New("new-error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserInvitationsRequest{
				Id: testUserEmail,
			}),
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
			want:    nil,
		},
		{
			name: "should return the list of invitations belonging to an user on success",
			setup: func(is *mocks.InvitationService) {
				var testInvitationList []invitation.Invitation
				for _, u := range testInvitationMap {
					if u.UserEmailID == testUserEmail {
						testInvitationList = append(testInvitationList, u)
					}
				}
				is.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), testUserEmail).Return(testInvitationList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListUserInvitationsRequest{
				Id: testUserEmail,
			}),
			want: connect.NewResponse(&frontierv1beta1.ListUserInvitationsResponse{Invitations: []*frontierv1beta1.Invitation{
				{
					Id:       testInvitation1ID.String(),
					OrgId:    testOrgID,
					UserId:   testUserEmail,
					GroupIds: []string{testGroupID},
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"group_ids": structpb.NewStringValue(testGroupID),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					ExpiresAt: timestamppb.New(time.Time{}),
				},
			}}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvitationSvc := new(mocks.InvitationService)
			if tt.setup != nil {
				tt.setup(mockInvitationSvc)
			}
			h := ConnectHandler{
				invitationService: mockInvitationSvc,
			}
			got, err := h.ListUserInvitations(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_CreateOrganizationInvitation(t *testing.T) {
	randomOrgID := utils.NewString()
	randomGroupID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.CreateOrganizationInvitationRequest]
		want    *connect.Response[frontierv1beta1.CreateOrganizationInvitationResponse]
		wantErr error
	}{
		{
			name: "should create an invitation on success and return the invitation",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), invitation.Invitation{
					OrgID:       testOrgID,
					UserEmailID: testUserEmail,
					GroupIDs:    []string{randomGroupID},
					CreatedAt:   time.Time{},
					ExpiresAt:   time.Time{},
				}).Return(testInvitationMap[testInvitation1ID.String()], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    testOrgID,
				UserIds:  []string{testUserEmail},
				GroupIds: []string{randomGroupID},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateOrganizationInvitationResponse{
				Invitations: []*frontierv1beta1.Invitation{
					{
						Id:       testInvitation1ID.String(),
						OrgId:    testOrgID,
						UserId:   testUserEmail,
						GroupIds: []string{testGroupID},
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"group_ids": structpb.NewStringValue(testGroupID),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						ExpiresAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return an error if user email is not provided",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), randomOrgID).Return(testOrgMap[randomOrgID], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    randomOrgID,
				UserIds:  []string{"not-an-email"},
				GroupIds: []string{randomGroupID},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrInvalidEmail),
		},
		{
			name: "should return an error if the invitation service fails",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), invitation.Invitation{
					OrgID:       testOrgID,
					UserEmailID: testUserEmail,
					GroupIDs:    []string{randomGroupID},
					CreatedAt:   time.Time{},
					ExpiresAt:   time.Time{},
				}).Return(invitation.Invitation{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    testOrgID,
				UserIds:  []string{testUserEmail},
				GroupIds: []string{randomGroupID},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should create a new invitation with the default expiration date",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), invitation.Invitation{
					OrgID:       testOrgID,
					UserEmailID: testUserEmail,
					GroupIDs:    []string{randomGroupID},
					CreatedAt:   time.Time{},
					ExpiresAt:   time.Time{},
				}).Return(testInvitationMap[testInvitation1ID.String()], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    testOrgID,
				UserIds:  []string{testUserEmail},
				GroupIds: []string{randomGroupID},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateOrganizationInvitationResponse{
				Invitations: []*frontierv1beta1.Invitation{
					{
						Id:       testInvitation1ID.String(),
						OrgId:    testOrgID,
						UserId:   testUserEmail,
						GroupIds: []string{testGroupID},
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"group_ids": structpb.NewStringValue(testGroupID),
							},
						},
						CreatedAt: timestamppb.New(time.Time{}),
						ExpiresAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &mocks.InvitationService{}
			os := &mocks.OrganizationService{}
			if tt.setup != nil {
				tt.setup(is, os)
			}
			h := &ConnectHandler{
				invitationService: is,
				orgService:        os,
			}
			got, err := h.CreateOrganizationInvitation(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_GetOrganizationInvitation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.GetOrganizationInvitationRequest]
		want    *connect.Response[frontierv1beta1.GetOrganizationInvitationResponse]
		wantErr error
	}{
		{
			name: "should return an invitation",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(testInvitationMap[testInvitation1ID.String()], nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want: connect.NewResponse(&frontierv1beta1.GetOrganizationInvitationResponse{
				Invitation: &frontierv1beta1.Invitation{
					Id:       testInvitation1ID.String(),
					OrgId:    testOrgID,
					UserId:   testUserEmail,
					GroupIds: []string{testGroupID},
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"group_ids": structpb.NewStringValue(testGroupID),
						},
					},
					CreatedAt: timestamppb.New(time.Time{}),
					ExpiresAt: timestamppb.New(time.Time{}),
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return an error if the invitation service fails",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(invitation.Invitation{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return an error if the invitation is not found",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(invitation.Invitation{}, invitation.ErrNotFound)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &mocks.InvitationService{}
			os := &mocks.OrganizationService{}
			if tt.setup != nil {
				tt.setup(is, os)
			}
			h := &ConnectHandler{
				invitationService: is,
				orgService:        os,
			}
			got, err := h.GetOrganizationInvitation(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_AcceptOrganizationInvitation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.AcceptOrganizationInvitationRequest]
		want    *connect.Response[frontierv1beta1.AcceptOrganizationInvitationResponse]
		wantErr error
	}{
		{
			name: "should return an error if invite not found",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(invitation.ErrNotFound)
			},
			request: connect.NewRequest(&frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrInvitationNotFound),
		},
		{
			name: "should return an error if unable to get user by id",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(user.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			name: "should return an internal error if unable to accept invitation",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return error if invitation is expired",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation3ID).Return(invitation.ErrInviteExpired)
			},
			request: connect.NewRequest(&frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation3ID.String(),
				OrgId: testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrInvitationExpired),
		},
		{
			name: "should accept an invitation on success",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.AcceptOrganizationInvitationResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &mocks.InvitationService{}
			us := &mocks.UserService{}
			gs := &mocks.GroupService{}
			os := &mocks.OrganizationService{}

			if tt.setup != nil {
				tt.setup(is, us, gs, os)
			}
			h := &ConnectHandler{
				invitationService: is,
				userService:       us,
				groupService:      gs,
				orgService:        os,
			}
			got, err := h.AcceptOrganizationInvitation(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_DeleteOrganizationInvitation(t *testing.T) {
	randomOrgID := uuid.New().String()
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *connect.Request[frontierv1beta1.DeleteOrganizationInvitationRequest]
		want    *connect.Response[frontierv1beta1.DeleteOrganizationInvitationResponse]
		wantErr error
	}{
		{
			name: "should return an internal server error if invitation service fails to delete the invite",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), randomOrgID).Return(testOrgMap[randomOrgID], nil)
				is.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: randomOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should delete an invitation on success",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), randomOrgID).Return(testOrgMap[randomOrgID], nil)
				is.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: randomOrgID,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteOrganizationInvitationResponse{}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &mocks.InvitationService{}
			os := &mocks.OrganizationService{}
			if tt.setup != nil {
				tt.setup(is, os)
			}
			h := &ConnectHandler{
				invitationService: is,
				orgService:        os,
			}
			got, err := h.DeleteOrganizationInvitation(context.Background(), tt.request)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}
