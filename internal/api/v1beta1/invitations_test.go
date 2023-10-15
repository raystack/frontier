package v1beta1

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testInvitation1ID = uuid.New()
	testInvitation2ID = uuid.New()
	testOrg2ID        = uuid.New().String()
	testUserEmail     = "test@raystack.org"
	testUser2Email    = "user2@raystack.org"
	testInvitationMap = map[string]invitation.Invitation{
		testInvitation1ID.String(): {
			ID:     testInvitation1ID,
			UserID: testUserEmail,
			OrgID:  testOrgID,
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
			ID:       testInvitation2ID,
			UserID:   testUser2Email,
			OrgID:    testOrg2ID,
			GroupIDs: []string{},
			Metadata: metadata.Metadata{
				"group_ids": "",
			},
			CreatedAt: time.Time{},
			ExpiresAt: time.Time{},
		},
	}
)

func TestHandler_ListOrganizationInvitations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *frontierv1beta1.ListOrganizationInvitationsRequest
		want    *frontierv1beta1.ListOrganizationInvitationsResponse
		wantErr error
	}{
		{
			name: "should return an error if listing invitation returns an error",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				context.Background()
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), invitation.Filter{
					OrgID: testOrgID,
				}).Return(nil, errors.New("new-error"))
			},
			request: &frontierv1beta1.ListOrganizationInvitationsRequest{
				OrgId: testOrgID,
			},
			wantErr: status.Error(codes.Internal, "new-error"),
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
			request: &frontierv1beta1.ListOrganizationInvitationsRequest{
				OrgId: testOrgID,
			},
			want: &frontierv1beta1.ListOrganizationInvitationsResponse{Invitations: []*frontierv1beta1.Invitation{
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
			}},
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
			h := Handler{
				invitationService: mockInvitationSvc,
				orgService:        mockOrgSvc,
			}
			got, err := h.ListOrganizationInvitations(context.Background(), tt.request)
			assert.EqualValues(t, err, tt.wantErr)
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func TestHandler_ListUserInvitations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService)
		request *frontierv1beta1.ListUserInvitationsRequest
		want    *frontierv1beta1.ListUserInvitationsResponse
		wantErr error
	}{
		{
			name: "should return an error if listing user invitation returns an error",
			setup: func(is *mocks.InvitationService) {
				is.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), testUserEmail).Return(nil, errors.New("new-error"))
			},
			request: &frontierv1beta1.ListUserInvitationsRequest{
				Id: testUserEmail,
			},
			wantErr: status.Error(codes.Internal, "new-error"),
			want:    nil,
		},
		{
			name: "should return the list of invitations belonging to an user on success",
			setup: func(is *mocks.InvitationService) {
				var testInvitationList []invitation.Invitation
				for _, u := range testInvitationMap {
					if u.UserID == testUserEmail {
						testInvitationList = append(testInvitationList, u)
					}
				}
				is.EXPECT().ListByUser(mock.AnythingOfType("context.backgroundCtx"), testUserEmail).Return(testInvitationList, nil)
			},
			request: &frontierv1beta1.ListUserInvitationsRequest{
				Id: testUserEmail,
			},
			want: &frontierv1beta1.ListUserInvitationsResponse{Invitations: []*frontierv1beta1.Invitation{
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
			}},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockInvitationSvc := new(mocks.InvitationService)
			if tt.setup != nil {
				tt.setup(mockInvitationSvc)
			}
			h := Handler{
				invitationService: mockInvitationSvc,
			}
			got, err := h.ListUserInvitations(context.Background(), tt.request)
			assert.EqualValues(t, err, tt.wantErr)
			assert.EqualValues(t, got, tt.want)
		})
	}
}

func TestHandler_CreateOrganizationInvitation(t *testing.T) {
	randomOrgID := utils.NewString()
	randomGroupID := utils.NewString()
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *frontierv1beta1.CreateOrganizationInvitationRequest
		want    *frontierv1beta1.CreateOrganizationInvitationResponse
		wantErr error
	}{
		{
			name: "should create an invitation on success and return the invitation",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), invitation.Invitation{
					OrgID:     testOrgID,
					UserID:    testUserEmail,
					GroupIDs:  []string{randomGroupID},
					CreatedAt: time.Time{},
					ExpiresAt: time.Time{},
				}).Return(testInvitationMap[testInvitation1ID.String()], nil)
			},
			request: &frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    testOrgID,
				UserIds:  []string{testUserEmail},
				GroupIds: []string{randomGroupID},
			},
			want: &frontierv1beta1.CreateOrganizationInvitationResponse{
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
			},
			wantErr: nil,
		},
		{
			name: "should return an error if user email is not provided",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), randomOrgID).Return(testOrgMap[randomOrgID], nil)
			},
			request: &frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    randomOrgID,
				UserIds:  []string{"not-an-email"},
				GroupIds: []string{randomGroupID},
			},
			want:    nil,
			wantErr: status.Error(codes.InvalidArgument, "invalid email"),
		},
		{
			name: "should return an error if the invitation service fails",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), invitation.Invitation{
					OrgID:     testOrgID,
					UserID:    testUserEmail,
					GroupIDs:  []string{randomGroupID},
					CreatedAt: time.Time{},
					ExpiresAt: time.Time{},
				}).Return(invitation.Invitation{}, errors.New("test error"))
			},
			request: &frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    testOrgID,
				UserIds:  []string{testUserEmail},
				GroupIds: []string{randomGroupID},
			},
			want:    nil,
			wantErr: status.Error(codes.Internal, "test error"),
		},
		{
			name: "should create a new invitation with the default expiration date",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), invitation.Invitation{
					OrgID:     testOrgID,
					UserID:    testUserEmail,
					GroupIDs:  []string{randomGroupID},
					CreatedAt: time.Time{},
					ExpiresAt: time.Time{},
				}).Return(testInvitationMap[testInvitation1ID.String()], nil)
			},
			request: &frontierv1beta1.CreateOrganizationInvitationRequest{
				OrgId:    testOrgID,
				UserIds:  []string{testUserEmail},
				GroupIds: []string{randomGroupID},
			},
			want: &frontierv1beta1.CreateOrganizationInvitationResponse{
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
			},
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
			h := &Handler{
				invitationService: is,
				orgService:        os,
			}
			got, err := h.CreateOrganizationInvitation(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_GetOrganizationInvitation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *frontierv1beta1.GetOrganizationInvitationRequest
		want    *frontierv1beta1.GetOrganizationInvitationResponse
		wantErr error
	}{
		{
			name: "should return an invitation",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(testInvitationMap[testInvitation1ID.String()], nil)
			},
			request: &frontierv1beta1.GetOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want: &frontierv1beta1.GetOrganizationInvitationResponse{
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
			},
			wantErr: nil,
		},
		{
			name: "should return an error if the invitation service fails",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(invitation.Invitation{}, errors.New("test error"))
			},
			request: &frontierv1beta1.GetOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: status.Error(codes.Internal, "test error"),
		},
		{
			name: "should return an error if the invitation is not found",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(invitation.Invitation{}, invitation.ErrNotFound)
			},
			request: &frontierv1beta1.GetOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: status.Error(codes.Internal, "invitation not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			is := &mocks.InvitationService{}
			os := &mocks.OrganizationService{}
			if tt.setup != nil {
				tt.setup(is, os)
			}
			h := &Handler{
				invitationService: is,
				orgService:        os,
			}
			got, err := h.GetOrganizationInvitation(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_AcceptOrganizationInvitation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService)
		request *frontierv1beta1.AcceptOrganizationInvitationRequest
		want    *frontierv1beta1.AcceptOrganizationInvitationResponse
		wantErr error
	}{
		{
			name: "should return an error if invite not found",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(invitation.ErrNotFound)
			},
			request: &frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcInvitationNotFoundError,
		},
		{
			name: "should return an error if unable to get user by id",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(user.ErrNotExist)
			},
			request: &frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcUserNotFoundError,
		},
		{
			name: "should return an internal error if unable to accept invitation",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(errors.New("test error"))
			},
			request: &frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should accept an invitation on success",
			setup: func(is *mocks.InvitationService, us *mocks.UserService, gs *mocks.GroupService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testOrgID).Return(testOrgMap[testOrgID], nil)
				is.EXPECT().Accept(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(nil)
			},
			request: &frontierv1beta1.AcceptOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: testOrgID,
			},
			want:    &frontierv1beta1.AcceptOrganizationInvitationResponse{},
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
			h := &Handler{
				invitationService: is,
				userService:       us,
				groupService:      gs,
				orgService:        os,
			}
			got, err := h.AcceptOrganizationInvitation(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_DeleteOrganizationInvitation(t *testing.T) {
	randomOrgID := uuid.New().String()
	tests := []struct {
		name    string
		setup   func(is *mocks.InvitationService, os *mocks.OrganizationService)
		request *frontierv1beta1.DeleteOrganizationInvitationRequest
		want    *frontierv1beta1.DeleteOrganizationInvitationResponse
		wantErr error
	}{
		{
			name: "should return an internal server error if invitation service fails to delete the invite",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), randomOrgID).Return(testOrgMap[randomOrgID], nil)
				is.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(errors.New("test error"))
			},
			request: &frontierv1beta1.DeleteOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: randomOrgID,
			},
			want:    nil,
			wantErr: status.Error(codes.Internal, "test error"),
		},
		{
			name: "should delete an invitation on success",
			setup: func(is *mocks.InvitationService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), randomOrgID).Return(testOrgMap[randomOrgID], nil)
				is.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), testInvitation1ID).Return(nil)
			},
			request: &frontierv1beta1.DeleteOrganizationInvitationRequest{
				Id:    testInvitation1ID.String(),
				OrgId: randomOrgID,
			},
			want:    &frontierv1beta1.DeleteOrganizationInvitationResponse{},
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
			h := &Handler{
				invitationService: is,
				orgService:        os,
			}
			got, err := h.DeleteOrganizationInvitation(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.EqualError(t, err, tt.wantErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.want, got)
		})
	}
}
