package v1beta1connect

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/internal/api/v1beta1connect/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_CreateAuditRecord(t *testing.T) {
	testTime := time.Now()
	testUUID := uuid.New().String()
	testOrgID := uuid.New().String()
	testRequestID := "req-123"
	testIdempotencyKey := uuid.New().String()

	tests := []struct {
		name         string
		setup        func(ars *mocks.AuditRecordService)
		request      *connect.Request[frontierv1beta1.CreateAuditRecordRequest]
		want         *connect.Response[frontierv1beta1.CreateAuditRecordResponse]
		wantErr      error
		checkHeaders bool
		wantHeader   string
	}{
		{
			name:  "should return invalid argument error when request validation fails",
			setup: func(ars *mocks.AuditRecordService) {},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "", // Empty event should fail validation
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "project",
					Name: "test-project",
				},
				OccurredAt: timestamppb.New(testTime),
				OrgId:      testOrgID,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("invalid CreateAuditRecordRequest.Event: value length must be at least 3 runes")),
		},
		{
			name:  "should return invalid argument error for invalid actor type",
			setup: func(ars *mocks.AuditRecordService) {},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "user.created",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: "invalid-type", // Invalid actor type
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "project",
					Name: "test-project",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: uuid.Nil.String(), // Use zero UUID to pass validation
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, ErrInvalidActorType),
		},
		{
			name: "should allow system actor with ZeroUUID",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "system.maintenance",
					Actor: auditrecord.Actor{
						ID:       uuid.Nil.String(),
						Type:     "system",
						Name:     "system",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "cluster",
						Name:     "main-cluster",
						Metadata: metadata.Metadata{},
					},
					Target:         nil,
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: uuid.Nil.String(),
				}

				returnedRecord := expectedRecord
				returnedRecord.ID = testUUID
				returnedRecord.CreatedAt = testTime

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(returnedRecord, false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "system.maintenance",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   uuid.Nil.String(),
					Type: "system",
					Name: "system",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "cluster",
					Name: "main-cluster",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: uuid.Nil.String(), // Use zero UUID to pass validation
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "system.maintenance",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       uuid.Nil.String(),
						Type:     "system",
						Name:     "system",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "cluster",
						Name:     "main-cluster",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					OrgId:      testOrgID,
					CreatedAt:  timestamppb.New(testTime),
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should create audit record successfully with minimal fields",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "user.created",
					Actor: auditrecord.Actor{
						ID:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "project",
						Name:     "test-project",
						Metadata: metadata.Metadata{},
					},
					Target:         nil,
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: uuid.Nil.String(),
				}

				returnedRecord := expectedRecord
				returnedRecord.ID = testUUID
				returnedRecord.CreatedAt = testTime

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(returnedRecord, false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "user.created",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "project",
					Name: "test-project",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: uuid.Nil.String(), // Use zero UUID to pass validation
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "user.created",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "project",
						Name:     "test-project",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					OrgId:      testOrgID,
					CreatedAt:  timestamppb.New(testTime),
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should create audit record with target",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "permission.granted",
					Actor: auditrecord.Actor{
						ID:       testUUID,
						Type:     schema.ServiceUserPrincipal,
						Name:     "service-user",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "role",
						Name:     "admin-role",
						Metadata: metadata.Metadata{},
					},
					Target: &auditrecord.Target{
						ID:   "target-user-123",
						Type: "user",
						Name: "target-user",
						Metadata: metadata.Metadata{
							"email": "user@example.com",
						},
					},
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: uuid.Nil.String(),
				}

				returnedRecord := expectedRecord
				returnedRecord.ID = testUUID
				returnedRecord.CreatedAt = testTime

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(returnedRecord, false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "permission.granted",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.ServiceUserPrincipal,
					Name: "service-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "role",
					Name: "admin-role",
				},
				Target: &frontierv1beta1.AuditRecordTarget{
					Id:   "target-user-123",
					Type: "user",
					Name: "target-user",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"email": {Kind: &structpb.Value_StringValue{StringValue: "user@example.com"}},
						},
					},
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: uuid.Nil.String(), // Use zero UUID to pass validation
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "permission.granted",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       testUUID,
						Type:     schema.ServiceUserPrincipal,
						Name:     "service-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "role",
						Name:     "admin-role",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Target: &frontierv1beta1.AuditRecordTarget{
						Id:   "target-user-123",
						Type: "user",
						Name: "target-user",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"email": {Kind: &structpb.Value_StringValue{StringValue: "user@example.com"}},
							},
						},
					},
					OccurredAt: timestamppb.New(testTime),
					OrgId:      testOrgID,
					CreatedAt:  timestamppb.New(testTime),
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should create audit record with request ID",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "api.called",
					Actor: auditrecord.Actor{
						ID:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "api",
						Name:     "create-project",
						Metadata: metadata.Metadata{},
					},
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					RequestID:      &testRequestID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: uuid.Nil.String(),
				}

				returnedRecord := expectedRecord
				returnedRecord.ID = testUUID
				returnedRecord.CreatedAt = testTime

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(returnedRecord, false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "api.called",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "api",
					Name: "create-project",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				RequestId:      testRequestID,
				IdempotencyKey: uuid.Nil.String(),
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "api.called",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "api",
						Name:     "create-project",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					OrgId:      testOrgID,
					RequestId:  testRequestID,
					CreatedAt:  timestamppb.New(testTime),
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			}),
			wantErr: nil,
		},
		{
			name: "should handle idempotency key and set header when replayed",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "user.updated",
					Actor: auditrecord.Actor{
						ID:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "user",
						Name:     "updated-user",
						Metadata: metadata.Metadata{},
					},
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: testIdempotencyKey,
				}

				returnedRecord := expectedRecord
				returnedRecord.ID = testUUID
				returnedRecord.CreatedAt = testTime

				// Return true for idempotentReply to indicate this is a replayed request
				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(returnedRecord, true, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "user.updated",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "user",
					Name: "updated-user",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: testIdempotencyKey,
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "user.updated",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "user",
						Name:     "updated-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					OrgId:      testOrgID,
					CreatedAt:  timestamppb.New(testTime),
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			}),
			wantErr:      nil,
			checkHeaders: true,
			wantHeader:   "true",
		},
		{
			name: "should return already exists error for idempotency key conflict",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "user.deleted",
					Actor: auditrecord.Actor{
						ID:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "user",
						Name:     "deleted-user",
						Metadata: metadata.Metadata{},
					},
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: testIdempotencyKey,
				}

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(auditrecord.AuditRecord{}, false, auditrecord.ErrIdempotencyKeyConflict)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "user.deleted",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "user",
					Name: "deleted-user",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: testIdempotencyKey,
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeAlreadyExists, auditrecord.ErrIdempotencyKeyConflict),
		},
		{
			name: "should return error for service failure",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "user.created",
					Actor: auditrecord.Actor{
						ID:       testUUID,
						Type:     schema.UserPrincipal,
						Name:     "test-user",
						Metadata: metadata.Metadata{},
					},
					Resource: auditrecord.Resource{
						ID:       "resource-123",
						Type:     "project",
						Name:     "test-project",
						Metadata: metadata.Metadata{},
					},
					OccurredAt:     testTime,
					OrgID:          testOrgID,
					Metadata:       metadata.Metadata{},
					IdempotencyKey: uuid.Nil.String(),
				}

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(auditrecord.AuditRecord{}, false, errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "user.created",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "project",
					Name: "test-project",
				},
				OccurredAt:     timestamppb.New(testTime),
				OrgId:          testOrgID,
				IdempotencyKey: uuid.Nil.String(),
			}),
			want:    nil,
			wantErr: errors.New("database error"),
		},
		{
			name: "should create audit record with all metadata fields",
			setup: func(ars *mocks.AuditRecordService) {
				expectedRecord := auditrecord.AuditRecord{
					Event: "complex.event",
					Actor: auditrecord.Actor{
						ID:   testUUID,
						Type: schema.UserPrincipal,
						Name: "test-user",
						Metadata: metadata.Metadata{
							"role": "admin",
							"ip":   "192.168.1.1",
						},
					},
					Resource: auditrecord.Resource{
						ID:   "resource-123",
						Type: "project",
						Name: "test-project",
						Metadata: metadata.Metadata{
							"version": "1.0",
							"owner":   "team-a",
						},
					},
					Target: &auditrecord.Target{
						ID:   "target-123",
						Type: "permission",
						Name: "read-write",
						Metadata: metadata.Metadata{
							"scope": "global",
						},
					},
					OccurredAt: testTime,
					OrgID:      testOrgID,
					RequestID:  &testRequestID,
					Metadata: metadata.Metadata{
						"action": "grant",
						"reason": "promotion",
					},
					IdempotencyKey: testIdempotencyKey,
				}

				returnedRecord := expectedRecord
				returnedRecord.ID = testUUID
				returnedRecord.CreatedAt = testTime

				ars.EXPECT().Create(mock.AnythingOfType("context.backgroundCtx"), mock.MatchedBy(func(r auditrecord.AuditRecord) bool {
					return r.Event == expectedRecord.Event &&
						reflect.DeepEqual(r.Actor, expectedRecord.Actor) &&
						reflect.DeepEqual(r.Resource, expectedRecord.Resource) &&
						reflect.DeepEqual(r.Target, expectedRecord.Target) &&
						r.OrgID == expectedRecord.OrgID &&
						r.IdempotencyKey == expectedRecord.IdempotencyKey &&
						r.OccurredAt.Unix() == expectedRecord.OccurredAt.Unix()
				})).Return(returnedRecord, false, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateAuditRecordRequest{
				Event: "complex.event",
				Actor: &frontierv1beta1.AuditRecordActor{
					Id:   testUUID,
					Type: schema.UserPrincipal,
					Name: "test-user",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"role": {Kind: &structpb.Value_StringValue{StringValue: "admin"}},
							"ip":   {Kind: &structpb.Value_StringValue{StringValue: "192.168.1.1"}},
						},
					},
				},
				Resource: &frontierv1beta1.AuditRecordResource{
					Id:   "resource-123",
					Type: "project",
					Name: "test-project",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"version": {Kind: &structpb.Value_StringValue{StringValue: "1.0"}},
							"owner":   {Kind: &structpb.Value_StringValue{StringValue: "team-a"}},
						},
					},
				},
				Target: &frontierv1beta1.AuditRecordTarget{
					Id:   "target-123",
					Type: "permission",
					Name: "read-write",
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"scope": {Kind: &structpb.Value_StringValue{StringValue: "global"}},
						},
					},
				},
				OccurredAt: timestamppb.New(testTime),
				OrgId:      testOrgID,
				RequestId:  testRequestID,
				Metadata: &structpb.Struct{
					Fields: map[string]*structpb.Value{
						"action": {Kind: &structpb.Value_StringValue{StringValue: "grant"}},
						"reason": {Kind: &structpb.Value_StringValue{StringValue: "promotion"}},
					},
				},
				IdempotencyKey: testIdempotencyKey,
			}),
			want: connect.NewResponse(&frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "complex.event",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:   testUUID,
						Type: schema.UserPrincipal,
						Name: "test-user",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"role": {Kind: &structpb.Value_StringValue{StringValue: "admin"}},
								"ip":   {Kind: &structpb.Value_StringValue{StringValue: "192.168.1.1"}},
							},
						},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:   "resource-123",
						Type: "project",
						Name: "test-project",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"version": {Kind: &structpb.Value_StringValue{StringValue: "1.0"}},
								"owner":   {Kind: &structpb.Value_StringValue{StringValue: "team-a"}},
							},
						},
					},
					Target: &frontierv1beta1.AuditRecordTarget{
						Id:   "target-123",
						Type: "permission",
						Name: "read-write",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"scope": {Kind: &structpb.Value_StringValue{StringValue: "global"}},
							},
						},
					},
					OccurredAt: timestamppb.New(testTime),
					OrgId:      testOrgID,
					RequestId:  testRequestID,
					CreatedAt:  timestamppb.New(testTime),
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"action": {Kind: &structpb.Value_StringValue{StringValue: "grant"}},
							"reason": {Kind: &structpb.Value_StringValue{StringValue: "promotion"}},
						},
					},
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuditRecordSrv := new(mocks.AuditRecordService)
			if tt.setup != nil {
				tt.setup(mockAuditRecordSrv)
			}
			handler := &ConnectHandler{auditRecordService: mockAuditRecordSrv}
			resp, err := handler.CreateAuditRecord(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.wantErr.Error(), err.Error())
			} else {
				assert.NoError(t, err)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want.Msg, resp.Msg)

				if tt.checkHeaders {
					assert.Equal(t, tt.wantHeader, resp.Header().Get(IdempotencyReplyHeader))
				}
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}

func TestTransformAuditRecordToPB(t *testing.T) {
	testTime := time.Now()
	testUUID := uuid.New().String()
	testRequestID := "req-123"

	tests := []struct {
		name    string
		record  auditrecord.AuditRecord
		want    *frontierv1beta1.CreateAuditRecordResponse
		wantErr bool
	}{
		{
			name: "should transform minimal audit record",
			record: auditrecord.AuditRecord{
				ID:    testUUID,
				Event: "user.created",
				Actor: auditrecord.Actor{
					ID:   "actor-123",
					Type: "user",
					Name: "test-user",
				},
				Resource: auditrecord.Resource{
					ID:   "resource-123",
					Type: "project",
					Name: "test-project",
				},
				OccurredAt: testTime,
				CreatedAt:  testTime,
				OrgID:      "org-123",
			},
			want: &frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "user.created",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       "actor-123",
						Type:     "user",
						Name:     "test-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "project",
						Name:     "test-project",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					CreatedAt:  timestamppb.New(testTime),
					OrgId:      "org-123",
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			},
			wantErr: false,
		},
		{
			name: "should transform audit record with target",
			record: auditrecord.AuditRecord{
				ID:    testUUID,
				Event: "permission.granted",
				Actor: auditrecord.Actor{
					ID:   "actor-123",
					Type: "user",
					Name: "test-user",
				},
				Resource: auditrecord.Resource{
					ID:   "resource-123",
					Type: "role",
					Name: "admin-role",
				},
				Target: &auditrecord.Target{
					ID:   "target-123",
					Type: "user",
					Name: "target-user",
				},
				OccurredAt: testTime,
				CreatedAt:  testTime,
				OrgID:      "org-123",
			},
			want: &frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "permission.granted",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       "actor-123",
						Type:     "user",
						Name:     "test-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "role",
						Name:     "admin-role",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Target: &frontierv1beta1.AuditRecordTarget{
						Id:       "target-123",
						Type:     "user",
						Name:     "target-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					CreatedAt:  timestamppb.New(testTime),
					OrgId:      "org-123",
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			},
			wantErr: false,
		},
		{
			name: "should transform audit record with request ID",
			record: auditrecord.AuditRecord{
				ID:    testUUID,
				Event: "api.called",
				Actor: auditrecord.Actor{
					ID:   "actor-123",
					Type: "user",
					Name: "test-user",
				},
				Resource: auditrecord.Resource{
					ID:   "resource-123",
					Type: "api",
					Name: "create-project",
				},
				OccurredAt: testTime,
				CreatedAt:  testTime,
				OrgID:      "org-123",
				RequestID:  &testRequestID,
			},
			want: &frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "api.called",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:       "actor-123",
						Type:     "user",
						Name:     "test-user",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:       "resource-123",
						Type:     "api",
						Name:     "create-project",
						Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					OccurredAt: timestamppb.New(testTime),
					CreatedAt:  timestamppb.New(testTime),
					OrgId:      "org-123",
					RequestId:  testRequestID,
					Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
				},
			},
			wantErr: false,
		},
		{
			name: "should transform audit record with all metadata",
			record: auditrecord.AuditRecord{
				ID:    testUUID,
				Event: "complex.event",
				Actor: auditrecord.Actor{
					ID:   "actor-123",
					Type: "user",
					Name: "test-user",
					Metadata: metadata.Metadata{
						"role": "admin",
					},
				},
				Resource: auditrecord.Resource{
					ID:   "resource-123",
					Type: "project",
					Name: "test-project",
					Metadata: metadata.Metadata{
						"version": "1.0",
					},
				},
				Target: &auditrecord.Target{
					ID:   "target-123",
					Type: "permission",
					Name: "read-write",
					Metadata: metadata.Metadata{
						"scope": "global",
					},
				},
				OccurredAt: testTime,
				CreatedAt:  testTime,
				OrgID:      "org-123",
				RequestID:  &testRequestID,
				Metadata: metadata.Metadata{
					"action": "grant",
				},
			},
			want: &frontierv1beta1.CreateAuditRecordResponse{
				AuditRecord: &frontierv1beta1.AuditRecord{
					Id:    testUUID,
					Event: "complex.event",
					Actor: &frontierv1beta1.AuditRecordActor{
						Id:   "actor-123",
						Type: "user",
						Name: "test-user",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"role": {Kind: &structpb.Value_StringValue{StringValue: "admin"}},
							},
						},
					},
					Resource: &frontierv1beta1.AuditRecordResource{
						Id:   "resource-123",
						Type: "project",
						Name: "test-project",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"version": {Kind: &structpb.Value_StringValue{StringValue: "1.0"}},
							},
						},
					},
					Target: &frontierv1beta1.AuditRecordTarget{
						Id:   "target-123",
						Type: "permission",
						Name: "read-write",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"scope": {Kind: &structpb.Value_StringValue{StringValue: "global"}},
							},
						},
					},
					OccurredAt: timestamppb.New(testTime),
					CreatedAt:  timestamppb.New(testTime),
					OrgId:      "org-123",
					RequestId:  testRequestID,
					Metadata: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"action": {Kind: &structpb.Value_StringValue{StringValue: "grant"}},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := TransformAuditRecordToPB(tt.record)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestHandler_ListAuditRecords(t *testing.T) {
	testTime := time.Now()
	testUUID1 := uuid.New().String()
	testUUID2 := uuid.New().String()
	testOrgID := uuid.New().String()

	tests := []struct {
		name    string
		setup   func(ars *mocks.AuditRecordService)
		request *connect.Request[frontierv1beta1.ListAuditRecordsRequest]
		want    *connect.Response[frontierv1beta1.ListAuditRecordsResponse]
		wantErr error
	}{
		{
			name:  "should return invalid argument error for invalid RQL query",
			setup: func(ars *mocks.AuditRecordService) {},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Filters: []*frontierv1beta1.RQLFilter{
						{
							Name:     "invalid_field",
							Operator: "eq",
							Value: &frontierv1beta1.RQLFilter_StringValue{
								StringValue: "value",
							},
						},
					},
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, errors.New("failed to read rql query")),
		},
		{
			name: "should list audit records successfully with basic query",
			setup: func(ars *mocks.AuditRecordService) {
				auditRecordsList := auditrecord.AuditRecordsList{
					AuditRecords: []auditrecord.AuditRecord{
						{
							ID:    testUUID1,
							Event: "user.created",
							Actor: auditrecord.Actor{
								ID:       "actor-123",
								Type:     "user",
								Name:     "test-user",
								Metadata: metadata.Metadata{},
							},
							Resource: auditrecord.Resource{
								ID:       "resource-123",
								Type:     "project",
								Name:     "test-project",
								Metadata: metadata.Metadata{},
							},
							OccurredAt: testTime,
							CreatedAt:  testTime,
							OrgID:      testOrgID,
							Metadata:   metadata.Metadata{},
						},
					},
					Page: utils.Page{
						Offset:     0,
						Limit:      10,
						TotalCount: 1,
					},
					Group: nil,
				}

				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditRecordsList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Filters: []*frontierv1beta1.RQLFilter{
						{
							Name:     "event",
							Operator: "eq",
							Value: &frontierv1beta1.RQLFilter_StringValue{
								StringValue: "user.created",
							},
						},
					},
					Limit:  10,
					Offset: 0,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListAuditRecordsResponse{
				AuditRecords: []*frontierv1beta1.AuditRecord{
					{
						Id:    testUUID1,
						Event: "user.created",
						Actor: &frontierv1beta1.AuditRecordActor{
							Id:       "actor-123",
							Type:     "user",
							Name:     "test-user",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						Resource: &frontierv1beta1.AuditRecordResource{
							Id:       "resource-123",
							Type:     "project",
							Name:     "test-project",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						OccurredAt: timestamppb.New(testTime),
						CreatedAt:  timestamppb.New(testTime),
						OrgId:      testOrgID,
						Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
				Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
					Offset:     0,
					Limit:      10,
					TotalCount: 1,
				},
			}),
			wantErr: nil,
		},
		{
			name: "should list multiple audit records successfully",
			setup: func(ars *mocks.AuditRecordService) {
				auditRecordsList := auditrecord.AuditRecordsList{
					AuditRecords: []auditrecord.AuditRecord{
						{
							ID:    testUUID1,
							Event: "user.created",
							Actor: auditrecord.Actor{
								ID:       "actor-123",
								Type:     "user",
								Name:     "test-user-1",
								Metadata: metadata.Metadata{},
							},
							Resource: auditrecord.Resource{
								ID:       "resource-123",
								Type:     "project",
								Name:     "test-project-1",
								Metadata: metadata.Metadata{},
							},
							OccurredAt: testTime,
							CreatedAt:  testTime,
							OrgID:      testOrgID,
							Metadata:   metadata.Metadata{},
						},
						{
							ID:    testUUID2,
							Event: "user.updated",
							Actor: auditrecord.Actor{
								ID:       "actor-456",
								Type:     "service",
								Name:     "test-service",
								Metadata: metadata.Metadata{},
							},
							Resource: auditrecord.Resource{
								ID:       "resource-456",
								Type:     "user",
								Name:     "updated-user",
								Metadata: metadata.Metadata{},
							},
							OccurredAt: testTime.Add(time.Hour),
							CreatedAt:  testTime.Add(time.Hour),
							OrgID:      testOrgID,
							Metadata:   metadata.Metadata{},
						},
					},
					Page: utils.Page{
						Offset:     0,
						Limit:      10,
						TotalCount: 2,
					},
					Group: nil,
				}

				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditRecordsList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Limit:  10,
					Offset: 0,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListAuditRecordsResponse{
				AuditRecords: []*frontierv1beta1.AuditRecord{
					{
						Id:    testUUID1,
						Event: "user.created",
						Actor: &frontierv1beta1.AuditRecordActor{
							Id:       "actor-123",
							Type:     "user",
							Name:     "test-user-1",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						Resource: &frontierv1beta1.AuditRecordResource{
							Id:       "resource-123",
							Type:     "project",
							Name:     "test-project-1",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						OccurredAt: timestamppb.New(testTime),
						CreatedAt:  timestamppb.New(testTime),
						OrgId:      testOrgID,
						Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
					{
						Id:    testUUID2,
						Event: "user.updated",
						Actor: &frontierv1beta1.AuditRecordActor{
							Id:       "actor-456",
							Type:     "service",
							Name:     "test-service",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						Resource: &frontierv1beta1.AuditRecordResource{
							Id:       "resource-456",
							Type:     "user",
							Name:     "updated-user",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						OccurredAt: timestamppb.New(testTime.Add(time.Hour)),
						CreatedAt:  timestamppb.New(testTime.Add(time.Hour)),
						OrgId:      testOrgID,
						Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
				Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
					Offset:     0,
					Limit:      10,
					TotalCount: 2,
				},
			}),
			wantErr: nil,
		},
		{
			name: "should list audit records with group data",
			setup: func(ars *mocks.AuditRecordService) {
				auditRecordsList := auditrecord.AuditRecordsList{
					AuditRecords: []auditrecord.AuditRecord{
						{
							ID:    testUUID1,
							Event: "user.created",
							Actor: auditrecord.Actor{
								ID:       "actor-123",
								Type:     "user",
								Name:     "test-user",
								Metadata: metadata.Metadata{},
							},
							Resource: auditrecord.Resource{
								ID:       "resource-123",
								Type:     "project",
								Name:     "test-project",
								Metadata: metadata.Metadata{},
							},
							OccurredAt: testTime,
							CreatedAt:  testTime,
							OrgID:      testOrgID,
							Metadata:   metadata.Metadata{},
						},
					},
					Page: utils.Page{
						Offset:     0,
						Limit:      10,
						TotalCount: 1,
					},
					Group: &utils.Group{
						Name: "event",
						Data: []utils.GroupData{
							{Name: "user.created", Count: 5},
							{Name: "user.updated", Count: 3},
						},
					},
				}

				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditRecordsList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					GroupBy: []string{"event"},
					Limit:   10,
					Offset:  0,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListAuditRecordsResponse{
				AuditRecords: []*frontierv1beta1.AuditRecord{
					{
						Id:    testUUID1,
						Event: "user.created",
						Actor: &frontierv1beta1.AuditRecordActor{
							Id:       "actor-123",
							Type:     "user",
							Name:     "test-user",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						Resource: &frontierv1beta1.AuditRecordResource{
							Id:       "resource-123",
							Type:     "project",
							Name:     "test-project",
							Metadata: &structpb.Struct{Fields: map[string]*structpb.Value{}},
						},
						OccurredAt: timestamppb.New(testTime),
						CreatedAt:  timestamppb.New(testTime),
						OrgId:      testOrgID,
						Metadata:   &structpb.Struct{Fields: map[string]*structpb.Value{}},
					},
				},
				Group: &frontierv1beta1.RQLQueryGroupResponse{
					Name: "event",
					Data: []*frontierv1beta1.RQLQueryGroupData{
						{Name: "user.created", Count: 5},
						{Name: "user.updated", Count: 3},
					},
				},
				Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
					Offset:     0,
					Limit:      10,
					TotalCount: 1,
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return empty list when no records found",
			setup: func(ars *mocks.AuditRecordService) {
				auditRecordsList := auditrecord.AuditRecordsList{
					AuditRecords: []auditrecord.AuditRecord{},
					Page: utils.Page{
						Offset:     0,
						Limit:      10,
						TotalCount: 0,
					},
					Group: nil,
				}

				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditRecordsList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Filters: []*frontierv1beta1.RQLFilter{
						{
							Name:     "event",
							Operator: "eq",
							Value: &frontierv1beta1.RQLFilter_StringValue{
								StringValue: "nonexistent.event",
							},
						},
					},
					Limit:  10,
					Offset: 0,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListAuditRecordsResponse{
				AuditRecords: []*frontierv1beta1.AuditRecord{},
				Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
					Offset:     0,
					Limit:      10,
					TotalCount: 0,
				},
			}),
			wantErr: nil,
		},
		{
			name: "should return internal error for database failure",
			setup: func(ars *mocks.AuditRecordService) {
				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditrecord.AuditRecordsList{}, errors.New("database connection failed"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Filters: []*frontierv1beta1.RQLFilter{
						{
							Name:     "event",
							Operator: "eq",
							Value: &frontierv1beta1.RQLFilter_StringValue{
								StringValue: "user.created",
							},
						},
					},
					Limit:  10,
					Offset: 0,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInternal, ErrInternalServerError),
		},
		{
			name: "should return invalid argument error for bad input",
			setup: func(ars *mocks.AuditRecordService) {
				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditrecord.AuditRecordsList{}, auditrecord.ErrRepositoryBadInput)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Filters: []*frontierv1beta1.RQLFilter{
						{
							Name:     "invalid_field",
							Operator: "eq",
							Value: &frontierv1beta1.RQLFilter_StringValue{
								StringValue: "value",
							},
						},
					},
					Limit:  10,
					Offset: 0,
				},
			}),
			want:    nil,
			wantErr: connect.NewError(connect.CodeInvalidArgument, auditrecord.ErrRepositoryBadInput),
		},
		{
			name: "should handle audit record with target and request ID",
			setup: func(ars *mocks.AuditRecordService) {
				testRequestID := "req-123"
				auditRecordsList := auditrecord.AuditRecordsList{
					AuditRecords: []auditrecord.AuditRecord{
						{
							ID:    testUUID1,
							Event: "permission.granted",
							Actor: auditrecord.Actor{
								ID:   "actor-123",
								Type: "user",
								Name: "test-user",
								Metadata: metadata.Metadata{
									"role": "admin",
								},
							},
							Resource: auditrecord.Resource{
								ID:   "resource-123",
								Type: "role",
								Name: "admin-role",
								Metadata: metadata.Metadata{
									"permissions": "read,write",
								},
							},
							Target: &auditrecord.Target{
								ID:   "target-123",
								Type: "user",
								Name: "target-user",
								Metadata: metadata.Metadata{
									"email": "target@example.com",
								},
							},
							OccurredAt: testTime,
							CreatedAt:  testTime,
							OrgID:      testOrgID,
							RequestID:  &testRequestID,
							Metadata: metadata.Metadata{
								"action": "grant",
							},
						},
					},
					Page: utils.Page{
						Offset:     0,
						Limit:      10,
						TotalCount: 1,
					},
					Group: nil,
				}

				ars.EXPECT().List(mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("*rql.Query")).
					Return(auditRecordsList, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListAuditRecordsRequest{
				Query: &frontierv1beta1.RQLRequest{
					Filters: []*frontierv1beta1.RQLFilter{
						{
							Name:     "event",
							Operator: "eq",
							Value: &frontierv1beta1.RQLFilter_StringValue{
								StringValue: "permission.granted",
							},
						},
					},
					Limit:  10,
					Offset: 0,
				},
			}),
			want: connect.NewResponse(&frontierv1beta1.ListAuditRecordsResponse{
				AuditRecords: []*frontierv1beta1.AuditRecord{
					{
						Id:    testUUID1,
						Event: "permission.granted",
						Actor: &frontierv1beta1.AuditRecordActor{
							Id:   "actor-123",
							Type: "user",
							Name: "test-user",
							Metadata: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"role": {Kind: &structpb.Value_StringValue{StringValue: "admin"}},
								},
							},
						},
						Resource: &frontierv1beta1.AuditRecordResource{
							Id:   "resource-123",
							Type: "role",
							Name: "admin-role",
							Metadata: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"permissions": {Kind: &structpb.Value_StringValue{StringValue: "read,write"}},
								},
							},
						},
						Target: &frontierv1beta1.AuditRecordTarget{
							Id:   "target-123",
							Type: "user",
							Name: "target-user",
							Metadata: &structpb.Struct{
								Fields: map[string]*structpb.Value{
									"email": {Kind: &structpb.Value_StringValue{StringValue: "target@example.com"}},
								},
							},
						},
						OccurredAt: timestamppb.New(testTime),
						CreatedAt:  timestamppb.New(testTime),
						OrgId:      testOrgID,
						RequestId:  "req-123",
						Metadata: &structpb.Struct{
							Fields: map[string]*structpb.Value{
								"action": {Kind: &structpb.Value_StringValue{StringValue: "grant"}},
							},
						},
					},
				},
				Pagination: &frontierv1beta1.RQLQueryPaginationResponse{
					Offset:     0,
					Limit:      10,
					TotalCount: 1,
				},
			}),
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuditRecordSrv := new(mocks.AuditRecordService)
			if tt.setup != nil {
				tt.setup(mockAuditRecordSrv)
			}
			handler := &ConnectHandler{auditRecordService: mockAuditRecordSrv}
			resp, err := handler.ListAuditRecords(context.Background(), tt.request)

			if tt.wantErr != nil {
				assert.Error(t, err)
				// Check error codes match for 'connect' errors
				var connectErr *connect.Error
				if errors.As(tt.wantErr, &connectErr) {
					var respErr *connect.Error
					if errors.As(err, &respErr) {
						assert.Equal(t, connectErr.Code(), respErr.Code())
					}
				}
			} else {
				assert.NoError(t, err)
			}

			if tt.want != nil {
				assert.Equal(t, tt.want.Msg, resp.Msg)
			} else {
				assert.Nil(t, resp)
			}
		})
	}
}
