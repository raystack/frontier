package v1beta1

import (
	"context"
	"net/mail"
	"strings"

	"github.com/raystack/frontier/core/organization"

	"github.com/raystack/frontier/core/project"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/pkg/utils"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/authenticate"
	"github.com/raystack/frontier/internal/bootstrap/schema"

	"github.com/pkg/errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/raystack/frontier/core/group"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/str"
	"github.com/raystack/frontier/pkg/telemetry"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
)

var grpcUserNotFoundError = status.Errorf(codes.NotFound, "user doesn't exist")

type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(ctx context.Context, user user.User) (user.User, error)
	List(ctx context.Context, flt user.Filter) ([]user.User, error)
	ListByOrg(ctx context.Context, orgID string, roleFilter string) ([]user.User, error)
	ListByGroup(ctx context.Context, groupID string, roleFilter string) ([]user.User, error)
	Update(ctx context.Context, toUpdate user.User) (user.User, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	IsSudo(ctx context.Context, id string, permissionName string) (bool, error)
	Sudo(ctx context.Context, id string, relationName string) error
}

func (h Handler) ListUsers(ctx context.Context, request *frontierv1beta1.ListUsersRequest) (*frontierv1beta1.ListUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, request.GetOrgId())

	var users []*frontierv1beta1.User
	usersList, err := h.userService.List(ctx, user.Filter{
		Limit:   request.GetPageSize(),
		Page:    request.GetPageNum(),
		Keyword: request.GetKeyword(),
		OrgID:   request.GetOrgId(),
		GroupID: request.GetGroupId(),
		State:   user.State(request.GetState()),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, user := range usersList {
		userPB, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, userPB)
	}

	auditor.Log(audit.UserListedEvent, audit.OrgTarget(request.GetOrgId()))
	return &frontierv1beta1.ListUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}, nil
}

func (h Handler) ListAllUsers(ctx context.Context, request *frontierv1beta1.ListAllUsersRequest) (*frontierv1beta1.ListAllUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*frontierv1beta1.User
	usersList, err := h.userService.List(ctx, user.Filter{
		Limit:   request.GetPageSize(),
		Page:    request.GetPageNum(),
		Keyword: request.GetKeyword(),
		OrgID:   request.GetOrgId(),
		GroupID: request.GetGroupId(),
		State:   user.State(request.GetState()),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, user := range usersList {
		userPB, err := transformUserToPB(user)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		users = append(users, userPB)
	}

	return &frontierv1beta1.ListAllUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}, nil
}

func (h Handler) CreateUser(ctx context.Context, request *frontierv1beta1.CreateUserRequest) (*frontierv1beta1.CreateUserResponse, error) {
	logger := grpczap.Extract(ctx)
	ctx, err := tag.New(ctx, tag.Insert(telemetry.KeyMethod, "CreateUser"))
	if err != nil {
		return nil, grpcInternalServerError
	}

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	email := strings.TrimSpace(request.GetBody().GetEmail())
	if email == "" {
		currentUserEmail, ok := authenticate.GetEmailFromContext(ctx)
		if !ok {
			return nil, grpcBadBodyError
		}

		currentUserEmail = strings.TrimSpace(currentUserEmail)
		if currentUserEmail == "" {
			logger.Error(ErrEmptyEmailID.Error())
			return nil, grpcBadBodyError
		}
		email = currentUserEmail
	}

	title := request.GetBody().GetTitle()
	name := strings.TrimSpace(request.GetBody().GetName())
	if name == "" {
		name = str.GenerateUserSlug(email)
	}

	var metaDataMap metadata.Metadata
	if request.GetBody().GetMetadata() != nil {
		metaDataMap = metadata.Build(request.GetBody().GetMetadata().AsMap())

		if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
			logger.Error(err.Error())
			return nil, grpcBadBodyMetaSchemaError
		}
	}

	// TODO might need to check the valid email form
	newUser, err := h.userService.Create(ctx, user.User{
		Title:    title,
		Email:    email,
		Name:     name,
		Avatar:   request.GetBody().GetAvatar(),
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrConflict):
			return nil, grpcConflictError
		case errors.Is(errors.Unwrap(err), user.ErrKeyDoesNotExists):
			missingKey := strings.Split(err.Error(), ":")
			if len(missingKey) == 2 {
				ctx, _ = tag.New(ctx, tag.Upsert(telemetry.KeyMissingKey, missingKey[1]))
			}
			stats.Record(ctx, telemetry.MMissingMetadataKeys.M(1))

			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	transformedUser, err := transformUserToPB(newUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	audit.GetAuditor(ctx, schema.PlatformOrgID.String()).
		Log(audit.UserCreatedEvent, audit.OrgTarget(schema.PlatformOrgID.String()))
	return &frontierv1beta1.CreateUserResponse{User: transformedUser}, nil
}

func (h Handler) GetUser(ctx context.Context, request *frontierv1beta1.GetUserRequest) (*frontierv1beta1.GetUserResponse, error) {
	logger := grpczap.Extract(ctx)

	fetchedUser, err := h.userService.GetByID(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidUUID), errors.Is(err, user.ErrInvalidID):
			return nil, grpcUserNotFoundError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(fetchedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, ErrInternalServer.Error())
	}

	return &frontierv1beta1.GetUserResponse{
		User: userPB,
	}, nil
}

func (h Handler) GetCurrentUser(ctx context.Context, request *frontierv1beta1.GetCurrentUserRequest) (*frontierv1beta1.GetCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if principal.Type == schema.ServiceUserPrincipal {
		return &frontierv1beta1.GetCurrentUserResponse{
			Serviceuser: &frontierv1beta1.ServiceUser{
				Id:        principal.ServiceUser.ID,
				Title:     principal.ServiceUser.Title,
				State:     principal.ServiceUser.State,
				Metadata:  nil,
				CreatedAt: timestamppb.New(principal.ServiceUser.CreatedAt),
				UpdatedAt: timestamppb.New(principal.ServiceUser.UpdatedAt),
			},
		}, nil
	}

	userPB, err := transformUserToPB(*principal.User)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &frontierv1beta1.GetCurrentUserResponse{
		User: userPB,
	}, nil
}

func (h Handler) UpdateUser(ctx context.Context, request *frontierv1beta1.UpdateUserRequest) (*frontierv1beta1.UpdateUserResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, schema.PlatformOrgID.String())
	var updatedUser user.User

	if strings.TrimSpace(request.GetId()) == "" {
		return nil, grpcUserNotFoundError
	}

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	email := strings.TrimSpace(request.GetBody().GetEmail())
	if email == "" {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}
	var err error
	id := request.GetId()
	// upsert by email
	if isValidEmail(id) {
		_, err = h.userService.GetByEmail(ctx, id)
		if err != nil {
			if err == user.ErrNotExist {
				createUserResponse, err := h.CreateUser(ctx, &frontierv1beta1.CreateUserRequest{Body: request.GetBody()})
				if err != nil {
					return nil, grpcInternalServerError
				}
				return &frontierv1beta1.UpdateUserResponse{User: createUserResponse.User}, nil
			} else {
				return nil, grpcInternalServerError
			}
		}
		// if email in request body is different from that of user getting updated
		if email != id {
			return nil, status.Errorf(codes.InvalidArgument, ErrEmailConflict.Error())
		}
	}

	updatedUser, err = h.userService.Update(ctx, user.User{
		ID:       request.GetId(),
		Title:    request.GetBody().GetTitle(),
		Email:    request.GetBody().GetEmail(),
		Avatar:   request.GetBody().GetAvatar(),
		Name:     request.GetBody().GetName(),
		Metadata: metaDataMap,
	})

	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist), errors.Is(err, user.ErrInvalidID), errors.Is(err, user.ErrInvalidUUID):
			return nil, grpcUserNotFoundError
		case errors.Is(err, user.ErrInvalidDetails):
			return nil, grpcBadBodyError
		case errors.Is(err, user.ErrConflict):
			return nil, grpcConflictError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, ErrInternalServer
	}

	auditor.Log(audit.UserUpdatedEvent, audit.OrgTarget(schema.PlatformOrgID.String()))
	return &frontierv1beta1.UpdateUserResponse{User: userPB}, nil
}

func (h Handler) UpdateCurrentUser(ctx context.Context, request *frontierv1beta1.UpdateCurrentUserRequest) (*frontierv1beta1.UpdateCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)
	auditor := audit.GetAuditor(ctx, schema.PlatformOrgID.String())
	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	// if email in request body is different from the email in the header
	if principal.User != nil && principal.User.Email != request.GetBody().GetEmail() {
		return nil, grpcBadBodyError
	}

	metaDataMap := metadata.Build(request.GetBody().GetMetadata().AsMap())

	if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	updatedUser, err := h.userService.Update(ctx, user.User{
		ID:       principal.ID,
		Title:    request.GetBody().GetTitle(),
		Avatar:   request.GetBody().GetAvatar(),
		Name:     request.GetBody().GetName(),
		Metadata: metaDataMap,
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, user.ErrNotExist):
			return nil, grpcUserNotFoundError
		case errors.Is(err, user.ErrInvalidDetails):
			return nil, grpcBadBodyError
		default:
			return nil, grpcInternalServerError
		}
	}

	userPB, err := transformUserToPB(updatedUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	auditor.Log(audit.UserUpdatedEvent, audit.OrgTarget(schema.PlatformOrgID.String()))
	return &frontierv1beta1.UpdateCurrentUserResponse{User: userPB}, nil
}

func (h Handler) ListUserGroups(ctx context.Context, request *frontierv1beta1.ListUserGroupsRequest) (*frontierv1beta1.ListUserGroupsResponse, error) {
	logger := grpczap.Extract(ctx)
	var groups []*frontierv1beta1.Group

	groupsList, err := h.groupService.ListByUser(ctx, request.GetId(), group.Filter{OrganizationID: request.GetOrgId()})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	for _, group := range groupsList {
		groupPB, err := transformGroupToPB(group)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		groups = append(groups, &groupPB)
	}

	return &frontierv1beta1.ListUserGroupsResponse{
		Groups: groups,
	}, nil
}

func (h Handler) ListCurrentUserGroups(ctx context.Context, request *frontierv1beta1.ListCurrentUserGroupsRequest) (*frontierv1beta1.ListCurrentUserGroupsResponse, error) {
	logger := grpczap.Extract(ctx)
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	var groupsPb []*frontierv1beta1.Group
	var accessPairsPb []*frontierv1beta1.ListCurrentUserGroupsResponse_AccessPair

	groupsList, err := h.groupService.ListByUser(ctx, principal.ID, group.Filter{
		OrganizationID: request.GetOrgId(),
	})
	if err != nil {
		logger.Error(err.Error())
		switch {
		case errors.Is(err, group.ErrInvalidID), errors.Is(err, group.ErrInvalidUUID):
			return nil, grpcGroupNotFoundErr
		default:
			return nil, grpcInternalServerError
		}
	}

	for _, group := range groupsList {
		groupPB, err := transformGroupToPB(group)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		groupsPb = append(groupsPb, &groupPB)
	}

	if len(request.WithPermissions) > 0 {
		resourceIds := utils.Map(groupsList, func(res group.Group) string {
			return res.ID
		})
		successCheckPairs, err := h.fetchAccessPairsOnResource(ctx, schema.GroupNamespace, resourceIds, request.GetWithPermissions())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		for _, successCheck := range successCheckPairs {
			resID := successCheck.Relation.Object.ID

			// find all permission checks on same resource
			pairsForCurrentGroup := utils.Filter(successCheckPairs, func(pair relation.CheckPair) bool {
				return pair.Relation.Object.ID == resID
			})
			// fetch permissions
			permissions := utils.Map(pairsForCurrentGroup, func(pair relation.CheckPair) string {
				return pair.Relation.RelationName
			})
			accessPairsPb = append(accessPairsPb, &frontierv1beta1.ListCurrentUserGroupsResponse_AccessPair{
				GroupId:     resID,
				Permissions: permissions,
			})
		}
	}
	return &frontierv1beta1.ListCurrentUserGroupsResponse{
		Groups:      groupsPb,
		AccessPairs: accessPairsPb,
	}, nil
}

func (h Handler) ListOrganizationsByUser(ctx context.Context, request *frontierv1beta1.ListOrganizationsByUserRequest) (*frontierv1beta1.ListOrganizationsByUserResponse, error) {
	logger := grpczap.Extract(ctx)

	orgList, err := h.orgService.ListByUser(ctx, request.GetId(), organization.Filter{})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var orgs []*frontierv1beta1.Organization
	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		orgs = append(orgs, orgPB)
	}

	principal, err := h.GetUser(ctx, &frontierv1beta1.GetUserRequest{Id: request.GetId()})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	joinableOrgIDs, err := h.domainService.ListJoinableOrgsByDomain(ctx, principal.User.Email)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var joinableOrgs []*frontierv1beta1.Organization
	for _, joinableOrg := range joinableOrgIDs {
		org, err := h.orgService.Get(ctx, joinableOrg)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		orgPB, err := transformOrgToPB(org)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		joinableOrgs = append(joinableOrgs, orgPB)
	}
	return &frontierv1beta1.ListOrganizationsByUserResponse{Organizations: orgs, JoinableViaDomain: joinableOrgs}, nil
}

func (h Handler) ListOrganizationsByCurrentUser(ctx context.Context, request *frontierv1beta1.ListOrganizationsByCurrentUserRequest) (*frontierv1beta1.ListOrganizationsByCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	orgFilter := organization.Filter{}
	if request.GetState() != "" {
		orgFilter.State = organization.State(request.GetState())
	}
	orgList, err := h.orgService.ListByUser(ctx, principal.ID, orgFilter)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var orgs []*frontierv1beta1.Organization
	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		orgs = append(orgs, orgPB)
	}

	joinableOrgIDs, err := h.domainService.ListJoinableOrgsByDomain(ctx, principal.User.Email)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var joinableOrgs []*frontierv1beta1.Organization
	for _, joinableOrg := range joinableOrgIDs {
		org, err := h.orgService.Get(ctx, joinableOrg)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		orgPB, err := transformOrgToPB(org)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		joinableOrgs = append(joinableOrgs, orgPB)
	}

	return &frontierv1beta1.ListOrganizationsByCurrentUserResponse{Organizations: orgs, JoinableViaDomain: joinableOrgs}, nil
}

func (h Handler) ListProjectsByUser(ctx context.Context, request *frontierv1beta1.ListProjectsByUserRequest) (*frontierv1beta1.ListProjectsByUserResponse, error) {
	logger := grpczap.Extract(ctx)

	projList, err := h.projectService.ListByUser(ctx, request.GetId(), project.Filter{})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var projects []*frontierv1beta1.Project
	for _, v := range projList {
		projPB, err := transformProjectToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		projects = append(projects, projPB)
	}
	return &frontierv1beta1.ListProjectsByUserResponse{Projects: projects}, nil
}

func (h Handler) ListProjectsByCurrentUser(ctx context.Context, request *frontierv1beta1.ListProjectsByCurrentUserRequest) (*frontierv1beta1.ListProjectsByCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)
	principal, err := h.GetLoggedInPrincipal(ctx)
	if err != nil {
		return nil, err
	}
	projList, err := h.projectService.ListByUser(ctx, principal.ID, project.Filter{
		OrgID: request.GetOrgId(),
	})
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var projects []*frontierv1beta1.Project
	var accessPairsPb []*frontierv1beta1.ListProjectsByCurrentUserResponse_AccessPair
	for _, v := range projList {
		projPB, err := transformProjectToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		projects = append(projects, projPB)
	}
	if len(request.WithPermissions) > 0 {
		resourceIds := utils.Map(projList, func(res project.Project) string {
			return res.ID
		})
		successCheckPairs, err := h.fetchAccessPairsOnResource(ctx, schema.ProjectNamespace, resourceIds, request.GetWithPermissions())
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		for _, successCheck := range successCheckPairs {
			resID := successCheck.Relation.Object.ID

			// find all permission checks on same resource
			pairsForCurrentGroup := utils.Filter(successCheckPairs, func(pair relation.CheckPair) bool {
				return pair.Relation.Object.ID == resID
			})
			// fetch permissions
			permissions := utils.Map(pairsForCurrentGroup, func(pair relation.CheckPair) string {
				return pair.Relation.RelationName
			})
			accessPairsPb = append(accessPairsPb, &frontierv1beta1.ListProjectsByCurrentUserResponse_AccessPair{
				ProjectId:   resID,
				Permissions: permissions,
			})
		}
	}
	return &frontierv1beta1.ListProjectsByCurrentUserResponse{
		Projects:    projects,
		AccessPairs: accessPairsPb,
	}, nil
}

func (h Handler) EnableUser(ctx context.Context, request *frontierv1beta1.EnableUserRequest) (*frontierv1beta1.EnableUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.userService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.EnableUserResponse{}, nil
}

func (h Handler) DisableUser(ctx context.Context, request *frontierv1beta1.DisableUserRequest) (*frontierv1beta1.DisableUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.userService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.DisableUserResponse{}, nil
}

func (h Handler) DeleteUser(ctx context.Context, request *frontierv1beta1.DeleteUserRequest) (*frontierv1beta1.DeleteUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.deleterService.DeleteUser(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &frontierv1beta1.DeleteUserResponse{}, nil
}

func transformUserToPB(usr user.User) (*frontierv1beta1.User, error) {
	metaData, err := usr.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &frontierv1beta1.User{
		Id:        usr.ID,
		Title:     usr.Title,
		Email:     usr.Email,
		Name:      usr.Name,
		Metadata:  metaData,
		Avatar:    usr.Avatar,
		State:     usr.State.String(),
		CreatedAt: timestamppb.New(usr.CreatedAt),
		UpdatedAt: timestamppb.New(usr.UpdatedAt),
	}, nil
}

func isValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}
