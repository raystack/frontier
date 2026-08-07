package e2e_test

import (
	"context"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/test/e2e/testbench"
)

func (s *APIRegressionTestSuite) TestDisabledOrganizationEnforcement() {
	ctxAdmin := testbench.ContextWithAuth(context.Background(), s.adminCookie)

	// fixtures: an org with a project and a group, one org member, one outsider
	createOrgResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
		Body: &frontierv1beta1.OrganizationRequestBody{
			Title: "org disabled gate",
			Name:  "org-disabled-gate",
		},
	}))
	s.Require().NoError(err)
	orgID := createOrgResp.Msg.GetOrganization().GetId()

	createProjResp, err := s.testBench.Client.CreateProject(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
		Body: &frontierv1beta1.ProjectRequestBody{
			Name:  "org-disabled-gate-proj",
			OrgId: orgID,
		},
	}))
	s.Require().NoError(err)
	projectID := createProjResp.Msg.GetProject().GetId()

	createGroupResp, err := s.testBench.Client.CreateGroup(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
		OrgId: orgID,
		Body: &frontierv1beta1.GroupRequestBody{
			Name: "org-disabled-gate-group",
		},
	}))
	s.Require().NoError(err)
	groupID := createGroupResp.Msg.GetGroup().GetId()

	createMemberResp, err := s.testBench.Client.CreateUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateUserRequest{
		Body: &frontierv1beta1.UserRequestBody{
			Title: "disabled gate member",
			Email: "disabled-gate-member@raystack.org",
			Name:  "disabled_gate_member",
		},
	}))
	s.Require().NoError(err)
	addMembersResp, err := s.testBench.AdminClient.AddOrganizationMembers(ctxAdmin, connect.NewRequest(&frontierv1beta1.AddOrganizationMembersRequest{
		OrgId: orgID,
		Members: []*frontierv1beta1.OrgMemberEntry{{
			UserId: createMemberResp.Msg.GetUser().GetId(),
			RoleId: s.orgOwnerRole,
		}},
	}))
	requireAddOrgMembersSuccess(s.T(), addMembersResp, err)

	memberCookie, err := testbench.AuthenticateUser(context.Background(), s.testBench.Client, "disabled-gate-member@raystack.org")
	s.Require().NoError(err)
	ctxMember := testbench.ContextWithAuth(context.Background(), memberCookie)

	createOutsiderResp, err := s.testBench.Client.CreateUser(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateUserRequest{
		Body: &frontierv1beta1.UserRequestBody{
			Title: "disabled gate outsider",
			Email: "disabled-gate-outsider@raystack.org",
			Name:  "disabled_gate_outsider",
		},
	}))
	s.Require().NoError(err)
	outsiderEmail := createOutsiderResp.Msg.GetUser().GetEmail()

	outsiderCookie, err := testbench.AuthenticateUser(context.Background(), s.testBench.Client, outsiderEmail)
	s.Require().NoError(err)
	ctxOutsider := testbench.ContextWithAuth(context.Background(), outsiderCookie)

	s.Run("1. member can use org resources while the org is enabled", func() {
		_, err := s.testBench.Client.GetProject(ctxMember, connect.NewRequest(&frontierv1beta1.GetProjectRequest{Id: projectID}))
		s.Assert().NoError(err)

		_, err = s.testBench.Client.GetGroup(ctxMember, connect.NewRequest(&frontierv1beta1.GetGroupRequest{Id: groupID}))
		s.Assert().NoError(err)

		_, err = s.testBench.Client.ListOrganizationProjects(ctxMember, connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{Id: orgID}))
		s.Assert().NoError(err)
	})

	s.Run("2. superuser disables the org", func() {
		_, err := s.testBench.Client.DisableOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.DisableOrganizationRequest{Id: orgID}))
		s.Assert().NoError(err)
	})

	s.Run("3. member gets failed precondition on every org resource RPC", func() {
		_, err := s.testBench.Client.GetOrganization(ctxMember, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{Id: orgID}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.GetProject(ctxMember, connect.NewRequest(&frontierv1beta1.GetProjectRequest{Id: projectID}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.UpdateProject(ctxMember, connect.NewRequest(&frontierv1beta1.UpdateProjectRequest{
			Id: projectID,
			Body: &frontierv1beta1.UpdateProjectRequestBody{
				Name:  "org-disabled-gate-proj",
				Title: "renamed while disabled",
			},
		}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.ListProjectResources(ctxMember, connect.NewRequest(&frontierv1beta1.ListProjectResourcesRequest{ProjectId: projectID}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.GetGroup(ctxMember, connect.NewRequest(&frontierv1beta1.GetGroupRequest{Id: groupID}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.CreateGroup(ctxMember, connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
			OrgId: orgID,
			Body:  &frontierv1beta1.GroupRequestBody{Name: "blocked-group"},
		}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.ListBillingAccounts(ctxMember, connect.NewRequest(&frontierv1beta1.ListBillingAccountsRequest{OrgId: orgID}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))

		_, err = s.testBench.Client.CreateOrganizationInvitation(ctxMember, connect.NewRequest(&frontierv1beta1.CreateOrganizationInvitationRequest{
			OrgId:   orgID,
			UserIds: []string{outsiderEmail},
		}))
		s.Assert().Equal(connect.CodeFailedPrecondition, connect.CodeOf(err))
	})

	s.Run("4. non-member still gets permission denied, not the org state", func() {
		_, err := s.testBench.Client.GetProject(ctxOutsider, connect.NewRequest(&frontierv1beta1.GetProjectRequest{Id: projectID}))
		s.Assert().Equal(connect.CodePermissionDenied, connect.CodeOf(err))

		// probing the disabled org by name must look the same as any other
		// denied request
		_, err = s.testBench.Client.GetOrganization(ctxOutsider, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{Id: "org-disabled-gate"}))
		s.Assert().Equal(connect.CodePermissionDenied, connect.CodeOf(err))
	})

	s.Run("5. superuser can still see, export and re-enable the org", func() {
		listResp, err := s.testBench.AdminClient.ListAllOrganizations(ctxAdmin, connect.NewRequest(&frontierv1beta1.ListAllOrganizationsRequest{State: "disabled"}))
		s.Require().NoError(err)
		found := false
		for _, org := range listResp.Msg.GetOrganizations() {
			if org.GetId() == orgID {
				found = true
				break
			}
		}
		s.Assert().True(found, "disabled org missing from admin listing")

		stream, err := s.testBench.AdminClient.ExportOrganizations(ctxAdmin, connect.NewRequest(&frontierv1beta1.ExportOrganizationsRequest{}))
		s.Require().NoError(err)
		exported := 0
		for stream.Receive() {
			exported += len(stream.Msg().GetData())
		}
		s.Assert().NoError(stream.Err())
		s.Assert().NoError(stream.Close())
		s.Assert().Greater(exported, 0)

		_, err = s.testBench.Client.GetOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.GetOrganizationRequest{Id: orgID}))
		s.Assert().NoError(err)

		_, err = s.testBench.Client.EnableOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.EnableOrganizationRequest{Id: orgID}))
		s.Assert().NoError(err)
	})

	s.Run("6. re-enable restores member access without any re-grants", func() {
		_, err := s.testBench.Client.GetProject(ctxMember, connect.NewRequest(&frontierv1beta1.GetProjectRequest{Id: projectID}))
		s.Assert().NoError(err)

		_, err = s.testBench.Client.GetGroup(ctxMember, connect.NewRequest(&frontierv1beta1.GetGroupRequest{Id: groupID}))
		s.Assert().NoError(err)
	})

	s.Run("7. a disabled org can still be deleted by the superuser", func() {
		createResp, err := s.testBench.Client.CreateOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: &frontierv1beta1.OrganizationRequestBody{
				Title: "org disabled delete",
				Name:  "org-disabled-delete",
			},
		}))
		s.Require().NoError(err)
		deleteOrgID := createResp.Msg.GetOrganization().GetId()

		_, err = s.testBench.Client.DisableOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.DisableOrganizationRequest{Id: deleteOrgID}))
		s.Require().NoError(err)

		_, err = s.testBench.Client.DeleteOrganization(ctxAdmin, connect.NewRequest(&frontierv1beta1.DeleteOrganizationRequest{Id: deleteOrgID}))
		s.Assert().NoError(err)

		// after the delete cascade the org must be gone from the admin listing
		listResp, err := s.testBench.AdminClient.ListAllOrganizations(ctxAdmin, connect.NewRequest(&frontierv1beta1.ListAllOrganizationsRequest{State: "disabled"}))
		s.Require().NoError(err)
		for _, org := range listResp.Msg.GetOrganizations() {
			s.Assert().NotEqual(deleteOrgID, org.GetId(), "deleted org still present in admin listing")
		}
	})
}
