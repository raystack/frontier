package cmd

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/MakeNowJust/heredoc"
	"github.com/raystack/frontier/config"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/salt/printer"
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
		Long:  "This command can be used to create an organization structure with predefined groups, projects, and resources. It bootstarps these data in the Frontier db, making it easier to get started.",
		Example: heredoc.Doc(`
			$ frontier seed
			$ frontier seed --header=X-Frontier-Email
		`),
		Annotations: map[string]string{
			"group":  "core",
			"client": "true",
		},

		RunE: func(cmd *cli.Command, args []string) error {
			if header == "" {
				appConfig, err := config.Load(configFile)
				if err != nil {
					panic(err)
				}
				if appConfig.App.IdentityProxyHeader == "" {
					return errors.New("identity proxy header missing in server config, pass key in the header flag \nexample: frontier seed -H X-Frontier-Email")
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
			fmt.Println("initialized sample data in frontier successfully")
			return nil
		},
	}

	bindFlagsFromClientConfig(cmd)
	cmd.Flags().StringVarP(&header, "header", "H", "", "Header <key>")
	cmd.Flags().StringVarP(&configFile, "config", "c", "", "config file path")
	return cmd
}

// create sample platform wide custom permissions and roles
func createCustomRolesAndPermissions(ctx context.Context, client frontierv1beta1.AdminServiceClient) error {
	var permissionBodies []*frontierv1beta1.PermissionRequestBody
	if err := json.Unmarshal(mockCustomPermissions, &permissionBodies); err != nil {
		return fmt.Errorf("failed to unmarshal custom permissions: %w", err)
	}

	if _, err := client.CreatePermission(ctx, &frontierv1beta1.CreatePermissionRequest{
		Bodies: permissionBodies,
	}); err != nil {
		return fmt.Errorf("failed to create custom permission: %w", err)
	}

	str := "created custom permissions : "
	for _, v := range permissionBodies {
		str = fmt.Sprintf("%s %s:%s", str, v.GetNamespace(), v.GetName())
		resourceNamespaces = append(resourceNamespaces, v.GetNamespace())
	}
	fmt.Println(str)

	var roles []*frontierv1beta1.RoleRequestBody
	if err := json.Unmarshal(mockCustomRoles, &roles); err != nil {
		return fmt.Errorf("failed to unmarshal custom roles: %w", err)
	}

	str = "created custom roles :"
	var roleResp *frontierv1beta1.CreateRoleResponse
	var err error
	for _, role := range roles {
		if roleResp, err = client.CreateRole(ctx, &frontierv1beta1.CreateRoleRequest{
			Body: role,
		}); err != nil {
			return fmt.Errorf("failed to create custom role: %w", err)
		}
		roleIDs = append(roleIDs, roleResp.GetRole().GetId())
		str = fmt.Sprintf("%s %s", str, role.GetName())
	}

	fmt.Println(str)
	return nil
}

func bootstrapData(ctx context.Context, client frontierv1beta1.FrontierServiceClient) error {
	var userBodies []*frontierv1beta1.UserRequestBody
	if err := json.Unmarshal(mockHumanUser, &userBodies); err != nil {
		return fmt.Errorf("failed to unmarshal user body: %w", err)
	}

	var orgBodies []*frontierv1beta1.OrganizationRequestBody
	if err := json.Unmarshal(mockOrganizations, &orgBodies); err != nil {
		return fmt.Errorf("error unmarshaling JSON: %w", err)
	}

	var projBodies []*frontierv1beta1.ProjectRequestBody
	if err := json.Unmarshal(mockProjects, &projBodies); err != nil {
		return fmt.Errorf("failed to unmarshal project body: %w", err)
	}

	var resourceBodies []*frontierv1beta1.ResourceRequestBody
	if err := json.Unmarshal(mockResource, &resourceBodies); err != nil {
		return fmt.Errorf("failed to unmarshal resource body: %w", err)
	}
	samplePolicyRole := []string{roleIDs[0], roleIDs[3]}
	samplePolicyNamespace := []string{resourceNamespaces[0], resourceNamespaces[5]}

	reportUser := [][]string{}
	reportUser = append(reportUser, []string{"USER_ID", "NAME", "EMAIL", "TITLE"})

	reportServiceUser := [][]string{}
	reportServiceUser = append(reportServiceUser, []string{"SERVICE_USER_ID", "ORG_ID", "TITLE"})

	reportServiceUserCred := [][]string{}
	reportServiceUserCred = append(reportServiceUserCred, []string{"ID", "SERVICE_USER_ID", "SECRET_HASH"})

	reportOrg := [][]string{}
	reportOrg = append(reportOrg, []string{"ORG_ID", "NAME", "ORG_ADMIN"})

	reportProject := [][]string{}
	reportProject = append(reportProject, []string{"PROJECT_ID", "PROJECT_NAME", "PROJECT_TITLE", "ORG_NAME"})

	reportResource := [][]string{}
	reportResource = append(reportResource, []string{"RESOURCE_ID", "RESOURCE_NAME", "RESOURCE_NAMESPACE", "PROJECT_NAME"})

	reportPolicy := [][]string{}
	reportPolicy = append(reportPolicy, []string{"CREATED_FOR", "ROLE", "RESOURCE"})

	for idx, orgBody := range orgBodies {
		userResp, err := client.CreateUser(ctx, &frontierv1beta1.CreateUserRequest{
			Body: userBodies[idx],
		})
		if err != nil {
			return fmt.Errorf("failed to create sample user: %w", err)
		}
		reportUser = append(reportUser, []string{
			userResp.GetUser().GetId(),
			userResp.GetUser().GetName(),
			userResp.GetUser().GetEmail(),
			userResp.GetUser().GetTitle(),
		})

		orgResp, err := client.CreateOrganization(ctx, &frontierv1beta1.CreateOrganizationRequest{
			Body: orgBody,
		})
		if err != nil {
			return fmt.Errorf("failed to create sample organization: %w", err)
		}
		reportOrg = append(reportOrg, []string{
			orgResp.GetOrganization().GetId(),
			orgResp.GetOrganization().GetName(),
			sampleSeedEmail,
		})

		// create service user for an org
		serviceUserResp, err := client.CreateServiceUser(ctx, &frontierv1beta1.CreateServiceUserRequest{
			Body:  &frontierv1beta1.ServiceUserRequestBody{Title: "sample service user"},
			OrgId: orgResp.GetOrganization().GetId(),
		})
		if err != nil {
			return fmt.Errorf("failed to create sample service user: %w", err)
		}
		reportServiceUser = append(reportServiceUser, []string{
			serviceUserResp.GetServiceuser().GetId(),
			orgResp.GetOrganization().GetId(),
			serviceUserResp.GetServiceuser().GetTitle(),
		})
		// create service user credentials for an org
		serviceUserCredentialResp, err := client.CreateServiceUserCredential(ctx, &frontierv1beta1.CreateServiceUserCredentialRequest{
			Id:    serviceUserResp.GetServiceuser().GetId(),
			Title: "service user id and pass",
		})
		if err != nil {
			return fmt.Errorf("failed to generate sample service user password: %w", err)
		}
		reportServiceUserCred = append(reportServiceUserCred, []string{
			serviceUserCredentialResp.GetSecret().GetId(),
			serviceUserResp.GetServiceuser().GetId(),
			serviceUserCredentialResp.GetSecret().GetSecret(),
		})

		// create project inside org
		projBodies[idx].OrgId = orgResp.GetOrganization().GetId()
		projResp, err := client.CreateProject(ctx, &frontierv1beta1.CreateProjectRequest{
			Body: projBodies[idx],
		})
		if err != nil {
			return fmt.Errorf("failed to create sample project: %w", err)
		}
		reportProject = append(reportProject, []string{
			projResp.GetProject().GetId(),
			projResp.GetProject().GetName(),
			projResp.GetProject().GetTitle(),
			orgResp.GetOrganization().GetName(),
		})

		// create resource inside project
		resourceBodies[idx].Principal = userResp.GetUser().GetId()
		resrcResp, err := client.CreateProjectResource(ctx, &frontierv1beta1.CreateProjectResourceRequest{
			ProjectId: projResp.GetProject().GetId(),
			Body:      resourceBodies[idx],
		})
		if err != nil {
			return fmt.Errorf("failed to create sample resource: %w", err)
		}
		reportResource = append(reportResource, []string{
			resrcResp.GetResource().GetId(),
			resrcResp.GetResource().GetName(),
			resrcResp.GetResource().GetNamespace(),
			projResp.GetProject().GetName(),
		})

		// create sample policy
		resource := fmt.Sprintf("%s:%s", samplePolicyNamespace[idx], resrcResp.GetResource().GetId())
		user := fmt.Sprintf("%s:%s", "app/user", userResp.GetUser().GetId())
		policyResp, err := client.CreatePolicy(ctx, &frontierv1beta1.CreatePolicyRequest{
			Body: &frontierv1beta1.PolicyRequestBody{
				RoleId:    samplePolicyRole[idx],
				Resource:  resource,
				Principal: user,
				Title:     "Sample Policy",
			},
		})
		if err != nil {
			return fmt.Errorf("failed to create sample policy %w", err)
		}
		reportPolicy = append(reportPolicy, []string{
			policyResp.GetPolicy().GetPrincipal(),
			policyResp.GetPolicy().GetRoleId(),
			policyResp.GetPolicy().GetResource(),
		})
	}
	fmt.Printf("\n")
	fmt.Println("Created User in frontier")
	printer.Table(os.Stdout, reportUser)
	fmt.Printf("\n")
	fmt.Println("Created Organization")
	printer.Table(os.Stdout, reportOrg)
	fmt.Printf("\n")
	fmt.Println("Created Service User in Org")
	printer.Table(os.Stdout, reportServiceUser)
	fmt.Printf("\n")
	fmt.Println("Created Service User Credentials in Org")
	printer.Table(os.Stdout, reportServiceUserCred)
	fmt.Printf("\n")
	fmt.Println("Created Project in Org")
	printer.Table(os.Stdout, reportProject)
	fmt.Printf("\n")
	fmt.Println("Created Resource in Project")
	printer.Table(os.Stdout, reportResource)
	fmt.Printf("\n")
	fmt.Println("Created Policy for User")
	printer.Table(os.Stdout, reportPolicy)
	return nil
}
