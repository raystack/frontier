package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestConnectHandler_CreatePolicy(t *testing.T) {
	fixedTime := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)
	testPolicyID := utils.NewString()
	testUserID := utils.NewString()
	testResourceID := utils.NewString()
	testGroupID := utils.NewString()

	tests := []struct {
		name    string
		setup   func(ps *mocks.PolicyService)
		request *connect.Request[frontierv1beta1.CreatePolicyRequest]
		want    *connect.Response[frontierv1beta1.CreatePolicyResponse]
		wantErr error
		errCode connect.Code
	}{
		{
			name: "should return invalid argument error when resource namespace splitting fails",
			setup: func(ps *mocks.PolicyService) {
				// No expectations as we return early on resource splitting error
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "admin",
					Resource:  "invalid-resource-format",
					Principal: "user:" + testUserID,
				},
			}),
			want:    nil,
			wantErr: ErrBadRequest,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return invalid argument error when principal namespace splitting fails",
			setup: func(ps *mocks.PolicyService) {
				// No expectations as we return early on principal splitting error
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "admin",
					Resource:  "project:" + testResourceID,
					Principal: "invalid-principal-format",
				},
			}),
			want:    nil,
			wantErr: ErrBadRequest,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return invalid argument error when role ID is invalid",
			setup: func(ps *mocks.PolicyService) {
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "invalid-role",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Metadata(nil),
				}).Return(policy.Policy{}, role.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "invalid-role",
					Resource:  "project:" + testResourceID,
					Principal: "user:" + testUserID,
				},
			}),
			want:    nil,
			wantErr: role.ErrInvalidID,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return invalid argument error when policy details are invalid",
			setup: func(ps *mocks.PolicyService) {
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "admin",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Metadata(nil),
				}).Return(policy.Policy{}, policy.ErrInvalidDetail)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "admin",
					Resource:  "project:" + testResourceID,
					Principal: "user:" + testUserID,
				},
			}),
			want:    nil,
			wantErr: ErrBadRequest,
			errCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return internal server error when policy service returns unknown error",
			setup: func(ps *mocks.PolicyService) {
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "admin",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Metadata(nil),
				}).Return(policy.Policy{}, errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "admin",
					Resource:  "project:" + testResourceID,
					Principal: "user:" + testUserID,
				},
			}),
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
		{
			name: "should successfully create policy with basic data",
			setup: func(ps *mocks.PolicyService) {
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "admin",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Metadata(nil),
				}).Return(policy.Policy{
					ID:            testPolicyID,
					RoleID:        "admin",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Metadata{},
					CreatedAt:     fixedTime,
					UpdatedAt:     fixedTime,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "admin",
					Resource:  "project:" + testResourceID,
					Principal: "user:" + testUserID,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreatePolicyResponse{
				Policy: &frontierv1beta1.Policy{
					Id:        testPolicyID,
					RoleId:    "admin",
					Resource:  "app/project:" + testResourceID,
					Principal: "app/user:" + testUserID,
					Metadata:  nil,
					CreatedAt: timestamppb.New(fixedTime),
					UpdatedAt: timestamppb.New(fixedTime),
				},
			}),
		},
		{
			name: "should successfully create policy with metadata",
			setup: func(ps *mocks.PolicyService) {
				metadataMap := map[string]interface{}{
					"description": "Test policy",
					"priority":    "high",
				}
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "viewer",
					ResourceID:    testResourceID,
					ResourceType:  "app/organization",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Build(metadataMap),
				}).Return(policy.Policy{
					ID:            testPolicyID,
					RoleID:        "viewer",
					ResourceID:    testResourceID,
					ResourceType:  "app/organization",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Build(metadataMap),
					CreatedAt:     fixedTime,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "viewer",
					Resource:  "organization:" + testResourceID,
					Principal: "user:" + testUserID,
					Metadata: func() *structpb.Struct {
						s, _ := structpb.NewStruct(map[string]interface{}{
							"description": "Test policy",
							"priority":    "high",
						})
						return s
					}(),
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreatePolicyResponse{
				Policy: &frontierv1beta1.Policy{
					Id:        testPolicyID,
					RoleId:    "viewer",
					Resource:  "app/organization:" + testResourceID,
					Principal: "app/user:" + testUserID,
					Metadata: func() *structpb.Struct {
						s, _ := structpb.NewStruct(map[string]interface{}{
							"description": "Test policy",
							"priority":    "high",
						})
						return s
					}(),
					CreatedAt: timestamppb.New(fixedTime),
				},
			}),
		},
		{
			name: "should successfully create policy for group principal",
			setup: func(ps *mocks.PolicyService) {
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "editor",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testGroupID,
					PrincipalType: "app/group",
					Metadata:      metadata.Metadata(nil),
				}).Return(policy.Policy{
					ID:            testPolicyID,
					RoleID:        "editor",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testGroupID,
					PrincipalType: "app/group",
					Metadata:      metadata.Metadata{},
					CreatedAt:     fixedTime,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "editor",
					Resource:  "project:" + testResourceID,
					Principal: "group:" + testGroupID,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreatePolicyResponse{
				Policy: &frontierv1beta1.Policy{
					Id:        testPolicyID,
					RoleId:    "editor",
					Resource:  "app/project:" + testResourceID,
					Principal: "app/group:" + testGroupID,
					Metadata:  nil,
					CreatedAt: timestamppb.New(fixedTime),
				},
			}),
		},
		{
			name: "should return internal error when transformPolicyToPB fails due to metadata error",
			setup: func(ps *mocks.PolicyService) {
				// Create policy with metadata that will fail structpb conversion
				invalidMetadata := metadata.Metadata{"invalid": make(chan int)} // channels can't be converted to protobuf
				ps.On("Create", mock.Anything, policy.Policy{
					RoleID:        "admin",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      metadata.Metadata(nil),
				}).Return(policy.Policy{
					ID:            testPolicyID,
					RoleID:        "admin",
					ResourceID:    testResourceID,
					ResourceType:  "app/project",
					PrincipalID:   testUserID,
					PrincipalType: "app/user",
					Metadata:      invalidMetadata,
					CreatedAt:     fixedTime,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreatePolicyRequest{
				Body: &frontierv1beta1.PolicyRequestBody{
					RoleId:    "admin",
					Resource:  "project:" + testResourceID,
					Principal: "user:" + testUserID,
				},
			}),
			want:    nil,
			wantErr: ErrInternalServerError,
			errCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockPolicyService := &mocks.PolicyService{}
			if tt.setup != nil {
				tt.setup(mockPolicyService)
			}

			handler := &ConnectHandler{
				policyService: mockPolicyService,
			}

			got, err := handler.CreatePolicy(context.Background(), tt.request)
			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.errCode, connect.CodeOf(err))
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}

			mockPolicyService.AssertExpectations(t)
		})
	}
}
