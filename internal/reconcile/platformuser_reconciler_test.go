package reconcile

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/structpb"
)

type fakePlatformUserAPI struct {
	users   []*frontierv1beta1.User
	sus     []*frontierv1beta1.ServiceUser
	added   []*frontierv1beta1.AddPlatformUserRequest
	removed []*frontierv1beta1.RemovePlatformUserRequest
}

func (f *fakePlatformUserAPI) ListPlatformUsers(_ context.Context, _ *connect.Request[frontierv1beta1.ListPlatformUsersRequest]) (*connect.Response[frontierv1beta1.ListPlatformUsersResponse], error) {
	return connect.NewResponse(&frontierv1beta1.ListPlatformUsersResponse{Users: f.users, Serviceusers: f.sus}), nil
}

func (f *fakePlatformUserAPI) AddPlatformUser(_ context.Context, req *connect.Request[frontierv1beta1.AddPlatformUserRequest]) (*connect.Response[frontierv1beta1.AddPlatformUserResponse], error) {
	f.added = append(f.added, req.Msg)
	return connect.NewResponse(&frontierv1beta1.AddPlatformUserResponse{}), nil
}

func (f *fakePlatformUserAPI) RemovePlatformUser(_ context.Context, req *connect.Request[frontierv1beta1.RemovePlatformUserRequest]) (*connect.Response[frontierv1beta1.RemovePlatformUserResponse], error) {
	f.removed = append(f.removed, req.Msg)
	return connect.NewResponse(&frontierv1beta1.RemovePlatformUserResponse{}), nil
}

func platformUserPB(t *testing.T, id, email, relation string) *frontierv1beta1.User {
	t.Helper()
	md, err := structpb.NewStruct(map[string]interface{}{"relation": relation})
	if err != nil {
		t.Fatalf("struct: %v", err)
	}
	return &frontierv1beta1.User{Id: id, Email: email, Metadata: md}
}

func TestPlatformUserReconciler_Reconcile(t *testing.T) {
	// desired: one new admin; the existing "drop" admin is absent -> removed.
	specYAML := []byte("- {type: user, ref: new@x.com, role: admin}\n")

	t.Run("applies adds and removes", func(t *testing.T) {
		api := &fakePlatformUserAPI{users: []*frontierv1beta1.User{
			platformUserPB(t, "drop-id", "drop@x.com", schema.AdminRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), specYAML, false)

		assert.NoError(t, err)
		assert.Equal(t, 2, rep.Applied)
		if assert.Len(t, api.removed, 1) {
			assert.Equal(t, "drop-id", api.removed[0].GetUserId())
		}
		if assert.Len(t, api.added, 1) {
			assert.Equal(t, "new@x.com", api.added[0].GetUserId())
			assert.Equal(t, schema.AdminRelationName, api.added[0].GetRelation())
		}
	})

	t.Run("dry-run plans without applying", func(t *testing.T) {
		api := &fakePlatformUserAPI{users: []*frontierv1beta1.User{
			platformUserPB(t, "drop-id", "drop@x.com", schema.AdminRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), specYAML, true)

		assert.NoError(t, err)
		assert.True(t, rep.DryRun)
		assert.NotEmpty(t, rep.Planned)
		assert.Equal(t, 0, rep.Applied)
		assert.Empty(t, api.added)
		assert.Empty(t, api.removed)
	})

	t.Run("no changes when already converged", func(t *testing.T) {
		api := &fakePlatformUserAPI{users: []*frontierv1beta1.User{
			platformUserPB(t, "new-id", "new@x.com", schema.AdminRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), specYAML, false)

		assert.NoError(t, err)
		assert.Empty(t, rep.Planned)
		assert.Empty(t, api.added)
		assert.Empty(t, api.removed)
	})
}
