package auditrecord_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/core/auditrecord/mocks"
	"github.com/raystack/frontier/core/serviceuser"
	"github.com/raystack/frontier/core/user"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func createMockServices(t *testing.T) (*mocks.Repository, *mocks.UserService, *mocks.ServiceUserService) {
	t.Helper() // This marks the function as a test helper
	return mocks.NewRepository(t), mocks.NewUserService(t), mocks.NewServiceUserService(t)
}

// Helper function to create base audit record for testing
func createBaseAuditRecord() auditrecord.AuditRecord {
	return auditrecord.AuditRecord{
		Event: "user.created",
		Actor: auditrecord.Actor{
			ID:   uuid.New().String(),
			Type: schema.UserPrincipal,
		},
		Resource: auditrecord.Resource{
			ID:   "resource-123",
			Type: "project",
		},
		OccurredAt:     time.Now(),
		OrgID:          "org-123",
		IdempotencyKey: "test-key",
	}
}

// TEST 1: Testing idempotency scenarios
func TestService_Create_Idempotency(t *testing.T) {
	tests := []struct {
		name                   string
		setupMocks             func(*mocks.Repository, *mocks.UserService, *mocks.ServiceUserService)
		auditRecord            auditrecord.AuditRecord
		expectIdempotent       bool
		expectError            error
		expectRepositoryCreate bool // Should repository.Create be called?
	}{
		{
			name: "no idempotency key - creates new record",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// Since no idempotency key, should skip GetByIdempotencyKey
				// Should enrich the actor and create a new record
				userSvc.EXPECT().GetByID(mock.Anything, mock.Anything).Return(user.User{
					ID: "user-123", Title: "Test User",
				}, nil)
				userSvc.EXPECT().IsSudo(mock.Anything, mock.Anything, schema.PlatformSudoPermission).Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					return ar.Actor.Name == "Test User" // Check enrichment happened
				})).Return(auditrecord.AuditRecord{ID: "new-id"}, nil)
			},
			auditRecord: func() auditrecord.AuditRecord {
				record := createBaseAuditRecord()
				record.IdempotencyKey = "" // No key
				return record
			}(),
			expectIdempotent:       false,
			expectError:            nil,
			expectRepositoryCreate: true,
		},
		{
			name: "idempotency key - record doesn't exist - creates new",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "unique-key").Return(auditrecord.AuditRecord{}, auditrecord.ErrNotFound)

				userSvc.EXPECT().GetByID(mock.Anything, mock.Anything).Return(user.User{
					ID: "user-123", Title: "Test User",
				}, nil)
				userSvc.EXPECT().IsSudo(mock.Anything, mock.Anything, schema.PlatformSudoPermission).Return(false, nil)
				repo.EXPECT().Create(mock.Anything, mock.Anything).Return(auditrecord.AuditRecord{ID: "new-id"}, nil)
			},
			auditRecord: func() auditrecord.AuditRecord {
				record := createBaseAuditRecord()
				record.IdempotencyKey = "unique-key"
				return record
			}(),
			expectIdempotent:       false,
			expectError:            nil,
			expectRepositoryCreate: true,
		},
		{
			name: "idempotency key - same record exists - returns existing (idempotent)",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// Existing record with the same content hash
				// The hash is computed from: Event, Actor.ID, Resource.ID, OrgID, Target.ID
				existingRecord := auditrecord.AuditRecord{
					ID:    "existing-123",
					Event: "user.created",
					Actor: auditrecord.Actor{
						ID:   "actor-123",
						Type: schema.UserPrincipal,
					},
					Resource: auditrecord.Resource{
						ID:   "resource-123",
						Type: "project",
					},
					OrgID:          "org-123",
					Target:         nil, // No target
					IdempotencyKey: "duplicate-key",
				}

				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "duplicate-key").Return(existingRecord, nil)
				// No Create call expected since we return existing
			},
			auditRecord: auditrecord.AuditRecord{
				Event: "user.created",
				Actor: auditrecord.Actor{
					ID:   "actor-123",
					Type: schema.UserPrincipal,
				},
				Resource: auditrecord.Resource{
					ID:   "resource-123",
					Type: "project",
				},
				OrgID:          "org-123",
				Target:         nil,
				OccurredAt:     time.Now(),
				IdempotencyKey: "duplicate-key",
			},
			expectIdempotent:       true,
			expectError:            nil,
			expectRepositoryCreate: false,
		},
		{
			name: "idempotency key - different record exists - conflict error",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				existingRecord := createBaseAuditRecord()
				existingRecord.Event = "user.deleted"
				existingRecord.IdempotencyKey = "conflict-key"

				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "conflict-key").Return(existingRecord, nil)
				// No Create call expected since we return error
			},
			auditRecord: func() auditrecord.AuditRecord {
				record := createBaseAuditRecord()
				record.Event = "user.created"
				record.IdempotencyKey = "conflict-key"
				return record
			}(),
			expectIdempotent:       false,
			expectError:            auditrecord.ErrIdempotencyKeyConflict,
			expectRepositoryCreate: false,
		},
		{
			name: "idempotency key - repository error on lookup",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// Repository fails with unexpected error (not ErrNotFound)
				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "error-key").Return(auditrecord.AuditRecord{}, errors.New("database error"))
				// No Create call expected since we return error early
			},
			auditRecord: func() auditrecord.AuditRecord {
				record := createBaseAuditRecord()
				record.IdempotencyKey = "error-key"
				return record
			}(),
			expectIdempotent:       false,
			expectError:            errors.New("database error"),
			expectRepositoryCreate: false,
		},
		{
			name: "hash computation - case insensitive and whitespace trimmed",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// Existing record with uppercase and spaces
				existingRecord := auditrecord.AuditRecord{
					ID:    "existing-123",
					Event: "USER.CREATED",
					Actor: auditrecord.Actor{
						ID:   "  ACTOR-123  ",
						Type: schema.UserPrincipal,
					},
					Resource: auditrecord.Resource{
						ID:   "  RESOURCE-123  ",
						Type: "project",
					},
					OrgID:          "  ORG-123  ",
					Target:         nil,
					IdempotencyKey: "normalize-key",
				}

				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "normalize-key").Return(existingRecord, nil)
				// Should return existing record because hash matches after normalization
			},
			auditRecord: auditrecord.AuditRecord{
				Event: "user.created",
				Actor: auditrecord.Actor{
					ID:   "actor-123",
					Type: schema.UserPrincipal,
				},
				Resource: auditrecord.Resource{
					ID:   "resource-123",
					Type: "project",
				},
				OrgID:          "org-123",
				Target:         nil,
				OccurredAt:     time.Now(),
				IdempotencyKey: "normalize-key",
			},
			expectIdempotent:       true, // Should be idempotent because hash matches after normalization
			expectError:            nil,
			expectRepositoryCreate: false,
		},
		{
			name: "hash computation - fields not in hash don't affect idempotency",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// Existing record with different non-hash fields
				existingRecord := auditrecord.AuditRecord{
					ID:    "existing-456",
					Event: "user.created",
					Actor: auditrecord.Actor{
						ID:   "actor-789",
						Type: "app/user",
						Name: "Original User Name",
						Metadata: metadata.Metadata{
							"original": "metadata",
						},
					},
					Resource: auditrecord.Resource{
						ID:   "resource-789",
						Type: "organization",
						Name: "Original Project",
					},
					OrgID:      "org-456",
					Target:     nil,
					OccurredAt: time.Now().Add(-24 * time.Hour),
					Metadata: metadata.Metadata{
						"request_id": "old-request",
					},
					IdempotencyKey: "fields-key",
				}

				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "fields-key").Return(existingRecord, nil)
			},
			auditRecord: auditrecord.AuditRecord{
				Event: "user.created",
				Actor: auditrecord.Actor{
					ID:   "actor-789",
					Type: schema.ServiceUserPrincipal,
					Name: "New User Name",
					Metadata: metadata.Metadata{
						"new": "metadata",
					},
				},
				Resource: auditrecord.Resource{
					ID:   "resource-789",
					Type: "project",
					Name: "New Project",
				},
				OrgID:      "org-456",
				Target:     nil,
				OccurredAt: time.Now(),
				Metadata: metadata.Metadata{
					"request_id": "new-request",
				},
				IdempotencyKey: "fields-key",
			},
			expectIdempotent:       true, // Should be idempotent because hash-relevant fields match
			expectError:            nil,
			expectRepositoryCreate: false,
		},
		{
			name: "hash computation - with target affects hash",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// Existing record WITHOUT target
				existingRecord := auditrecord.AuditRecord{
					ID:    "existing-789",
					Event: "permission.granted",
					Actor: auditrecord.Actor{
						ID:   "admin-123",
						Type: schema.UserPrincipal,
					},
					Resource: auditrecord.Resource{
						ID:   "role-456",
						Type: "role",
					},
					OrgID:          "org-789",
					Target:         nil,
					IdempotencyKey: "target-key",
				}

				repo.EXPECT().GetByIdempotencyKey(mock.Anything, "target-key").Return(existingRecord, nil)
			},
			auditRecord: auditrecord.AuditRecord{
				Event: "permission.granted",
				Actor: auditrecord.Actor{
					ID:   "admin-123",
					Type: schema.UserPrincipal,
				},
				Resource: auditrecord.Resource{
					ID:   "role-456",
					Type: "role",
				},
				OrgID: "org-789",
				Target: &auditrecord.Target{
					ID:   "user-999",
					Type: "user",
					Name: "Target User",
				},
				OccurredAt:     time.Now(),
				IdempotencyKey: "target-key",
			},
			expectIdempotent:       false,
			expectError:            auditrecord.ErrIdempotencyKeyConflict,
			expectRepositoryCreate: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, userSvc, serviceuserSvc := createMockServices(t)

			tt.setupMocks(repo, userSvc, serviceuserSvc)

			service := auditrecord.NewService(repo, userSvc, serviceuserSvc)

			result, isIdempotent, err := service.Create(context.Background(), tt.auditRecord)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
			}

			assert.Equal(t, tt.expectIdempotent, isIdempotent, "Idempotent flag should match expectation")
		})
	}
}

// TEST 2: Testing actor enrichment scenarios
func TestService_Create_ActorEnrichment(t *testing.T) {
	tests := []struct {
		name          string
		setupMocks    func(*mocks.Repository, *mocks.UserService, *mocks.ServiceUserService)
		inputRecord   auditrecord.AuditRecord
		expectedActor auditrecord.Actor
		expectError   error
	}{
		{
			name: "system actor - nil UUID gets enriched",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				// System actors don't need external service calls
				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					return ar.Actor.Type == "system" && ar.Actor.Name == "system"
				})).Return(auditrecord.AuditRecord{ID: "created"}, nil)
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "system.maintenance",
				Actor: auditrecord.Actor{
					ID:   uuid.Nil.String(), // This triggers system actor logic
					Type: "original-type",   // This will be overwritten
				},
				Resource:   auditrecord.Resource{ID: "res-123", Type: "system"},
				OccurredAt: time.Now(),
				OrgID:      "org-123",
			},
			expectedActor: auditrecord.Actor{
				ID:   uuid.Nil.String(),
				Type: "system",
				Name: "system",
			},
			expectError: nil,
		},
		{
			name: "user actor - gets enriched with user details",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				userSvc.EXPECT().GetByID(mock.Anything, "user-456").Return(user.User{
					ID: "user-456", Title: "John Doe",
				}, nil)
				userSvc.EXPECT().IsSudo(mock.Anything, "user-456", schema.PlatformSudoPermission).Return(false, nil)

				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					return ar.Actor.Name == "John Doe" && ar.Actor.Type == schema.UserPrincipal
				})).Return(auditrecord.AuditRecord{ID: "created"}, nil)
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "user.login",
				Actor: auditrecord.Actor{
					ID:   "user-456",
					Type: schema.UserPrincipal,
				},
				Resource:   auditrecord.Resource{ID: "res-123", Type: "session"},
				OccurredAt: time.Now(),
				OrgID:      "org-123",
			},
			expectedActor: auditrecord.Actor{
				ID:   "user-456",
				Type: schema.UserPrincipal,
				Name: "John Doe",
			},
			expectError: nil,
		},
		{
			name: "super user - gets enriched with sudo metadata",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				userSvc.EXPECT().GetByID(mock.Anything, "admin-789").Return(user.User{
					ID: "admin-789", Title: "Super Admin",
				}, nil)
				userSvc.EXPECT().IsSudo(mock.Anything, "admin-789", schema.PlatformSudoPermission).Return(true, nil) // IS sudo!

				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					// Verify super user metadata was added
					isSudo, exists := ar.Actor.Metadata[auditrecord.SuperUserActorMetadataKey]
					return ar.Actor.Name == "Super Admin" && exists && isSudo == true
				})).Return(auditrecord.AuditRecord{ID: "created"}, nil)
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "admin.delete_org",
				Actor: auditrecord.Actor{
					ID:   "admin-789",
					Type: schema.UserPrincipal,
				},
				Resource:   auditrecord.Resource{ID: "org-999", Type: "organization"},
				OccurredAt: time.Now(),
				OrgID:      "platform",
			},
			expectedActor: auditrecord.Actor{
				ID:   "admin-789",
				Type: schema.UserPrincipal,
				Name: "Super Admin",
				Metadata: metadata.Metadata{
					auditrecord.SuperUserActorMetadataKey: true,
				},
			},
			expectError: nil,
		},
		{
			name: "service user - gets enriched with service user details",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				serviceuserSvc.EXPECT().Get(mock.Anything, "service-999").Return(serviceuser.ServiceUser{
					ID: "service-999", Title: "API Service",
				}, nil)

				repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
					return ar.Actor.Name == "API Service" && ar.Actor.Type == schema.ServiceUserPrincipal
				})).Return(auditrecord.AuditRecord{ID: "created"}, nil)
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "api.call",
				Actor: auditrecord.Actor{
					ID:   "service-999",
					Type: schema.ServiceUserPrincipal,
				},
				Resource:   auditrecord.Resource{ID: "endpoint-123", Type: "api"},
				OccurredAt: time.Now(),
				OrgID:      "org-123",
			},
			expectedActor: auditrecord.Actor{
				ID:   "service-999",
				Type: schema.ServiceUserPrincipal,
				Name: "API Service",
			},
			expectError: nil,
		},
		{
			name: "user service error - should return error",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				userSvc.EXPECT().GetByID(mock.Anything, "missing-user").Return(user.User{}, errors.New("user not found"))
				// No repo.Create call expected due to error
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "user.action",
				Actor: auditrecord.Actor{
					ID:   "missing-user",
					Type: schema.UserPrincipal,
				},
				Resource:   auditrecord.Resource{ID: "res-123", Type: "project"},
				OccurredAt: time.Now(),
				OrgID:      "org-123",
			},
			expectError: errors.New("user not found"),
		},
		{
			name: "sudo check error - should return error",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				userSvc.EXPECT().GetByID(mock.Anything, "user-123").Return(user.User{
					ID: "user-123", Title: "Test User",
				}, nil)
				userSvc.EXPECT().IsSudo(mock.Anything, "user-123", schema.PlatformSudoPermission).Return(false, errors.New("permission check failed"))
				// No repo.Create call expected due to error
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "user.action",
				Actor: auditrecord.Actor{
					ID:   "user-123",
					Type: schema.UserPrincipal,
				},
				Resource:   auditrecord.Resource{ID: "res-123", Type: "project"},
				OccurredAt: time.Now(),
				OrgID:      "org-123",
			},
			expectError: errors.New("permission check failed"),
		},
		{
			name: "service user error - should return error",
			setupMocks: func(repo *mocks.Repository, userSvc *mocks.UserService, serviceuserSvc *mocks.ServiceUserService) {
				serviceuserSvc.EXPECT().Get(mock.Anything, "missing-service").Return(serviceuser.ServiceUser{}, errors.New("service user not found"))
				// No repo.Create call expected due to error
			},
			inputRecord: auditrecord.AuditRecord{
				Event: "api.call",
				Actor: auditrecord.Actor{
					ID:   "missing-service",
					Type: schema.ServiceUserPrincipal,
				},
				Resource:   auditrecord.Resource{ID: "res-123", Type: "api"},
				OccurredAt: time.Now(),
				OrgID:      "org-123",
			},
			expectError: errors.New("service user not found"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, userSvc, serviceuserSvc := createMockServices(t)
			tt.setupMocks(repo, userSvc, serviceuserSvc)

			service := auditrecord.NewService(repo, userSvc, serviceuserSvc)
			result, isIdempotent, err := service.Create(context.Background(), tt.inputRecord)

			if tt.expectError != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectError.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, result)
				assert.False(t, isIdempotent) // All these tests create new records
			}
		})
	}
}

// TEST 3: Testing repository errors
func TestService_Create_RepositoryErrors(t *testing.T) {
	repo, userSvc, serviceuserSvc := createMockServices(t)

	userSvc.EXPECT().GetByID(mock.Anything, mock.Anything).Return(user.User{
		ID: "user-123", Title: "Test User",
	}, nil)
	userSvc.EXPECT().IsSudo(mock.Anything, mock.Anything, schema.PlatformSudoPermission).Return(false, nil)
	repo.EXPECT().Create(mock.Anything, mock.Anything).Return(auditrecord.AuditRecord{}, errors.New("database connection failed"))

	service := auditrecord.NewService(repo, userSvc, serviceuserSvc)
	record := createBaseAuditRecord()
	record.IdempotencyKey = ""

	result, isIdempotent, err := service.Create(context.Background(), record)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "database connection failed")
	assert.Empty(t, result)
	assert.False(t, isIdempotent)
}

// TEST 4: Testing edge cases and boundary conditions
func TestService_Create_EdgeCases(t *testing.T) {
	t.Run("actor with existing metadata gets sudo metadata added", func(t *testing.T) {
		repo, userSvc, serviceuserSvc := createMockServices(t)

		userSvc.EXPECT().GetByID(mock.Anything, "user-123").Return(user.User{
			ID: "user-123", Title: "Super User",
		}, nil)
		userSvc.EXPECT().IsSudo(mock.Anything, "user-123", schema.PlatformSudoPermission).Return(true, nil)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
			// Verify both original and sudo metadata exist
			originalValue, hasOriginal := ar.Actor.Metadata["existing_key"]
			sudoValue, hasSudo := ar.Actor.Metadata[auditrecord.SuperUserActorMetadataKey]
			return hasOriginal && originalValue == "existing_value" && hasSudo && sudoValue == true
		})).Return(auditrecord.AuditRecord{ID: "created"}, nil)

		service := auditrecord.NewService(repo, userSvc, serviceuserSvc)
		record := auditrecord.AuditRecord{
			Event: "admin.action",
			Actor: auditrecord.Actor{
				ID:   "user-123",
				Type: schema.UserPrincipal,
				Metadata: metadata.Metadata{
					"existing_key": "existing_value",
				},
			},
			Resource:   auditrecord.Resource{ID: "res-123", Type: "project"},
			OccurredAt: time.Now(),
			OrgID:      "org-123",
		}

		result, isIdempotent, err := service.Create(context.Background(), record)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.False(t, isIdempotent)
	})

	t.Run("actor with nil metadata gets sudo metadata map created", func(t *testing.T) {
		repo, userSvc, serviceuserSvc := createMockServices(t)

		userSvc.EXPECT().GetByID(mock.Anything, "user-456").Return(user.User{
			ID: "user-456", Title: "Super User",
		}, nil)
		userSvc.EXPECT().IsSudo(mock.Anything, "user-456", schema.PlatformSudoPermission).Return(true, nil)

		repo.EXPECT().Create(mock.Anything, mock.MatchedBy(func(ar auditrecord.AuditRecord) bool {
			// Verify metadata map was created and sudo flag added
			sudoValue, hasSudo := ar.Actor.Metadata[auditrecord.SuperUserActorMetadataKey]
			return hasSudo && sudoValue == true && ar.Actor.Metadata != nil
		})).Return(auditrecord.AuditRecord{ID: "created"}, nil)

		service := auditrecord.NewService(repo, userSvc, serviceuserSvc)
		record := auditrecord.AuditRecord{
			Event: "admin.action",
			Actor: auditrecord.Actor{
				ID:       "user-456",
				Type:     schema.UserPrincipal,
				Metadata: nil,
			},
			Resource:   auditrecord.Resource{ID: "res-123", Type: "project"},
			OccurredAt: time.Now(),
			OrgID:      "org-123",
		}

		result, isIdempotent, err := service.Create(context.Background(), record)
		assert.NoError(t, err)
		assert.NotEmpty(t, result)
		assert.False(t, isIdempotent)
	})
}
