package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
)

var (
	//go:embed seed/permissions.json
	mockCustomPermissions []byte
	//go:embed seed/roles.json
	mockCustomRoles []byte
	//go:embed seed/user.json
	mockHumanUser []byte
	//go:embed seed/organization.json
	mockOrganization []byte
	//go:embed seed/project.json
	mockProject []byte
	//go:embed seed/resource.json
	mockResource []byte

	orgAdminEmail  = "sampleAdmin@raystack.org"
	identityHeader = "X-Shield-Email"
)

func SeedCommand(cliConfig *Config) *cli.Command {
	cmd := &cli.Command{
		Use:   "seed",
		Short: "Seed the database with initial data",
		Args:  cli.NoArgs,
		Long:  "This command can be used to create an organization structure with predefined groups, projects, and resources. It bootstarps these data in the Shield db, making it easier to get started.",
		Example: heredoc.Doc(`
			$ shield seed
		`),
		Annotations: map[string]string{
			"group": "core",
		},

		RunE: func(cmd *cli.Command, args []string) error {
			adminClient, cancel, err := createAdminClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			// create custom roles and permissions for the sample data
			if err := createCustomRolesAndPermissions(adminClient); err != nil {
				return fmt.Errorf("failed to create custom permissions: %w", err)
			}

			client, cancel, err := createClient(cmd.Context(), cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			if err := bootstrapData(client); err != nil {
				return fmt.Errorf("failed to bootstrap data: %w", err)
			}

			fmt.Println("initialized sample data in shield successfully")
			return nil
		},
	}

	bindFlagsFromClientConfig(cmd)

	return cmd
}

// create compute/order namespace and custom roles and permissions
func createCustomRolesAndPermissions(client shieldv1beta1.AdminServiceClient) error {
	ctx := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		identityHeader: orgAdminEmail,
	}))

	var permissionBodies []*shieldv1beta1.PermissionRequestBody
	if err := json.Unmarshal(mockCustomPermissions, &permissionBodies); err != nil {
		fmt.Println("Error unmarshaling JSON :", err)
		return err
	}

	// create custom permission
	if _, err := client.CreatePermission(ctx, &shieldv1beta1.CreatePermissionRequest{
		Bodies: permissionBodies,
	}); err != nil {
		return fmt.Errorf("failed to create custom permission: %w", err)
	}

	fmt.Printf("created custom permissions for compute order service: %s, %s, %s, %s, %s\n", "create", "get", "update", "delete", "list")

	var roles []*shieldv1beta1.RoleRequestBody
	if err := json.Unmarshal(mockCustomRoles, &roles); err != nil {
		return fmt.Errorf("failed to unmarshal custom roles: %w", err)
	}

	for _, role := range roles {
		if _, err := client.CreateRole(ctx, &shieldv1beta1.CreateRoleRequest{
			Body: role,
		}); err != nil {
			return fmt.Errorf("failed to create custom role: %w", err)
		}
	}

	fmt.Printf("created custom roles %s and %s\n", "compute_order_manager", "compute_order_viewer")

	return nil
}

func bootstrapData(client shieldv1beta1.ShieldServiceClient) error {
	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		identityHeader: orgAdminEmail,
	}))

	// create (human) user for org admin
	var userBody *shieldv1beta1.UserRequestBody
	if err := json.Unmarshal(mockHumanUser, &userBody); err != nil {
		return fmt.Errorf("failed to unmarshal user body: %w", err)
	}
	userResp, err := client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
		Body: userBody,
	})

	if err != nil {
		return fmt.Errorf("failed to create sample user: %w", err)
	}

	fmt.Printf("created user with email %s in shield\n", userResp.User.Email)

	// create orgBody
	var orgBody *shieldv1beta1.OrganizationRequestBody
	if err := json.Unmarshal(mockOrganization, &orgBody); err != nil {
		fmt.Println("Error unmarshaling JSON :", err)
		return err
	}

	orgResp, err := client.CreateOrganization(ctxOrgAdminAuth, &shieldv1beta1.CreateOrganizationRequest{
		Body: orgBody,
	})
	if err != nil {
		return fmt.Errorf("failed to create sample organization: %w", err)
	}
	fmt.Printf("created organization name %s with user %s as the org admin \n", orgResp.Organization.Name, orgAdminEmail)

	// create projBody inside org
	var projBody *shieldv1beta1.ProjectRequestBody
	if err := json.Unmarshal(mockProject, &projBody); err != nil {
		return fmt.Errorf("failed to unmarshal project body: %w", err)
	}
	projBody.OrgId = orgResp.Organization.Id

	projResp, err := client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
		Body: projBody,
	})
	if err != nil {
		return fmt.Errorf("failed to create sample project: %w", err)
	}

	fmt.Printf("created project in org %s with name %s \n", orgResp.Organization.Name, projResp.Project.Name)

	// create resource inside project
	var resourceBody *shieldv1beta1.ResourceRequestBody
	if err := json.Unmarshal(mockResource, &resourceBody); err != nil {
		return fmt.Errorf("failed to unmarshal resource body: %w", err)
	}
	resourceBody.Principal = userResp.User.Id

	resrcResp, err := client.CreateProjectResource(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectResourceRequest{
		ProjectId: projResp.Project.Id,
		Body:      resourceBody,
	})
	if err != nil {
		return fmt.Errorf("failed to create sample resource: %w", err)
	}

	fmt.Printf("created resource in project %s with name %s \n", projResp.Project.Name, resrcResp.Resource.Name)
	return nil
}
