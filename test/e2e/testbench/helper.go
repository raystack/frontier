package testbench

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"net"

	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

var (
	//go:embed testdata/mocks/mock-user.json
	mockUserFixture []byte
	//go:embed testdata/mocks/mock-group.json
	mockGroupFixture []byte
	//go:embed testdata/mocks/mock-organization.json
	mockOrganizationFixture []byte
	//go:embed testdata/mocks/mock-project.json
	mockProjectFixture []byte
)

const (
	OrgAdminEmail  = "admin1-group1-org1@raystack.org"
	IdentityHeader = "X-Frontier-Email"
)

func GetFreePort() (int, error) {
	addr, err := net.ResolveTCPAddr("tcp", "localhost:0")
	if err != nil {
		return 0, err
	}

	l, err := net.ListenTCP("tcp", addr)
	if err != nil {
		return 0, err
	}
	defer l.Close()
	return l.Addr().(*net.TCPAddr).Port, nil
}

func createConnection(ctx context.Context, host string) (*grpc.ClientConn, error) {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	}

	return grpc.DialContext(ctx, host, opts...)
}

func CreateClient(ctx context.Context, host string) (frontierv1beta1.FrontierServiceClient, func() error, error) {
	conn, err := createConnection(ctx, host)
	if err != nil {
		return nil, nil, err
	}
	client := frontierv1beta1.NewFrontierServiceClient(conn)
	return client, conn.Close, nil
}

func CreateAdminClient(ctx context.Context, host string) (frontierv1beta1.AdminServiceClient, func() error, error) {
	conn, err := createConnection(ctx, host)
	if err != nil {
		return nil, nil, err
	}
	client := frontierv1beta1.NewAdminServiceClient(conn)
	return client, conn.Close, nil
}

func BootstrapUsers(ctx context.Context, cl frontierv1beta1.FrontierServiceClient, creatorEmail string) error {
	var data []*frontierv1beta1.UserRequestBody
	if err := json.Unmarshal(mockUserFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			IdentityHeader: creatorEmail,
		}))
		if _, err := cl.CreateUser(ctx, &frontierv1beta1.CreateUserRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListUsers(ctx, &frontierv1beta1.ListUsersRequest{})
	if err != nil {
		return err
	}
	// +1 for counting admin user
	if len(data)+1 != len(uRes.GetUsers()) {
		return errors.New("failed to validate number of users created")
	}
	return nil
}

func BootstrapOrganizations(ctx context.Context, cl frontierv1beta1.FrontierServiceClient, creatorEmail string) error {
	var data []*frontierv1beta1.OrganizationRequestBody
	if err := json.Unmarshal(mockOrganizationFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			IdentityHeader: creatorEmail,
		}))
		if _, err := cl.CreateOrganization(ctx, &frontierv1beta1.CreateOrganizationRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListOrganizations(ctx, &frontierv1beta1.ListOrganizationsRequest{})
	if err != nil {
		return err
	}
	if len(data) != len(uRes.GetOrganizations()) {
		return errors.New("failed to validate number of organizations created")
	}
	return nil
}

func BootstrapProject(ctx context.Context, cl frontierv1beta1.FrontierServiceClient, creatorEmail string) error {
	orgResp, err := cl.ListOrganizations(ctx, &frontierv1beta1.ListOrganizationsRequest{})
	if err != nil {
		return err
	}

	if len(orgResp.GetOrganizations()) < 1 {
		return errors.New("no organization found")
	}

	var data []*frontierv1beta1.ProjectRequestBody
	if err = json.Unmarshal(mockProjectFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		d.OrgId = orgResp.GetOrganizations()[0].GetId()
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			IdentityHeader: creatorEmail,
		}))
		if _, err := cl.CreateProject(ctx, &frontierv1beta1.CreateProjectRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListOrganizationProjects(ctx, &frontierv1beta1.ListOrganizationProjectsRequest{
		Id: orgResp.GetOrganizations()[0].GetId(),
	})
	if err != nil {
		return err
	}
	if len(data) != len(uRes.GetProjects()) {
		return errors.New("failed to validate number of projects created")
	}
	return nil
}

func BootstrapGroup(ctx context.Context, cl frontierv1beta1.FrontierServiceClient, creatorEmail string) error {
	orgResp, err := cl.ListOrganizations(ctx, &frontierv1beta1.ListOrganizationsRequest{})
	if err != nil {
		return err
	}

	if len(orgResp.GetOrganizations()) < 1 {
		return errors.New("no organization found")
	}

	var data []*frontierv1beta1.GroupRequestBody
	if err = json.Unmarshal(mockGroupFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			IdentityHeader: creatorEmail,
		}))
		if _, err := cl.CreateGroup(ctx, &frontierv1beta1.CreateGroupRequest{
			Body:  d,
			OrgId: orgResp.GetOrganizations()[0].GetId(),
		}); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListOrganizationGroups(ctx, &frontierv1beta1.ListOrganizationGroupsRequest{
		OrgId: orgResp.GetOrganizations()[0].GetId(),
	})
	if err != nil {
		return err
	}
	if len(data) != len(uRes.GetGroups()) {
		return errors.New("failed to validate number of groups created")
	}
	return nil
}
