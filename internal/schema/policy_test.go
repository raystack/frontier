package schema

import (
	"github.com/odpf/shield/model"
)

func (s *ServiceTestSuite) TestGenerateSchema_ShouldReturnValidSchema() {
	policies := []model.Policy{
		{
			Namespace: model.Namespace{Name: "Project", Id: "project"},
			Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
			Action:    model.Action{Name: "Read", Id: "read"},
		},
		{
			Namespace: model.Namespace{Name: "Project", Id: "project"},
			Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
			Action:    model.Action{Name: "Write", Id: "write"},
		},
		{
			Namespace: model.Namespace{Name: "Project", Id: "project"},
			Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User"}},
			Action:    model.Action{Name: "Delete", Id: "delete"},
		},
	}
	expectedSchema := []string{
		"definition project {\n\trelation admin: User\n\tpermission read = admin\n\tpermission write = admin\n\tpermission delete = admin\n}",
		"definition user {}",
	}

	generatedSchema, err := s.service.generateSchema(policies)

	s.Nil(err)
	s.Equal(expectedSchema, generatedSchema)
}

func (s *ServiceTestSuite) TestGenerateSchema_ShouldReturnError_WhenPolicyIsInvalid() {
	policies := []model.Policy{
		{
			Namespace: model.Namespace{Name: "Project", Id: "project"},
			Role:      model.Role{Name: "Admin", Id: "admin", Types: []string{"User", "Team#members"}},
			Action:    model.Action{Name: "Read", Id: "read", NamespaceId: "org"},
		},
	}
	expectedSchema := make([]string, 0)
	generatedSchema, err := s.service.generateSchema(policies)

	s.NotNil(err)
	s.Equal(expectedSchema, generatedSchema)
}
