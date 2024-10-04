package postgres

import (
	"fmt"
	"testing"

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
