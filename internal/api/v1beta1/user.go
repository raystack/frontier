package v1beta1

import (
	"context"
	"net/mail"
	"strings"

	"github.com/pkg/errors"

	grpczap "github.com/grpc-ecosystem/go-grpc-middleware/logging/zap/ctxzap"
	"go.opencensus.io/stats"
	"go.opencensus.io/tag"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	"github.com/raystack/shield/core/user"
	"github.com/raystack/shield/pkg/metadata"
	"github.com/raystack/shield/pkg/str"
	"github.com/raystack/shield/pkg/telemetry"
	shieldv1beta1 "github.com/raystack/shield/proto/v1beta1"
)

var grpcUserNotFoundError = status.Errorf(codes.NotFound, "user doesn't exist")

//go:generate mockery --name=UserService -r --case underscore --with-expecter --structname UserService --filename user_service.go --output=./mocks
type UserService interface {
	GetByID(ctx context.Context, id string) (user.User, error)
	GetByIDs(ctx context.Context, userIDs []string) ([]user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
	Create(ctx context.Context, user user.User) (user.User, error)
	List(ctx context.Context, flt user.Filter) ([]user.User, error)
	ListByOrg(ctx context.Context, orgID string, permissionFilter string) ([]user.User, error)
	Update(ctx context.Context, toUpdate user.User) (user.User, error)
	UpdateByEmail(ctx context.Context, toUpdate user.User) (user.User, error)
	FetchCurrentUser(ctx context.Context) (user.User, error)
	Enable(ctx context.Context, id string) error
	Disable(ctx context.Context, id string) error
	Delete(ctx context.Context, id string) error
	IsSudo(ctx context.Context, id string) (bool, error)
}

func (h Handler) ListUsers(ctx context.Context, request *shieldv1beta1.ListUsersRequest) (*shieldv1beta1.ListUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*shieldv1beta1.User
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

	return &shieldv1beta1.ListUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}, nil
}

func (h Handler) ListAllUsers(ctx context.Context, request *shieldv1beta1.ListAllUsersRequest) (*shieldv1beta1.ListAllUsersResponse, error) {
	logger := grpczap.Extract(ctx)
	var users []*shieldv1beta1.User
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

	return &shieldv1beta1.ListAllUsersResponse{
		Count: int32(len(users)),
		Users: users,
	}, nil
}

func (h Handler) CreateUser(ctx context.Context, request *shieldv1beta1.CreateUserRequest) (*shieldv1beta1.CreateUserResponse, error) {
	logger := grpczap.Extract(ctx)
	ctx, err := tag.New(ctx, tag.Insert(telemetry.KeyMethod, "CreateUser"))

	currentUserEmail, ok := user.GetEmailFromContext(ctx)
	if !ok {
		return nil, grpcUnauthenticated
	}

	currentUserEmail = strings.TrimSpace(currentUserEmail)
	if currentUserEmail == "" {
		logger.Error(ErrEmptyEmailID.Error())
		return nil, grpcUnauthenticated
	}

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	email := strings.TrimSpace(request.GetBody().GetEmail())
	if email == "" {
		email = currentUserEmail
	}

	title := request.GetBody().GetTitle()
	name := strings.TrimSpace(request.GetBody().GetName())
	if name == "" {
		name = str.GenerateUserSlug(email)
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	// TODO might need to check the valid email form
	newUser, err := h.userService.Create(ctx, user.User{
		Title:    title,
		Email:    email,
		Name:     name,
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
	return &shieldv1beta1.CreateUserResponse{User: transformedUser}, nil
}

func (h Handler) GetUser(ctx context.Context, request *shieldv1beta1.GetUserRequest) (*shieldv1beta1.GetUserResponse, error) {
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

	return &shieldv1beta1.GetUserResponse{
		User: userPB,
	}, nil
}

func (h Handler) GetCurrentUser(ctx context.Context, request *shieldv1beta1.GetCurrentUserRequest) (*shieldv1beta1.GetCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	currentUser, err := h.GetLoggedInUser(ctx)
	if err != nil {
		logger.Error(err.Error())
		return nil, err
	}

	if err = h.setUserContextTokenInHeaders(ctx, currentUser); err != nil {
		logger.Error(err.Error())
		return nil, status.Error(codes.Internal, err.Error())
	}

	userPB, err := transformUserToPB(currentUser)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}
	return &shieldv1beta1.GetCurrentUserResponse{
		User: userPB,
	}, nil
}

func (h Handler) UpdateUser(ctx context.Context, request *shieldv1beta1.UpdateUserRequest) (*shieldv1beta1.UpdateUserResponse, error) {
	logger := grpczap.Extract(ctx)
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

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	id := request.GetId()
	// upsert by email
	if isValidEmail(id) {
		_, err = h.userService.GetByEmail(ctx, id)
		if err != nil {
			if err == user.ErrNotExist {
				createUserResponse, err := h.CreateUser(ctx, &shieldv1beta1.CreateUserRequest{Body: request.GetBody()})
				if err != nil {
					return nil, grpcInternalServerError
				}
				return &shieldv1beta1.UpdateUserResponse{User: createUserResponse.User}, nil
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

	return &shieldv1beta1.UpdateUserResponse{User: userPB}, nil
}

func (h Handler) UpdateCurrentUser(ctx context.Context, request *shieldv1beta1.UpdateCurrentUserRequest) (*shieldv1beta1.UpdateCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)

	email, ok := user.GetEmailFromContext(ctx)
	if !ok {
		return nil, grpcUnauthenticated
	}

	email = strings.TrimSpace(email)
	if email == "" {
		logger.Error(ErrEmptyEmailID.Error())
		return nil, grpcUnauthenticated
	}

	if request.GetBody() == nil {
		return nil, grpcBadBodyError
	}

	metaDataMap, err := metadata.Build(request.GetBody().GetMetadata().AsMap())
	if err != nil {
		return nil, grpcBadBodyError
	}

	if err := h.metaSchemaService.Validate(metaDataMap, userMetaSchema); err != nil {
		logger.Error(err.Error())
		return nil, grpcBadBodyMetaSchemaError
	}

	// if email in request body is different from the email in the header
	if request.GetBody().GetEmail() != email {
		return nil, grpcBadBodyError
	}

	updatedUser, err := h.userService.UpdateByEmail(ctx, user.User{
		Title:    request.GetBody().GetTitle(),
		Email:    email,
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

	return &shieldv1beta1.UpdateCurrentUserResponse{User: userPB}, nil
}

func (h Handler) ListUserGroups(ctx context.Context, request *shieldv1beta1.ListUserGroupsRequest) (*shieldv1beta1.ListUserGroupsResponse, error) {
	logger := grpczap.Extract(ctx)
	var groups []*shieldv1beta1.Group

	groupsList, err := h.groupService.ListByUser(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	for _, group := range groupsList {
		groupPB, err := transformGroupToPB(group)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}

		groups = append(groups, &groupPB)
	}

	return &shieldv1beta1.ListUserGroupsResponse{
		Groups: groups,
	}, nil
}

func (h Handler) GetOrganizationsByUser(ctx context.Context, request *shieldv1beta1.GetOrganizationsByUserRequest) (*shieldv1beta1.GetOrganizationsByUserResponse, error) {
	logger := grpczap.Extract(ctx)

	orgList, err := h.orgService.ListByUser(ctx, request.GetId())
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var orgs []*shieldv1beta1.Organization
	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		orgs = append(orgs, orgPB)
	}
	return &shieldv1beta1.GetOrganizationsByUserResponse{Organizations: orgs}, nil
}

func (h Handler) EnableUser(ctx context.Context, request *shieldv1beta1.EnableUserRequest) (*shieldv1beta1.EnableUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.userService.Enable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.EnableUserResponse{}, nil
}

func (h Handler) DisableUser(ctx context.Context, request *shieldv1beta1.DisableUserRequest) (*shieldv1beta1.DisableUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.userService.Disable(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.DisableUserResponse{}, nil
}

func (h Handler) DeleteUser(ctx context.Context, request *shieldv1beta1.DeleteUserRequest) (*shieldv1beta1.DeleteUserResponse, error) {
	logger := grpczap.Extract(ctx)
	if err := h.userService.Delete(ctx, request.GetId()); err != nil {
		logger.Error(err.Error())
		return nil, status.Errorf(codes.Internal, err.Error())
	}
	return &shieldv1beta1.DeleteUserResponse{}, nil
}

func (h Handler) ListCurrentUserGroups(ctx context.Context, request *shieldv1beta1.ListCurrentUserGroupsRequest) (*shieldv1beta1.ListCurrentUserGroupsResponse, error) {
	// TODO
	return &shieldv1beta1.ListCurrentUserGroupsResponse{}, grpcOperationUnsupported
}

func (h Handler) GetOrganizationsByCurrentUser(ctx context.Context, request *shieldv1beta1.GetOrganizationsByCurrentUserRequest) (*shieldv1beta1.GetOrganizationsByCurrentUserResponse, error) {
	logger := grpczap.Extract(ctx)
	currentUser, err := h.GetLoggedInUser(ctx)
	if err != nil {
		return nil, err
	}
	orgList, err := h.orgService.ListByUser(ctx, currentUser.ID)
	if err != nil {
		logger.Error(err.Error())
		return nil, grpcInternalServerError
	}

	var orgs []*shieldv1beta1.Organization
	for _, v := range orgList {
		orgPB, err := transformOrgToPB(v)
		if err != nil {
			logger.Error(err.Error())
			return nil, grpcInternalServerError
		}
		orgs = append(orgs, orgPB)
	}
	return &shieldv1beta1.GetOrganizationsByCurrentUserResponse{Organizations: orgs}, nil
}

func transformUserToPB(usr user.User) (*shieldv1beta1.User, error) {
	metaData, err := usr.Metadata.ToStructPB()
	if err != nil {
		return nil, err
	}

	return &shieldv1beta1.User{
		Id:        usr.ID,
		Title:     usr.Title,
		Email:     usr.Email,
		Name:      usr.Name,
		Metadata:  metaData,
		CreatedAt: timestamppb.New(usr.CreatedAt),
		UpdatedAt: timestamppb.New(usr.UpdatedAt),
	}, nil
}

func isValidEmail(str string) bool {
	_, err := mail.ParseAddress(str)
	return err == nil
}
