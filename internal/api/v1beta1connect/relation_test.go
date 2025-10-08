package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	testRelationV2 = relation.Relation{
		ID: "relation-id-1",
		Subject: relation.Subject{
			ID:              "subject-id",
			Namespace:       "ns1",
			SubRelationName: "member",
		},
		Object: relation.Object{
			ID:        "object-id",
			Namespace: "ns2",
		},
		RelationName: "relation1",
	}

	testRelationPB = &frontierv1beta1.Relation{
		Id:                 "relation-id-1",
		Object:             schema.JoinNamespaceAndResourceID("ns2", "object-id"),
		Subject:            schema.JoinNamespaceAndResourceID("ns1", "subject-id"),
		SubjectSubRelation: "member",
		Relation:           "relation1",
	}
)

func TestHandler_ListRelations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService)
		want    *connect.Response[frontierv1beta1.ListRelationsResponse]
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), relation.Filter{}).Return([]relation.Relation{}, errors.New("test error"))
			},
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return relations if relation service return nil error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), relation.Filter{}).Return([]relation.Relation{
					testRelationV2,
				}, nil)
			},
			want: connect.NewResponse(&frontierv1beta1.ListRelationsResponse{
				Relations: []*frontierv1beta1.Relation{
					testRelationPB,
				},
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv)
			}
			mockDep := &ConnectHandler{relationService: mockRelationSrv}
			resp, err := mockDep.ListRelations(context.Background(), connect.NewRequest(&frontierv1beta1.ListRelationsRequest{}))
			assert.Equal(t, tt.want, resp)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateRelation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService, res *mocks.ResourceService, us *mocks.UserService)
		request *connect.Request[frontierv1beta1.CreateRelationRequest]
		want    *connect.Response[frontierv1beta1.CreateRelationResponse]
		wantErr error
	}{
		{
			name: "should return bad request error if request body is nil",
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: nil,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return bad request error if subject is not in namepsace:uuid format",
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: &frontierv1beta1.RelationRequestBody{
					Subject: "subject",
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if object is not in namepsace:uuid format",
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: &frontierv1beta1.RelationRequestBody{
					Subject: schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
					Object:  "object",
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return error if unable to get the user id from the user email in case subject namespace is app/user",
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: &frontierv1beta1.RelationRequestBody{
					Subject: schema.JoinNamespaceAndResourceID(schema.UserPrincipal, "not-a-valid-email"),
					Object:  schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
				},
			}),
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService, us *mocks.UserService) {
				us.EXPECT().GetByEmail(mock.AnythingOfType("context.backgroundCtx"), "not-a-valid-email").Return(user.User{}, user.ErrNotExist)
			},
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrUserNotExist),
		},
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService, us *mocks.UserService) {
				us.EXPECT().GetByEmail(mock.AnythingOfType("context.backgroundCtx"), "user@raystack.org").Return(user.User{
					ID: "subject-id",
				}, nil)
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), relation.Relation{
					Subject: relation.Subject{
						ID:              testRelationV2.Subject.ID,
						Namespace:       "app/user",
						SubRelationName: "",
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
					RelationName: testRelationV2.Subject.SubRelationName,
				}).Return(relation.Relation{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: &frontierv1beta1.RelationRequestBody{
					Object:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
					Subject:  schema.JoinNamespaceAndResourceID("app/user", "user@raystack.org"),
					Relation: testRelationV2.Subject.SubRelationName,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService, us *mocks.UserService) {
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), relation.Relation{
					Subject: relation.Subject{
						ID:              testRelationV2.Subject.ID,
						Namespace:       testRelationV2.Subject.Namespace,
						SubRelationName: "",
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
					RelationName: testRelationV2.Subject.SubRelationName,
				}).Return(relation.Relation{}, relation.ErrInvalidDetail)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: &frontierv1beta1.RelationRequestBody{
					Object:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
					Subject:  schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
					Relation: testRelationV2.Subject.SubRelationName,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrBadRequest),
		},
		{
			name: "should return success if relation service return nil",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService, us *mocks.UserService) {
				rs.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), relation.Relation{
					Subject: relation.Subject{
						ID:              testRelationV2.Subject.ID,
						Namespace:       testRelationV2.Subject.Namespace,
						SubRelationName: "",
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
					RelationName: testRelationV2.Subject.SubRelationName,
				}).Return(testRelationV2, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateRelationRequest{
				Body: &frontierv1beta1.RelationRequestBody{
					Object:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
					Subject:  schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
					Relation: testRelationV2.Subject.SubRelationName,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateRelationResponse{
				Relation: testRelationPB,
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			mockResourceSrv := new(mocks.ResourceService)
			mockUserSrc := new(mocks.UserService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv, mockResourceSrv, mockUserSrc)
			}
			mockDep := ConnectHandler{relationService: mockRelationSrv, resourceService: mockResourceSrv, userService: mockUserSrc}
			resp, err := mockDep.CreateRelation(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetRelation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService)
		request *connect.Request[frontierv1beta1.GetRelationRequest]
		want    *connect.Response[frontierv1beta1.GetRelationResponse]
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testRelationV2.ID).Return(relation.Relation{}, errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetRelationRequest{
				Id: testRelationV2.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "").Return(relation.Relation{}, relation.ErrInvalidID)
			},
			request: connect.NewRequest(&frontierv1beta1.GetRelationRequest{}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), "some-id").Return(relation.Relation{}, relation.ErrInvalidUUID)
			},
			request: connect.NewRequest(&frontierv1beta1.GetRelationRequest{
				Id: "some-id",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testRelationV2.ID).Return(relation.Relation{}, relation.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetRelationRequest{
				Id: testRelationV2.ID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return success if relation service return nil error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("context.backgroundCtx"), testRelationV2.ID).Return(testRelationV2, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetRelationRequest{
				Id: testRelationV2.ID,
			}),
			want: connect.NewResponse(&frontierv1beta1.GetRelationResponse{
				Relation: testRelationPB,
			}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv)
			}
			mockDep := ConnectHandler{relationService: mockRelationSrv}
			resp, err := mockDep.GetRelation(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_DeleteRelation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService, res *mocks.ResourceService)
		request *connect.Request[frontierv1beta1.DeleteRelationRequest]
		want    *connect.Response[frontierv1beta1.DeleteRelationResponse]
		wantErr error
	}{
		{
			name: "should return bad request error if subject is not in namepsace:uuid format",
			request: connect.NewRequest(&frontierv1beta1.DeleteRelationRequest{
				Subject: "not-namespace-uuid-format",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return bad request error if object is not in namepsace:uuid format",
			request: connect.NewRequest(&frontierv1beta1.DeleteRelationRequest{
				Subject: schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
				Object:  "not-namespace-uuid-format",
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrNamespaceSplitNotation),
		},
		{
			name: "should return not found error when relation service returns not exist error while deletion",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), relation.Relation{
					Subject: relation.Subject{
						Namespace: testRelationV2.Subject.Namespace,
						ID:        testRelationV2.Subject.ID,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
					RelationName: testRelationV2.Subject.SubRelationName,
				}).Return(relation.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteRelationRequest{
				Object:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
				Subject:  schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
				Relation: testRelationV2.Subject.SubRelationName,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeNotFound, ErrNotFound),
		},
		{
			name: "should return internal server error when relation service returns some error while deletion",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), relation.Relation{
					Subject: relation.Subject{
						Namespace: testRelationV2.Subject.Namespace,
						ID:        testRelationV2.Subject.ID,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
					RelationName: testRelationV2.Subject.SubRelationName,
				}).Return(errors.New("test error"))
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteRelationRequest{
				Object:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
				Subject:  schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
				Relation: testRelationV2.Subject.SubRelationName,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should successfully delete when relation exist",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				rs.EXPECT().Delete(mock.AnythingOfType("context.backgroundCtx"), relation.Relation{
					Subject: relation.Subject{
						Namespace: testRelationV2.Subject.Namespace,
						ID:        testRelationV2.Subject.ID,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
					RelationName: testRelationV2.Subject.SubRelationName,
				}).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.DeleteRelationRequest{
				Object:   schema.JoinNamespaceAndResourceID(testRelationV2.Object.Namespace, testRelationV2.Object.ID),
				Subject:  schema.JoinNamespaceAndResourceID(testRelationV2.Subject.Namespace, testRelationV2.Subject.ID),
				Relation: testRelationV2.Subject.SubRelationName,
			}),
			want:    connect.NewResponse(&frontierv1beta1.DeleteRelationResponse{}),
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv, mockResourceSrv)
			}
			mockDep := ConnectHandler{relationService: mockRelationSrv, resourceService: mockResourceSrv}
			resp, err := mockDep.DeleteRelation(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
