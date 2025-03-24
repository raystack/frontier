package prospect_test

import (
	"context"
	"errors"
	"testing"

	"github.com/raystack/frontier/core/prospect"
	"github.com/raystack/frontier/core/prospect/mocks"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var sampleError = errors.New("sample error")

func mockService(t *testing.T) *mocks.Repository {
	t.Helper()

	return mocks.NewRepository(t)
}

func TestService_Get(t *testing.T) {
	ctx := context.Background()
	prospectId := "123"

	type args struct {
		ctx        context.Context
		prospectID string
	}
	tests := []struct {
		name    string
		args    args
		want    prospect.Prospect
		wantErr error
		setup   func() *prospect.Service
	}{
		{
			name: "return error if repository.Get returns error",
			args: args{
				ctx:        ctx,
				prospectID: prospectId,
			},
			wantErr: sampleError,
			want:    prospect.Prospect{},
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Get", ctx, prospectId).Return(prospect.Prospect{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return prospect object if repository.Get returns success",
			args: args{
				ctx:        ctx,
				prospectID: prospectId,
			},
			want:    prospect.Prospect{ID: "123"},
			wantErr: nil,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Get", ctx, prospectId).Return(prospect.Prospect{ID: "123"}, nil).Once()
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			_, err := s.Get(ctx, tt.args.prospectID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_Delete(t *testing.T) {
	ctx := context.Background()
	prospectId := "123"

	type args struct {
		ctx        context.Context
		prospectID string
	}
	tests := []struct {
		name    string
		args    args
		wantErr error
		setup   func() *prospect.Service
	}{
		{
			name: "return error if repository.Delete returns error",
			args: args{
				ctx:        ctx,
				prospectID: prospectId,
			},
			wantErr: sampleError,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Delete", ctx, prospectId).Return(sampleError).Once()
				return service
			},
		},
		{
			name: "return nil if repository.Delete returns success",
			args: args{
				ctx:        ctx,
				prospectID: prospectId,
			},
			wantErr: nil,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Delete", ctx, prospectId).Return(nil).Once()
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			err := s.Delete(ctx, tt.args.prospectID)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	testProspect := prospect.Prospect{
		Name:  "John Doe",
		Email: "john@example.com",
	}

	type args struct {
		ctx      context.Context
		prospect prospect.Prospect
	}
	tests := []struct {
		name    string
		args    args
		want    prospect.Prospect
		wantErr error
		setup   func() *prospect.Service
	}{
		{
			name: "return error if repository.Create returns error",
			args: args{
				ctx:      ctx,
				prospect: testProspect,
			},
			want:    prospect.Prospect{},
			wantErr: sampleError,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Create", ctx, testProspect).Return(prospect.Prospect{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return prospect if repository.Create returns success",
			args: args{
				ctx:      ctx,
				prospect: testProspect,
			},
			want:    testProspect,
			wantErr: nil,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Create", ctx, testProspect).Return(testProspect, nil).Once()
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(tt.args.ctx, tt.args.prospect)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_List(t *testing.T) {
	ctx := context.Background()

	type args struct {
		ctx   context.Context
		query *rql.Query
	}
	testQuery := &rql.Query{}

	tests := []struct {
		name    string
		args    args
		want    prospect.ListProspects
		wantErr error
		setup   func() *prospect.Service
	}{
		{
			name: "return error if repository.List returns error",
			args: args{
				ctx:   ctx,
				query: testQuery,
			},
			wantErr: sampleError,
			want:    prospect.ListProspects{},
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("List", ctx, testQuery).Return(prospect.ListProspects{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return list of prospects if repository.List returns success",
			args: args{
				ctx:   ctx,
				query: testQuery,
			},
			want: prospect.ListProspects{
				Prospects: []prospect.Prospect{
					{ID: "123", Name: "John Doe", Email: "john@example.com"},
					{ID: "456", Name: "Jane Doe", Email: "jane@example.com"}},
			},
			wantErr: nil,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("List", ctx, testQuery).Return(prospect.ListProspects{
					Prospects: []prospect.Prospect{
						{ID: "123", Name: "John Doe", Email: "john@example.com"},
						{ID: "456", Name: "Jane Doe", Email: "jane@example.com"}},
				}, nil).Once()
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.List(tt.args.ctx, tt.args.query)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()
	testProspect := prospect.Prospect{
		ID:    "123",
		Name:  "John Doe",
		Email: "john@example.com",
	}

	type args struct {
		ctx      context.Context
		prospect prospect.Prospect
	}
	tests := []struct {
		name    string
		args    args
		want    prospect.Prospect
		wantErr error
		setup   func() *prospect.Service
	}{
		{
			name: "return error if repository.Update returns error",
			args: args{
				ctx:      ctx,
				prospect: testProspect,
			},
			want:    prospect.Prospect{},
			wantErr: sampleError,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Update", ctx, testProspect).Return(prospect.Prospect{}, sampleError).Once()
				return service
			},
		},
		{
			name: "return updated prospect if repository.Update returns success",
			args: args{
				ctx:      ctx,
				prospect: testProspect,
			},
			want:    testProspect,
			wantErr: nil,
			setup: func() *prospect.Service {
				repository := mockService(t)
				service := prospect.NewService(repository)

				repository.On("Update", ctx, testProspect).Return(testProspect, nil).Once()
				return service
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Update(tt.args.ctx, tt.args.prospect)

			if tt.wantErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}
