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

// platformUserPBRelations stamps the full "relations" list the way ListPlatformUsers does.
func platformUserPBRelations(t *testing.T, id, email string, relations ...string) *frontierv1beta1.User {
	t.Helper()
	vals := make([]interface{}, len(relations))
	for i, r := range relations {
		vals[i] = r
	}
	md, err := structpb.NewStruct(map[string]interface{}{"relations": vals})
	if err != nil {
		t.Fatalf("struct: %v", err)
	}
	return &frontierv1beta1.User{Id: id, Email: email, Metadata: md}
}

// serviceUserPBRelations builds a platform service-user list entry.
func serviceUserPBRelations(t *testing.T, id string, relations ...string) *frontierv1beta1.ServiceUser {
	t.Helper()
	vals := make([]interface{}, len(relations))
	for i, r := range relations {
		vals[i] = r
	}
	md, err := structpb.NewStruct(map[string]interface{}{"relations": vals})
	if err != nil {
		t.Fatalf("struct: %v", err)
	}
	return &frontierv1beta1.ServiceUser{Id: id, Metadata: md}
}

func TestPlatformUserReconciler_Reconcile(t *testing.T) {
	// desired: one new admin; the existing "drop" admin is absent -> removed.
	specYAML := []byte("- {type: user, ref: new@x.com, relation: admin}\n")

	t.Run("applies adds and removes", func(t *testing.T) {
		api := &fakePlatformUserAPI{users: []*frontierv1beta1.User{
			platformUserPB(t, "drop-id", "drop@x.com", schema.AdminRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), specYAML, false)

		assert.NoError(t, err)
		assert.Equal(t, 2, rep.Applied)
		if assert.Len(t, api.removed, 1) {
			assert.Equal(t, "drop-id", api.removed[0].GetUserId())
			assert.Equal(t, schema.AdminRelationName, api.removed[0].GetRelation())
		}
		if assert.Len(t, api.added, 1) {
			assert.Equal(t, "new@x.com", api.added[0].GetUserId())
			assert.Equal(t, schema.AdminRelationName, api.added[0].GetRelation())
		}
	})

	t.Run("an unknown field in the spec fails the plan", func(t *testing.T) {
		api := &fakePlatformUserAPI{}
		// `relatio` instead of `relation`: must fail, not be silently ignored
		spec := []byte("- {type: user, ref: a@x.com, relatio: admin}\n")
		_, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), spec, true)
		assert.ErrorContains(t, err, "parse PlatformUser spec")
		assert.Empty(t, api.added)
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

	t.Run("no changes when already matching", func(t *testing.T) {
		api := &fakePlatformUserAPI{users: []*frontierv1beta1.User{
			platformUserPB(t, "new-id", "new@x.com", schema.AdminRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), specYAML, false)

		assert.NoError(t, err)
		assert.Empty(t, rep.Planned)
		assert.Empty(t, api.added)
		assert.Empty(t, api.removed)
	})

	t.Run("never removes the bootstrap service account (fixed id)", func(t *testing.T) {
		// an empty desired state would remove everyone, but the bootstrap SA is
		// matched by its fixed id and must be left out and untouched.
		api := &fakePlatformUserAPI{sus: []*frontierv1beta1.ServiceUser{
			serviceUserPBRelations(t, schema.BootstrapServiceUserID, schema.AdminRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), []byte("[]\n"), false)

		assert.NoError(t, err)
		assert.Empty(t, rep.Planned)
		assert.Empty(t, api.added)
		assert.Empty(t, api.removed)
	})

	t.Run("export lists every principal-relation pair, sorted, without the bootstrap SA", func(t *testing.T) {
		// unsorted input, a user without an email (falls back to id), a user and a
		// service user that each hold both relations (one entry per relation), and
		// the bootstrap SA (must be skipped).
		api := &fakePlatformUserAPI{
			users: []*frontierv1beta1.User{
				platformUserPBRelations(t, "zoe-id", "zoe@x.com", schema.MemberRelationName),
				platformUserPBRelations(t, "noemail-id", "", schema.AdminRelationName),
				platformUserPBRelations(t, "alice-id", "alice@x.com", schema.MemberRelationName, schema.AdminRelationName),
			},
			sus: []*frontierv1beta1.ServiceUser{
				serviceUserPBRelations(t, schema.BootstrapServiceUserID, schema.AdminRelationName),
				serviceUserPBRelations(t, "sa-id", schema.MemberRelationName, schema.AdminRelationName),
			},
		}
		spec, err := NewPlatformUserReconciler(api, "").Export(context.Background())

		assert.NoError(t, err)
		assert.Equal(t, []PlatformUserSpec{
			{Type: "serviceuser", Ref: "sa-id", Relation: schema.AdminRelationName},
			{Type: "serviceuser", Ref: "sa-id", Relation: schema.MemberRelationName},
			{Type: "user", Ref: "alice@x.com", Relation: schema.AdminRelationName},
			{Type: "user", Ref: "alice@x.com", Relation: schema.MemberRelationName},
			{Type: "user", Ref: "noemail-id", Relation: schema.AdminRelationName},
			{Type: "user", Ref: "zoe@x.com", Relation: schema.MemberRelationName},
		}, spec)
	})

	t.Run("reconciling an exported document plans no changes", func(t *testing.T) {
		// A user without an email and a service user both export by id, so their ids
		// must be real uuids for the exported document to pass reconcile validation.
		api := &fakePlatformUserAPI{
			users: []*frontierv1beta1.User{
				platformUserPBRelations(t, "alice-id", "alice@x.com", schema.AdminRelationName, schema.MemberRelationName),
				platformUserPBRelations(t, "33333333-3333-3333-3333-333333333333", "", schema.MemberRelationName),
			},
			sus: []*frontierv1beta1.ServiceUser{
				serviceUserPBRelations(t, "44444444-4444-4444-4444-444444444444", schema.MemberRelationName, schema.AdminRelationName),
			},
		}
		registry := map[string]Reconciler{KindPlatformUser: NewPlatformUserReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPlatformUser)
		assert.NoError(t, err)

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("a user with a display-name email round-trips via its address", func(t *testing.T) {
		// The server can store a display-name email. Export normalizes it to the bare
		// address, and re-reconcile matches by address, so the round-trip plans nothing.
		api := &fakePlatformUserAPI{
			users: []*frontierv1beta1.User{
				platformUserPBRelations(t, "55555555-5555-5555-5555-555555555555", "Alice <alice@x.com>", schema.AdminRelationName),
			},
		}
		registry := map[string]Reconciler{KindPlatformUser: NewPlatformUserReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPlatformUser)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "alice@x.com") // the address, normalized
		assert.NotContains(t, string(out), "Alice")    // display name dropped

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("a user whose email is another user's id exports by id and round-trips", func(t *testing.T) {
		// Emails are unvalidated, so a user's email can be another user's id. Export
		// must reference such a user by its own id, not emit the id-shaped email as
		// the ref (which would grant the other user on re-reconcile).
		aID := "aaaaaaaa-1111-1111-1111-111111111111"
		bID := "bbbbbbbb-2222-2222-2222-222222222222"
		api := &fakePlatformUserAPI{
			users: []*frontierv1beta1.User{
				platformUserPBRelations(t, aID, bID, schema.MemberRelationName), // A's email is B's id
				platformUserPBRelations(t, bID, "b@x.com", schema.AdminRelationName),
			},
		}
		registry := map[string]Reconciler{KindPlatformUser: NewPlatformUserReconciler(api, "")}

		out, err := Export(context.Background(), registry, KindPlatformUser)
		assert.NoError(t, err)

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("export of an empty platform yields a document Run accepts", func(t *testing.T) {
		registry := map[string]Reconciler{KindPlatformUser: NewPlatformUserReconciler(&fakePlatformUserAPI{}, "")}

		out, err := Export(context.Background(), registry, KindPlatformUser)
		assert.NoError(t, err)
		assert.Contains(t, string(out), "spec: []")

		reports, err := Run(context.Background(), registry, out, true)
		assert.NoError(t, err)
		if assert.Len(t, reports, 1) {
			assert.Empty(t, reports[0].Planned)
		}
	})

	t.Run("strips the extra relation when the principal holds both (relations list)", func(t *testing.T) {
		// alice is desired as admin only but currently holds admin + member;
		// the reconciler must read both relations and revoke member exactly.
		yamlSpec := []byte("- {type: user, ref: alice@x.com, relation: admin}\n")
		api := &fakePlatformUserAPI{users: []*frontierv1beta1.User{
			platformUserPBRelations(t, "alice-id", "alice@x.com", schema.AdminRelationName, schema.MemberRelationName),
		}}
		rep, err := NewPlatformUserReconciler(api, "").Reconcile(context.Background(), yamlSpec, false)

		assert.NoError(t, err)
		assert.Equal(t, 1, rep.Applied)
		assert.Empty(t, api.added)
		if assert.Len(t, api.removed, 1) {
			assert.Equal(t, "alice-id", api.removed[0].GetUserId())
			assert.Equal(t, schema.MemberRelationName, api.removed[0].GetRelation())
		}
	})
}
