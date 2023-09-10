package v1beta1

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestHandler_ListOrganizationAuditLogs(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.AuditService, os *mocks.OrganizationService)
		request *frontierv1beta1.ListOrganizationAuditLogsRequest
		want    *frontierv1beta1.ListOrganizationAuditLogsResponse
		wantErr error
	}{
		{
			name: "should return list of audit logs",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
				as.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), audit.Filter{
					OrgID:     testOrgMap[testOrgID].ID,
					Source:    "guardian-service",
					Action:    "project.create",
					StartTime: time.Time{},
					EndTime:   time.Time{},
				}).Return([]audit.Log{
					{
						ID:        "test-id",
						OrgID:     "org-id",
						Source:    "guardian-service",
						Action:    "project.create",
						CreatedAt: time.Time{},
						Actor: audit.Actor{
							ID:   "test-actor-id",
							Type: "user",
							Name: "test-actor-name",
						},
						Target: audit.Target{
							ID:   "test-target-id",
							Type: "",
							Name: "test-target-name",
						},
					},
				}, nil)
			},
			request: &frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId:     "org-id",
				Source:    "guardian-service",
				Action:    "project.create",
				StartTime: timestamppb.New(time.Time{}),
				EndTime:   timestamppb.New(time.Time{}),
			},
			want: &frontierv1beta1.ListOrganizationAuditLogsResponse{
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:        "test-id",
						Source:    "guardian-service",
						Action:    "project.create",
						CreatedAt: timestamppb.New(time.Time{}),
						Actor: &frontierv1beta1.AuditLogActor{
							Id:   "test-actor-id",
							Type: "user",
							Name: "test-actor-name",
						},
						Target: &frontierv1beta1.AuditLogTarget{
							Id:   "test-target-id",
							Type: "",
							Name: "test-target-name",
						},
					},
				},
			},
		},
		{
			name: "should return error when audit service returns error",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
				as.EXPECT().List(mock.AnythingOfType("*context.emptyCtx"), audit.Filter{
					OrgID:     testOrgMap[testOrgID].ID,
					Source:    "guardian-service",
					Action:    "project.create",
					StartTime: time.Time{},
					EndTime:   time.Time{},
				}).Return(nil, errors.New("test-error"))
			},
			request: &frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId:     "org-id",
				Source:    "guardian-service",
				Action:    "project.create",
				StartTime: timestamppb.New(time.Time{}),
				EndTime:   timestamppb.New(time.Time{}),
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
		{
			name: "should return error when org is disabled",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: &frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId:     "org-id",
				Source:    "guardian-service",
				Action:    "project.create",
				StartTime: timestamppb.New(time.Time{}),
				EndTime:   timestamppb.New(time.Time{}),
			},
			want:    nil,
			wantErr: grpcOrgDisabledErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuditSrv := new(mocks.AuditService)
			mockOrgSrv := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockAuditSrv, mockOrgSrv)
			}
			mockDep := Handler{auditService: mockAuditSrv, orgService: mockOrgSrv}
			resp, err := mockDep.ListOrganizationAuditLogs(context.Background(), tt.request)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_CreateOrganizationAuditLogs(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.AuditService, os *mocks.OrganizationService)
		req     *frontierv1beta1.CreateOrganizationAuditLogsRequest
		want    *frontierv1beta1.CreateOrganizationAuditLogsResponse
		wantErr error
	}{
		{
			name: "should create audit logs on success and return nil error",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), &audit.Log{
					ID:    "test-id",
					OrgID: testOrgMap[testOrgID].ID,

					Source:    "guardian-service",
					Action:    "project.create",
					CreatedAt: time.Time{},
					Actor: audit.Actor{
						ID:   "test-actor-id",
						Type: "user",
						Name: "test-actor-name",
					},
					Target: audit.Target{
						ID:   "test-target-id",
						Type: "project",
						Name: "test-target-name",
					},
				}).Return(nil)
			},
			req: &frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:        "test-id",
						Source:    "guardian-service",
						Action:    "project.create",
						CreatedAt: timestamppb.New(time.Time{}),
						Actor: &frontierv1beta1.AuditLogActor{
							Id:   "test-actor-id",
							Type: "user",
							Name: "test-actor-name",
						},
						Target: &frontierv1beta1.AuditLogTarget{
							Id:   "test-target-id",
							Type: "project",
							Name: "test-target-name",
						},
					},
				},
			},
			want:    &frontierv1beta1.CreateOrganizationAuditLogsResponse{},
			wantErr: nil,
		},
		{
			name: "should return error when log source and action is empty",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
			},
			req: &frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:        "test-id",
						Source:    "",
						Action:    "",
						CreatedAt: timestamppb.New(time.Time{}),
						Actor: &frontierv1beta1.AuditLogActor{
							Id:   "test-actor-id",
							Type: "user",
							Name: "test-actor-name",
						},
						Target: &frontierv1beta1.AuditLogTarget{
							Id:   "test-target-id",
							Type: "project",
							Name: "test-target-name",
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcBadBodyError,
		},
		{
			name: "should return error when audit service returns error",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
				as.EXPECT().Create(mock.AnythingOfType("*context.emptyCtx"), &audit.Log{
					ID:        "test-id",
					OrgID:     testOrgMap[testOrgID].ID,
					Source:    "guardian-service",
					Action:    "project.create",
					CreatedAt: time.Time{},
					Metadata: map[string]string{
						"test-key": "test-value",
					},
				}).Return(errors.New("test-error"))
			},
			req: &frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:        "test-id",
						Source:    "guardian-service",
						Action:    "project.create",
						CreatedAt: timestamppb.New(time.Time{}),
						Context: map[string]string{
							"test-key": "test-value",
						},
					},
				},
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockAuditSrv := new(mocks.AuditService)
			if tt.setup != nil {
				tt.setup(mockAuditSrv, mockOrgService)
			}
			mockDep := Handler{auditService: mockAuditSrv, orgService: mockOrgService}
			resp, err := mockDep.CreateOrganizationAuditLogs(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}

func TestHandler_GetOrganizationAuditLog(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(as *mocks.AuditService, os *mocks.OrganizationService)
		req     *frontierv1beta1.GetOrganizationAuditLogRequest
		want    *frontierv1beta1.GetOrganizationAuditLogResponse
		wantErr error
	}{
		{
			name: "should return audit log on success",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
				as.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), "test-id").Return(audit.Log{
					ID:        "test-id",
					OrgID:     testOrgMap[testOrgID].ID,
					Source:    "guardian-service",
					Action:    "project.create",
					CreatedAt: time.Time{},
					Actor: audit.Actor{
						ID:   "test-actor-id",
						Type: "user",
						Name: "test-actor-name",
					},
					Target: audit.Target{
						ID:   "test-target-id",
						Type: "project",
						Name: "test-target-name",
					},
					Metadata: map[string]string{
						"test-key": "test-value",
					},
				}, nil)
			},
			req: &frontierv1beta1.GetOrganizationAuditLogRequest{
				Id:    "test-id",
				OrgId: "org-id",
			},
			want: &frontierv1beta1.GetOrganizationAuditLogResponse{
				Log: &frontierv1beta1.AuditLog{
					Id:        "test-id",
					Source:    "guardian-service",
					Action:    "project.create",
					CreatedAt: timestamppb.New(time.Time{}),
					Actor: &frontierv1beta1.AuditLogActor{
						Id:   "test-actor-id",
						Type: "user",
						Name: "test-actor-name",
					},
					Target: &frontierv1beta1.AuditLogTarget{
						Id:   "test-target-id",
						Type: "project",
						Name: "test-target-name",
					},
					Context: map[string]string{
						"test-key": "test-value",
					},
				},
			},
			wantErr: nil,
		},
		{
			name: "should return error when audit service returns error",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.AnythingOfType("*context.emptyCtx"), "org-id").Return(testOrgMap[testOrgID], nil)
				as.EXPECT().GetByID(mock.AnythingOfType("*context.emptyCtx"), "test-id").Return(audit.Log{}, errors.New("test-error"))
			},
			req: &frontierv1beta1.GetOrganizationAuditLogRequest{
				Id:    "test-id",
				OrgId: "org-id",
			},
			want:    nil,
			wantErr: grpcInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockOrgService := new(mocks.OrganizationService)
			mockAuditSrv := new(mocks.AuditService)
			if tt.setup != nil {
				tt.setup(mockAuditSrv, mockOrgService)
			}
			mockDep := Handler{auditService: mockAuditSrv, orgService: mockOrgService}
			resp, err := mockDep.GetOrganizationAuditLog(context.Background(), tt.req)
			assert.EqualValues(t, tt.want, resp)
			assert.EqualValues(t, tt.wantErr, err)
		})
	}
}
