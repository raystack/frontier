package credit_test

import (
	"context"
	"errors"
	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/credit/mocks"
	"testing"
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
		{
			name: "should return an error if there is an error in fetching transactions",
			args: args{
				accountID: "1",
			},
			want:    0,
			wantErr: true,
			setup: func() *credit.Service {
				s, mockTransaction := mockService(t)
				mockTransaction.EXPECT().GetBalance(ctx, "1").Return(0, errors.New("An error occurred"))
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

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	type args struct {
		transactionID string
	}
	tests := []struct {
		name    string
		args    args
		want    credit.Transaction
		wantErr bool
		setup   func() *credit.Service
	}{
		{
			name: "should return transaction if the transaction id passed is valid",
			args: args{
				transactionID: "1",
			},
			want: credit.Transaction{
				ID:         "1",
				CustomerID: "abc",
			},
			wantErr: false,
			setup: func() *credit.Service {
				s, mockTransaction := mockService(t)
				mockTransaction.EXPECT().GetByID(ctx, "1").Return(credit.Transaction{
					ID:         "1",
					CustomerID: "abc",
				}, nil)
				return s
			},
		},
		{
			name: "should return error if there is an error in fetching the transaction",
			args: args{
				transactionID: "1",
			},
			want:    credit.Transaction{},
			wantErr: true,
			setup: func() *credit.Service {
				s, mockTransaction := mockService(t)
				mockTransaction.EXPECT().GetByID(ctx, "1").Return(credit.Transaction{}, errors.New("An error occurred"))
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.GetByID(ctx, tt.args.transactionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetBalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got.ID != tt.want.ID) || (got.CustomerID != tt.want.CustomerID) {
				t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
			}
		})
	}
}
