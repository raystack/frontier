package v1beta1

import (
	"context"
	"reflect"
	"testing"

	"github.com/odpf/shield/core/resource"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

func TestHandler_ListResources(t *testing.T) {
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
		request *shieldv1beta1.ListResourcesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.ListResourcesResponse
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
			got, err := h.ListResources(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.ListResources() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.ListResources() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_CreateResource(t *testing.T) {
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
		request *shieldv1beta1.CreateResourceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.CreateResourceResponse
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
			got, err := h.CreateResource(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.CreateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.CreateResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_GetResource(t *testing.T) {
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
		request *shieldv1beta1.GetResourceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.GetResourceResponse
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
			got, err := h.GetResource(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.GetResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.GetResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHandler_UpdateResource(t *testing.T) {
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
		request *shieldv1beta1.UpdateResourceRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *shieldv1beta1.UpdateResourceResponse
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
			got, err := h.UpdateResource(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler.UpdateResource() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Handler.UpdateResource() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_transformResourceToPB(t *testing.T) {
	type args struct {
		from resource.Resource
	}
	tests := []struct {
		name    string
		args    args
		want    shieldv1beta1.Resource
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := transformResourceToPB(tt.args.from)
			if (err != nil) != tt.wantErr {
				t.Errorf("transformResourceToPB() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("transformResourceToPB() = %v, want %v", got, tt.want)
			}
		})
	}
}
