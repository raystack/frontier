package postgres_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/raystack/frontier/core/kyc"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/raystack/frontier/core/invitation"
	"github.com/raystack/frontier/core/permission"

	"github.com/ory/dockertest"
	"github.com/ory/dockertest/docker"
	"github.com/raystack/frontier/cmd"
	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/namespace"
	"github.com/raystack/frontier/core/policy"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/resource"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/log"
)

const (
	logLevelDebug = "debug"
	pg_uname      = "test_user"
	pg_passwd     = "test_pass"
	pg_dbname     = "test_db"
)

func newTestClient(logger log.Logger) (*db.Client, *dockertest.Pool, *dockertest.Resource, error) {
	opts := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "12",
		Env: []string{
			"POSTGRES_PASSWORD=" + pg_passwd,
			"POSTGRES_USER=" + pg_uname,
			"POSTGRES_DB=" + pg_dbname,
		},
	}

	// uses a sensible default on windows (tcp/http) and linux/osx (socket)
	pool, err := dockertest.NewPool("")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not create dockertest pool: %w", err)
	}

	// pulls an image, creates a container based on it and runs it
	resource, err := pool.RunWithOptions(opts, func(config *docker.HostConfig) {
		// set AutoRemove to true so that stopped container goes away by itself
		config.AutoRemove = true
		config.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		return nil, nil, nil, fmt.Errorf("could not start resource: %w", err)
	}

	pg_port := resource.GetPort("5432/tcp")
	if err != nil {
		return nil, nil, nil, fmt.Errorf("cannot parse external port of container to int: %w", err)
	}

	// attach terminal logger to container if exists
	// for debugging purpose
	if logger.Level() == logLevelDebug {
		logWaiter, err := pool.Client.AttachToContainerNonBlocking(docker.AttachToContainerOptions{
			Container:    resource.Container.ID,
			OutputStream: logger.Writer(),
			ErrorStream:  logger.Writer(),
			Stderr:       true,
			Stdout:       true,
			Stream:       true,
		})
		if err != nil {
			logger.Fatal("could not connect to postgres container log output", "error", err)
		}
		defer func() {
			err = logWaiter.Close()
			if err != nil {
				logger.Fatal("could not close container log", "error", err)
			}

			err = logWaiter.Wait()
			if err != nil {
				logger.Fatal("could not wait for container log to close", "error", err)
			}
		}()
	}

	// Tell docker to hard kill the container in 120 seconds
	if err := resource.Expire(120); err != nil {
		return nil, nil, nil, err
	}

	// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
	pool.MaxWait = 60 * time.Second

	pgConfig := db.Config{
		Driver:          "pgx",
		URL:             fmt.Sprintf("postgres://%s:%s@localhost:%s/%s?sslmode=disable", pg_uname, pg_passwd, pg_port, pg_dbname),
		MaxIdleConns:    10,
		MaxOpenConns:    10,
		ConnMaxLifeTime: time.Second * 60,
		MaxQueryTimeout: time.Millisecond * 1000,
	}
	var pgClient *db.Client
	if err = pool.Retry(func() error {
		pgClient, err = db.New(pgConfig)
		if err != nil {
			return err
		}
		return nil
	}); err != nil {
		return nil, nil, nil, fmt.Errorf("could not connect to docker: %w", err)
	}

	err = setup(context.Background(), logger, pgClient, pgConfig)
	if err != nil {
		logger.Fatal("failed to setup and migrate DB", "error", err)
	}
	return pgClient, pool, resource, nil
}

func purgeDocker(pool *dockertest.Pool, resource *dockertest.Resource) error {
	if err := pool.Purge(resource); err != nil {
		return fmt.Errorf("could not purge resource: %w", err)
	}
	return nil
}

func setup(ctx context.Context, logger log.Logger, client *db.Client, cfg db.Config) (err error) {
	var queries = []string{
		"DROP SCHEMA public CASCADE",
		"CREATE SCHEMA public",
	}

	err = execQueries(ctx, client, queries)
	if err != nil {
		return
	}

	err = cmd.RunMigrations(logger, cfg)
	return
}

// ExecQueries is used for executing list of db query
func execQueries(ctx context.Context, client *db.Client, queries []string) error {
	for _, query := range queries {
		if _, err := client.DB.ExecContext(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

func bootstrapPermissions(client *db.Client) ([]permission.Permission, error) {
	actionRepository := postgres.NewPermissionRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-permission.json")
	if err != nil {
		return nil, err
	}

	var data []permission.Permission
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []permission.Permission
	for _, d := range data {
		act, err := actionRepository.Upsert(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, act)
	}
	return insertedData, nil
}

func bootstrapNamespace(client *db.Client) ([]namespace.Namespace, error) {
	namespaceRepository := postgres.NewNamespaceRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-namespace.json")
	if err != nil {
		return nil, err
	}

	var data []namespace.Namespace
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []namespace.Namespace
	for _, d := range data {
		domain, err := namespaceRepository.Upsert(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapUser(client *db.Client) ([]user.User, error) {
	userRepository := postgres.NewUserRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-user.json")
	if err != nil {
		return nil, err
	}

	var data []user.User
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []user.User
	for _, d := range data {
		domain, err := userRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapRole(client *db.Client, orgID string) ([]role.Role, error) {
	roleRepository := postgres.NewRoleRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-role.json")
	if err != nil {
		return nil, err
	}

	var data []role.Role
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []role.Role
	for _, d := range data {
		d.OrgID = orgID
		domain, err := roleRepository.Upsert(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapPolicy(client *db.Client, orgID string, role role.Role, userID string) ([]policy.Policy, error) {
	policyRepository := postgres.NewPolicyRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-policy.json")
	if err != nil {
		return nil, err
	}

	var data []policy.Policy
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []policy.Policy
	for _, d := range data {
		d.PrincipalID = userID
		d.ResourceID = orgID
		d.RoleID = role.ID
		pol, err := policyRepository.Upsert(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, pol)
	}

	return insertedData, nil
}

func bootstrapRelation(client *db.Client) ([]relation.Relation, error) {
	relationRepository := postgres.NewRelationRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-relation.json")
	if err != nil {
		return nil, err
	}

	var data []relation.Relation
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []relation.Relation
	for _, d := range data {
		domain, err := relationRepository.Upsert(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapOrganization(client *db.Client) ([]organization.Organization, error) {
	orgRepository := postgres.NewOrganizationRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-organization.json")
	if err != nil {
		return nil, err
	}

	var data []organization.Organization
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []organization.Organization
	for _, d := range data {
		domain, err := orgRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapOrganizationKYC(ctx context.Context, client *db.Client, orgs []organization.Organization) ([]kyc.KYC, error) {
	orgKycRepository := postgres.NewOrgKycRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-organization-kyc.json")
	if err != nil {
		return nil, err
	}

	var data []kyc.KYC
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	var insertedData []kyc.KYC
	for idx, d := range data {
		d.OrgID = orgs[idx].ID
		domain, err := orgKycRepository.Upsert(ctx, d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapProject(client *db.Client, orgs []organization.Organization) ([]project.Project, error) {
	projectRepository := postgres.NewProjectRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-project.json")
	if err != nil {
		return nil, err
	}

	var data []project.Project
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	data[0].Organization = organization.Organization{ID: orgs[0].ID}
	data[1].Organization = organization.Organization{ID: orgs[0].ID}
	data[2].Organization = organization.Organization{ID: orgs[1].ID}

	var insertedData []project.Project
	for _, d := range data {
		domain, err := projectRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapGroup(client *db.Client, orgs []organization.Organization) ([]group.Group, error) {
	groupRepository := postgres.NewGroupRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-group.json")
	if err != nil {
		return nil, err
	}

	var data []group.Group
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	data[0].OrganizationID = orgs[0].ID
	data[1].OrganizationID = orgs[0].ID
	data[2].OrganizationID = orgs[1].ID
	data[3].OrganizationID = orgs[1].ID

	var insertedData []group.Group
	for _, d := range data {
		domain, err := groupRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}

func bootstrapInvitation(client *db.Client, users []user.User, orgs []organization.Organization, groups []group.Group) ([]invitation.Invitation, error) {
	invitationRepository := postgres.NewInvitationRepository(log.NewLogrus(), client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-invitation.json")
	if err != nil {
		return nil, err
	}

	var data []invitation.Invitation
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	data[0].OrgID = orgs[0].ID
	data[0].UserEmailID = users[0].Email
	data[0].GroupIDs = []string{groups[0].ID}
	data[1].OrgID = orgs[1].ID
	data[1].UserEmailID = users[1].Email
	data[1].GroupIDs = []string{groups[1].ID}

	var insertedData []invitation.Invitation
	for _, d := range data {
		invite, err := invitationRepository.Set(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, invite)
	}

	return insertedData, nil
}

func bootstrapResource(
	client *db.Client,
	projects []project.Project,
	namespaces []namespace.Namespace,
	users []user.User) ([]resource.Resource, error) {
	resRepository := postgres.NewResourceRepository(client)
	testFixtureJSON, err := os.ReadFile("./testdata/mock-resource.json")
	if err != nil {
		return nil, err
	}

	var data []resource.Resource
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return nil, err
	}

	data[0].ProjectID = projects[0].ID
	data[0].NamespaceID = namespaces[0].Name
	data[0].PrincipalID = users[0].ID
	data[0].PrincipalType = schema.UserPrincipal

	data[1].ProjectID = projects[1].ID
	data[1].NamespaceID = namespaces[1].Name
	data[1].PrincipalID = users[1].ID
	data[1].PrincipalType = schema.UserPrincipal

	data[2].ProjectID = projects[2].ID
	data[2].NamespaceID = namespaces[1].Name
	data[2].PrincipalID = users[1].ID
	data[2].PrincipalType = schema.UserPrincipal

	var insertedData []resource.Resource
	for _, d := range data {
		domain, err := resRepository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}

		insertedData = append(insertedData, domain)
	}

	return insertedData, nil
}
