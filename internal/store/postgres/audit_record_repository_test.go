package postgres_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/salt/log"
	"github.com/stretchr/testify/suite"
)

type AuditRecordRepositoryTestSuite struct {
	suite.Suite
	ctx        context.Context
	client     *db.Client
	pool       *dockertest.Pool
	resource   *dockertest.Resource
	repository *postgres.AuditRecordRepository
}

func (s *AuditRecordRepositoryTestSuite) SetupSuite() {
	var err error

	logger := log.NewZap()
	s.client, s.pool, s.resource, err = newTestClient(logger)
	if err != nil {
		s.T().Fatal(err)
	}

	s.ctx = context.TODO()
	s.repository = postgres.NewAuditRecordRepository(s.client)
}

func (s *AuditRecordRepositoryTestSuite) SetupTest() {
	// For audit records, we typically don't need shared test data
	// Each test will create its own records
}

func (s *AuditRecordRepositoryTestSuite) TearDownSuite() {
	if err := purgeDocker(s.pool, s.resource); err != nil {
		s.T().Fatal(err)
	}
}

func (s *AuditRecordRepositoryTestSuite) TearDownTest() {
	if err := s.cleanup(); err != nil {
		s.T().Fatal(err)
	}
}

func (s *AuditRecordRepositoryTestSuite) cleanup() error {
	queries := []string{
		fmt.Sprintf("TRUNCATE TABLE %s RESTART IDENTITY CASCADE", postgres.TABLE_AUDITRECORDS),
	}
	return execQueries(context.TODO(), s.client, queries)
}

func (s *AuditRecordRepositoryTestSuite) createValidAuditRecord() auditrecord.AuditRecord {
	return auditrecord.AuditRecord{
		Event: "user.created",
		Actor: auditrecord.Actor{
			ID:   uuid.New().String(),
			Type: "app/user",
			Name: "Test User",
			Metadata: metadata.Metadata{
				"role": "admin",
			},
		},
		Resource: auditrecord.Resource{
			ID:   "resource-" + uuid.New().String(),
			Type: "project",
			Name: "Test Project",
			Metadata: metadata.Metadata{
				"env": "production",
			},
		},
		Target: &auditrecord.Target{
			ID:   "target-" + uuid.New().String(),
			Type: "organization",
			Name: "Test Org",
			Metadata: metadata.Metadata{
				"plan": "enterprise",
			},
		},
		OccurredAt: time.Now().UTC(),
		OrgID:      uuid.New().String(),
		RequestID:  stringPtr("req-" + uuid.New().String()),
		Metadata: metadata.Metadata{
			"ip_address": "192.168.1.1",
		},
		IdempotencyKey: uuid.New().String(),
	}
}

func stringPtr(s string) *string {
	return &s
}

// TEST 1: Create audit record - Success cases
func (s *AuditRecordRepositoryTestSuite) TestCreate_Success() {
	tests := []struct {
		name         string
		modifyRecord func(*auditrecord.AuditRecord)
		description  string
	}{
		{
			name: "create complete audit record with all fields",
			modifyRecord: func(ar *auditrecord.AuditRecord) {
				// Keep all fields as is
			},
			description: "Should successfully create a record with all fields populated",
		},
		{
			name: "create audit record without target",
			modifyRecord: func(ar *auditrecord.AuditRecord) {
				ar.Target = nil
			},
			description: "Target is optional - should handle nil target",
		},
		{
			name: "create audit record without metadata",
			modifyRecord: func(ar *auditrecord.AuditRecord) {
				ar.Metadata = nil
				ar.Actor.Metadata = nil
				ar.Resource.Metadata = nil
				if ar.Target != nil {
					ar.Target.Metadata = nil
				}
			},
			description: "Metadata fields are optional - should handle nil metadata",
		},
		{
			name: "create audit record without request ID",
			modifyRecord: func(ar *auditrecord.AuditRecord) {
				ar.RequestID = nil
			},
			description: "RequestID is optional - should handle nil",
		},
		{
			name: "create system actor audit record",
			modifyRecord: func(ar *auditrecord.AuditRecord) {
				ar.Actor.ID = uuid.Nil.String() // System actor
				ar.Actor.Type = "system"
				ar.Actor.Name = "system"
			},
			description: "Should handle system actor (nil UUID)",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			record := s.createValidAuditRecord()
			tt.modifyRecord(&record)

			created, err := s.repository.Create(s.ctx, record)

			s.NoError(err, tt.description)
			s.NotEmpty(created.ID, "Should generate an ID")
			s.NotZero(created.CreatedAt, "Should set CreatedAt")

			s.Equal(record.Event, created.Event)
			s.Equal(record.Actor.ID, created.Actor.ID)
			s.Equal(record.Actor.Type, created.Actor.Type)
			s.Equal(record.Actor.Name, created.Actor.Name)
			s.Equal(record.Resource.ID, created.Resource.ID)
			s.Equal(record.OrgID, created.OrgID)
			s.Equal(record.IdempotencyKey, created.IdempotencyKey)

			if record.Metadata != nil {
				s.Equal(record.Metadata, created.Metadata)
			} else {
				s.Empty(created.Metadata)
			}

			if record.Target == nil {
				s.Nil(created.Target)
			} else {
				s.NotNil(created.Target)
				s.Equal(record.Target.ID, created.Target.ID)
				s.Equal(record.Target.Type, created.Target.Type)
				s.Equal(record.Target.Name, created.Target.Name)
			}
		})
	}
}

// TEST 2: Create audit record - Error cases
func (s *AuditRecordRepositoryTestSuite) TestCreate_Errors() {
	tests := []struct {
		name        string
		setupRecord func() auditrecord.AuditRecord
		expectedErr error
		description string
	}{
		{
			name: "duplicate idempotency key",
			setupRecord: func() auditrecord.AuditRecord {
				record := s.createValidAuditRecord()
				_, err := s.repository.Create(s.ctx, record)
				s.NoError(err)

				// Return a new record with the same idempotency key
				newRecord := s.createValidAuditRecord()
				newRecord.IdempotencyKey = record.IdempotencyKey
				return newRecord
			},
			expectedErr: auditrecord.ErrIdempotencyKeyConflict,
			description: "Should return conflict error for duplicate idempotency key",
		},
		{
			name: "invalid actor UUID",
			setupRecord: func() auditrecord.AuditRecord {
				record := s.createValidAuditRecord()
				record.Actor.ID = "not-a-uuid"
				return record
			},
			expectedErr: nil, // Will be a parse error
			description: "Should return error for invalid actor UUID",
		},
		{
			name: "invalid org UUID",
			setupRecord: func() auditrecord.AuditRecord {
				record := s.createValidAuditRecord()
				record.OrgID = "not-a-uuid"
				return record
			},
			expectedErr: nil, // Will be a parse error
			description: "Should return error for invalid org UUID",
		},
		{
			name: "invalid idempotency key UUID",
			setupRecord: func() auditrecord.AuditRecord {
				record := s.createValidAuditRecord()
				record.IdempotencyKey = "not-a-uuid"
				return record
			},
			expectedErr: nil, // Will be a parse error
			description: "Should return error for invalid idempotency key UUID",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			record := tt.setupRecord()

			created, err := s.repository.Create(s.ctx, record)

			s.Error(err, tt.description)
			s.Empty(created.ID, "Should not create record on error")

			if tt.expectedErr != nil {
				s.ErrorIs(err, tt.expectedErr)
			}
		})
	}
}

// TEST 3: GetByID functionality
func (s *AuditRecordRepositoryTestSuite) TestGetByID() {
	record := s.createValidAuditRecord()
	created, err := s.repository.Create(s.ctx, record)
	s.NoError(err)

	tests := []struct {
		name        string
		id          string
		expectError error
		description string
	}{
		{
			name:        "get existing record by ID",
			id:          created.ID,
			expectError: nil,
			description: "Should retrieve the record successfully",
		},
		{
			name:        "get non-existent record",
			id:          uuid.New().String(),
			expectError: auditrecord.ErrNotFound,
			description: "Should return not found error",
		},
		{
			name:        "get with invalid UUID",
			id:          "not-a-uuid",
			expectError: auditrecord.ErrInvalidUUID,
			description: "Should return invalid UUID error",
		},
		{
			name:        "get with empty ID",
			id:          "",
			expectError: auditrecord.ErrNotFound,
			description: "Should return not found for empty ID",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			retrieved, err := s.repository.GetByID(s.ctx, tt.id)

			if tt.expectError != nil {
				s.Error(err, tt.description)
				s.ErrorIs(err, tt.expectError)
				s.Empty(retrieved.ID)
			} else {
				s.NoError(err, tt.description)
				s.Equal(created.ID, retrieved.ID)
				s.Equal(created.Event, retrieved.Event)
				s.Equal(created.Actor.ID, retrieved.Actor.ID)
				s.Equal(created.Resource.ID, retrieved.Resource.ID)
			}
		})
	}
}

// TEST 4: GetByIdempotencyKey functionality
func (s *AuditRecordRepositoryTestSuite) TestGetByIdempotencyKey() {
	// Create a record with a known idempotency key
	record := s.createValidAuditRecord()
	idempotencyKey := uuid.New().String()
	record.IdempotencyKey = idempotencyKey

	created, err := s.repository.Create(s.ctx, record)
	s.NoError(err)

	tests := []struct {
		name        string
		key         string
		expectError error
		description string
	}{
		{
			name:        "get existing record by idempotency key",
			key:         idempotencyKey,
			expectError: nil,
			description: "Should retrieve the record successfully",
		},
		{
			name:        "get non-existent idempotency key",
			key:         uuid.New().String(),
			expectError: auditrecord.ErrNotFound,
			description: "Should return not found error",
		},
		{
			name:        "get with invalid UUID key",
			key:         "not-a-uuid",
			expectError: auditrecord.ErrInvalidUUID,
			description: "Should return invalid UUID error",
		},
	}

	for _, tt := range tests {
		s.Run(tt.name, func() {
			retrieved, err := s.repository.GetByIdempotencyKey(s.ctx, tt.key)

			if tt.expectError != nil {
				s.Error(err, tt.description)
				s.ErrorIs(err, tt.expectError)
				s.Empty(retrieved.ID)
			} else {
				s.NoError(err, tt.description)
				s.Equal(created.ID, retrieved.ID)
				s.Equal(idempotencyKey, retrieved.IdempotencyKey)
				s.Equal(created.Event, retrieved.Event)
			}
		})
	}
}

// TEST 5: Edge cases and special scenarios
func (s *AuditRecordRepositoryTestSuite) TestEdgeCases() {
	s.Run("empty metadata stored as NULL not empty JSON", func() {
		record := s.createValidAuditRecord()
		record.Metadata = metadata.Metadata{} // Empty metadata
		record.Actor.Metadata = nil
		record.Resource.Metadata = metadata.Metadata{}

		created, err := s.repository.Create(s.ctx, record)
		s.NoError(err)

		retrieved, err := s.repository.GetByID(s.ctx, created.ID)
		s.NoError(err)

		// Empty metadata should be returned as an empty map, not nil. Implemented through the nullJSONTextToMetadata helper function
		s.NotNil(retrieved.Metadata)
		s.Empty(retrieved.Metadata)
		s.NotNil(retrieved.Actor.Metadata)
		s.Empty(retrieved.Actor.Metadata)
	})

	s.Run("nil target handled correctly", func() {
		record := s.createValidAuditRecord()
		record.Target = nil

		created, err := s.repository.Create(s.ctx, record)
		s.NoError(err)

		retrieved, err := s.repository.GetByID(s.ctx, created.ID)
		s.NoError(err)

		// Target should be nil when originally nil
		s.Nil(retrieved.Target)
	})

	s.Run("very long strings handled", func() {
		record := s.createValidAuditRecord()
		longString := func(prefix string, length int) string {
			pattern := "abcdefghijklmnopqrstuvwxyz0123456789"
			result := prefix
			for len(result) < length {
				result += pattern
			}
			return result[:length]
		}

		record.Event = longString("event.long.", 250)
		record.Actor.Name = longString("Very Long User Name ", 200)
		record.Resource.Name = longString("Resource ", 200)

		created, err := s.repository.Create(s.ctx, record)
		s.NoError(err)

		retrieved, err := s.repository.GetByID(s.ctx, created.ID)
		s.NoError(err)
		s.Equal(record.Event, retrieved.Event)
		s.Equal(record.Actor.Name, retrieved.Actor.Name)
		s.Equal(record.Resource.Name, retrieved.Resource.Name)
	})

	s.Run("special characters in strings for SQL injection", func() {
		record := s.createValidAuditRecord()
		record.Event = "user.created'; DROP TABLE audit_records; --"
		record.Actor.Name = "User with 'quotes' and \"double quotes\""
		record.Resource.Name = "Resource\nwith\nnewlines\tand\ttabs"

		created, err := s.repository.Create(s.ctx, record)
		s.NoError(err, "SQL injection attempt should be safely handled")

		retrieved, err := s.repository.GetByID(s.ctx, created.ID)
		s.NoError(err)
		s.Equal(record.Event, retrieved.Event)
		s.Equal(record.Actor.Name, retrieved.Actor.Name)
		s.Equal(record.Resource.Name, retrieved.Resource.Name)
	})

	s.Run("unicode characters handled correctly", func() {
		record := s.createValidAuditRecord()
		record.Event = "user.created"
		record.Actor.Name = "José García"
		record.Resource.Name = "München Office"

		created, err := s.repository.Create(s.ctx, record)
		s.NoError(err)

		retrieved, err := s.repository.GetByID(s.ctx, created.ID)
		s.NoError(err)
		s.Equal(record.Event, retrieved.Event)
		s.Equal(record.Actor.Name, retrieved.Actor.Name)
		s.Equal(record.Resource.Name, retrieved.Resource.Name)
	})
}

// TEST 6: Concurrent operations
func (s *AuditRecordRepositoryTestSuite) TestConcurrency() {
	s.Run("concurrent creates with different idempotency keys", func() {
		const numGoroutines = 10
		errChan := make(chan error, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(index int) {
				record := s.createValidAuditRecord()
				record.IdempotencyKey = uuid.New().String() // Each goroutine gets unique UUID

				_, err := s.repository.Create(s.ctx, record)
				errChan <- err
			}(i)
		}

		for i := 0; i < numGoroutines; i++ {
			err := <-errChan
			s.NoError(err, "Concurrent create should succeed")
		}
	})

	s.Run("concurrent creates with same idempotency key", func() {
		sharedKey := uuid.New().String()
		const numGoroutines = 5
		errChan := make(chan error, numGoroutines)
		successes := 0
		conflicts := 0

		for i := 0; i < numGoroutines; i++ {
			go func() {
				record := s.createValidAuditRecord()
				record.IdempotencyKey = sharedKey // All use the same key

				_, err := s.repository.Create(s.ctx, record)
				errChan <- err
			}()
		}

		// Collect results
		for i := 0; i < numGoroutines; i++ {
			err := <-errChan
			if err == nil {
				successes++
			} else if errors.Is(err, auditrecord.ErrIdempotencyKeyConflict) {
				conflicts++
			}
		}

		// Exactly one should succeed, others should get conflict
		s.Equal(1, successes, "Exactly one create should succeed")
		s.Equal(numGoroutines-1, conflicts, "Others should get conflict error")
	})
}

// TestAuditRecordRepositoryTestSuite is the entry point for the test suite
func TestAuditRecordRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AuditRecordRepositoryTestSuite))
}
