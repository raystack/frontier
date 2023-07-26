package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/shield/config"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
	cli "github.com/spf13/cobra"
)

var (
	//go:embed seed/permissions.json
	mockCustomPermissions []byte
	//go:embed seed/roles.json
	mockCustomRoles []byte
	//go:embed seed/users.json
	mockHumanUser []byte
	//go:embed seed/organizations.json
	mockOrganizations []byte
	//go:embed seed/projects.json
	mockProjects []byte
	//go:embed seed/resource.json
	mockResource []byte

	sampleSeedEmail    = "sample.admin@raystack.org"
	resourceNamespaces []string
	roleIDs            []string
)

func SeedCommand(cliConfig *Config) *cli.Command {
	var header, configFile string
	cmd := &cli.Command{
		Use:   "seed",
		Short: "Seed the database with initial data",
		Args:  cli.NoArgs,
		Long:  "This command can be used to create an organization structure with predefined groups, projects, and resources. It bootstarps these data in the Shield db, making it easier to get started.",
		Example: heredoc.Doc(`
			$ shield seed
			$ shield seed --header=X-Shield-Email
		`),
		Annotations: map[string]string{
			"group": "core",
		},

		RunE: func(cmd *cli.Command, args []string) error {
			if header == "" {
				appConfig, err := config.Load(configFile)
				if err != nil {
					panic(err)
				}
				if appConfig.App.IdentityProxyHeader == "" {
					return errors.New("identity proxy header missing in server config, pass key in the header flag \nexample: shield seed -H X-Shield-Email")
				}
				header = appConfig.App.IdentityProxyHeader
			}
			header := fmt.Sprintf("%s:%s", header, sampleSeedEmail)
			ctx := setCtxHeader(cmd.Context(), header)

			adminClient, cancel, err := createAdminClient(ctx, cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			if err := createCustomRolesAndPermissions(ctx, adminClient); err != nil {
				return fmt.Errorf("failed to create custom permissions: %w", err)
			}

			client, cancel, err := createClient(ctx, cliConfig.Host)
			if err != nil {
				return err
			}
			defer cancel()

			if err := bootstrapData(ctx, client); err != nil {
				return fmt.Errorf("failed to bootstrap data: %w", err)
			}
			fmt.Println("initialized sample data in shield successfully")
			return nil
		},
	}

	bindFlagsFromClientConfig(cmd)
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
	return cmd
}

// create sample platform wide custom permissions and roles
func createCustomRolesAndPermissions(ctx context.Context, client shieldv1beta1.AdminServiceClient) error {
	var permissionBodies []*shieldv1beta1.PermissionRequestBody
	if err := json.Unmarshal(mockCustomPermissions, &permissionBodies); err != nil {
		return fmt.Errorf("failed to unmarshal custom permissions: %w", err)
	}

	if _, err := client.CreatePermission(ctx, &shieldv1beta1.CreatePermissionRequest{
		Bodies: permissionBodies,
	}); err != nil {
		return fmt.Errorf("failed to create custom permission: %w", err)
	}

	str := "created custom permissions : "
	for _, v := range permissionBodies {
		str = fmt.Sprintf("%s %s:%s", str, v.Namespace, v.Name)
		resourceNamespaces = append(resourceNamespaces, v.Namespace)
	}
	fmt.Println(str)

	var roles []*shieldv1beta1.RoleRequestBody
	if err := json.Unmarshal(mockCustomRoles, &roles); err != nil {
		return fmt.Errorf("failed to unmarshal custom roles: %w", err)
	}

	str = "created custom roles :"
	var roleResp *shieldv1beta1.CreateRoleResponse
	var err error
	for _, role := range roles {
		if roleResp, err = client.CreateRole(ctx, &shieldv1beta1.CreateRoleRequest{
			Body: role,
		}); err != nil {
			return fmt.Errorf("failed to create custom role: %w", err)
		}
		roleIDs = append(roleIDs, roleResp.Role.Id)
		str = fmt.Sprintf("%s %s", str, role.Name)
	}

	fmt.Println(str)
	return nil
}

func bootstrapData(ctx context.Context, client shieldv1beta1.ShieldServiceClient) error {
	var userBodies []*shieldv1beta1.UserRequestBody
	if err := json.Unmarshal(mockHumanUser, &userBodies); err != nil {
		return fmt.Errorf("failed to unmarshal user body: %w", err)
	}

	var orgBodies []*shieldv1beta1.OrganizationRequestBody
	if err := json.Unmarshal(mockOrganizations, &orgBodies); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	var projBodies []*shieldv1beta1.ProjectRequestBody
	if err := json.Unmarshal(mockProjects, &projBodies); err != nil {
		return fmt.Errorf("failed to unmarshal project body: %w", err)
	}

	var resourceBodies []*shieldv1beta1.ResourceRequestBody
	if err := json.Unmarshal(mockResource, &resourceBodies); err != nil {
		return fmt.Errorf("failed to unmarshal resource body: %w", err)
	}
	samplePolicyRole := []string{roleIDs[0], roleIDs[3]}
	samplePolicyNamespace := []string{resourceNamespaces[0], resourceNamespaces[5]}

	var i = 0
	for _, orgBody := range orgBodies {
		userResp, err := client.CreateUser(ctx, &shieldv1beta1.CreateUserRequest{
			Body: userBodies[i],
		})
		if err != nil {
			return fmt.Errorf("failed to create sample user: %w", err)
		}

		fmt.Printf("created user with email %s in shield\n", userResp.User.Email)

		orgResp, err := client.CreateOrganization(ctx, &shieldv1beta1.CreateOrganizationRequest{
			Body: orgBody,
		})
		if err != nil {
			return fmt.Errorf("failed to create sample organization: %w", err)
		}
		fmt.Printf("created organization name %s with user %s as the org admin \n", orgResp.Organization.Name, sampleSeedEmail)

		// create service user for an org
		serviceUserResp, err := client.CreateServiceUser(ctx, &shieldv1beta1.CreateServiceUserRequest{
			Body:  &shieldv1beta1.ServiceUserRequestBody{Title: "sample service user"},
			OrgId: orgResp.Organization.Id,
		})
		if err != nil {
			return fmt.Errorf("failed to create sample service user: %w", err)
		}

		fmt.Printf("created service user with id %s in org %s\n", serviceUserResp.Serviceuser.Id, orgResp.Organization.Name)

		// create project inside org
		projBodies[i].OrgId = orgResp.Organization.Id

		projResp, err := client.CreateProject(ctx, &shieldv1beta1.CreateProjectRequest{
			Body: projBodies[i],
		})
		if err != nil {
			return fmt.Errorf("failed to create sample project: %w", err)
		}

		fmt.Printf("created project in org %s with name %s \n", orgResp.Organization.Name, projResp.Project.Name)

		// create resource inside project
		resourceBodies[i].Principal = userResp.User.Id

		resrcResp, err := client.CreateProjectResource(ctx, &shieldv1beta1.CreateProjectResourceRequest{
			ProjectId: projResp.Project.Id,
			Body:      resourceBodies[i],
		})
		if err != nil {
			return fmt.Errorf("failed to create sample resource: %w", err)
		}

		fmt.Printf("created resource in project %s with name %s \n", projResp.Project.Name, resrcResp.Resource.Name)

		//create sample policy
		resource := fmt.Sprintf("%s:%s", samplePolicyNamespace[i], resrcResp.Resource.Id)
		user := fmt.Sprintf("%s:%s", "app/user", userResp.User.Id)
		policyResp, err := client.CreatePolicy(ctx, &shieldv1beta1.CreatePolicyRequest{
			Body: &shieldv1beta1.PolicyRequestBody{
				RoleId:    samplePolicyRole[i],
				Resource:  resource,
				Principal: user,
				Title:     "Sample Policy",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create sample policy %w", err)
		}
		fmt.Printf("sample policy created for user %s with role %s for resource %s\n", policyResp.Policy.Principal, policyResp.Policy.RoleId, policyResp.Policy.Resource)
		i++
	}
	return nil
}
