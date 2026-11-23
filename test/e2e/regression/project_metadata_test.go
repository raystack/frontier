package e2e_test

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *APIRegressionTestSuite) TestUpdateProjectMetadata() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	newMetadata := func(fields map[string]any) *structpb.Struct {
		metadata, err := structpb.NewStruct(fields)
		s.Require().NoError(err)
		return metadata
	}

	getMetadata := func(projectID string) map[string]any {
		resp, err := s.testBench.Client.GetProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetProjectRequest{Id: projectID}))
		s.Require().NoError(err)
		return resp.Msg.GetProject().GetMetadata().AsMap()
	}

	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Title: "project metadata update org",
			Name:  "project-metadata-update-org",
		},
	}))
	s.Require().NoError(err)

	createResp, err := s.testBench.Client.CreateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Title:    "project metadata update",
			Name:     "project-metadata-update",
			OrgId:    createOrgResp.Msg.GetOrganization().GetId(),
			Metadata: newMetadata(map[string]any{"team": "core", "size": "10"}),
		},
	}))
	s.Require().NoError(err)
	projectID := createResp.Msg.GetProject().GetId()

	s.Run("1. update without metadata keeps the stored metadata", func() {
		resp, err := s.testBench.Client.UpdateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
			Id: projectID,
			Body: &frontierv1beta1.UpdateProjectRequestBody{
				Title: "project metadata update renamed",
				Name:  "project-metadata-update",
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal("project metadata update renamed", resp.Msg.GetProject().GetTitle())
		s.Assert().Equal(map[string]any{"team": "core", "size": "10"}, getMetadata(projectID))
	})

	s.Run("2. update with metadata replaces the stored metadata", func() {
		_, err := s.testBench.Client.UpdateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
			Id: projectID,
			Body: &frontierv1beta1.UpdateProjectRequestBody{
				Title:    "project metadata update renamed",
				Name:     "project-metadata-update",
				Metadata: newMetadata(map[string]any{"size": "20"}),
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal(map[string]any{"size": "20"}, getMetadata(projectID))
	})

	s.Run("3. update with empty metadata clears the stored metadata", func() {
		_, err := s.testBench.Client.UpdateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
			Id: projectID,
			Body: &frontierv1beta1.UpdateProjectRequestBody{
				Title:    "project metadata update renamed",
				Name:     "project-metadata-update",
				Metadata: &structpb.Struct{},
			},
		}))
		s.Require().NoError(err)
		s.Assert().Empty(getMetadata(projectID))
	})

	s.Run("update project by name", func() {
		projectName := "project-metadata-update-by-name"
		createByNameResp, err := s.testBench.Client.CreateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
			Body: &frontierv1beta1.ProjectRequestBody{
				Title:    "project metadata update by name",
				Name:     projectName,
				OrgId:    createOrgResp.Msg.GetOrganization().GetId(),
				Metadata: newMetadata(map[string]any{"team": "core", "size": "10"}),
			},
		}))
		s.Require().NoError(err)
		projectByNameID := createByNameResp.Msg.GetProject().GetId()

		s.Run("1. update without metadata keeps the stored metadata", func() {
			resp, err := s.testBench.Client.UpdateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: projectName,
				Body: &frontierv1beta1.UpdateProjectRequestBody{
					Title: "project metadata update by name renamed",
					Name:  projectName,
				},
			}))
			s.Require().NoError(err)
			s.Assert().Equal("project metadata update by name renamed", resp.Msg.GetProject().GetTitle())
			s.Assert().Equal(map[string]any{"team": "core", "size": "10"}, getMetadata(projectByNameID))
		})

		s.Run("2. update with metadata replaces the stored metadata", func() {
			_, err := s.testBench.Client.UpdateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: projectName,
				Body: &frontierv1beta1.UpdateProjectRequestBody{
					Title:    "project metadata update by name renamed",
					Name:     projectName,
					Metadata: newMetadata(map[string]any{"size": "20"}),
				},
			}))
			s.Require().NoError(err)
			s.Assert().Equal(map[string]any{"size": "20"}, getMetadata(projectByNameID))
		})

		s.Run("3. update with empty metadata clears the stored metadata", func() {
			_, err := s.testBench.Client.UpdateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
				Id: projectName,
				Body: &frontierv1beta1.UpdateProjectRequestBody{
					Title:    "project metadata update by name renamed",
					Name:     projectName,
					Metadata: &structpb.Struct{},
				},
			}))
			s.Require().NoError(err)
			s.Assert().Empty(getMetadata(projectByNameID))
		})
	})
}
