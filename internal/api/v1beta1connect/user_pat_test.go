package v1beta1connect

import (
	"context"
	"fmt"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/core/userpat"
	paterrors "github.com/raystack/frontier/core/userpat/errors"
	"github.com/raystack/frontier/core/userpat/models"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CreateCurrentUserPAT(t *testing.T) {
	testTime := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	testCreatedAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	testOrgID := "9f256f86-31a3-11ec-8d3d-0242ac130003"
	testUserID := "8e256f86-31a3-11ec-8d3d-0242ac130003"
	testRoleID := "7d256f86-31a3-11ec-8d3d-0242ac130003"

	tests := []struct {
		name    string
		setup   func(ps *mocks.UserPATService, as *mocks.AuthnService)
		request *connect.Request[frontierv1beta1.CreateCurrentUserPATRequest]
		want    *frontierv1beta1.CreateCurrentUserPATResponse
		wantErr error
	}{
		{
			name: "should return unauthenticated error when GetLoggedInPrincipal fails",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return permission denied when principal is not a user",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   "sv-1",
					Type: schema.ServiceUserPrincipal,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated),
		},
		{
			name: "should return invalid argument when expiry is in the past",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(paterrors.ErrExpiryInPast)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, paterrors.ErrExpiryInPast),
		},
		{
			name: "should return invalid argument when expiry exceeds max lifetime",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(paterrors.ErrExpiryExceeded)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(time.Now().Add(48 * time.Hour)),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, paterrors.ErrExpiryExceeded),
		},
		{
			name: "should return failed precondition when PAT is disabled",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", paterrors.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeFailedPrecondition, paterrors.ErrDisabled),
		},
		{
			name: "should return already exists when title conflicts",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", paterrors.ErrConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, paterrors.ErrConflict),
		},
		{
			name: "should return resource exhausted when limit exceeded",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", paterrors.ErrLimitExceeded)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeResourceExhausted, paterrors.ErrLimitExceeded),
		},
		{
			name: "should return invalid argument when role is not found",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", fmt.Errorf("fetching roles: %w", paterrors.ErrRoleNotFound))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, paterrors.ErrRoleNotFound),
		},
		{
			name: "should return invalid argument when role is denied",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", fmt.Errorf("creating policies: %w", paterrors.ErrDeniedRole))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, paterrors.ErrDeniedRole),
		},
		{
			name: "should return invalid argument when role scope is unsupported",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", fmt.Errorf("creating policies: %w", paterrors.ErrUnsupportedScope))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, paterrors.ErrUnsupportedScope),
		},
		{
			name: "should return internal error for unknown service failure",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{}, "", errors.New("unexpected error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should create PAT successfully and return response",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.MatchedBy(func(req userpat.CreateRequest) bool {
					return req.UserID == testUserID &&
						req.OrgID == testOrgID &&
						req.Title == "my-token" &&
						len(req.RoleIDs) == 1 && req.RoleIDs[0] == testRoleID
				})).Return(models.PAT{
					ID:        "pat-1",
					UserID:    testUserID,
					OrgID:     testOrgID,
					Title:     "my-token",
					ExpiresAt: testTime,
					CreatedAt: testCreatedAt,
					UpdatedAt: testCreatedAt,
				}, "fpt_abc123", nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
			}),
			want: &frontierv1beta1.CreateCurrentUserPATResponse{
				Pat: &frontierv1beta1.PAT{
					Id:        "pat-1",
					UserId:    testUserID,
					OrgId:     testOrgID,
					Title:     "my-token",
					Token:     "fpt_abc123",
					ExpiresAt: timestamppb.New(testTime),
					CreatedAt: timestamppb.New(testCreatedAt),
					UpdatedAt: timestamppb.New(testCreatedAt),
				},
			},
			wantErr: nil,
		},
		{
			name: "should create PAT with metadata",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().ValidateExpiry(mock.AnythingOfType("time.Time")).Return(nil)
				ps.EXPECT().Create(mock.Anything, mock.AnythingOfType("userpat.CreateRequest")).
					Return(models.PAT{
						ID:        "pat-1",
						UserID:    testUserID,
						OrgID:     testOrgID,
						Title:     "my-token",
						ExpiresAt: testTime,
						CreatedAt: testCreatedAt,
						UpdatedAt: testCreatedAt,
						Metadata:  metadata.Metadata{"env": "staging"},
					}, "fpt_xyz789", nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateCurrentUserPATRequest{
				Title:     "my-token",
				OrgId:     testOrgID,
				RoleIds:   []string{testRoleID},
				ExpiresAt: timestamppb.New(testTime),
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"env": structpb.NewStringValue("staging"),
					},
				},
			}),
			want: &frontierv1beta1.CreateCurrentUserPATResponse{
				Pat: &frontierv1beta1.PAT{
					Id:        "pat-1",
					UserId:    testUserID,
					OrgId:     testOrgID,
					Title:     "my-token",
					Token:     "fpt_xyz789",
					ExpiresAt: timestamppb.New(testTime),
					CreatedAt: timestamppb.New(testCreatedAt),
					UpdatedAt: timestamppb.New(testCreatedAt),
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"env": structpb.NewStringValue("staging"),
						},
					},
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPATSrv := new(mocks.UserPATService)
			mockAuthnSrv := new(mocks.AuthnService)

			if tt.setup != nil {
				tt.setup(mockPATSrv, mockAuthnSrv)
			}

			handler := &ConnectHandler{
				userPATService: mockPATSrv,
				authnService:   mockAuthnSrv,
			}

			resp, err := handler.CreateCurrentUserPAT(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, resp.Msg)
			} else {
				assert.Nil(t, resp)
			}

			mockPATSrv.AssertExpectations(t)
			mockAuthnSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_GetCurrentUserPAT(t *testing.T) {
	testCreatedAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	testExpiry := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	testUserID := "8e256f86-31a3-11ec-8d3d-0242ac130003"
	testPATID := "6c256f86-31a3-11ec-8d3d-0242ac130003"

	tests := []struct {
		name    string
		setup   func(ps *mocks.UserPATService, as *mocks.AuthnService)
		request *connect.Request[frontierv1beta1.GetCurrentUserPATRequest]
		want    *frontierv1beta1.GetCurrentUserPATResponse
		wantErr error
	}{
		{
			name: "should return unauthenticated error when GetLoggedInPrincipal fails",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
			},
			request: connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
				Id: testPATID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return permission denied when principal is not a user",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   "sv-1",
					Type: schema.ServiceUserPrincipal,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
				Id: testPATID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated),
		},
		{
			name: "should return failed precondition when PAT is disabled",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Get(mock.Anything, testUserID, testPATID).
					Return(models.PAT{}, paterrors.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
				Id: testPATID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeFailedPrecondition, paterrors.ErrDisabled),
		},
		{
			name: "should return not found when PAT does not exist",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Get(mock.Anything, testUserID, testPATID).
					Return(models.PAT{}, paterrors.ErrNotFound)
			},
			request: connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
				Id: testPATID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, paterrors.ErrNotFound),
		},
		{
			name: "should return internal error for unknown failure",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Get(mock.Anything, testUserID, testPATID).
					Return(models.PAT{}, errors.New("unexpected error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
				Id: testPATID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return PAT successfully",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Get(mock.Anything, testUserID, testPATID).
					Return(models.PAT{
						ID:        testPATID,
						UserID:    testUserID,
						OrgID:     "org-1",
						Title:     "my-token",
						RoleIDs:   []string{"role-1"},
						ExpiresAt: testExpiry,
						CreatedAt: testCreatedAt,
						UpdatedAt: testCreatedAt,
					}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetCurrentUserPATRequest{
				Id: testPATID,
			}),
			want: &frontierv1beta1.GetCurrentUserPATResponse{
				Pat: &frontierv1beta1.PAT{
					Id:        testPATID,
					UserId:    testUserID,
					OrgId:     "org-1",
					Title:     "my-token",
					RoleIds:   []string{"role-1"},
					ExpiresAt: timestamppb.New(testExpiry),
					CreatedAt: timestamppb.New(testCreatedAt),
					UpdatedAt: timestamppb.New(testCreatedAt),
				},
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPATSrv := new(mocks.UserPATService)
			mockAuthnSrv := new(mocks.AuthnService)

			if tt.setup != nil {
				tt.setup(mockPATSrv, mockAuthnSrv)
			}

			handler := &ConnectHandler{
				userPATService: mockPATSrv,
				authnService:   mockAuthnSrv,
			}

			resp, err := handler.GetCurrentUserPAT(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want, resp.Msg)
			} else {
				assert.Nil(t, resp)
			}

			mockPATSrv.AssertExpectations(t)
			mockAuthnSrv.AssertExpectations(t)
		})
	}
}

func TestTransformPATToPB(t *testing.T) {
	testTime := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	testCreatedAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)
	testLastUsed := time.Date(2026, 3, 1, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name     string
		pat      models.PAT
		patValue string
		want     *frontierv1beta1.PAT
	}{
		{
			name: "should transform minimal PAT",
			pat: models.PAT{
				ID:        "pat-1",
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				ExpiresAt: testTime,
				CreatedAt: testCreatedAt,
				UpdatedAt: testCreatedAt,
			},
			patValue: "",
			want: &frontierv1beta1.PAT{
				Id:        "pat-1",
				UserId:    "user-1",
				OrgId:     "org-1",
				Title:     "my-token",
				ExpiresAt: timestamppb.New(testTime),
				CreatedAt: timestamppb.New(testCreatedAt),
				UpdatedAt: timestamppb.New(testCreatedAt),
			},
		},
		{
			name: "should include token value when provided",
			pat: models.PAT{
				ID:        "pat-1",
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				ExpiresAt: testTime,
				CreatedAt: testCreatedAt,
				UpdatedAt: testCreatedAt,
			},
			patValue: "fpt_abc123",
			want: &frontierv1beta1.PAT{
				Id:        "pat-1",
				UserId:    "user-1",
				OrgId:     "org-1",
				Title:     "my-token",
				Token:     "fpt_abc123",
				ExpiresAt: timestamppb.New(testTime),
				CreatedAt: timestamppb.New(testCreatedAt),
				UpdatedAt: timestamppb.New(testCreatedAt),
			},
		},
		{
			name: "should include last_used_at when set",
			pat: models.PAT{
				ID:         "pat-1",
				UserID:     "user-1",
				OrgID:      "org-1",
				Title:      "my-token",
				ExpiresAt:  testTime,
				CreatedAt:  testCreatedAt,
				UpdatedAt:  testCreatedAt,
				LastUsedAt: &testLastUsed,
			},
			patValue: "",
			want: &frontierv1beta1.PAT{
				Id:         "pat-1",
				UserId:     "user-1",
				OrgId:      "org-1",
				Title:      "my-token",
				ExpiresAt:  timestamppb.New(testTime),
				CreatedAt:  timestamppb.New(testCreatedAt),
				UpdatedAt:  timestamppb.New(testCreatedAt),
				LastUsedAt: timestamppb.New(testLastUsed),
			},
		},
		{
			name: "should include metadata when set",
			pat: models.PAT{
				ID:        "pat-1",
				UserID:    "user-1",
				OrgID:     "org-1",
				Title:     "my-token",
				ExpiresAt: testTime,
				CreatedAt: testCreatedAt,
				UpdatedAt: testCreatedAt,
				Metadata:  metadata.Metadata{"env": "prod"},
			},
			patValue: "fpt_xyz",
			want: &frontierv1beta1.PAT{
				Id:        "pat-1",
				UserId:    "user-1",
				OrgId:     "org-1",
				Title:     "my-token",
				Token:     "fpt_xyz",
				ExpiresAt: timestamppb.New(testTime),
				CreatedAt: timestamppb.New(testCreatedAt),
				UpdatedAt: timestamppb.New(testCreatedAt),
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"env": structpb.NewStringValue("prod"),
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := transformPATToPB(tt.pat, tt.patValue)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestHandler_ListRolesForPAT(t *testing.T) {
	testCreatedAt := time.Date(2026, 2, 10, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		setup   func(ps *mocks.UserPATService)
		want    *frontierv1beta1.ListRolesForPATResponse
		wantErr error
	}{
		{
			name: "should return failed precondition when PAT is disabled",
			setup: func(ps *mocks.UserPATService) {
				ps.EXPECT().ListAllowedRoles(mock.Anything, mock.Anything).Return(nil, paterrors.ErrDisabled)
			},
			wantErr: connect.NewError(connect.CodeFailedPrecondition, paterrors.ErrDisabled),
		},
		{
			name: "should return invalid argument on unsupported scope",
			setup: func(ps *mocks.UserPATService) {
				ps.EXPECT().ListAllowedRoles(mock.Anything, mock.Anything).
					Return(nil, fmt.Errorf("scope %q: %w", "group", paterrors.ErrUnsupportedScope))
			},
			wantErr: connect.NewError(connect.CodeInvalidArgument, fmt.Errorf("scope %q: %w", "group", paterrors.ErrUnsupportedScope)),
		},
		{
			name: "should return internal error on service failure",
			setup: func(ps *mocks.UserPATService) {
				ps.EXPECT().ListAllowedRoles(mock.Anything, mock.Anything).Return(nil, errors.New("db error"))
			},
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return roles successfully",
			setup: func(ps *mocks.UserPATService) {
				ps.EXPECT().ListAllowedRoles(mock.Anything, mock.Anything).Return([]role.Role{
					{
						ID:          "role-1",
						Name:        "org_viewer",
						Title:       "Organization Viewer",
						Permissions: []string{"app_organization_get"},
						Scopes:      []string{schema.OrganizationNamespace},
						OrgID:       schema.PlatformOrgID.String(),
						CreatedAt:   testCreatedAt,
						UpdatedAt:   testCreatedAt,
					},
					{
						ID:          "role-2",
						Name:        "proj_viewer",
						Title:       "Project Viewer",
						Permissions: []string{"app_project_get"},
						Scopes:      []string{schema.ProjectNamespace},
						OrgID:       schema.PlatformOrgID.String(),
						CreatedAt:   testCreatedAt,
						UpdatedAt:   testCreatedAt,
					},
				}, nil)
			},
			want: func() *frontierv1beta1.ListRolesForPATResponse {
				emptyMeta, _ := metadata.Metadata(nil).ToStructPB()
				return &frontierv1beta1.ListRolesForPATResponse{
					Roles: []*frontierv1beta1.Role{
						{
							Id:          "role-1",
							Name:        "org_viewer",
							Title:       "Organization Viewer",
							Permissions: []string{"app_organization_get"},
							Scopes:      []string{schema.OrganizationNamespace},
							OrgId:       schema.PlatformOrgID.String(),
							Metadata:    emptyMeta,
							CreatedAt:   timestamppb.New(testCreatedAt),
							UpdatedAt:   timestamppb.New(testCreatedAt),
						},
						{
							Id:          "role-2",
							Name:        "proj_viewer",
							Title:       "Project Viewer",
							Permissions: []string{"app_project_get"},
							Scopes:      []string{schema.ProjectNamespace},
							OrgId:       schema.PlatformOrgID.String(),
							Metadata:    emptyMeta,
							CreatedAt:   timestamppb.New(testCreatedAt),
							UpdatedAt:   timestamppb.New(testCreatedAt),
						},
					},
				}
			}(),
		},
		{
			name: "should return empty response when no roles available",
			setup: func(ps *mocks.UserPATService) {
				ps.EXPECT().ListAllowedRoles(mock.Anything, mock.Anything).Return([]role.Role{}, nil)
			},
			want: &frontierv1beta1.ListRolesForPATResponse{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPATSrv := new(mocks.UserPATService)
			if tt.setup != nil {
				tt.setup(mockPATSrv)
			}

			handler := &ConnectHandler{
				userPATService: mockPATSrv,
			}

			resp, err := handler.ListRolesForPAT(context.Background(), connect.NewRequest(&frontierv1beta1.ListRolesForPATRequest{}))

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, resp.Msg)
			}

			mockPATSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_DeleteCurrentUserPAT(t *testing.T) {
	testUserID := "8e256f86-31a3-11ec-8d3d-0242ac130003"
	testPATID := "6c256f86-31a3-11ec-8d3d-0242ac130003"

	tests := []struct {
		name    string
		setup   func(ps *mocks.UserPATService, as *mocks.AuthnService)
		request *connect.Request[frontierv1beta1.DeleteCurrentUserPATRequest]
		wantErr error
	}{
		{
			name: "should return unauthenticated error when GetLoggedInPrincipal fails",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
				Id: testPATID,
			}),
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return permission denied when principal is not a user",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   "sv-1",
					Type: schema.ServiceUserPrincipal,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
				Id: testPATID,
			}),
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated),
		},
		{
			name: "should return failed precondition when PAT is disabled",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Delete(mock.Anything, testUserID, testPATID).
					Return(paterrors.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
				Id: testPATID,
			}),
			wantErr: connect.NewError(connect.CodeFailedPrecondition, paterrors.ErrDisabled),
		},
		{
			name: "should return not found when PAT does not exist",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Delete(mock.Anything, testUserID, testPATID).
					Return(paterrors.ErrNotFound)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
				Id: testPATID,
			}),
			wantErr: connect.NewError(connect.CodeNotFound, paterrors.ErrNotFound),
		},
		{
			name: "should return internal error for unknown failure",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Delete(mock.Anything, testUserID, testPATID).
					Return(errors.New("unexpected error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
				Id: testPATID,
			}),
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should delete PAT successfully",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().Delete(mock.Anything, testUserID, testPATID).
					Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteCurrentUserPATRequest{
				Id: testPATID,
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPATSrv := new(mocks.UserPATService)
			mockAuthnSrv := new(mocks.AuthnService)

			if tt.setup != nil {
				tt.setup(mockPATSrv, mockAuthnSrv)
			}

			handler := &ConnectHandler{
				userPATService: mockPATSrv,
				authnService:   mockAuthnSrv,
			}

			resp, err := handler.DeleteCurrentUserPAT(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}

			mockPATSrv.AssertExpectations(t)
			mockAuthnSrv.AssertExpectations(t)
		})
	}
}

func TestHandler_CheckCurrentUserPATTitle(t *testing.T) {
	testUserID := "8e256f86-31a3-11ec-8d3d-0242ac130003"
	testOrgID := "9f256f86-31a3-11ec-8d3d-0242ac130003"

	tests := []struct {
		name    string
		setup   func(ps *mocks.UserPATService, as *mocks.AuthnService)
		request *connect.Request[frontierv1beta1.CheckCurrentUserPATTitleRequest]
		wantErr error
		wantAvl bool
	}{
		{
			name: "should return unauthenticated error when GetLoggedInPrincipal fails",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{}, errors.ErrUnauthenticated)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
				OrgId: testOrgID,
				Title: "my-token",
			}),
			wantErr: connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated),
		},
		{
			name: "should return permission denied when principal is not a user",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   "sv-1",
					Type: schema.ServiceUserPrincipal,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
				OrgId: testOrgID,
				Title: "my-token",
			}),
			wantErr: connect.NewError(connect.CodePermissionDenied, ErrUnauthenticated),
		},
		{
			name: "should return failed precondition when PAT is disabled",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().IsTitleAvailable(mock.Anything, testUserID, testOrgID, "my-token").
					Return(false, paterrors.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
				OrgId: testOrgID,
				Title: "my-token",
			}),
			wantErr: connect.NewError(connect.CodeFailedPrecondition, paterrors.ErrDisabled),
		},
		{
			name: "should return internal error for unknown failure",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().IsTitleAvailable(mock.Anything, testUserID, testOrgID, "my-token").
					Return(false, errors.New("db error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
				OrgId: testOrgID,
				Title: "my-token",
			}),
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return available true when title is free",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().IsTitleAvailable(mock.Anything, testUserID, testOrgID, "new-token").
					Return(true, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
				OrgId: testOrgID,
				Title: "new-token",
			}),
			wantAvl: true,
		},
		{
			name: "should return available false when title is taken",
			setup: func(ps *mocks.UserPATService, as *mocks.AuthnService) {
				as.EXPECT().GetPrincipal(mock.Anything).Return(authenticate.Principal{
					ID:   testUserID,
					Type: schema.UserPrincipal,
					User: &user.User{ID: testUserID},
				}, nil)
				ps.EXPECT().IsTitleAvailable(mock.Anything, testUserID, testOrgID, "existing-token").
					Return(false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CheckCurrentUserPATTitleRequest{
				OrgId: testOrgID,
				Title: "existing-token",
			}),
			wantAvl: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPATSrv := new(mocks.UserPATService)
			mockAuthnSrv := new(mocks.AuthnService)

			if tt.setup != nil {
				tt.setup(mockPATSrv, mockAuthnSrv)
			}

			handler := &ConnectHandler{
				userPATService: mockPATSrv,
				authnService:   mockAuthnSrv,
			}

			resp, err := handler.CheckCurrentUserPATTitle(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.wantAvl, resp.Msg.GetAvailable())
			}

			mockPATSrv.AssertExpectations(t)
			mockAuthnSrv.AssertExpectations(t)
		})
	}
}
