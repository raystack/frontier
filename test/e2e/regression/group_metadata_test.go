package e2e_test

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *APIRegressionTestSuite) TestUpdateGroupMetadata() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	newMetadata := func(fields map[string]any) *structpb.Struct {
		metadata, err := structpb.NewStruct(fields)
		s.Require().NoError(err)
		return metadata
	}

	getMetadata := func(groupID string) map[string]any {
		resp, err := s.testBench.Client.GetGroup(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetGroupRequest{Id: groupID}))
		s.Require().NoError(err)
		return resp.Msg.GetGroup().GetMetadata().AsMap()
	}

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Title: "group metadata update org",
			Name:  "group-metadata-update-org",
		},
	}))
	s.Require().NoError(err)

	createResp, err := s.testBench.Client.CreateGroup(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
		OrgId: createOrgResp.Msg.GetOrganization().GetId(),
		Body: &frontierv1beta1.GroupRequestBody{
			Title:    "group metadata update",
			Name:     "group-metadata-update",
			Metadata: newMetadata(map[string]any{"description": "core team", "size": "10"}),
		},
	}))
	s.Require().NoError(err)
	groupID := createResp.Msg.GetGroup().GetId()

	s.Run("1. update without metadata keeps the stored metadata", func() {
		resp, err := s.testBench.Client.UpdateGroup(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
			Id: groupID,
			Body: &frontierv1beta1.GroupRequestBody{
				Title: "group metadata update renamed",
				Name:  "group-metadata-update",
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal("group metadata update renamed", resp.Msg.GetGroup().GetTitle())
		s.Assert().Equal(map[string]any{"description": "core team", "size": "10"}, getMetadata(groupID))
	})

	s.Run("2. update with metadata replaces the stored metadata", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
			Id: groupID,
			Body: &frontierv1beta1.GroupRequestBody{
				Title:    "group metadata update renamed",
				Name:     "group-metadata-update",
				Metadata: newMetadata(map[string]any{"size": "20"}),
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal(map[string]any{"size": "20"}, getMetadata(groupID))
	})

	s.Run("3. update with empty metadata clears the stored metadata", func() {
		_, err := s.testBench.Client.UpdateGroup(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateGroupRequest{
			Id: groupID,
			Body: &frontierv1beta1.GroupRequestBody{
				Title:    "group metadata update renamed",
				Name:     "group-metadata-update",
				Metadata: &structpb.Struct{},
			},
		}))
		s.Require().NoError(err)
		s.Assert().Empty(getMetadata(groupID))
	})
}
