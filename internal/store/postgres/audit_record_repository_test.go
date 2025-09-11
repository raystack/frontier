package postgres_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/ory/dockertest"
	"github.com/raystack/frontier/core/auditrecord"
	"github.com/raystack/frontier/internal/store/postgres"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/raystack/frontier/pkg/utils"
	"github.com/raystack/salt/log"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/suite"
)

const (
	defaultListLimit  = 50
	defaultListOffset = 0
)

type AuditRecordRepositoryTestSuite struct {
	suite.Suite
	ctx          context.Context
	client       *db.Client
	pool         *dockertest.Pool
	resource     *dockertest.Resource
	repository   *postgres.AuditRecordRepository
	auditRecords []auditrecord.AuditRecord
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
	// Only bootstrap data for List tests
	testName := s.T().Name()
	if strings.Contains(testName, "TestList_") {
		var err error
		s.auditRecords, err = s.bootstrapAuditRecords()
		if err != nil {
			s.T().Fatal(err)
		}
	}
}

func (s *AuditRecordRepositoryTestSuite) bootstrapAuditRecords() ([]auditrecord.AuditRecord, error) {
	testFixtureJSON, err := os.ReadFile("./testdata/mock-audit-records.json")
	if err != nil {
		return nil, err
	}

	var fixtureData []auditrecord.AuditRecord
	if err = json.Unmarshal(testFixtureJSON, &fixtureData); err != nil {
		return nil, err
	}

	var insertedData []auditrecord.AuditRecord
	for _, d := range fixtureData {
		created, err := s.repository.Create(context.Background(), d)
		if err != nil {
			return nil, err
		}
		insertedData = append(insertedData, created)
	}

	return insertedData, nil
}

// setupListTestData ensures test data is available for List tests
func (s *AuditRecordRepositoryTestSuite) setupListTestData() {
	if len(s.auditRecords) == 0 {
		var err error
		s.auditRecords, err = s.bootstrapAuditRecords()
		if err != nil {
			s.T().Fatal(err)
		}
	}
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

// TEST 7: List functionality - Basic operations
func (s *AuditRecordRepositoryTestSuite) TestList_BasicOperations() {
	s.setupListTestData()

	type testCase struct {
		Description string
		Setup       func(t *testing.T) *rql.Query
		Expected    auditrecord.AuditRecordsList
		Err         error
	}

	testCases := []testCase{
		{
			Description: "should list all audit records with default pagination",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			Expected: auditrecord.AuditRecordsList{
				AuditRecords: s.auditRecords,
				Page: utils.Page{
					Limit:      defaultListLimit,
					Offset:     defaultListOffset,
					TotalCount: int64(len(s.auditRecords)),
				},
			},
		},
		{
			Description: "should return paginated results",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", 1, 2, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			Expected: auditrecord.AuditRecordsList{
				AuditRecords: s.auditRecords[1:3],
				Page: utils.Page{
					Limit:      2,
					Offset:     1,
					TotalCount: int64(len(s.auditRecords)),
				},
			},
		},
		{
			Description: "should handle high offset returning empty results",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", 100, 10, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			Expected: auditrecord.AuditRecordsList{
				AuditRecords: []auditrecord.AuditRecord{},
				Page: utils.Page{
					Limit:      10,
					Offset:     100,
					TotalCount: int64(len(s.auditRecords)),
				},
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			query := tc.Setup(s.T())
			result, err := s.repository.List(s.ctx, query)

			if tc.Err != nil {
				s.Error(err)
				s.ErrorIs(err, tc.Err)
				return
			}

			s.NoError(err)
			s.Equal(tc.Expected.Page.TotalCount, result.Page.TotalCount)
			s.Equal(tc.Expected.Page.Limit, result.Page.Limit)
			s.Equal(tc.Expected.Page.Offset, result.Page.Offset)
			s.Len(result.AuditRecords, len(tc.Expected.AuditRecords))
		})
	}
}

// TEST 8: List functionality - Filtering
func (s *AuditRecordRepositoryTestSuite) TestList_Filtering() {
	s.setupListTestData()

	type testCase struct {
		Description   string
		Setup         func(t *testing.T) *rql.Query
		ExpectedCount int
		ValidateFunc  func([]auditrecord.AuditRecord) bool
		Err           error
	}

	testCases := []testCase{
		{
			Description: "should filter by event type",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
					Name:     "event",
					Operator: "eq",
					Value:    "user.created",
				}}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 3, // Alice, Jack, Kelly
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if r.Event != "user.created" {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should filter by actor type",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
					Name:     "actor_type",
					Operator: "eq",
					Value:    "system",
				}}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 1, // System cleanup record
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if r.Actor.Type != "system" {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should filter by organization ID",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
					Name:     "organization_id",
					Operator: "eq",
					Value:    "22222222-2222-2222-2222-222222222222",
				}}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 4, // Alice, Bob, Eve login, Eve logout
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if r.OrgID != "22222222-2222-2222-2222-222222222222" {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should filter by resource type",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
					Name:     "resource_type",
					Operator: "eq",
					Value:    "project",
				}}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 3, // Alpha, Beta, Gamma projects
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if r.Resource.Type != "project" {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should filter by date range",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
					Name:     "occurred_at",
					Operator: "gte",
					Value:    "2024-01-02T00:00:00Z",
				}}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 9, // Records from Jan 2nd and 3rd
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				cutoff := time.Date(2024, 1, 2, 0, 0, 0, 0, time.UTC)
				for _, r := range records {
					if r.OccurredAt.Before(cutoff) {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should handle complex AND filtering",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{
					{
						Name:     "event",
						Operator: "eq",
						Value:    "user.created",
					},
					{
						Name:     "organization_id",
						Operator: "eq",
						Value:    "22222222-2222-2222-2222-222222222229",
					},
				}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 2, // Jack and Kelly
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if r.Event != "user.created" || r.OrgID != "22222222-2222-2222-2222-222222222229" {
						return false
					}
				}
				return true
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			query := tc.Setup(s.T())
			result, err := s.repository.List(s.ctx, query)

			if tc.Err != nil {
				s.Error(err)
				s.ErrorIs(err, tc.Err)
				return
			}

			s.NoError(err)
			s.Len(result.AuditRecords, tc.ExpectedCount)

			if tc.ValidateFunc != nil {
				s.True(tc.ValidateFunc(result.AuditRecords), "Validation function should pass")
			}
		})
	}
}

// TEST 9: List functionality - Search
func (s *AuditRecordRepositoryTestSuite) TestList_Search() {
	s.setupListTestData()

	type testCase struct {
		Description   string
		Setup         func(t *testing.T) *rql.Query
		ExpectedCount int
		ValidateFunc  func([]auditrecord.AuditRecord) bool
	}

	testCases := []testCase{
		{
			Description: "should search by actor name",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("Alice", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 1,
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				return len(records) > 0 && strings.Contains(strings.ToLower(records[0].Actor.Name), "alice")
			},
		},
		{
			Description: "should search by event name",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("permission", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 2, // permission.granted and permission.revoked
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if !strings.Contains(strings.ToLower(r.Event), "permission") {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should search by resource name",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("Project", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 3, // Project Alpha, Beta, Gamma
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				for _, r := range records {
					if !strings.Contains(r.Resource.Name, "Project") {
						return false
					}
				}
				return true
			},
		},
		{
			Description: "should handle unicode search",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("García", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 1,
			ValidateFunc: func(records []auditrecord.AuditRecord) bool {
				return len(records) > 0 && strings.Contains(records[0].Actor.Name, "García")
			},
		},
		{
			Description: "should return empty for no matches",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("NonExistentTerm", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})
			},
			ExpectedCount: 0,
			ValidateFunc:  nil,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			query := tc.Setup(s.T())
			result, err := s.repository.List(s.ctx, query)

			s.NoError(err)
			s.Len(result.AuditRecords, tc.ExpectedCount)

			if tc.ValidateFunc != nil {
				s.True(tc.ValidateFunc(result.AuditRecords), "Validation function should pass")
			}
		})
	}
}

// TEST 10: List functionality - Grouping
func (s *AuditRecordRepositoryTestSuite) TestList_Grouping() {
	s.setupListTestData()

	type testCase struct {
		Description    string
		Setup          func(t *testing.T) *rql.Query
		ExpectedGroups int
		ValidateFunc   func(*utils.Group) bool
	}

	testCases := []testCase{
		{
			Description: "should group by event",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{"event"})
			},
			ExpectedGroups: 12, // Unique events in test data
			ValidateFunc: func(group *utils.Group) bool {
				return group.Name == "event" && len(group.Data) > 0
			},
		},
		{
			Description: "should group by actor_type",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{"actor_type"})
			},
			ExpectedGroups: 3, // app/user, app/serviceuser, system
			ValidateFunc: func(group *utils.Group) bool {
				return group.Name == "actor_type" && len(group.Data) == 3
			},
		},
		{
			Description: "should group by resource_type",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{"resource_type"})
			},
			ExpectedGroups: 8, // Different resource types in test data
			ValidateFunc: func(group *utils.Group) bool {
				return group.Name == "resource_type" && len(group.Data) > 0
			},
		},
		{
			Description: "should group by organization_id",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{"organization_id"})
			},
			ExpectedGroups: 7, // Different organizations in test data
			ValidateFunc: func(group *utils.Group) bool {
				return group.Name == "organization_id" && len(group.Data) > 0
			},
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			query := tc.Setup(s.T())
			result, err := s.repository.List(s.ctx, query)

			s.NoError(err)
			s.NotNil(result.Group, "Group should not be nil")

			if tc.ValidateFunc != nil {
				s.True(tc.ValidateFunc(result.Group), "Group validation should pass")
			}

			// Verify group data structure
			s.NotEmpty(result.Group.Data, "Group data should not be empty")
			for _, groupItem := range result.Group.Data {
				s.NotEmpty(groupItem.Name, "Group item name should not be empty")
				s.Greater(groupItem.Count, 0, "Group item count should be positive")
			}
		})
	}
}

// TEST 11: List functionality - Complex scenarios
func (s *AuditRecordRepositoryTestSuite) TestList_ComplexScenarios() {
	s.setupListTestData()

	s.Run("combined filter, search, sort, and pagination", func() {
		query := utils.NewRQLQuery("user", 0, 3, []rql.Filter{{
			Name:     "actor_type",
			Operator: "eq",
			Value:    "app/user",
		}}, []rql.Sort{
			{Name: "occurred_at", Order: "desc"},
		}, []string{})

		result, err := s.repository.List(s.ctx, query)
		s.NoError(err)
		s.LessOrEqual(len(result.AuditRecords), 3)
		s.Equal(3, result.Page.Limit)
		s.Equal(0, result.Page.Offset)

		// Verify sorting
		if len(result.AuditRecords) > 1 {
			for i := 1; i < len(result.AuditRecords); i++ {
				s.False(result.AuditRecords[i].OccurredAt.After(result.AuditRecords[i-1].OccurredAt),
					"Records should be sorted by occurred_at descending")
			}
		}
	})

	s.Run("filter with grouping", func() {
		query := utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
			Name:     "resource_type",
			Operator: "eq",
			Value:    "project",
		}}, []rql.Sort{}, []string{"event"})

		result, err := s.repository.List(s.ctx, query)
		s.NoError(err)
		s.NotNil(result.Group)
		s.Equal("event", result.Group.Name)

		// Verify all returned records are projects
		for _, record := range result.AuditRecords {
			s.Equal("project", record.Resource.Type)
		}
	})

	s.Run("empty result set with grouping", func() {
		query := utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
			Name:     "event",
			Operator: "eq",
			Value:    "nonexistent.event",
		}}, []rql.Sort{}, []string{"actor_type"})

		result, err := s.repository.List(s.ctx, query)

		// This might return ErrNotFound or empty results depending on implementation
		if err == nil {
			s.Empty(result.AuditRecords)
			if result.Group != nil {
				s.Empty(result.Group.Data)
			}
		} else {
			s.ErrorIs(err, auditrecord.ErrNotFound)
		}
	})
}

// TEST 12: List functionality - Error cases
func (s *AuditRecordRepositoryTestSuite) TestList_ErrorCases() {
	s.setupListTestData()

	type testCase struct {
		Description string
		Setup       func(t *testing.T) *rql.Query
		ExpectedErr error
	}

	testCases := []testCase{
		{
			Description: "should return error for invalid filter column",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{{
					Name:     "invalid_column",
					Operator: "eq",
					Value:    "value",
				}}, []rql.Sort{}, []string{})
			},
			ExpectedErr: auditrecord.ErrRepositoryBadInput,
		},
		{
			Description: "should return error for invalid group column",
			Setup: func(t *testing.T) *rql.Query {
				t.Helper()
				return utils.NewRQLQuery("", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{"invalid_column"})
			},
			ExpectedErr: auditrecord.ErrRepositoryBadInput,
		},
	}

	for _, tc := range testCases {
		s.Run(tc.Description, func() {
			query := tc.Setup(s.T())
			result, err := s.repository.List(s.ctx, query)

			if tc.ExpectedErr != nil {
				s.Error(err)
				s.ErrorIs(err, tc.ExpectedErr)
				s.Empty(result.AuditRecords)
			} else {
				s.NoError(err)
			}
		})
	}
}

// TEST 13: List functionality - Edge cases
func (s *AuditRecordRepositoryTestSuite) TestList_EdgeCases() {
	s.setupListTestData()

	s.Run("very high limit", func() {
		query := utils.NewRQLQuery("", defaultListOffset, 10000, []rql.Filter{}, []rql.Sort{}, []string{})

		result, err := s.repository.List(s.ctx, query)
		s.NoError(err)
		s.LessOrEqual(len(result.AuditRecords), len(s.auditRecords)) // Should not exceed actual data
		s.GreaterOrEqual(result.Page.TotalCount, int64(len(s.auditRecords)))
	})

	s.Run("unicode search", func() {
		query := utils.NewRQLQuery("Léon", defaultListOffset, defaultListLimit, []rql.Filter{}, []rql.Sort{}, []string{})

		result, err := s.repository.List(s.ctx, query)
		s.NoError(err)
		// Should find the record with unicode characters
		if len(result.AuditRecords) > 0 {
			found := false
			for _, record := range result.AuditRecords {
				if strings.Contains(record.Actor.Name, "Léon") {
					found = true
					break
				}
			}
			s.True(found, "Should find record with unicode characters")
		}
	})

	s.Run("nil query parameter", func() {
		result, err := s.repository.List(s.ctx, nil)
		s.NoError(err)
		s.GreaterOrEqual(len(result.AuditRecords), 0)
		s.GreaterOrEqual(result.Page.TotalCount, int64(0))
	})
}

// TestAuditRecordRepositoryTestSuite is the entry point for the test suite
func TestAuditRecordRepositoryTestSuite(t *testing.T) {
	suite.Run(t, new(AuditRecordRepositoryTestSuite))
}
