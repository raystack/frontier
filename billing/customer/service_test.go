package customer_test

import (
	"context"
	"github.com/google/go-cmp/cmp"
	"github.com/raystack/frontier/billing"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/customer/mocks"
	stripemock "github.com/raystack/frontier/billing/stripetest/mocks"
	"github.com/stripe/stripe-go/v79"
	"github.com/stripe/stripe-go/v79/client"
	"testing"
)

func mockService(t *testing.T) (*client.API, *stripemock.Backend, *mocks.Repository) {
	t.Helper()
	mockRepository := mocks.NewRepository(t)
	mockBackend := stripemock.NewBackend(t)

	stripeClient := client.New("key_123", &stripe.Backends{
		API: mockBackend,
	})

	return stripeClient, mockBackend, mockRepository
}

func TestService_Create(t *testing.T) {
	ctx := context.Background()
	type args struct {
		customer customer.Customer
		offline  bool
	}
	tests := []struct {
		name    string
		args    args
		want    customer.Customer
		wantErr bool
		setup   func() *customer.Service
	}{
		{
			name: "should return error if active customer already exists",
			args: args{
				customer: customer.Customer{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
					State: customer.ActiveState,
				},
				offline: false,
			},
			want:    customer.Customer{},
			wantErr: true,
			setup: func() *customer.Service {
				stripeClient, _, mockRepo := mockService(t)

				mockRepo.EXPECT().List(ctx, customer.Filter{
					OrgID: "org1",
					State: customer.ActiveState,
				}).Return([]customer.Customer{{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
					State: customer.ActiveState,
				},
				}, nil) // existing active accounts

				cfg := billing.Config{}

				return customer.NewService(stripeClient, mockRepo, cfg)
			},
		},
		{
			name: "should create customer if no existing active accounts",
			args: args{
				customer: customer.Customer{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
				},
				offline: false,
			},
			want: customer.Customer{
				ID:    "1",
				Name:  "customer1",
				OrgID: "org1",
				State: customer.ActiveState,
			},
			wantErr: false,
			setup: func() *customer.Service {
				stripeClient, mockStripeBackend, mockRepo := mockService(t)

				mockRepo.EXPECT().List(ctx, customer.Filter{
					OrgID: "org1",
					State: customer.ActiveState,
				}).Return([]customer.Customer{}, nil) // No existing active accounts

				mockStripeBackend.EXPECT().Call("POST", "/v1/customers", "key_123",
					&stripe.CustomerParams{
						Params: stripe.Params{
							Context: ctx,
						},
						Address: &stripe.AddressParams{
							City:       stripe.String(""),
							Country:    stripe.String(""),
							Line1:      stripe.String(""),
							Line2:      stripe.String(""),
							PostalCode: stripe.String(""),
							State:      stripe.String(""),
						},
						Email:     stripe.String(""),
						Name:      stripe.String("customer1"),
						Phone:     stripe.String(""),
						TaxIDData: nil,
						Metadata: map[string]string{
							"managed_by": "frontier",
							"org_id":     "org1",
						},
					}, &stripe.Customer{ID: ""}).Return(nil)

				mockRepo.EXPECT().Create(ctx, customer.Customer{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
					State: customer.ActiveState,
				}).Return(customer.Customer{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
					State: customer.ActiveState,
				}, nil)

				cfg := billing.Config{}

				return customer.NewService(stripeClient, mockRepo, cfg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Create(ctx, tt.args.customer, tt.args.offline)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-)want (+)got:\n%s", diff)
			}
		})
	}
}

func TestService_Update(t *testing.T) {
	ctx := context.Background()
	type args struct {
		customer customer.Customer
	}
	tests := []struct {
		name    string
		args    args
		want    customer.Customer
		wantErr bool
		setup   func() *customer.Service
	}{
		{
			name: "successfully update existing customer",
			args: args{
				customer: customer.Customer{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
				},
			},
			want: customer.Customer{
				ID:    "1",
				Name:  "updated_customer1",
				OrgID: "org1",
			},
			setup: func() *customer.Service {
				stripeClient, mockStripeBackend, mockRepo := mockService(t)

				mockRepo.EXPECT().GetByID(ctx, "1").Return(customer.Customer{
					ID:    "1",
					Name:  "customer1",
					OrgID: "org1",
				}, nil)

				mockStripeBackend.EXPECT().Call("POST", "/v1/customers/", "key_123",
					&stripe.CustomerParams{
						Params: stripe.Params{
							Context: ctx,
						},
						Address: &stripe.AddressParams{
							City:       stripe.String(""),
							Country:    stripe.String(""),
							Line1:      stripe.String(""),
							Line2:      stripe.String(""),
							PostalCode: stripe.String(""),
							State:      stripe.String(""),
						},
						Email:     stripe.String(""),
						Name:      stripe.String("customer1"),
						Phone:     stripe.String(""),
						TaxIDData: nil,
						Metadata: map[string]string{
							"managed_by": "frontier",
							"org_id":     "org1",
						},
					}, &stripe.Customer{ID: ""}).Return(nil)

				mockRepo.EXPECT().
					UpdateByID(ctx, customer.Customer{
						ID:    "1",
						Name:  "customer1",
						OrgID: "org1"}).
					Return(customer.Customer{
						ID:    "1",
						Name:  "updated_customer1",
						OrgID: "org1",
					}, nil)

				cfg := billing.Config{}

				return customer.NewService(stripeClient, mockRepo, cfg)
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.Update(ctx, tt.args.customer)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-)want (+)got:\n%s", diff)
			}
		})
	}
}

func TestService_GetByID(t *testing.T) {
	ctx := context.Background()
	type args struct {
		id string
	}
	tests := []struct {
		name    string
		args    args
		want    customer.Customer
		wantErr bool
		setup   func() *customer.Service
	}{
		{
			name: "successfully get customer by id",
			args: args{
				id: "1",
			},
			want: customer.Customer{
				ID:    "1",
				Name:  "customer1",
				OrgID: "org1",
			},
			wantErr: false,
			setup: func() *customer.Service {
				stripeClient, _, mockRepo := mockService(t)
				mockRepo.EXPECT().GetByID(ctx, "1").Return(
					customer.Customer{
						ID:    "1",
						Name:  "customer1",
						OrgID: "org1",
					}, nil)

				cfg := billing.Config{}

				return customer.NewService(stripeClient, mockRepo, cfg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.GetByID(ctx, tt.args.id)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-)want (+)got:\n%s", diff)
			}
		})
	}

}

func TestService_List(t *testing.T) {
	ctx := context.Background()
	type args struct {
		filter customer.Filter
	}
	tests := []struct {
		name    string
		args    args
		want    []customer.Customer
		wantErr bool
		setup   func() *customer.Service
	}{
		{
			name: "List all billing customers",
			args: args{
				filter: customer.Filter{},
			},
			want: []customer.Customer{
				{ID: "1",
					Name:  "customer1",
					OrgID: "org1"},
				{ID: "2",
					Name:  "customer2",
					OrgID: "org1"},
				{ID: "3",
					Name:  "customer3",
					OrgID: "org1"},
			},
			wantErr: false,
			setup: func() *customer.Service {
				stripeClient, _, mockRepo := mockService(t)
				mockRepo.EXPECT().List(ctx, customer.Filter{}).Return(
					[]customer.Customer{
						{ID: "1",
							Name:  "customer1",
							OrgID: "org1"},
						{ID: "2",
							Name:  "customer2",
							OrgID: "org1"},
						{ID: "3",
							Name:  "customer3",
							OrgID: "org1"},
					}, nil)

				cfg := billing.Config{}

				return customer.NewService(stripeClient, mockRepo, cfg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := tt.setup()
			got, err := s.List(ctx, tt.args.filter)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if diff := cmp.Diff(tt.want, got); diff != "" {
				t.Errorf("diff (-)want (+)got:\n%s", diff)
			}
		})
	}

}
