package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CreateProspectPublic(t *testing.T) {
	now := time.Now()
	// fixedTime := timestamppb.New(now)
	args := &frontierv1beta1.CreateProspectPublicRequest{
		Email:    "test@example.com",
		Activity: "newsletter",
		Metadata: nil,
	}
	tests := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.CreateProspectPublicRequest
		want  *frontierv1beta1.CreateProspectPublicResponse
		err   error
	}{
		{
			title: "should return internal error in if prospect service return some error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{}, grpcInternalServerError)
				return ctx
			},
			req: &frontierv1beta1.CreateProspectPublicRequest{
				Email:    "test@example.com",
				Activity: "newsletter",
				Metadata: nil,
			},
			want: &frontierv1beta1.CreateProspectPublicResponse{},
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad schema error if meta schema service gives error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(errors.New("grpcBadBodyMetaSchemaError"))
				return ctx
			},
			req:  args,
			want: nil,
			err:  grpcBadBodyMetaSchemaError,
		},
		{
			title: "should return bad request error if there is no email in request",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{
					Email:    "",
					Activity: "newsletter",
					Status:   prospect.Status(frontierv1beta1.Prospect_STATUS_SUBSCRIBED),
					Metadata: nil}, grpcEmailInvalidError)
				return ctx
			},
			req: &frontierv1beta1.CreateProspectPublicRequest{
				Email:    "",
				Activity: "newsletter",
				Metadata: nil,
			},
			want: nil,
			err:  grpcEmailInvalidError,
		},
		{
			title: "should ignore error if prospect service returns conflict",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{}, nil)
				return ctx
			},
			req:  args,
			want: &frontierv1beta1.CreateProspectPublicResponse{},
			err:  nil,
		},
		{
			title: "should return success if prospect service return nil error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{
					ID:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    prospect.Subscribed,
					Metadata:  metadata.Metadata{"medium": "test"},
					Source:    "test",
					CreatedAt: now,
					UpdatedAt: now,
					ChangedAt: now,
				}, nil)
				return ctx
			},
			req: &frontierv1beta1.CreateProspectPublicRequest{
				Email:    "test@example.com",
				Activity: "newsletter",
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"medium": structpb.NewStringValue("test"),
					},
				},
			},
			want: &frontierv1beta1.CreateProspectPublicResponse{},
			err:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockProspectSrv := new(mocks.ProspectService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.TODO()
			if tt.setup != nil {
				ctx = tt.setup(context.Background(), mockProspectSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{prospectService: mockProspectSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.CreateProspectPublic(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestHandler_CreateProspect(t *testing.T) {
	now := time.Now()
	fixedTime := timestamppb.New(now)
	args := &frontierv1beta1.CreateProspectRequest{
		Email:    "test@example.com",
		Activity: "newsletter",
		Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
		Metadata: nil,
	}
	tests := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.CreateProspectRequest
		want  *frontierv1beta1.CreateProspectResponse
		err   error
	}{
		{
			title: "should return internal error in if prospect service return some error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{}, grpcInternalServerError)
				return ctx
			},
			req: &frontierv1beta1.CreateProspectRequest{
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
				Metadata: nil,
			},
			want: &frontierv1beta1.CreateProspectResponse{},
			err:  grpcInternalServerError,
		},
		{
			title: "should return bad schema error if meta schema service gives error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(errors.New("grpcBadBodyMetaSchemaError"))
				return ctx
			},
			req:  args,
			want: nil,
			err:  grpcBadBodyMetaSchemaError,
		},
		{
			title: "should return bad request error if there is no email in request",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{
					Email:    "",
					Activity: "newsletter",
					Status:   prospect.Status(frontierv1beta1.Prospect_STATUS_SUBSCRIBED),
					Metadata: nil}, grpcEmailInvalidError)
				return ctx
			},
			req: &frontierv1beta1.CreateProspectRequest{
				Email:    "",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
				Metadata: nil,
			},
			want: nil,
			err:  grpcEmailInvalidError,
		},
		{
			title: "should return already exist error if prospect service return error conflict",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{}, prospect.ErrEmailActivityAlreadyExists)
				return ctx
			},
			req:  args,
			want: &frontierv1beta1.CreateProspectResponse{},
			err:  grpcConflictError,
		},
		{
			title: "should return success if prospect service return nil error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ms.On("Validate", mock.AnythingOfType("metadata.Metadata"), prospectMetaSchema).Return(nil)
				ps.On("Create", mock.Anything, mock.Anything).Return(prospect.Prospect{
					ID:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    prospect.Subscribed,
					Metadata:  metadata.Metadata{"medium": "test"},
					Source:    "test",
					CreatedAt: now,
					UpdatedAt: now,
					ChangedAt: now,
				}, nil)
				return ctx
			},
			req: &frontierv1beta1.CreateProspectRequest{
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"medium": structpb.NewStringValue("test"),
					},
				},
			},
			want: &frontierv1beta1.CreateProspectResponse{
				Prospect: &frontierv1beta1.Prospect{
					Id:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
					ChangedAt: fixedTime,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"medium": structpb.NewStringValue("test"),
						}},
					Source:    "test",
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
				},
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockProspectSrv := new(mocks.ProspectService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)

			ctx := context.TODO()
			if tt.setup != nil {
				ctx = tt.setup(context.Background(), mockProspectSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{prospectService: mockProspectSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.CreateProspect(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestHandler_ListProspects(t *testing.T) {
	now := time.Now()
	fixedTime := timestamppb.New(now)
	tests := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.ListProspectsRequest
		want  *frontierv1beta1.ListProspectsResponse
		err   error
	}{
		{
			title: "should return internal error in if prospect service return some error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("List", mock.Anything, mock.Anything).Return(prospect.ListProspects{Prospects: nil,
					Page: utils.Page{
						Limit:      0,
						Offset:     0,
						TotalCount: 0,
					},
					Group: nil}, grpcInternalServerError)
				return ctx
			},
			req:  &frontierv1beta1.ListProspectsRequest{},
			want: nil,
			err:  grpcInternalServerError,
		},
		{
			title: "should return success if prospect service return nil error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("List", mock.Anything, mock.Anything).
					Return(prospect.ListProspects{
						Prospects: []prospect.Prospect{
							{
								ID:        "id-1",
								Email:     "test@example.com",
								Activity:  "newsletter",
								Status:    prospect.Subscribed,
								ChangedAt: now,
								Name:      "",
								Phone:     "",
								Source:    "test",
								Verified:  true,
								CreatedAt: now,
								UpdatedAt: now,
								Metadata:  metadata.Metadata{},
							},
						},
						Page: utils.Page{
							Limit:      1,
							Offset:     0,
							TotalCount: 1,
						},
						Group: &utils.Group{
							Name: "",
							Data: []utils.GroupData{},
						}}, nil)
				return ctx
			},
			req: &frontierv1beta1.ListProspectsRequest{
				Query: nil,
			},
			want: &frontierv1beta1.ListProspectsResponse{
				Prospects: []*frontierv1beta1.Prospect{
					{
						Id:        "id-1",
						Email:     "test@example.com",
						Activity:  "newsletter",
						Status:    frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
						ChangedAt: fixedTime,
						Name:      "",
						Phone:     "",
						Source:    "test",
						Verified:  true,
						CreatedAt: fixedTime,
						UpdatedAt: fixedTime,
						Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
				Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
					Offset:     0,
					Limit:      1,
					TotalCount: 1,
				},
				Group: nil,
			},
			err: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockProspectSrv := new(mocks.ProspectService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.TODO()
			if tt.setup != nil {
				ctx = tt.setup(context.Background(), mockProspectSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{prospectService: mockProspectSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.ListProspects(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestHandler_GetProspect(t *testing.T) {
	now := time.Now()
	fixedTime := timestamppb.New(now)
	tests := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.GetProspectRequest
		want  *frontierv1beta1.GetProspectResponse
		err   error
	}{
		{
			title: "should return bad request error in if prospect ID is null",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("Get", mock.Anything, mock.Anything).Return(prospect.Prospect{}, grpcProspectIdRequiredError)
				return ctx
			},
			req:  nil,
			want: nil,
			err:  grpcProspectIdRequiredError,
		},
		{
			title: "should return success if service returns nil error",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("Get", mock.Anything, "id-1").Return(prospect.Prospect{
					ID:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    prospect.Subscribed,
					ChangedAt: now,
					Name:      "",
					Phone:     "",
					Source:    "test",
					Verified:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Metadata:  metadata.Metadata{},
				}, nil)
				return ctx
			},
			req: &frontierv1beta1.GetProspectRequest{
				Id: "id-1",
			},
			want: &frontierv1beta1.GetProspectResponse{
				Prospect: &frontierv1beta1.Prospect{
					Id:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
					ChangedAt: fixedTime,
					Name:      "",
					Phone:     "",
					Source:    "test",
					Verified:  true,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
					Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			},
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockProspectSrv := new(mocks.ProspectService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)

			ctx := context.TODO()
			if tt.setup != nil {
				ctx = tt.setup(context.Background(), mockProspectSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{prospectService: mockProspectSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.GetProspect(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestHandler_UpdateProspect(t *testing.T) {
	now := time.Now()
	fixedTime := timestamppb.New(now)
	tests := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.UpdateProspectRequest
		want  *frontierv1beta1.UpdateProspectResponse
		err   error
	}{
		{
			title: "should return error if prospect ID is empty",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				return ctx
			},
			req: &frontierv1beta1.UpdateProspectRequest{
				Id:       "",
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
			},
			want: nil,
			err:  grpcProspectIdRequiredError,
		},
		{
			title: "should return error if email is invalid",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				return ctx
			},
			req: &frontierv1beta1.UpdateProspectRequest{
				Id:       "id-1",
				Email:    "invalid-email",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
			},
			want: nil,
			err:  grpcEmailInvalidError,
		},
		{
			title: "should return error if activity is empty",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				return ctx
			},
			req: &frontierv1beta1.UpdateProspectRequest{
				Id:       "id-1",
				Email:    "test@example.com",
				Activity: "",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
			},
			want: nil,
			err:  grpcActivityRequiredError,
		},
		{
			title: "should return error if status is unspecified",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				return ctx
			},
			req: &frontierv1beta1.UpdateProspectRequest{
				Id:       "id-1",
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_UNSPECIFIED,
			},
			want: nil,
			err:  grpcStatusRequiredError,
		},
		{
			title: "should return success if update is successful",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("Update", mock.Anything, mock.AnythingOfType("prospect.Prospect")).Return(prospect.Prospect{
					ID:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    prospect.Subscribed,
					ChangedAt: now,
					Name:      "Test User",
					Phone:     "",
					Source:    "test",
					Verified:  true,
					CreatedAt: now,
					UpdatedAt: now,
					Metadata:  metadata.Metadata{},
				}, nil)
				ms.On("Validate", mock.Anything, prospectMetaSchema).Return(nil)
				return ctx
			},
			req: &frontierv1beta1.UpdateProspectRequest{
				Id:       "id-1",
				Email:    "test@example.com",
				Name:     "Test User",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
				Verified: true,
				Source:   "test",
			},
			want: &frontierv1beta1.UpdateProspectResponse{
				Prospect: &frontierv1beta1.Prospect{
					Id:        "id-1",
					Email:     "test@example.com",
					Activity:  "newsletter",
					Status:    frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
					ChangedAt: fixedTime,
					Name:      "Test User",
					Phone:     "",
					Source:    "test",
					Verified:  true,
					CreatedAt: fixedTime,
					UpdatedAt: fixedTime,
					Metadata:  &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			},
			err: nil,
		},
		{
			title: "should return not found error if prospect doesn't exist",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("Update", mock.Anything, mock.AnythingOfType("prospect.Prospect")).Return(prospect.Prospect{}, prospect.ErrNotExist)
				ms.On("Validate", mock.Anything, prospectMetaSchema).Return(nil)
				return ctx
			},
			req: &frontierv1beta1.UpdateProspectRequest{
				Id:       "non-existent",
				Email:    "test@example.com",
				Activity: "newsletter",
				Status:   frontierv1beta1.Prospect_STATUS_SUBSCRIBED,
			},
			want: nil,
			err:  grpcProspectNotFoundError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockProspectSrv := new(mocks.ProspectService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.Background()

			if tt.setup != nil {
				ctx = tt.setup(ctx, mockProspectSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{prospectService: mockProspectSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.UpdateProspect(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}

func TestHandler_DeleteProspect(t *testing.T) {
	tests := []struct {
		title string
		setup func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context
		req   *frontierv1beta1.DeleteProspectRequest
		want  *frontierv1beta1.DeleteProspectResponse
		err   error
	}{
		{
			title: "should return error if prospect ID is empty",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				return ctx
			},
			req:  &frontierv1beta1.DeleteProspectRequest{Id: ""},
			want: nil,
			err:  grpcProspectIdRequiredError,
		},
		{
			title: "should return success if delete is successful",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("Delete", mock.Anything, "id-1").Return(nil)
				return ctx
			},
			req:  &frontierv1beta1.DeleteProspectRequest{Id: "id-1"},
			want: &frontierv1beta1.DeleteProspectResponse{},
			err:  nil,
		},
		{
			title: "should return not found error if prospect doesn't exist",
			setup: func(ctx context.Context, ps *mocks.ProspectService, ms *mocks.MetaSchemaService) context.Context {
				ps.On("Delete", mock.Anything, "non-existent").Return(prospect.ErrNotExist)
				return ctx
			},
			req:  &frontierv1beta1.DeleteProspectRequest{Id: "non-existent"},
			want: nil,
			err:  grpcProspectNotFoundError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.title, func(t *testing.T) {
			mockProspectSrv := new(mocks.ProspectService)
			mockMetaSchemaSvc := new(mocks.MetaSchemaService)
			ctx := context.Background()

			if tt.setup != nil {
				ctx = tt.setup(ctx, mockProspectSrv, mockMetaSchemaSvc)
			}
			mockDep := Handler{prospectService: mockProspectSrv, metaSchemaService: mockMetaSchemaSvc}
			resp, err := mockDep.DeleteProspect(ctx, tt.req)
			assert.EqualValues(t, resp, tt.want)
			assert.EqualValues(t, err, tt.err)
		})
	}
}
