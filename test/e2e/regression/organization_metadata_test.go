package e2e_test

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *APIRegressionTestSuite) TestUpdateOrganizationMetadata() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	newMetadata := func(fields map[string]any) *structpb.Struct {
		metadata, err := structpb.NewStruct(fields)
		s.Require().NoError(err)
		return metadata
	}

	getMetadata := func(orgID string) map[string]any {
		resp, err := s.testBench.Client.GetOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{Id: orgID}))
		s.Require().NoError(err)
		return resp.Msg.GetOrganization().GetMetadata().AsMap()
	}

	createResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Title:    "org metadata update",
			Name:     "org-metadata-update",
			Metadata: newMetadata(map[string]any{"poc": "owner@raystack.org", "size": "10"}),
		},
	}))
	s.Require().NoError(err)
	orgID := createResp.Msg.GetOrganization().GetId()

	s.Run("1. update without metadata keeps the stored metadata", func() {
		resp, err := s.testBench.Client.UpdateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
			Id: orgID,
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org metadata update renamed",
				Name:  "org-metadata-update",
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal("org metadata update renamed", resp.Msg.GetOrganization().GetTitle())
		s.Assert().Equal(map[string]any{"poc": "owner@raystack.org", "size": "10"}, getMetadata(orgID))
	})

	s.Run("2. update with metadata replaces the stored metadata", func() {
		_, err := s.testBench.Client.UpdateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
			Id: orgID,
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title:    "org metadata update renamed",
				Name:     "org-metadata-update",
				Metadata: newMetadata(map[string]any{"size": "20"}),
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal(map[string]any{"size": "20"}, getMetadata(orgID))
	})

	s.Run("3. update with empty metadata clears the stored metadata", func() {
		_, err := s.testBench.Client.UpdateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
			Id: orgID,
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title:    "org metadata update renamed",
				Name:     "org-metadata-update",
				Metadata: &structpb.Struct{},
			},
		}))
		s.Require().NoError(err)
		s.Assert().Empty(getMetadata(orgID))
	})
}
