package e2e_test

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
	"google.golang.org/protobuf/types/known/structpb"
)

func (s *APIRegressionTestSuite) TestUpdateProjectByNameTargetsRequestID() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	_, userCtx := signUpUser(s.T(), s.testBench.Client, "update-project-by-name@raystack.org")

	otherOrg, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Title: "other org", Name: "update-project-by-name-other-org"},
	}))
	s.Require().NoError(err)
	otherMetadata, err := structpb.NewStruct(map[string]any{"owner": "other org"})
	s.Require().NoError(err)
	otherProject, err := s.testBench.Client.CreateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Title:    "other project",
			Name:     "update-project-by-name-other",
			OrgId:    otherOrg.Msg.GetOrganization().GetId(),
			Metadata: otherMetadata,
		},
	}))
	s.Require().NoError(err)
	otherProjectID := otherProject.Msg.GetProject().GetId()

	ownOrg, err := s.testBench.Client.CreateOrganization(userCtx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Title: "own org", Name: "update-project-by-name-own-org"},
	}))
	s.Require().NoError(err)
	ownProject, err := s.testBench.Client.CreateProject(userCtx, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Title: "own project",
			Name:  "update-project-by-name-own",
			OrgId: ownOrg.Msg.GetOrganization().GetId(),
		},
	}))
	s.Require().NoError(err)
	ownProjectID := ownProject.Msg.GetProject().GetId()

	s.Run("updating the other project directly is refused", func() {
		_, err := s.testBench.Client.UpdateProject(userCtx, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
			Id:   "update-project-by-name-other",
			Body: &frontierv1beta1.UpdateProjectRequestBody{Name: "update-project-by-name-other", Title: "changed"},
		}))
		s.Assert().Equal(connect.CodePermissionDenied, connect.CodeOf(err))
	})

	s.Run("a body name for another project updates only the project in the request id", func() {
		resp, err := s.testBench.Client.UpdateProject(userCtx, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
			Id: "update-project-by-name-own",
			Body: &frontierv1beta1.UpdateProjectRequestBody{
				Name:     "update-project-by-name-other",
				Title:    "changed",
				Metadata: &structpb.Struct{},
			},
		}))
		s.Require().NoError(err)
		s.Assert().Equal(ownProjectID, resp.Msg.GetProject().GetId())
		s.Assert().Equal("update-project-by-name-own", resp.Msg.GetProject().GetName())
		s.Assert().Equal("changed", resp.Msg.GetProject().GetTitle())

		other, err := s.testBench.Client.GetProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetProjectRequest{Id: otherProjectID}))
		s.Require().NoError(err)
		s.Assert().Equal("other project", other.Msg.GetProject().GetTitle())
		s.Assert().Equal(map[string]any{"owner": "other org"}, other.Msg.GetProject().GetMetadata().AsMap())
	})
}

func (s *APIRegressionTestSuite) TestUpdateOrganizationByNameTargetsRequestID() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)
	_, userCtx := signUpUser(s.T(), s.testBench.Client, "update-org-by-name@raystack.org")

	otherOrg, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Title: "other org", Name: "update-org-by-name-other"},
	}))
	s.Require().NoError(err)
	otherOrgID := otherOrg.Msg.GetOrganization().GetId()

	ownOrg, err := s.testBench.Client.CreateOrganization(userCtx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{Title: "own org", Name: "update-org-by-name-own"},
	}))
	s.Require().NoError(err)
	ownOrgID := ownOrg.Msg.GetOrganization().GetId()

	s.Run("updating the other org directly is refused", func() {
		_, err := s.testBench.Client.UpdateOrganization(userCtx, connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
			Id:   "update-org-by-name-other",
			Body: &frontierv1beta1.OrganizationRequestBody{Name: "update-org-by-name-other", Title: "changed"},
		}))
		s.Assert().Equal(connect.CodePermissionDenied, connect.CodeOf(err))
	})

	s.Run("a body name for another org updates only the org in the request id", func() {
		resp, err := s.testBench.Client.UpdateOrganization(userCtx, connect.NewRequest(&frontierv1beta1.UpdateOrganizationRequest{
			Id:   "update-org-by-name-own",
			Body: &frontierv1beta1.OrganizationRequestBody{Name: "update-org-by-name-other", Title: "changed"},
		}))
		s.Require().NoError(err)
		s.Assert().Equal(ownOrgID, resp.Msg.GetOrganization().GetId())
		s.Assert().Equal("update-org-by-name-own", resp.Msg.GetOrganization().GetName())
		s.Assert().Equal("changed", resp.Msg.GetOrganization().GetTitle())

		other, err := s.testBench.Client.GetOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{Id: otherOrgID}))
		s.Require().NoError(err)
		s.Assert().Equal("other org", other.Msg.GetOrganization().GetTitle())
		s.Assert().Equal(otherOrg.Msg.GetOrganization().GetAvatar(), other.Msg.GetOrganization().GetAvatar())
		s.Assert().Equal(otherOrg.Msg.GetOrganization().GetMetadata().AsMap(), other.Msg.GetOrganization().GetMetadata().AsMap())
	})
}
