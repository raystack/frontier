package v1beta1

import (
	"context"
	"errors"
	"testing"

	"github.com/odpf/shield/core/permission"
	"github.com/odpf/shield/internal/bootstrap/schema"

	"github.com/odpf/shield/core/relation"
	"github.com/odpf/shield/core/resource"
	"github.com/odpf/shield/internal/api/v1beta1/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	shieldv1beta1 "github.com/odpf/shield/proto/v1beta1"
)

var (
	testRelationV2 = relation.RelationV2{
		ID: "relation-id-1",
		Subject: relation.Subject{
			ID:        "subject-id",
			Namespace: "ns1",
		},
		Object: relation.Object{
			ID:        "object-id",
			Namespace: "ns2",
		},
		RelationName: "relation1",
	}

	testRelationPB = &shieldv1beta1.Relation{
		Id:               "relation-id-1",
		ObjectId:         "object-id",
		ObjectNamespace:  "ns2",
		SubjectId:        "subject-id",
		SubjectNamespace: "ns1",
		RelationName:     "relation1",
	}
)

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
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]relation.RelationV2{}, errors.New("some error"))
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return relations if relation service return nil error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().List(mock.AnythingOfType("*context.emptyCtx")).Return([]relation.RelationV2{
					testRelationV2,
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
		setup   func(rs *mocks.RelationService, res *mocks.ResourceService)
		request *shieldv1beta1.CreateRelationRequest
		want    *shieldv1beta1.CreateRelationResponse
		wantErr error
	}{
		{
			name: "should return internal error if relation service return some error",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), relation.RelationV2{
					Subject: relation.Subject{
						ID:              testRelationV2.Subject.ID,
						Namespace:       testRelationV2.Subject.Namespace,
						SubRelationName: testRelationV2.Subject.SubRelationName,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
				}).Return(relation.RelationV2{}, errors.New("some error"))
			},
			request: &shieldv1beta1.CreateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					ObjectId:         testRelationV2.Object.ID,
					ObjectNamespace:  testRelationV2.Object.Namespace,
					SubjectId:        testRelationV2.Subject.ID,
					SubjectNamespace: testRelationV2.Subject.Namespace,
					RelationName:     testRelationV2.Subject.SubRelationName,
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return bad request error if field value not exist in foreign reference",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					Name:        testRelationV2.Object.ID,
					NamespaceID: testRelationV2.Object.Namespace,
				}, permission.Permission{ID: schema.UpdatePermission}).Return(true, nil)

				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), relation.RelationV2{
					Subject: relation.Subject{
						ID:              testRelationV2.Subject.ID,
						Namespace:       testRelationV2.Subject.Namespace,
						SubRelationName: testRelationV2.Subject.SubRelationName,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
				}).Return(relation.RelationV2{}, relation.ErrInvalidDetail)
			},
			request: &shieldv1beta1.CreateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					ObjectId:         testRelationV2.Object.ID,
					ObjectNamespace:  testRelationV2.Object.Namespace,
					SubjectId:        testRelationV2.Subject.ID,
					SubjectNamespace: testRelationV2.Subject.Namespace,
					RelationName:     testRelationV2.Subject.SubRelationName,
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return success if relation service return nil",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					Name:        testRelationV2.Object.ID,
					NamespaceID: testRelationV2.Object.Namespace,
				}, permission.Permission{ID: schema.UpdatePermission}).Return(true, nil)

				rs.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), relation.RelationV2{
					Subject: relation.Subject{
						ID:              testRelationV2.Subject.ID,
						Namespace:       testRelationV2.Subject.Namespace,
						SubRelationName: testRelationV2.Subject.SubRelationName,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
				}).Return(testRelationV2, nil)
			},
			request: &shieldv1beta1.CreateRelationRequest{
				Body: &shieldv1beta1.RelationRequestBody{
					ObjectId:         testRelationV2.Object.ID,
					ObjectNamespace:  testRelationV2.Object.Namespace,
					SubjectId:        testRelationV2.Subject.ID,
					SubjectNamespace: testRelationV2.Subject.Namespace,
					RelationName:     testRelationV2.Subject.SubRelationName,
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
			mockResourceSrv := new(mocks.ResourceService)
			if tt.setup != nil {
				tt.setup(mockRelationSrv, mockResourceSrv)
			}

			mockDep := Handler{relationService: mockRelationSrv, resourceService: mockResourceSrv}
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
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRelationV2.ID).Return(relation.RelationV2{}, errors.New("some error"))
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: testRelationV2.ID,
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return not found error if id is empty",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "").Return(relation.RelationV2{}, relation.ErrInvalidID)
			},
			request: &shieldv1beta1.GetRelationRequest{},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return not found error if id is not uuid",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "some-id").Return(relation.RelationV2{}, relation.ErrInvalidUUID)
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
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRelationV2.ID).Return(relation.RelationV2{}, relation.ErrNotExist)
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: testRelationV2.ID,
			},
			want:    nil,
			wantErr: grpcRelationNotFoundErr,
		},
		{
			name: "should return success if relation service return nil error",
			setup: func(rs *mocks.RelationService) {
				rs.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), testRelationV2.ID).Return(testRelationV2, nil)
			},
			request: &shieldv1beta1.GetRelationRequest{
				Id: testRelationV2.ID,
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

func TestHandler_DeleteRelation(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(rs *mocks.RelationService, res *mocks.ResourceService)
		request *shieldv1beta1.DeleteRelationRequest
		want    *shieldv1beta1.DeleteRelationResponse
		wantErr error
	}{
		{
			name: "should return internal server error when relation service returns some error while deletion",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					Name:        testRelationV2.Object.ID,
					NamespaceID: testRelationV2.Object.Namespace,
				}, permission.Permission{ID: schema.UpdatePermission}).Return(true, nil)

				rs.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), relation.RelationV2{
					Subject: relation.Subject{
						Namespace:       testRelationV2.Subject.Namespace,
						ID:              testRelationV2.Subject.ID,
						SubRelationName: testRelationV2.Subject.SubRelationName,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
				}).Return(nil)
			},
			request: &shieldv1beta1.DeleteRelationRequest{
				ObjectId:         testRelationV2.Object.ID,
				ObjectNamespace:  testRelationV2.Object.Namespace,
				SubjectId:        testRelationV2.Subject.ID,
				SubjectNamespace: testRelationV2.Subject.Namespace,
				Relation:         testRelationV2.Subject.SubRelationName,
			},
			want: &shieldv1beta1.DeleteRelationResponse{
				Message: "relation deleted",
			},
			wantErr: nil,
		},
		{
			name: "should successfully delete when relation exist and user has permission to edit it",
			setup: func(rs *mocks.RelationService, res *mocks.ResourceService) {
				res.EXPECT().CheckAuthz(mock.AnythingOfType("*context.emptyCtx"), resource.Resource{
					Name:        testRelationV2.Object.ID,
					NamespaceID: testRelationV2.Object.Namespace,
				}, permission.Permission{ID: schema.UpdatePermission}).Return(true, nil)

				rs.EXPECT().Delete(mock.AnythingOfType("*context.emptyCtx"), relation.RelationV2{
					Subject: relation.Subject{
						Namespace:       testRelationV2.Subject.Namespace,
						ID:              testRelationV2.Subject.ID,
						SubRelationName: testRelationV2.Subject.SubRelationName,
					},
					Object: relation.Object{
						ID:        testRelationV2.Object.ID,
						Namespace: testRelationV2.Object.Namespace,
					},
				}).Return(nil)
			},
			request: &shieldv1beta1.DeleteRelationRequest{
				ObjectId:         testRelationV2.Object.ID,
				ObjectNamespace:  testRelationV2.Object.Namespace,
				SubjectId:        testRelationV2.Subject.ID,
				SubjectNamespace: testRelationV2.Subject.Namespace,
				Relation:         testRelationV2.Subject.SubRelationName,
			},
			want: &shieldv1beta1.DeleteRelationResponse{
				Message: "relation deleted",
			},
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
			mockDep := Handler{relationService: mockRelationSrv, resourceService: mockResourceSrv}
			resp, err := mockDep.DeleteRelation(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
