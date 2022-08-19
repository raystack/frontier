package e2e_test

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
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

func createClient(ctx context.Context, host string) (shieldv1beta1.ShieldServiceClient, func(), error) {
	conn, err := createConnection(context.Background(), host)
	if err != nil {
		return nil, nil, err
	}

	cancel := func() {
		conn.Close()
	}

	client := shieldv1beta1.NewShieldServiceClient(conn)
	return client, cancel, nil
}

func bootstrapUser(ctx context.Context, cl shieldv1beta1.ShieldServiceClient, creatorEmail string) error {
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mocks/mock-user.json")
	if err != nil {
		return err
	}

	var data []*shieldv1beta1.UserRequestBody
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"X-Shield-Email": creatorEmail,
		}))
		if _, err := cl.CreateUser(ctx, &shieldv1beta1.CreateUserRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	return nil
}

func bootstrapOrganization(ctx context.Context, cl shieldv1beta1.ShieldServiceClient, creatorEmail string) error {
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mocks/mock-organization.json")
	if err != nil {
		return err
	}

	var data []*shieldv1beta1.OrganizationRequestBody
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"X-Shield-Email": creatorEmail,
		}))
		if _, err := cl.CreateOrganization(ctx, &shieldv1beta1.CreateOrganizationRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	return nil
}

func bootstrapProject(ctx context.Context, cl shieldv1beta1.ShieldServiceClient, creatorEmail string) error {
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mocks/mock-project.json")
	if err != nil {
		return err
	}

	orgResp, err := cl.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
	if err != nil {
		return err
	}

	if len(orgResp.GetOrganizations()) < 1 {
		return errors.New("no organization found")
	}

	var data []*shieldv1beta1.ProjectRequestBody
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return err
	}

	data[0].OrgId = orgResp.GetOrganizations()[0].GetId()

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"X-Shield-Email": creatorEmail,
		}))
		if _, err := cl.CreateProject(ctx, &shieldv1beta1.CreateProjectRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	return nil
}

func bootstrapGroup(ctx context.Context, cl shieldv1beta1.ShieldServiceClient, creatorEmail string) error {
	testFixtureJSON, err := ioutil.ReadFile("./testdata/mocks/mock-group.json")
	if err != nil {
		return err
	}

	orgResp, err := cl.ListOrganizations(ctx, &shieldv1beta1.ListOrganizationsRequest{})
	if err != nil {
		return err
	}

	if len(orgResp.GetOrganizations()) < 1 {
		return errors.New("no organization found")
	}

	var data []*shieldv1beta1.GroupRequestBody
	if err = json.Unmarshal(testFixtureJSON, &data); err != nil {
		return err
	}

	data[0].OrgId = orgResp.GetOrganizations()[0].GetId()
	data[1].OrgId = orgResp.GetOrganizations()[0].GetId()
	data[2].OrgId = orgResp.GetOrganizations()[0].GetId()

	for _, d := range data {
		ctx = metadata.NewOutgoingContext(ctx, metadata.New(map[string]string{
			"X-Shield-Email": creatorEmail,
		}))
		if _, err := cl.CreateGroup(ctx, &shieldv1beta1.CreateGroupRequest{
			Body: d,
		}); err != nil {
			return err
		}
	}

	return nil
}
