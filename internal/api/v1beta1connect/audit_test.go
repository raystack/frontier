package v1beta1connect

import (
	"context"
	"errors"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/raystack/frontier/core/audit"
	"github.com/raystack/frontier/core/organization"
	"github.com/raystack/frontier/internal/api/v1beta1/mocks"
	frontierv1beta1 "github.com/raystack/frontier/proto/v1beta1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	testOrg = organization.Organization{
		ID:    "test-org-id",
		Name:  "test-org",
		State: organization.Enabled,
	}
)

func TestConnectHandler_ListOrganizationAuditLogs(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(as *mocks.AuditService, os *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.ListOrganizationAuditLogsRequest]
		wantErr     bool
		wantErrCode connect.Code
		wantCount   int
	}{
		{
			name: "should return list of audit logs successfully",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().List(mock.Anything, audit.Filter{
					OrgID:        testOrg.ID,
					Source:       "guardian-service",
					Action:       "project.create",
					StartTime:    time.Time{},
					EndTime:      time.Time{},
					IgnoreSystem: false,
				}).Return([]audit.Log{
					{
						ID:        "test-id",
						OrgID:     "test-org-id",
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
							"key": "value",
						},
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId:        "test-org-id",
				Source:       "guardian-service",
				Action:       "project.create",
				StartTime:    timestamppb.New(time.Time{}),
				EndTime:      timestamppb.New(time.Time{}),
				IgnoreSystem: false,
			}),
			wantErr:   false,
			wantCount: 1,
		},
		{
			name: "should return empty list when no logs found",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().List(mock.Anything, mock.AnythingOfType("audit.Filter")).Return([]audit.Log{}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId:  "test-org-id",
				Source: "test-source",
			}),
			wantErr:   false,
			wantCount: 0,
		},
		{
			name: "should return error when organization is disabled",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "disabled-org").Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId: "disabled-org",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeFailedPrecondition,
		},
		{
			name: "should return error when organization not found",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "nonexistent-org").Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId: "nonexistent-org",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
		},
		{
			name: "should return internal error when audit service fails",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().List(mock.Anything, mock.AnythingOfType("audit.Filter")).Return(nil, errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.ListOrganizationAuditLogsRequest{
				OrgId: "test-org-id",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuditSvc := new(mocks.AuditService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockAuditSvc, mockOrgSvc)
			}

			handler := &ConnectHandler{
				auditService: mockAuditSvc,
				orgService:   mockOrgSvc,
			}

			resp, err := handler.ListOrganizationAuditLogs(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					var connectErr *connect.Error
					if assert.True(t, errors.As(err, &connectErr)) {
						assert.Equal(t, tt.wantErrCode, connectErr.Code())
					}
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Len(t, resp.Msg.GetLogs(), tt.wantCount)
			}
		})
	}
}

func TestConnectHandler_CreateOrganizationAuditLogs(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(as *mocks.AuditService, os *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.CreateOrganizationAuditLogsRequest]
		wantErr     bool
		wantErrCode connect.Code
	}{
		{
			name: "should create audit logs successfully",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().Create(mock.Anything, mock.AnythingOfType("*audit.Log")).Return(nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "test-org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:        "test-log-id",
						Source:    "test-source",
						Action:    "test-action",
						CreatedAt: timestamppb.New(time.Now()),
						Actor: &frontierv1beta1.AuditLogActor{
							Id:   "actor-id",
							Type: "user",
							Name: "actor-name",
						},
						Target: &frontierv1beta1.AuditLogTarget{
							Id:   "target-id",
							Type: "resource",
							Name: "target-name",
						},
						Context: map[string]string{
							"key": "value",
						},
					},
				},
			}),
			wantErr: false,
		},
		{
			name: "should return error when log source is empty",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "test-org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:     "test-log-id",
						Source: "", // Empty source
						Action: "test-action",
					},
				},
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return error when log action is empty",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "test-org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:     "test-log-id",
						Source: "test-source",
						Action: "", // Empty action
					},
				},
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInvalidArgument,
		},
		{
			name: "should return error when organization is disabled",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "disabled-org").Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "disabled-org",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:     "test-log-id",
						Source: "test-source",
						Action: "test-action",
					},
				},
			}),
			wantErr:     true,
			wantErrCode: connect.CodeFailedPrecondition,
		},
		{
			name: "should return error when audit service create fails",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().Create(mock.Anything, mock.AnythingOfType("*audit.Log")).Return(errors.New("database error"))
			},
			request: connect.NewRequest(&frontierv1beta1.CreateOrganizationAuditLogsRequest{
				OrgId: "test-org-id",
				Logs: []*frontierv1beta1.AuditLog{
					{
						Id:     "test-log-id",
						Source: "test-source",
						Action: "test-action",
					},
				},
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuditSvc := new(mocks.AuditService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockAuditSvc, mockOrgSvc)
			}

			handler := &ConnectHandler{
				auditService: mockAuditSvc,
				orgService:   mockOrgSvc,
			}

			resp, err := handler.CreateOrganizationAuditLogs(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					var connectErr *connect.Error
					if assert.True(t, errors.As(err, &connectErr)) {
						assert.Equal(t, tt.wantErrCode, connectErr.Code())
					}
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			}
		})
	}
}

func TestConnectHandler_GetOrganizationAuditLog(t *testing.T) {
	tests := []struct {
		name        string
		setup       func(as *mocks.AuditService, os *mocks.OrganizationService)
		request     *connect.Request[frontierv1beta1.GetOrganizationAuditLogRequest]
		wantErr     bool
		wantErrCode connect.Code
	}{
		{
			name: "should return audit log successfully",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().GetByID(mock.Anything, "test-log-id").Return(audit.Log{
					ID:        "test-log-id",
					OrgID:     "test-org-id",
					Source:    "test-source",
					Action:    "test-action",
					CreatedAt: time.Now(),
					Actor: audit.Actor{
						ID:   "actor-id",
						Type: "user",
						Name: "actor-name",
					},
					Target: audit.Target{
						ID:   "target-id",
						Type: "resource",
						Name: "target-name",
					},
					Metadata: map[string]string{
						"key": "value",
					},
				}, nil)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationAuditLogRequest{
				OrgId: "test-org-id",
				Id:    "test-log-id",
			}),
			wantErr: false,
		},
		{
			name: "should return error when organization is disabled",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "disabled-org").Return(organization.Organization{}, organization.ErrDisabled)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationAuditLogRequest{
				OrgId: "disabled-org",
				Id:    "test-log-id",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeFailedPrecondition,
		},
		{
			name: "should return error when organization not found",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "nonexistent-org").Return(organization.Organization{}, organization.ErrNotExist)
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationAuditLogRequest{
				OrgId: "nonexistent-org",
				Id:    "test-log-id",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeNotFound,
		},
		{
			name: "should return error when audit log not found",
			setup: func(as *mocks.AuditService, os *mocks.OrganizationService) {
				os.EXPECT().Get(mock.Anything, "test-org-id").Return(testOrg, nil)
				as.EXPECT().GetByID(mock.Anything, "nonexistent-log").Return(audit.Log{}, errors.New("log not found"))
			},
			request: connect.NewRequest(&frontierv1beta1.GetOrganizationAuditLogRequest{
				OrgId: "test-org-id",
				Id:    "nonexistent-log",
			}),
			wantErr:     true,
			wantErrCode: connect.CodeInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockAuditSvc := new(mocks.AuditService)
			mockOrgSvc := new(mocks.OrganizationService)
			if tt.setup != nil {
				tt.setup(mockAuditSvc, mockOrgSvc)
			}

			handler := &ConnectHandler{
				auditService: mockAuditSvc,
				orgService:   mockOrgSvc,
			}

			resp, err := handler.GetOrganizationAuditLog(context.Background(), tt.request)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.wantErrCode != 0 {
					var connectErr *connect.Error
					if assert.True(t, errors.As(err, &connectErr)) {
						assert.Equal(t, tt.wantErrCode, connectErr.Code())
					}
				}
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.NotNil(t, resp.Msg.GetLog())
				assert.Equal(t, "test-log-id", resp.Msg.GetLog().GetId())
			}
		})
	}
}

func TestTransformAuditLogToPB(t *testing.T) {
	now := time.Now()
	log := audit.Log{
		ID:        "test-id",
		Source:    "test-source",
		Action:    "test-action",
		CreatedAt: now,
		Actor: audit.Actor{
			ID:   "actor-id",
			Type: "user",
			Name: "actor-name",
		},
		Target: audit.Target{
			ID:   "target-id",
			Type: "resource",
			Name: "target-name",
		},
		Metadata: map[string]string{
			"key":   "value",
			"count": "42",
		},
	}

	pbLog := transformAuditLogToPB(log)

	assert.Equal(t, "test-id", pbLog.GetId())
	assert.Equal(t, "test-source", pbLog.GetSource())
	assert.Equal(t, "test-action", pbLog.GetAction())
	assert.Equal(t, now.Unix(), pbLog.GetCreatedAt().GetSeconds())
	assert.Equal(t, "actor-id", pbLog.GetActor().GetId())
	assert.Equal(t, "user", pbLog.GetActor().GetType())
	assert.Equal(t, "actor-name", pbLog.GetActor().GetName())
	assert.Equal(t, "target-id", pbLog.GetTarget().GetId())
	assert.Equal(t, "resource", pbLog.GetTarget().GetType())
	assert.Equal(t, "target-name", pbLog.GetTarget().GetName())
	assert.NotNil(t, pbLog.GetContext())
}
