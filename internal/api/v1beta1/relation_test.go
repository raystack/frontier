package v1beta1

import (
	"context"
	"reflect"
	"testing"

	"github.com/odpf/shield/core/relation"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

func TestHandler_ListRelations(t *testing.T) {
	type fields struct {
		UnimplementedShieldServiceServer shieldv1beta1.UnimplementedShieldServiceServer
		orgService                       OrganizationService
		projectService                   ProjectService
		groupService                     GroupService
		roleService                      RoleService
		policyService                    PolicyService
		userService                      UserService
		namespaceService                 NamespaceService
		actionService                    ActionService
		relationService                  RelationService
		resourceService                  ResourceService
		ruleService                      RuleService
	}
	type args struct {
		ctx     context.Context
		request *shieldv1beta1.ListRelationsRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.ListRelationsResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handler{
				UnimplementedShieldServiceServer: tt.fields.UnimplementedShieldServiceServer,
				orgService:                       tt.fields.orgService,
				projectService:                   tt.fields.projectService,
				groupService:                     tt.fields.groupService,
				roleService:                      tt.fields.roleService,
				policyService:                    tt.fields.policyService,
				userService:                      tt.fields.userService,
				namespaceService:                 tt.fields.namespaceService,
				actionService:                    tt.fields.actionService,
				relationService:                  tt.fields.relationService,
				resourceService:                  tt.fields.resourceService,
				ruleService:                      tt.fields.ruleService,
			}
			got, err := h.ListRelations(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.ListRelations() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.ListRelations() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_CreateRelation(t *testing.T) {
	type fields struct {
		UnimplementedShieldServiceServer shieldv1beta1.UnimplementedShieldServiceServer
		orgService                       OrganizationService
		projectService                   ProjectService
		groupService                     GroupService
		roleService                      RoleService
		policyService                    PolicyService
		userService                      UserService
		namespaceService                 NamespaceService
		actionService                    ActionService
		relationService                  RelationService
		resourceService                  ResourceService
		ruleService                      RuleService
	}
	type args struct {
		ctx     context.Context
		request *shieldv1beta1.CreateRelationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.CreateRelationResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handler{
				UnimplementedShieldServiceServer: tt.fields.UnimplementedShieldServiceServer,
				orgService:                       tt.fields.orgService,
				projectService:                   tt.fields.projectService,
				groupService:                     tt.fields.groupService,
				roleService:                      tt.fields.roleService,
				policyService:                    tt.fields.policyService,
				userService:                      tt.fields.userService,
				namespaceService:                 tt.fields.namespaceService,
				actionService:                    tt.fields.actionService,
				relationService:                  tt.fields.relationService,
				resourceService:                  tt.fields.resourceService,
				ruleService:                      tt.fields.ruleService,
			}
			got, err := h.CreateRelation(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.CreateRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.CreateRelation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_GetRelation(t *testing.T) {
	type fields struct {
		UnimplementedShieldServiceServer shieldv1beta1.UnimplementedShieldServiceServer
		orgService                       OrganizationService
		projectService                   ProjectService
		groupService                     GroupService
		roleService                      RoleService
		policyService                    PolicyService
		userService                      UserService
		namespaceService                 NamespaceService
		actionService                    ActionService
		relationService                  RelationService
		resourceService                  ResourceService
		ruleService                      RuleService
	}
	type args struct {
		ctx     context.Context
		request *shieldv1beta1.GetRelationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.GetRelationResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handler{
				UnimplementedShieldServiceServer: tt.fields.UnimplementedShieldServiceServer,
				orgService:                       tt.fields.orgService,
				projectService:                   tt.fields.projectService,
				groupService:                     tt.fields.groupService,
				roleService:                      tt.fields.roleService,
				policyService:                    tt.fields.policyService,
				userService:                      tt.fields.userService,
				namespaceService:                 tt.fields.namespaceService,
				actionService:                    tt.fields.actionService,
				relationService:                  tt.fields.relationService,
				resourceService:                  tt.fields.resourceService,
				ruleService:                      tt.fields.ruleService,
			}
			got, err := h.GetRelation(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.GetRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.GetRelation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_UpdateRelation(t *testing.T) {
	type fields struct {
		UnimplementedShieldServiceServer shieldv1beta1.UnimplementedShieldServiceServer
		orgService                       OrganizationService
		projectService                   ProjectService
		groupService                     GroupService
		roleService                      RoleService
		policyService                    PolicyService
		userService                      UserService
		namespaceService                 NamespaceService
		actionService                    ActionService
		relationService                  RelationService
		resourceService                  ResourceService
		ruleService                      RuleService
	}
	type args struct {
		ctx     context.Context
		request *shieldv1beta1.UpdateRelationRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.UpdateRelationResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := Handler{
				UnimplementedShieldServiceServer: tt.fields.UnimplementedShieldServiceServer,
				orgService:                       tt.fields.orgService,
				projectService:                   tt.fields.projectService,
				groupService:                     tt.fields.groupService,
				roleService:                      tt.fields.roleService,
				policyService:                    tt.fields.policyService,
				userService:                      tt.fields.userService,
				namespaceService:                 tt.fields.namespaceService,
				actionService:                    tt.fields.actionService,
				relationService:                  tt.fields.relationService,
				resourceService:                  tt.fields.resourceService,
				ruleService:                      tt.fields.ruleService,
			}
			got, err := h.UpdateRelation(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.UpdateRelation() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.UpdateRelation() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transformRelationToPB(t *testing.T) {
	type args struct {
		relation relation.Relation
	}
	tests := []struct {
		name    string
		args    args
		want    shieldv1beta1.Relation
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformRelationToPB(tt.args.relation)
			if (err != nil) != tt.wantErr {
				t.Errorf("transformRelationToPB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transformRelationToPB() = %v, want %v", got, tt.want)
			}
		})
	}
}
