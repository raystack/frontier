package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/odpf/shield/core/namespace"
	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/role"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var testRelation = relation.Relation{
	ID: "relation-id-1",
	SubjectNamespace: namespace.Namespace{
		ID: "ns1",
	},
	SubjectNamespaceID: "ns1",
	SubjectID:          "subject-id",
	SubjectRoleID:      "role1",
	ObjectNamespace: namespace.Namespace{
		ID: "ns2",
	},
	ObjectNamespaceID: "ns2",
	ObjectID:          "object-id",
	RoleID:            "role1",
	Role: role.Role{
		ID: "role1",
	},
	RelationType: "role",
}

var testRelationPB = &shieldv1beta1.Relation{
	Id: "relation-id-1",
	SubjectType: &shieldv1beta1.Namespace{
		Id:        "ns1",
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	},
	SubjectId: "subject-id",
	ObjectType: &shieldv1beta1.Namespace{
		Id:        "ns2",
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	},
	ObjectId:  "object-id",
	CreatedAt: timestamppb.New(time.Time{}),
	UpdatedAt: timestamppb.New(time.Time{}),
	Role: &shieldv1beta1.Role{
		Id: "role1",
		Namespace: &shieldv1beta1.Namespace{
			CreatedAt: timestamppb.New(time.Time{}),
			UpdatedAt: timestamppb.New(time.Time{}),
		},
		Metadata: &structpb.Struct{
			Fields: make(map[string]*structpb.Value),
		},
		CreatedAt: timestamppb.New(time.Time{}),
		UpdatedAt: timestamppb.New(time.Time{}),
	},
}

func TestHandler_ListRelations(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService)
		want    *shieldv1beta1.ListRelationsResponse
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]relation.Relation{}, errors.New("some error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return relations if relation service return nil error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]relation.Relation{
					testRelation,
				}, nil)
			},
			want: &shieldv1beta1.ListRelationsResponse{
				Relations: []*shieldv1beta1.Relation{
					testRelationPB,
				},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv)
			}
			mockDep := Handler{relationService: mockRelationSrv}
			resp, err := mockDep.ListRelations(context.Background(), &shieldv1beta1.ListRelationsRequest{})
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateRelation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService)
		request *shieldv1beta1.CreateRelationRequest
		want    *shieldv1beta1.CreateRelationResponse
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, errors.New("some error"))
			},
			request: &shieldv1beta1.CreateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, relation.ErrInvalidDetail)
			},
			request: &shieldv1beta1.CreateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if relation service return nil",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(testRelation, nil)
			},
			request: &shieldv1beta1.CreateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want: &shieldv1beta1.CreateRelationResponse{
				Relation: testRelationPB,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv)
			}
			mockDep := Handler{relationService: mockRelationSrv}
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
		request *shieldv1beta1.GetRelationRequest
		want    *shieldv1beta1.GetRelationResponse
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRelation.ID).Return(relation.Relation{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: testRelation.ID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(relation.Relation{}, relation.ErrInvalidID)
			},
			request: &shieldv1beta1.GetRelationRequest{},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(relation.Relation{}, relation.ErrInvalidUUID)
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: "some-id",
			},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return not found error if id not exist",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRelation.ID).Return(relation.Relation{}, relation.ErrNotExist)
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: testRelation.ID,
			},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return success if relation service return nil error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRelation.ID).Return(testRelation, nil)
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: testRelation.ID,
			},
			want: &shieldv1beta1.GetRelationResponse{
				Relation: testRelationPB,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv)
			}
			mockDep := Handler{relationService: mockRelationSrv}
			resp, err := mockDep.GetRelation(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_UpdateRelation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService)
		request *shieldv1beta1.UpdateRelationRequest
		want    *shieldv1beta1.UpdateRelationResponse
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					ID:                 testRelation.ID,
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, errors.New("some error"))
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Id: testRelation.ID,
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, relation.ErrInvalidID)
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return not found error if id is not exist",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, relation.ErrNotExist)
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, relation.ErrInvalidUUID)
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					ID:                 testRelation.ID,
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, relation.ErrInvalidDetail)
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Id: testRelation.ID,
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return already exist error if relation service return err conflict",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					ID:                 testRelation.ID,
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(relation.Relation{}, relation.ErrConflict)
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Id: testRelation.ID,
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want:    nil,
			wantErr: grpcConflictError,
		},
		{
			name: "should return success if relation service return nil",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Update(mock.AnythingOfType("*context.emptyCtx"), relation.Relation{
					ID:                 testRelation.ID,
					SubjectNamespaceID: testRelation.SubjectNamespaceID,
					SubjectID:          testRelation.SubjectID,
					ObjectNamespaceID:  testRelation.ObjectNamespaceID,
					ObjectID:           testRelation.ObjectID,
					RoleID:             testRelation.RoleID,
				}).Return(testRelation, nil)
			},
			request: &shieldv1beta1.UpdateRelationRequest{
				Id: testRelation.ID,
				Body: &shieldv1beta1.RelationRequestBody{
					SubjectType: testRelation.SubjectNamespaceID,
					SubjectId:   testRelation.SubjectID,
					ObjectType:  testRelation.ObjectNamespaceID,
					ObjectId:    testRelation.ObjectID,
					RoleId:      testRelation.RoleID,
				},
			},
			want: &shieldv1beta1.UpdateRelationResponse{
				Relation: testRelationPB,
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockRelationSrv := new(mocks.RelationService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv)
			}
			mockDep := Handler{relationService: mockRelationSrv}
			resp, err := mockDep.UpdateRelation(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
