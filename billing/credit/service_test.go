package credit_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/credit/mocks"
	"github.com/raystack/frontier/internal/bootstrap/schema"
	"github.com/raystack/frontier/pkg/metadata"
	"github.com/stretchr/testify/mock"
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

func TestService_Add(t *testing.T) {
	ctx := context.Background()
	dummyError := errors.New("dummy error")
	type args struct {
		cred credit.Credit
	}
	tests := []struct {
		name  string
		args  args
		want  error
		setup func() *credit.Service
	}{
		{
			name: "should return an error if credit id is empty",
			args: args{
				cred: credit.Credit{
					ID: "",
				},
			},
			want: errors.New("credit id is empty, it is required to create a transaction"),
			setup: func() *credit.Service {
				s, _ := mockService(t)
				return s
			},
		},
		{
			name: "should return an error if credit amount less than zero",
			args: args{
				cred: credit.Credit{
					ID:     "12",
					Amount: -10,
				},
			},
			want: errors.New("credit amount is negative"),
			setup: func() *credit.Service {
				s, _ := mockService(t)
				return s
			},
		},
		{
			name: "should return an error if a transaction has already been created with that id",
			args: args{
				cred: credit.Credit{
					ID:     "12",
					Amount: 10,
				},
			},
			want: credit.ErrAlreadyApplied,
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetByID(ctx, "12").Return(credit.Transaction{ID: "12"}, nil)
				return s
			},
		},
		{
			name: "should return an error if there is an error in checking if the transaction already exists",
			args: args{
				cred: credit.Credit{
					ID:     "12",
					Amount: 10,
				},
			},
			want: dummyError,
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetByID(ctx, "12").Return(credit.Transaction{}, dummyError)
				return s
			},
		},
		{
			name: "should return an error if there is an error in creating transaction entry",
			args: args{
				cred: credit.Credit{
					ID:     "12",
					Amount: 10,
				},
			},
			want: errors.New(fmt.Sprintf("transactionRepository.CreateEntry: %v", dummyError)),
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetByID(ctx, "12").Return(credit.Transaction{}, nil)
				mockTransactionRepo.EXPECT().CreateEntry(ctx, mock.Anything, mock.Anything).Return([]credit.Transaction{}, dummyError)
				return s
			},
		},
		{
			name: "should create a transaction entry if parameters passed are correct",
			args: args{
				cred: credit.Credit{
					ID:         "12",
					Amount:     10,
					CustomerID: "",
					Metadata: metadata.Metadata{
						"a": "a",
					},
				},
			},
			want: nil,
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetByID(ctx, "12").Return(credit.Transaction{}, nil)
				mockTransactionRepo.EXPECT().CreateEntry(ctx, credit.Transaction{CustomerID: schema.PlatformOrgID.String(), Type: credit.DebitType, Amount: 10, Source: "system", Metadata: metadata.Metadata{"a": "a"}}, credit.Transaction{Type: credit.CreditType, Amount: 10, ID: "12", Source: "system", Metadata: metadata.Metadata{"a": "a"}}).Return([]credit.Transaction{}, nil)
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got := s.Add(ctx, tt.args.cred)
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
				}
			} else {
				if got == nil || got.Error() != tt.want.Error() {
					t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}

func TestService_Deduct(t *testing.T) {
	ctx := context.Background()
	dummyError := errors.New("dummy error")
	type args struct {
		cred credit.Credit
	}
	tests := []struct {
		name  string
		args  args
		want  error
		setup func() *credit.Service
	}{
		{
			name: "should return an error if credit id is empty",
			args: args{
				cred: credit.Credit{
					ID: "",
				},
			},
			want: errors.New("credit id is empty, it is required to create a transaction"),
			setup: func() *credit.Service {
				s, _ := mockService(t)
				return s
			},
		},
		{
			name: "should return an error if credit amount less than zero",
			args: args{
				cred: credit.Credit{
					ID:     "12",
					Amount: -10,
				},
			},
			want: errors.New("credit amount is negative"),
			setup: func() *credit.Service {
				s, _ := mockService(t)
				return s
			},
		},
		{
			name: "should return an error if balance cannot be fetched",
			args: args{
				cred: credit.Credit{
					ID:         "12",
					CustomerID: "customer_id",
					Amount:     10,
				},
			},
			want: errors.New(fmt.Sprintf("failed to apply transaction: %v", dummyError)),
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetBalance(ctx, "customer_id").Return(0, dummyError)
				return s
			},
		},
		{
			name: "should return ErrInsufficientCredits error if customer's balance is less than transaction amount",
			args: args{
				cred: credit.Credit{
					ID:         "12",
					CustomerID: "customer_id",
					Amount:     10,
				},
			},
			want: credit.ErrInsufficientCredits,
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetBalance(ctx, "customer_id").Return(5, nil)
				return s
			},
		},
		{
			name: "should return an error if there is an error in creating transaction entry",
			args: args{
				cred: credit.Credit{
					ID:         "12",
					Amount:     10,
					CustomerID: "customer_id",
				},
			},
			want: errors.New(fmt.Sprintf("failed to deduct credits: %v", dummyError)),
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetBalance(ctx, "customer_id").Return(20, nil)
				mockTransactionRepo.EXPECT().CreateEntry(ctx, mock.Anything, mock.Anything).Return([]credit.Transaction{}, dummyError)
				return s
			},
		},
		{
			name: "should deduct amount if parameters passed are correct",
			args: args{
				cred: credit.Credit{
					ID:         "12",
					Amount:     10,
					CustomerID: "customer_id",
					Metadata: metadata.Metadata{
						"a": "a",
					},
				},
			},
			want: nil,
			setup: func() *credit.Service {
				s, mockTransactionRepo := mockService(t)
				mockTransactionRepo.EXPECT().GetBalance(ctx, "customer_id").Return(20, nil)
				mockTransactionRepo.EXPECT().CreateEntry(ctx, credit.Transaction{ID: "12", CustomerID: "customer_id", Type: credit.DebitType, Amount: 10, Source: "system", Metadata: metadata.Metadata{"a": "a"}}, credit.Transaction{Type: credit.CreditType, CustomerID: schema.PlatformOrgID.String(), Amount: 10, Source: "system", Metadata: metadata.Metadata{"a": "a"}}).Return([]credit.Transaction{}, nil)
				return s
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got := s.Deduct(ctx, tt.args.cred)
			if tt.want == nil {
				if got != nil {
					t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
				}
			} else {
				if got == nil || got.Error() != tt.want.Error() {
					t.Errorf("GetBalance() got = %v, want %v", got, tt.want)
				}
			}
		})
	}
}
