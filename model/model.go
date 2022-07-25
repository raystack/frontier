package model

import (
	"time"
)

type Project struct {
	Id           string
	Name         string
	Slug         string
	Organization Organization
	Metadata     map[string]any
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Organization struct {
	Id        string
	Name      string
	Slug      string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Group struct {
	Id             string
	Name           string
	Slug           string
	Organization   Organization
	OrganizationId string `json:"orgId"`
	Metadata       map[string]any
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Role struct {
	Id          string
	Name        string
	Types       []string
	Namespace   Namespace
	NamespaceId string
	Metadata    map[string]any
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Action struct {
	Id          string
	Name        string
	NamespaceId string
	Namespace   Namespace
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type Namespace struct {
	Id        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Policy struct {
	Id          string
	Role        Role
	RoleId      string `json:"role_id"`
	Namespace   Namespace
	NamespaceId string `json:"namespace_id"`
	Action      Action
	ActionId    string `json:"action_id"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

type User struct {
	Id        string
	Name      string
	Email     string
	Metadata  map[string]any
	CreatedAt time.Time
	UpdatedAt time.Time
}

type PagedUsers struct {
	Count int32
	Users []User
}

type Relation struct {
	Id                 string
	SubjectNamespace   Namespace
	SubjectNamespaceId string `json:"subject_namespace_id"`
	SubjectId          string `json:"subject_id"`
	SubjectRoleId      string `json:"subject_role_id"`
	ObjectNamespace    Namespace
	ObjectNamespaceId  string `json:"object_namespace_id"`
	ObjectId           string `json:"object_id"`
	Role               Role
	RoleId             string       `json:"role_id"`
	RelationType       RelationType `json:"role_type"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Resource struct {
	Idxa           string
	Urn            string
	Name           string
	ProjectId      string `json:"project_id"`
	Project        Project
	GroupId        string `json:"group_id"`
	Group          Group
	OrganizationId string `json:"organization_id"`
	Organization   Organization
	NamespaceId    string `json:"namespace_id"`
	Namespace      Namespace
	User           User
	UserId         string `json:"user_id"`
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type ResourceFilters struct {
	ProjectId      string `json:"project_id"`
	GroupId        string `json:"group_id"`
	OrganizationId string `json:"org_id"`
	NamespaceId    string `json:"namespace_id"`
}

type Permission struct {
	Name string
}

type RelationType string

var RelationTypes = struct {
	Role      RelationType
	Namespace RelationType
}{
	Role:      "role",
	Namespace: "namespace",
}
