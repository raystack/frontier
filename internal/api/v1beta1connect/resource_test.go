package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testResourceID = utils.NewString()
	testResource   = resource.Resource{
		ID:            testResourceID,
		URN:           "res-urn",
		Name:          "a resource name",
		ProjectID:     testProjectID,
		NamespaceID:   testNSID,
		PrincipalID:   testUserID,
		PrincipalType: schema.UserPrincipal,
	}
	testResourcePB = &frontierv1beta1.Resource{
		Id:        testResource.ID,
		Name:      testResource.Name,
		Urn:       testResource.URN,
		ProjectId: testProjectID,
		Namespace: testNSID,
		Principal: schema.JoinNamespaceAndResourceID(testResource.PrincipalType, testResource.PrincipalID),
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	}
)

func TestHandler_ListResources(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.ListResourcesRequest]
		want    *connect.Response[frontierv1beta1.ListResourcesResponse]
		wantErr error
	}{
		{
			name: "should return internal error if resource service return some error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{}).Return([]resource.Resource{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return resources if resource service return nil error",
			setup: func(rs *mocks.ResourceService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), resource.Filter{}).Return([]resource.Resource{
					testResource,
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListResourcesRequest{}),
			want: connect.NewResponse(&frontierv1beta1.ListResourcesResponse{
				Resources: []*frontierv1beta1.Resource{
					testResourcePB,
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockResourceSrv)
			}
			mockDep := ConnectHandler{resourceService: mockResourceSrv}
			resp, err := mockDep.ListResources(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
