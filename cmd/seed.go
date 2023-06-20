package cmd

import (
	"context"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/salt/printer"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/structpb"
)

var orgAdminEmail = "johndoe@raystack.org"
var orgUserEmail = "johndee@raystack.org"

const computeOrderNamespace = "compute/order"

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
			spinner := printer.Spin("")
			defer spinner.Stop()

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

			spinner.Stop()
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
		"X-Shield-Email": orgAdminEmail,
	}))

	// create custom permission
	if _, err := client.CreatePermission(ctx, &shieldv1beta1.CreatePermissionRequest{
		Bodies: []*shieldv1beta1.PermissionRequestBody{
			{
				Name:      "create",
				Title:     "Create Order",
				Namespace: computeOrderNamespace,
				Metadata:  &structpb.Struct{},
			},
			{
				Name:      "get",
				Title:     "Read Order",
				Namespace: computeOrderNamespace,
				Metadata:  &structpb.Struct{},
			},
			{
				Name:      "update",
				Title:     "Update Order",
				Namespace: computeOrderNamespace,
				Metadata:  &structpb.Struct{},
			},
			{
				Name:      "delete",
				Title:     "Delete Order",
				Namespace: computeOrderNamespace,
				Metadata:  &structpb.Struct{},
			},
			{
				Name:      "list",
				Title:     "List Orders",
				Namespace: computeOrderNamespace,
				Metadata:  &structpb.Struct{},
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to create custom permission: %w", err)
	}

	fmt.Printf("created custom permissions for compute order service: %s, %s, %s, %s, %s\n", "create", "get", "update", "delete", "list")

	// create custom roles for compute order manager and viewer
	if _, err := client.CreateRole(ctx, &shieldv1beta1.CreateRoleRequest{
		Body: &shieldv1beta1.RoleRequestBody{
			Name:     "compute_order_manager",
			Title:    "Compute Order Manager",
			Metadata: &structpb.Struct{},
			Permissions: []string{
				computeOrderNamespace + ":create",
				computeOrderNamespace + ":get",
				computeOrderNamespace + ":update",
				computeOrderNamespace + ":delete",
				computeOrderNamespace + ":list",
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to create custom role: %w", err)
	}

	if _, err := client.CreateRole(ctx, &shieldv1beta1.CreateRoleRequest{
		Body: &shieldv1beta1.RoleRequestBody{
			Name:     "compute_order_viewer",
			Title:    "Compute Order Viewer",
			Metadata: &structpb.Struct{},
			Permissions: []string{
				computeOrderNamespace + ":get",
				computeOrderNamespace + ":list",
			},
		},
	}); err != nil {
		return fmt.Errorf("failed to create custom organization role: %w", err)
	}

	fmt.Printf("created custom roles %s and %s\n", "compute_order_manager", "compute_order_viewer")

	return nil
}

func bootstrapData(client shieldv1beta1.ShieldServiceClient) error {
	// create (human) user for org admin
	userData := &shieldv1beta1.UserRequestBody{
		Email: orgUserEmail,
	}

	ctxOrgAdminAuth := metadata.NewOutgoingContext(context.Background(), metadata.New(map[string]string{
		"X-Shield-Email": orgAdminEmail,
	}))

	userResp, err := client.CreateUser(ctxOrgAdminAuth, &shieldv1beta1.CreateUserRequest{
		Body: userData,
	})

	if err != nil {
		return fmt.Errorf("failed to create sample user: %w", err)
	}

	fmt.Printf("created user with email %s in shield\n", userResp.User.Email)

	// create org
	orgData := &shieldv1beta1.OrganizationRequestBody{
		Name:  "raystack-09111",
		Title: "Raystack",
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"description": structpb.NewStringValue("Description"),
			},
		},
	}

	orgResp, err := client.CreateOrganization(ctxOrgAdminAuth, &shieldv1beta1.CreateOrganizationRequest{
		Body: orgData,
	})
	if err != nil {
		return fmt.Errorf("failed to create sample organization: %w", err)
	}
	fmt.Printf("created organization name %s with user %s as the org admin \n", orgResp.Organization.Name, orgAdminEmail)

	// create project inside org
	projData := &shieldv1beta1.ProjectRequestBody{
		Name:  "raystack-09111",
		Title: "Raystack",
		OrgId: orgResp.Organization.Id,
		Metadata: &structpb.Struct{
			Fields: map[string]*structpb.Value{
				"description": structpb.NewStringValue("Description"),
			},
		},
	}

	projResp, err := client.CreateProject(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectRequest{
		Body: projData,
	})
	if err != nil {
		return fmt.Errorf("failed to create sample project: %w", err)
	}

	fmt.Printf("created project in org %s with name %s \n", orgResp.Organization.Name, projResp.Project.Name)

	// create resource inside project
	resrcResp, err := client.CreateProjectResource(ctxOrgAdminAuth, &shieldv1beta1.CreateProjectResourceRequest{
		ProjectId: projResp.Project.Id,
		Body: &shieldv1beta1.ResourceRequestBody{
			Name:      "resource-1",
			Namespace: computeOrderNamespace,
			Principal: userResp.User.Id,
			Metadata:  &structpb.Struct{},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to create sample resource: %w", err)
	}
	fmt.Printf("created resource in project %s with name %s \n", projResp.Project.Name, resrcResp.Resource.Name)

	return nil
}
