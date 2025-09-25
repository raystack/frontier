package v1beta1connect

import (
	"time"

	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/metadata"
)

var (
	testUserID  = "9f256f86-31a3-11ec-8d3d-0242ac130003"
	testUserMap = map[string]user.User{
		"9f256f86-31a3-11ec-8d3d-0242ac130003": {
			ID:    "9f256f86-31a3-11ec-8d3d-0242ac130003",
			Title: "User 1",
			Name:  "user1",
			Email: "test@test.com",
			Metadata: metadata.Metadata{
				"foo":    "bar",
				"age":    21,
				"intern": true,
			},
			CreatedAt: time.Time{},
			UpdatedAt: time.Time{},
		},
	}
)
