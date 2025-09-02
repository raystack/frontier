package v1beta1connect

import (
	"context"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testProjectIDList = []string{"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71", "c7772c63-fca4-4c7c-bf93-c8f85115de4b"}
	testProjectMap    = map[string]project.Project{
		"ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71": {
			ID:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
			Name: "prj-1",
			Metadata: metadata.Metadata{
				"email": "org1@org1.com",
			},
			Organization: organization.Organization{
				ID: testOrgID,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
		"c7772c63-fca4-4c7c-bf93-c8f85115de4b": {
			ID:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
			Name: "prj-2",
			Metadata: metadata.Metadata{
				"email": "org1@org2.com",
			},
			Organization: organization.Organization{
				ID: testOrgID,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)

func TestListProjects(t *testing.T) {
	table := []struct {
		title string
		setup func(ps *mocks.ProjectService)
		req   *connect.Request[frontierv1beta1.ListProjectsRequest]
		want  *connect.Response[frontierv1beta1.ListProjectsResponse]
		err   error
	}{
		{
			title: "should return internal error if project service return some error",
			req:   connect.NewRequest(&frontierv1beta1.ListProjectsRequest{}),
			setup: func(ps *mocks.ProjectService) {
				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{}).Return([]project.Project{}, errors.New("test error"))
			},
			want: nil,
			err:  connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			title: "should return success if project return nil error",
			req:   connect.NewRequest(&frontierv1beta1.ListProjectsRequest{}),
			setup: func(ps *mocks.ProjectService) {
				var prjs []project.Project

				for _, projectID := range testProjectIDList {
					prjs = append(prjs, testProjectMap[projectID])
				}

				ps.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), project.Filter{}).Return(prjs, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListProjectsResponse{Projects: []*frontierv1beta1.Project{
				{
					Id:   "ab657ae7-8c9e-45eb-9862-dd9ceb6d5c71",
					Name: "prj-1",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org1.com"),
						},
					},
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
				{
					Id:   "c7772c63-fca4-4c7c-bf93-c8f85115de4b",
					Name: "prj-2",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": structpb.NewStringValue("org1@org2.com"),
						},
					},
					OrgId:     "9f256f86-31a3-11ec-8d3d-0242ac130003",
					CreatedAt: timestamppb.New(time.Time{}),
					UpdatedAt: timestamppb.New(time.Time{}),
				},
			}}),
			err: nil,
		},
	}

	for _, tt := range table {
		t.Run(tt.title, func(t *testing.T) {
			mockProjectSrv := new(mocks.ProjectService)
			if tt.setup != nil {
				tt.setup(mockProjectSrv)
			}
			mockDep := ConnectHandler{projectService: mockProjectSrv}
			resp, err := mockDep.ListProjects(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.err, err)
		})
	}
}
