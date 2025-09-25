package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testServiceUserID = "su-9f256f86-31a3-11ec-8d3d-0242ac130003"
	testServiceUserMap = map[string]serviceuser.ServiceUser{
		"su-9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
			Title: "Test Service User",
			OrgID: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
			State: "enabled",
			Metadata: metadata.Metadata{
				"purpose": "testing",
				"team":    "backend",
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)

func TestHandler_ListServiceUsers(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(sus *mocks.ServiceUserService)
		request *connect.Request[frontierv1beta1.ListServiceUsersRequest]
		want    *connect.Response[frontierv1beta1.ListServiceUsersResponse]
		wantErr bool
	}{
		{
			name: "should list service users successfully",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, serviceuser.Filter{
					OrgID: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
					State: serviceuser.State("enabled"),
				}).Return([]serviceuser.ServiceUser{
					testServiceUserMap["su-9f256f86-31a3-11ec-8d3d-0242ac130003"],
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
				State: "enabled",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{
					{
						Id:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Test Service User",
						OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
						State: "enabled",
						Metadata: func() *structpb.Struct {
							md, _ := metadata.Metadata{
								"purpose": "testing",
								"team":    "backend",
							}.ToStructPB()
							return md
						}(),
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: false,
		},
		{
			name: "should list service users with only org filter",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, serviceuser.Filter{
					OrgID: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
					State: serviceuser.State(""),
				}).Return([]serviceuser.ServiceUser{
					testServiceUserMap["su-9f256f86-31a3-11ec-8d3d-0242ac130003"],
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{
					{
						Id:    "su-9f256f86-31a3-11ec-8d3d-0242ac130003",
						Title: "Test Service User",
						OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
						State: "enabled",
						Metadata: func() *structpb.Struct {
							md, _ := metadata.Metadata{
								"purpose": "testing",
								"team":    "backend",
							}.ToStructPB()
							return md
						}(),
						CreatedAt: timestamppb.New(time.Time{}),
						UpdatedAt: timestamppb.New(time.Time{}),
					},
				},
			}),
			wantErr: false,
		},
		{
			name: "should return empty list when no service users found",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, serviceuser.Filter{
					OrgID: "org-nonexistent",
					State: serviceuser.State(""),
				}).Return([]serviceuser.ServiceUser{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-nonexistent",
			}),
			want: connect.NewResponse(&frontierv1beta1.ListServiceUsersResponse{
				Serviceusers: []*frontierv1beta1.ServiceUser{},
			}),
			wantErr: false,
		},
		{
			name: "should return internal error when service fails",
			setup: func(sus *mocks.ServiceUserService) {
				sus.On("List", mock.Anything, mock.Anything).Return(nil, errors.New("service error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListServiceUsersRequest{
				OrgId: "org-9f256f86-31a3-11ec-8d3d-0242ac130003",
			}),
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			serviceUserService := &mocks.ServiceUserService{}
			if tt.setup != nil {
				tt.setup(serviceUserService)
			}

			h := &ConnectHandler{
				serviceUserService: serviceUserService,
			}

			got, err := h.ListServiceUsers(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, got)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, got)
				if tt.want != nil {
					assert.Equal(t, len(tt.want.Msg.Serviceusers), len(got.Msg.Serviceusers))
					for i, expectedSU := range tt.want.Msg.Serviceusers {
						actualSU := got.Msg.Serviceusers[i]
						assert.Equal(t, expectedSU.Id, actualSU.Id)
						assert.Equal(t, expectedSU.Title, actualSU.Title)
						assert.Equal(t, expectedSU.OrgId, actualSU.OrgId)
						assert.Equal(t, expectedSU.State, actualSU.State)
					}
				}
			}

			serviceUserService.AssertExpectations(t)
		})
	}
}