package v1beta1connect

import (
	"context"
	"errors"
	"testing"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/relation"
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
			ID:        "subject-id",
			Namespace: "ns1",
		},
		Object: relation.Object{
			ID:        "object-id",
			Namespace: "ns2",
		},
		RelationName: "relation1",
	}

	testRelationPB = &frontierv1beta1.Relation{
		Id:       "relation-id-1",
		Object:   schema.JoinNamespaceAndResourceID("ns2", "object-id"),
		Subject:  schema.JoinNamespaceAndResourceID("ns1", "subject-id"),
		Relation: "relation1",
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
