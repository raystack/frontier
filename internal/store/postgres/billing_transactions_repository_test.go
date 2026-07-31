package postgres

import (
	"errors"
	"fmt"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
)

func Test_isSufficientBalance(t *testing.T) {
	type args struct {
		customerMinLimit int64
		currentBalance   int64
		txAmount         int64
	}
	tests := []struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		{
			name: "sufficient balance with 0 limit",
			args: args{
				customerMinLimit: 0,
				currentBalance:   1000,
				txAmount:         100,
			},
			wantErr: assert.NoError,
		},
		{
			name: "sufficient balance with positive limit",
			args: args{
				customerMinLimit: 100,
				currentBalance:   1000,
				txAmount:         100,
			},
			wantErr: assert.NoError,
		},
		{
			name: "sufficient balance with negative limit",
			args: args{
				customerMinLimit: -100,
				currentBalance:   1000,
				txAmount:         100,
			},
			wantErr: assert.NoError,
		},
		{
			name: "insufficient balance with positive limit",
			args: args{
				customerMinLimit: 100,
				currentBalance:   80,
				txAmount:         100,
			},
			wantErr: assert.Error,
		},
		{
			name: "insufficient balance with 0 limit",
			args: args{
				customerMinLimit: 0,
				currentBalance:   80,
				txAmount:         100,
			},
			wantErr: assert.Error,
		},
		{
			name: "insufficient balance with sufficient negative limit",
			args: args{
				customerMinLimit: -100,
				currentBalance:   80,
				txAmount:         100,
			},
			wantErr: assert.NoError,
		},
		{
			name: "insufficient balance with insufficient negative limit",
			args: args{
				customerMinLimit: -100,
				currentBalance:   80,
				txAmount:         200,
			},
			wantErr: assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.wantErr(t, isSufficientBalance(tt.args.customerMinLimit, tt.args.currentBalance, tt.args.txAmount), fmt.Sprintf("isSufficientBalance(%v, %v, %v)", tt.args.customerMinLimit, tt.args.currentBalance, tt.args.txAmount))
		})
	}
}

func TestCheckPostgresError(t *testing.T) {
	tests := []struct {
		name string
		code string
		want error
	}{
		{"unique violation", pgerrcode.UniqueViolation, ErrDuplicateKey},
		{"check violation", pgerrcode.CheckViolation, ErrCheckViolation},
		{"foreign key violation", pgerrcode.ForeignKeyViolation, ErrForeignKeyViolation},
		{"invalid text representation", pgerrcode.InvalidTextRepresentation, ErrInvalidTextRepresentation},
		{"user-defined immutability code", pgUserDefinedImmutabilityError, ErrImmutableRecord},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := checkPostgresError(&pgconn.PgError{Code: tt.code})
			assert.ErrorIs(t, got, tt.want)
		})
	}

	t.Run("wrapped pg error still matches", func(t *testing.T) {
		wrapped := fmt.Errorf("exec failed: %w", &pgconn.PgError{Code: pgerrcode.UniqueViolation})
		assert.ErrorIs(t, checkPostgresError(wrapped), ErrDuplicateKey)
	})

	t.Run("unmapped pg code passes through unchanged", func(t *testing.T) {
		in := &pgconn.PgError{Code: pgerrcode.SerializationFailure}
		assert.Equal(t, error(in), checkPostgresError(in))
	})

	t.Run("non-postgres error passes through unchanged", func(t *testing.T) {
		in := errors.New("plain error")
		assert.Equal(t, in, checkPostgresError(in))
	})
}
