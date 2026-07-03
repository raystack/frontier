package v1beta1connect

import (
	"context"
	"fmt"

	"connectrpc.com/connect"

	"log/slog"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/core/avatar"
	"github.com/raystack/frontier/core/membership"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/role"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/errors"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/pagination"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	orgMetaSchema = "organization"
)

func (h *ConnectHandler) GetOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.GetOrganizationRequest]) (*connect.Response[frontierv1beta1.GetOrganizationResponse], error) {
	fetchedOrg, err := h.orgService.GetRaw(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, err)
		case errors.Is(err, organization.ErrInvalidUUID):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetOrganization.GetRaw: org_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	orgPB, err := transformOrgToPB(fetchedOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("GetOrganization: entity_id=%s: %w", fetchedOrg.ID, err))
	}

	return connect.NewResponse(&frontierv1beta1.GetOrganizationResponse{
		Organization: orgPB,
	}), nil
}

func (h *ConnectHandler) ListOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationsResponse], error) {
	orgs, _, err := h.searchOrgs(ctx, request, request.Msg.GetUserId(), request.Msg.GetState(), request.Msg.GetPageNum(), request.Msg.GetPageSize(), "ListOrganizations")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.ListOrganizationsResponse{
		Organizations: orgs,
	}), nil
}

func (h *ConnectHandler) ListAllOrganizations(ctx context.Context, request *connect.Request[frontierv1beta1.ListAllOrganizationsRequest]) (*connect.Response[frontierv1beta1.ListAllOrganizationsResponse], error) {
	orgs, count, err := h.searchOrgs(ctx, request, request.Msg.GetUserId(), request.Msg.GetState(), request.Msg.GetPageNum(), request.Msg.GetPageSize(), "ListAllOrganizations")
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&frontierv1beta1.ListAllOrganizationsResponse{
		Organizations: orgs,
		Count:         count,
	}), nil
}

func (h *ConnectHandler) searchOrgs(ctx context.Context, req connect.AnyRequest, userID, stateStr string, pageNum, pageSize int32, rpcName string) ([]*frontierv1beta1.Organization, int32, error) {
	paginate := pagination.NewPagination(pageNum, pageSize)
	filter := organization.Filter{
		State:      organization.State(stateStr),
		Pagination: paginate,
	}
	if userID != "" {
		filter.Principal = &authenticate.Principal{ID: userID, Type: schema.UserPrincipal}
	}

	orgList, err := h.orgService.List(ctx, filter)
	if err != nil {
		return nil, 0, connect.NewError(connect.CodeInternal, fmt.Errorf("%s.List: state=%s user_id=%s: %w", rpcName, stateStr, userID, err))
	}

	orgs := make([]*frontierv1beta1.Organization, 0, len(orgList))
	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			return nil, 0, connect.NewError(connect.CodeInternal, fmt.Errorf("%s: entity_id=%s: %w", rpcName, v.ID, err))
		}
		orgs = append(orgs, orgPB)
	}
	return orgs, paginate.Count, nil
}

func (h *ConnectHandler) CreateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.CreateOrganizationRequest]) (*connect.Response[frontierv1beta1.CreateOrganizationResponse], error) {
	errorLogger := NewErrorLogger()

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newOrg, err := h.orgService.Create(ctx, organization.Organization{
		Name:     request.Msg.GetBody().GetName(),
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	})
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeUnauthenticated, ErrUnauthenticated)
		case errors.Is(err, organization.ErrUserPrincipalOnly):
			return nil, connect.NewError(connect.CodePermissionDenied, err)
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		case errors.Is(err, relation.ErrSubjectNotAllowed):
			errorLogger.LogServiceError(ctx, request, "CreateOrganization.Create", err,
				"org_name", request.Msg.GetBody().GetName(),
				"org_title", request.Msg.GetBody().GetTitle())
			return nil, connect.NewError(connect.CodePermissionDenied, ErrUnauthorized)
		case errors.Is(err, avatar.ErrInvalid):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateOrganization.Create: org_name=%s org_title=%s: %w", request.Msg.GetBody().GetName(), request.Msg.GetBody().GetTitle(), err))
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("CreateOrganization: entity_id=%s: %w", newOrg.ID, err))
	}

	if err := audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	}); err != nil {
		errorLogger.LogServiceError(ctx, request, "CreateOrganization.AuditLog", err,
			"org_id", newOrg.ID)
	}
	return connect.NewResponse(&frontierv1beta1.CreateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) AdminCreateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.AdminCreateOrganizationRequest]) (*connect.Response[frontierv1beta1.AdminCreateOrganizationResponse], error) {
	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	newOrg, err := h.orgService.AdminCreate(ctx, organization.Organization{
		Name:     request.Msg.GetBody().GetName(),
		Title:    request.Msg.GetBody().GetTitle(),
		Avatar:   request.Msg.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	}, request.Msg.GetBody().GetOrgOwnerEmail())
	if err != nil {
		switch {
		case errors.Is(err, user.ErrInvalidEmail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrInvalidDetail):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		case errors.Is(err, avatar.ErrInvalid):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AdminCreateOrganization.AdminCreate: org_name=%s org_title=%s owner_email=%s: %w", request.Msg.GetBody().GetName(), request.Msg.GetBody().GetTitle(), request.Msg.GetBody().GetOrgOwnerEmail(), err))
		}
	}

	orgPB, err := transformOrgToPB(newOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("AdminCreateOrganization: entity_id=%s: %w", newOrg.ID, err))
	}

	audit.GetAuditor(ctx, newOrg.ID).LogWithAttrs(audit.OrgCreatedEvent, audit.OrgTarget(newOrg.ID), map[string]string{
		"title": newOrg.Title,
		"name":  newOrg.Name,
	})
	return connect.NewResponse(&frontierv1beta1.AdminCreateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) UpdateOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.UpdateOrganizationRequest]) (*connect.Response[frontierv1beta1.UpdateOrganizationResponse], error) {
	if request.Msg.GetBody() == nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadRequest)
	}

	metaDataMap := metadata.Build(request.Msg.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, orgMetaSchema); err != nil {
		return nil, connect.NewError(connect.CodeInvalidArgument, ErrBadBodyMetaSchemaError)
	}

	var updatedOrg organization.Organization
	var err error
	if utils.IsValidUUID(request.Msg.GetId()) {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			ID:       request.Msg.GetId(),
			Name:     request.Msg.GetBody().GetName(),
			Title:    request.Msg.GetBody().GetTitle(),
			Avatar:   request.Msg.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	} else {
		updatedOrg, err = h.orgService.Update(ctx, organization.Organization{
			Name:     request.Msg.GetBody().GetName(),
			Title:    request.Msg.GetBody().GetTitle(),
			Avatar:   request.Msg.GetBody().GetAvatar(),
			Metadata: metaDataMap,
		})
	}
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrNotExist), errors.Is(err, organization.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, organization.ErrConflict):
			return nil, connect.NewError(connect.CodeAlreadyExists, ErrConflictRequest)
		case errors.Is(err, avatar.ErrInvalid):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateOrganization.Update: org_id=%s org_name=%s org_title=%s: %w", request.Msg.GetId(), request.Msg.GetBody().GetName(), request.Msg.GetBody().GetTitle(), err))
		}
	}

	orgPB, err := transformOrgToPB(updatedOrg)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("UpdateOrganization: entity_id=%s: %w", updatedOrg.ID, err))
	}

	audit.GetAuditor(ctx, updatedOrg.ID).Log(audit.OrgUpdatedEvent, audit.OrgTarget(updatedOrg.ID))
	return connect.NewResponse(&frontierv1beta1.UpdateOrganizationResponse{Organization: orgPB}), nil
}

func (h *ConnectHandler) ListOrganizationProjects(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationProjectsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationProjectsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationProjects.Get: org_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	projects, err := h.projectService.List(ctx, project.Filter{
		OrgID:           orgResp.ID,
		WithMemberCount: request.Msg.GetWithMemberCount(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationProjects.List: org_id=%s with_member_count=%v: %w", orgResp.ID, request.Msg.GetWithMemberCount(), err))
	}

	var projectPB []*frontierv1beta1.Project
	for _, rel := range projects {
		u, err := transformProjectToPB(rel)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationProjects: entity_id=%s: %w", rel.ID, err))
		}

		projectPB = append(projectPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationProjectsResponse{Projects: projectPB}), nil
}

func (h *ConnectHandler) ListOrganizationAdmins(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationAdminsRequest]) (*connect.Response[frontierv1beta1.ListOrganizationAdminsResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationAdmins.Get: org_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	ownerRole, err := h.roleService.Get(ctx, organization.AdminRole)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationAdmins.roleService.Get: org_id=%s role=%s: %w", orgResp.ID, organization.AdminRole, err))
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, orgResp.ID, schema.OrganizationNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
		RoleIDs:       []string{ownerRole.ID},
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationAdmins.ListPrincipalsByResource: org_id=%s: %w", orgResp.ID, err))
	}

	adminIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	admins, err := h.userService.GetByIDs(ctx, adminIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationAdmins.GetByIDs: org_id=%s: %w", orgResp.ID, err))
	}

	var adminsPB []*frontierv1beta1.User
	for _, user := range admins {
		u, err := transformUserToPB(user)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationAdmins: entity_id=%s: %w", user.ID, err))
		}

		adminsPB = append(adminsPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationAdminsResponse{Users: adminsPB}), nil
}

func (h *ConnectHandler) ListOrganizationUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationUsersRequest]) (*connect.Response[frontierv1beta1.ListOrganizationUsersResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationUsers.Get: org_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	members, err := h.membershipService.ListPrincipalsByResource(ctx, orgResp.ID, schema.OrganizationNamespace, membership.MemberFilter{
		PrincipalType: schema.UserPrincipal,
		RoleIDs:       request.Msg.GetRoleIds(),
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationUsers.ListPrincipalsByResource: org_id=%s role_ids=%v: %w", orgResp.ID, request.Msg.GetRoleIds(), err))
	}

	userIDs := utils.Map(members, func(m membership.Member) string { return m.PrincipalID })
	users, err := h.userService.GetByIDs(ctx, userIDs)
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationUsers.GetByIDs: org_id=%s: %w", orgResp.ID, err))
	}

	var rolePairPBs []*frontierv1beta1.ListOrganizationUsersResponse_RolePair
	for _, m := range members {
		rolesPb := utils.Filter(utils.Map(m.Roles, func(r role.Role) *frontierv1beta1.Role {
			pb, err := transformRoleToPB(r)
			if err != nil {
				slog.ErrorContext(ctx, "failed to transform role for user", "error", err)
				return nil
			}
			return &pb
		}), func(r *frontierv1beta1.Role) bool {
			return r != nil
		})
		rolePairPBs = append(rolePairPBs, &frontierv1beta1.ListOrganizationUsersResponse_RolePair{
			UserId: m.PrincipalID,
			Roles:  rolesPb,
		})
	}

	var usersPB []*frontierv1beta1.User
	for _, rel := range users {
		u, err := transformUserToPB(rel)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationUsers: entity_id=%s: %w", rel.ID, err))
		}
		usersPB = append(usersPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationUsersResponse{
		Users:     usersPB,
		RolePairs: rolePairPBs,
	}), nil
}

func (h *ConnectHandler) RemoveOrganizationMember(ctx context.Context, request *connect.Request[frontierv1beta1.RemoveOrganizationMemberRequest]) (*connect.Response[frontierv1beta1.RemoveOrganizationMemberResponse], error) {
	orgID := request.Msg.GetOrgId()
	principalID := request.Msg.GetPrincipalId()
	principalType := request.Msg.GetPrincipalType()

	// service users are bound to a single org — use DeleteServiceUser instead
	if principalType == schema.ServiceUserPrincipal {
		return nil, connect.NewError(connect.CodeInvalidArgument,
			errors.New("cannot remove service user from org, use DeleteServiceUser instead"))
	}

	if err := h.membershipService.RemoveOrganizationMember(ctx, orgID, principalID, principalType); err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, membership.ErrInvalidPrincipalType):
			return nil, connect.NewError(connect.CodeInvalidArgument, err)
		case errors.Is(err, membership.ErrNotMember):
			return nil, connect.NewError(connect.CodeFailedPrecondition, membership.ErrNotMember)
		case errors.Is(err, membership.ErrLastOwnerRole):
			return nil, connect.NewError(connect.CodeFailedPrecondition, membership.ErrLastOwnerRole)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("RemoveOrganizationMember: org_id=%s principal_id=%s principal_type=%s: %w", orgID, principalID, principalType, err))
		}
	}

	return connect.NewResponse(&frontierv1beta1.RemoveOrganizationMemberResponse{}), nil
}

func (h *ConnectHandler) SetOrganizationMemberRole(ctx context.Context, request *connect.Request[frontierv1beta1.SetOrganizationMemberRoleRequest]) (*connect.Response[frontierv1beta1.SetOrganizationMemberRoleResponse], error) {
	errorLogger := NewErrorLogger()

	orgID := request.Msg.GetOrgId()
	userID := request.Msg.GetUserId()
	roleID := request.Msg.GetRoleId()

	if err := h.membershipService.SetOrganizationMemberRole(ctx, orgID, userID, schema.UserPrincipal, roleID); err != nil {
		errorLogger.LogServiceError(ctx, request, "SetOrganizationMemberRole", err,
			"org_id", orgID,
			"user_id", userID,
			"role_id", roleID)

		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		case errors.Is(err, user.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrUserNotExist)
		case errors.Is(err, user.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, err)
		case errors.Is(err, membership.ErrNotMember):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrNotMember)
		case errors.Is(err, role.ErrNotExist), errors.Is(err, role.ErrInvalidID):
			return nil, connect.NewError(connect.CodeNotFound, ErrInvalidRoleID)
		case errors.Is(err, membership.ErrInvalidOrgRole):
			return nil, connect.NewError(connect.CodeInvalidArgument, ErrInvalidOrgRole)
		case errors.Is(err, membership.ErrLastOwnerRole):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrLastOwnerRole)
		default:
			return nil, connect.NewError(connect.CodeInternal, err)
		}
	}

	return connect.NewResponse(&frontierv1beta1.SetOrganizationMemberRoleResponse{}), nil
}

func (h *ConnectHandler) AddOrganizationMembers(ctx context.Context, request *connect.Request[frontierv1beta1.AddOrganizationMembersRequest]) (*connect.Response[frontierv1beta1.AddOrganizationMembersResponse], error) {
	errorLogger := NewErrorLogger()
	orgID := request.Msg.GetOrgId()

	var results []*frontierv1beta1.OrgMemberResult
	for _, member := range request.Msg.GetMembers() {
		result := &frontierv1beta1.OrgMemberResult{
			UserId: member.GetUserId(),
			RoleId: member.GetRoleId(),
		}

		if err := h.membershipService.AddOrganizationMember(ctx, orgID, member.GetUserId(), schema.UserPrincipal, member.GetRoleId()); err != nil {
			result.Success = false
			result.Error = toClientError(err)
			if !isDomainError(err) {
				errorLogger.LogServiceError(ctx, request, "AddOrganizationMembers", err,
					"org_id", orgID,
					"user_id", member.GetUserId(),
					"role_id", member.GetRoleId())
			}
		} else {
			result.Success = true
		}

		results = append(results, result)
	}

	return connect.NewResponse(&frontierv1beta1.AddOrganizationMembersResponse{
		Results: results,
	}), nil
}

// isDomainError returns true if the error is a known domain error safe to expose to clients.
func isDomainError(err error) bool {
	knownErrors := []error{
		membership.ErrAlreadyMember,
		membership.ErrInvalidOrgRole,
		organization.ErrNotExist,
		organization.ErrDisabled,
		user.ErrNotExist,
		user.ErrDisabled,
		role.ErrNotExist,
		role.ErrInvalidID,
	}
	for _, known := range knownErrors {
		if errors.Is(err, known) {
			return true
		}
	}
	return false
}

// toClientError returns a client-safe error message.
func toClientError(err error) string {
	if isDomainError(err) {
		return err.Error()
	}
	return ErrInternalServerError.Error()
}

func (h *ConnectHandler) EnableOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.EnableOrganizationRequest]) (*connect.Response[frontierv1beta1.EnableOrganizationResponse], error) {
	if err := h.orgService.Enable(ctx, request.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("EnableOrganization.Enable: org_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.EnableOrganizationResponse{}), nil
}

func (h *ConnectHandler) DisableOrganization(ctx context.Context, request *connect.Request[frontierv1beta1.DisableOrganizationRequest]) (*connect.Response[frontierv1beta1.DisableOrganizationResponse], error) {
	if err := h.orgService.Disable(ctx, request.Msg.GetId()); err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("DisableOrganization.Disable: org_id=%s: %w", request.Msg.GetId(), err))
	}
	return connect.NewResponse(&frontierv1beta1.DisableOrganizationResponse{}), nil
}

func (h *ConnectHandler) ListOrganizationServiceUsers(ctx context.Context, request *connect.Request[frontierv1beta1.ListOrganizationServiceUsersRequest]) (*connect.Response[frontierv1beta1.ListOrganizationServiceUsersResponse], error) {
	orgResp, err := h.orgService.Get(ctx, request.Msg.GetId())
	if err != nil {
		switch {
		case errors.Is(err, organization.ErrDisabled):
			return nil, connect.NewError(connect.CodeFailedPrecondition, ErrOrgDisabled)
		case errors.Is(err, organization.ErrNotExist):
			return nil, connect.NewError(connect.CodeNotFound, ErrNotFound)
		default:
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationServiceUsers.Get: org_id=%s: %w", request.Msg.GetId(), err))
		}
	}

	usersList, err := h.serviceUserService.List(ctx, serviceuser.Filter{
		OrgID: orgResp.ID,
	})
	if err != nil {
		return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationServiceUsers.List: org_id=%s: %w", orgResp.ID, err))
	}

	var usersPB []*frontierv1beta1.ServiceUser
	for _, rel := range usersList {
		u, err := transformServiceUserToPB(rel)
		if err != nil {
			return nil, connect.NewError(connect.CodeInternal, fmt.Errorf("ListOrganizationServiceUsers: entity_id=%s: %w", rel.ID, err))
		}

		usersPB = append(usersPB, u)
	}

	return connect.NewResponse(&frontierv1beta1.ListOrganizationServiceUsersResponse{Serviceusers: usersPB}), nil
}

func transformOrgToPB(org organization.Organization) (*frontierv1beta1.Organization, error) {
	metaData, err := org.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.Organization{
		Id:        org.ID,
		Name:      org.Name,
		Title:     org.Title,
		Metadata:  metaData,
		State:     org.State.String(),
		Avatar:    org.Avatar,
		CreatedAt: timestamppb.New(org.CreatedAt),
		UpdatedAt: timestamppb.New(org.UpdatedAt),
	}, nil
}
