package model

import "time"

type Project struct {
	Id           string
	Name         string
	Slug         string
	Organization Organization
	Metadata     map[string]string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type Organization struct {
	Id        string
	Name      string
	Slug      string
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Group struct {
	Id             string
	Name           string
	Slug           string
	Organization   Organization
	OrganizationId string `json:"orgId"`
	Metadata       map[string]string
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type Role struct {
	Id          string
	Name        string
	Types       []string
	Namespace   Namespace
	NamespaceId string
	Metadata    map[string]string
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
	Metadata  map[string]string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Relation struct {
	Id                 string
	SubjectNamespace   Namespace
	SubjectNamespaceId string `json:"subject_namespace_id"`
	SubjectId          string `json:"subject_id"`
	ObjectNamespace    Namespace
	ObjectNamespaceId  string `json:"object_namespace_id"`
	ObjectId           string `json:"object_id"`
	Role               Role
	RoleId             string `json:"role_id"`
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

type Resource struct {
	Id             string
	Name           string
	ProjectId      string `json:"project_id"`
	Project        Project
	GroupId        string `json:"group_id"`
	Group          Group
	OrganizationId string `json:"organization_id"`
	Organization   Organization
	NamespaceId    string `json:"namespace_id"`
	Namespace      Namespace
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
