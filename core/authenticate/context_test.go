package authenticate_test

import (
	"testing"

	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	patModels "github.com/raystack/frontier/core/userpat/models"
	"github.com/stretchr/testify/assert"
)

func TestGetPrincipalNameAndTitle(t *testing.T) {
	tests := []struct {
		name          string
		principal     *authenticate.Principal
		expectedName  string
		expectedTitle string
	}{
		{
			name:          "nil principal returns empty",
			principal:     nil,
			expectedName:  "",
			expectedTitle: "",
		},
		{
			name: "user principal returns user name and title",
			principal: &authenticate.Principal{
				ID:   "user-123",
				Type: "app/user",
				User: &user.User{
					ID:    "user-123",
					Name:  "John Doe",
					Title: "Engineer",
				},
			},
			expectedName:  "John Doe",
			expectedTitle: "Engineer",
		},
		{
			name: "service user principal returns service user title",
			principal: &authenticate.Principal{
				ID:   "su-123",
				Type: "app/serviceuser",
				ServiceUser: &serviceuser.ServiceUser{
					ID:    "su-123",
					Title: "API Bot",
				},
			},
			expectedName:  "",
			expectedTitle: "API Bot",
		},
		{
			name: "PAT principal returns PAT title",
			principal: &authenticate.Principal{
				ID:   "pat-123",
				Type: "app/pat",
				PAT: &patModels.PAT{
					ID:     "pat-123",
					UserID: "user-456",
					Title:  "My Deploy Token",
				},
				User: &user.User{
					ID:    "user-456",
					Name:  "Jane Doe",
					Title: "DevOps",
				},
			},
			expectedName:  "",
			expectedTitle: "My Deploy Token",
		},
		{
			name: "PAT principal takes precedence over user",
			principal: &authenticate.Principal{
				ID:   "pat-789",
				Type: "app/pat",
				PAT: &patModels.PAT{
					ID:    "pat-789",
					Title: "CI Token",
				},
				User: &user.User{
					ID:    "user-owner",
					Name:  "Owner Name",
					Title: "Owner Title",
				},
			},
			expectedName:  "",
			expectedTitle: "CI Token",
		},
		{
			name: "principal with no user, serviceuser, or PAT returns empty",
			principal: &authenticate.Principal{
				ID:   "unknown-123",
				Type: "app/unknown",
			},
			expectedName:  "",
			expectedTitle: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, title := authenticate.GetPrincipalNameAndTitle(tt.principal)
			assert.Equal(t, tt.expectedName, name)
			assert.Equal(t, tt.expectedTitle, title)
		})
	}
}
