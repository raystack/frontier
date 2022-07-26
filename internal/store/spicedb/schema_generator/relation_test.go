package schema_generator

import (
	"testing"

	v1 "github.com/authzed/authzed-go/proto/authzed/api/v1"
	"github.com/authzed/spicedb/pkg/tuple"
	"github.com/odpf/shield/core/relation"
	shieldRole "github.com/odpf/shield/core/role"
	"github.com/stretchr/testify/assert"
)

func TestTransformRelation(t *testing.T) {
	t.Run("should generate empty tuple from relation model", func(t *testing.T) {
		input := relation.Relation{}
		output, _ := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "",
				ObjectType: "",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "",
					ObjectType: "",
				},
				OptionalRelation: "",
			},
			Relation: "",
		}

		relString := tuple.RelString(output)
		expectedString := ":#@:"
		assert.EqualValues(t, expected, output)
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should generate tuple from relation model", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "team",
			ObjectID:           "team_1",
			SubjectNamespaceID: "user",
			SubjectID:          "user_1",
			Role: shieldRole.Role{
				ID:          "team_member",
				NamespaceID: "team",
			},
		}
		output, _ := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "team_1",
				ObjectType: "team",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "user_1",
					ObjectType: "user",
				},
				OptionalRelation: "",
			},
			Relation: "team_member",
		}

		relString := tuple.RelString(output)
		expectedString := "team:team_1#team_member@user:user_1"
		assert.EqualValues(t, expected, output)
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should generate tuple from relation model", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "project",
			ObjectID:           "project_1",
			SubjectNamespaceID: "team",
			SubjectID:          "team_1",
			Role: shieldRole.Role{
				ID:          "editor",
				NamespaceID: "project",
			},
		}
		output, _ := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "project_1",
				ObjectType: "project",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "team_1",
					ObjectType: "team",
				},
				OptionalRelation: "",
			},
			Relation: "editor",
		}

		relString := tuple.RelString(output)
		expectedString := "project:project_1#editor@team:team_1"
		assert.EqualValues(t, expected, output)
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should should throw error if role doesnt exist in object", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "project",
			ObjectID:           "project_1",
			SubjectNamespaceID: "team",
			SubjectID:          "team_1",
			Role: shieldRole.Role{
				ID:          "editor",
				NamespaceID: "org",
			},
		}
		output, err := TransformRelation(input)
		expected := &v1.Relationship{}
		assert.EqualValues(t, expected, output)
		assert.EqualError(t, err, "Role editor doesnt exist in project")
	})

	t.Run("should add org to team", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "team",
			ObjectID:           "team_1",
			SubjectNamespaceID: "organization",
			SubjectID:          "org_1",
			Role: shieldRole.Role{
				ID:          "organization",
				NamespaceID: "team",
			},
		}
		output, err := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "team_1",
				ObjectType: "team",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "org_1",
					ObjectType: "organization",
				},
				OptionalRelation: "",
			},
			Relation: "organization",
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)

		relString := tuple.RelString(output)
		expectedString := "team:team_1#organization@organization:org_1"
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should team to resource", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "resource/dagger",
			ObjectID:           "dagger_1",
			SubjectNamespaceID: "team",
			SubjectID:          "team_1",
			Role: shieldRole.Role{
				ID:          "team",
				NamespaceID: "resource/dagger",
			},
		}
		output, err := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "dagger_1",
				ObjectType: "resource/dagger",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "team_1",
					ObjectType: "team",
				},
				OptionalRelation: "",
			},
			Relation: "team",
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)

		relString := tuple.RelString(output)
		expectedString := "resource/dagger:dagger_1#team@team:team_1"
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should project to resource", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "resource/dagger",
			ObjectID:           "dagger_1",
			SubjectNamespaceID: "project",
			SubjectID:          "project_1",
			Role: shieldRole.Role{
				ID:          "project",
				NamespaceID: "resource/dagger",
			},
		}
		output, err := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "dagger_1",
				ObjectType: "resource/dagger",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "project_1",
					ObjectType: "project",
				},
				OptionalRelation: "",
			},
			Relation: "project",
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)

		relString := tuple.RelString(output)
		expectedString := "resource/dagger:dagger_1#project@project:project_1"
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should editor role to team members", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "resource/dagger",
			ObjectID:           "dagger_1",
			SubjectNamespaceID: "team",
			SubjectID:          "team_1",
			SubjectRoleID:      "team_member",
			Role: shieldRole.Role{
				ID:          "editor",
				NamespaceID: "resource/dagger",
			},
		}
		output, err := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "dagger_1",
				ObjectType: "resource/dagger",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "team_1",
					ObjectType: "team",
				},
				OptionalRelation: "team_member",
			},
			Relation: "editor",
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)

		relString := tuple.RelString(output)
		expectedString := "resource/dagger:dagger_1#editor@team:team_1#team_member"
		assert.Equal(t, expectedString, relString)
	})

	t.Run("should admin to team", func(t *testing.T) {
		input := relation.Relation{
			ObjectNamespaceID:  "team",
			ObjectID:           "team_1",
			SubjectNamespaceID: "user",
			SubjectID:          "user_1",
			Role: shieldRole.Role{
				ID:          "team_admin",
				NamespaceID: "team",
			},
		}
		output, err := TransformRelation(input)
		expected := &v1.Relationship{
			Resource: &v1.ObjectReference{
				ObjectId:   "team_1",
				ObjectType: "team",
			},
			Subject: &v1.SubjectReference{
				Object: &v1.ObjectReference{
					ObjectId:   "user_1",
					ObjectType: "user",
				},
				OptionalRelation: "",
			},
			Relation: "team_admin",
		}
		assert.EqualValues(t, expected, output)
		assert.NoError(t, err)

		relString := tuple.RelString(output)
		expectedString := "team:team_1#team_admin@user:user_1"
		assert.Equal(t, expectedString, relString)
	})
}
