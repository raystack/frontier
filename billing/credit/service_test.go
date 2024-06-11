package credit_test

import (
	"context"
	"testing"

	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/credit/mocks"
)

func mockService(t *testing.T) (*credit.Service, *mocks.TransactionRepository) {
	t.Helper()
	mockTransaction := mocks.NewTransactionRepository(t)
	return credit.NewService(mockTransaction), mockTransaction
}

func TestService_GetBalance(t *testing.T) {
	ctx := context.Background()
	type args struct {
		accountID string
	}
	tests := []struct {
		name    string
		args    args
		want    int64
		wantErr bool
		setup   func() *credit.Service
	}{
		{
			name: "should return balance if transaction repository returns balance",
			args: args{
				accountID: "1",
			},
			want:    100,
			wantErr: false,
			setup: func() *credit.Service {
				s, mockTransaction := mockService(t)
				mockTransaction.EXPECT().GetBalance(ctx, "1").Return(100, nil)
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.GetBalance(ctx, tt.args.accountID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}
