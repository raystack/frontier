package testbench

import (
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"

	"connectrpc.com/connect"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/raystack/frontier/proto/v1beta1/frontierv1beta1connect"
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

// headersKey is the context key for storing headers to be sent with ConnectRPC requests.
type headersKey struct{}

// ContextWithHeaders returns a new context with the given headers.
// These headers will be automatically applied to ConnectRPC requests
// by the headerInterceptor.
func ContextWithHeaders(ctx context.Context, headers map[string]string) context.Context {
	return context.WithValue(ctx, headersKey{}, headers)
}

// HeadersFromContext returns headers stored in the context, if any.
func HeadersFromContext(ctx context.Context) map[string]string {
	if h, ok := ctx.Value(headersKey{}).(map[string]string); ok {
		return h
	}
	return nil
}

// headerInterceptor is a ConnectRPC unary interceptor that reads headers
// from the context and sets them on the outgoing request.
func headerInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return connect.UnaryFunc(func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			if headers := HeadersFromContext(ctx); headers != nil {
				for k, v := range headers {
					req.Header().Set(k, v)
				}
			}
			return next(ctx, req)
		})
	})
}

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

func CreateClient(host string) (frontierv1beta1connect.FrontierServiceClient, error) {
	return frontierv1beta1connect.NewFrontierServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://%s", host),
		connect.WithInterceptors(headerInterceptor()),
	), nil
}

func CreateAdminClient(host string) (frontierv1beta1connect.AdminServiceClient, error) {
	return frontierv1beta1connect.NewAdminServiceClient(
		http.DefaultClient,
		fmt.Sprintf("http://%s", host),
		connect.WithInterceptors(headerInterceptor()),
	), nil
}

func BootstrapUsers(ctx context.Context, cl frontierv1beta1connect.FrontierServiceClient, creatorEmail string) error {
	var data []*frontierv1beta1.UserRequestBody
	if err := json.Unmarshal(mockUserFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = ContextWithHeaders(ctx, map[string]string{
			IdentityHeader: creatorEmail,
		})
		if _, err := cl.CreateUser(ctx, connect.NewRequest(&frontierv1beta1.CreateUserRequest{
			Body: d,
		})); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListUsers(ctx, connect.NewRequest(&frontierv1beta1.ListUsersRequest{}))
	if err != nil {
		return err
	}
	// +1 for counting admin user
	if len(data)+1 != len(uRes.Msg.GetUsers()) {
		return errors.New("failed to validate number of users created")
	}
	return nil
}

func BootstrapOrganizations(ctx context.Context, cl frontierv1beta1connect.FrontierServiceClient, creatorEmail string) error {
	var data []*frontierv1beta1.OrganizationRequestBody
	if err := json.Unmarshal(mockOrganizationFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = ContextWithHeaders(ctx, map[string]string{
			IdentityHeader: creatorEmail,
		})
		if _, err := cl.CreateOrganization(ctx, connect.NewRequest(&frontierv1beta1.CreateOrganizationRequest{
			Body: d,
		})); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListOrganizations(ctx, connect.NewRequest(&frontierv1beta1.ListOrganizationsRequest{}))
	if err != nil {
		return err
	}
	if len(data) != len(uRes.Msg.GetOrganizations()) {
		return errors.New("failed to validate number of organizations created")
	}
	return nil
}

func BootstrapProject(ctx context.Context, cl frontierv1beta1connect.FrontierServiceClient, creatorEmail string) error {
	orgResp, err := cl.ListOrganizations(ctx, connect.NewRequest(&frontierv1beta1.ListOrganizationsRequest{}))
	if err != nil {
		return err
	}

	if len(orgResp.Msg.GetOrganizations()) < 1 {
		return errors.New("no organization found")
	}

	var data []*frontierv1beta1.ProjectRequestBody
	if err = json.Unmarshal(mockProjectFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		d.OrgId = orgResp.Msg.GetOrganizations()[0].GetId()
		ctx = ContextWithHeaders(ctx, map[string]string{
			IdentityHeader: creatorEmail,
		})
		if _, err := cl.CreateProject(ctx, connect.NewRequest(&frontierv1beta1.CreateProjectRequest{
			Body: d,
		})); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListOrganizationProjects(ctx, connect.NewRequest(&frontierv1beta1.ListOrganizationProjectsRequest{
		Id: orgResp.Msg.GetOrganizations()[0].GetId(),
	}))
	if err != nil {
		return err
	}
	if len(data) != len(uRes.Msg.GetProjects()) {
		return errors.New("failed to validate number of projects created")
	}
	return nil
}

func BootstrapGroup(ctx context.Context, cl frontierv1beta1connect.FrontierServiceClient, creatorEmail string) error {
	orgResp, err := cl.ListOrganizations(ctx, connect.NewRequest(&frontierv1beta1.ListOrganizationsRequest{}))
	if err != nil {
		return err
	}

	if len(orgResp.Msg.GetOrganizations()) < 1 {
		return errors.New("no organization found")
	}

	var data []*frontierv1beta1.GroupRequestBody
	if err = json.Unmarshal(mockGroupFixture, &data); err != nil {
		return err
	}

	for _, d := range data {
		ctx = ContextWithHeaders(ctx, map[string]string{
			IdentityHeader: creatorEmail,
		})
		if _, err := cl.CreateGroup(ctx, connect.NewRequest(&frontierv1beta1.CreateGroupRequest{
			Body:  d,
			OrgId: orgResp.Msg.GetOrganizations()[0].GetId(),
		})); err != nil {
			return err
		}
	}

	// validate
	uRes, err := cl.ListOrganizationGroups(ctx, connect.NewRequest(&frontierv1beta1.ListOrganizationGroupsRequest{
		OrgId: orgResp.Msg.GetOrganizations()[0].GetId(),
	}))
	if err != nil {
		return err
	}
	if len(data) != len(uRes.Msg.GetGroups()) {
		return errors.New("failed to validate number of groups created")
	}
	return nil
}
