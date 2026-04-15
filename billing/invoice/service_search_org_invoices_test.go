package invoice

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/raystack/frontier/billing/credit"
	"github.com/raystack/frontier/billing/customer"
	"github.com/raystack/frontier/billing/product"
	"github.com/raystack/frontier/pkg/db"
	"github.com/raystack/salt/rql"
	"github.com/stretchr/testify/assert"
)

type searchOrgInvoicesRepoStub struct {
	searchFn func(ctx context.Context, customerID string, rqlQuery *rql.Query) (SearchOrgInvoicesResult, error)
}

func (s searchOrgInvoicesRepoStub) Create(context.Context, Invoice) (Invoice, error) {
	return Invoice{}, errors.New("not implemented")
}
func (s searchOrgInvoicesRepoStub) GetByID(context.Context, string) (Invoice, error) {
	return Invoice{}, errors.New("not implemented")
}
func (s searchOrgInvoicesRepoStub) List(context.Context, Filter) ([]Invoice, error) {
	return nil, errors.New("not implemented")
}
func (s searchOrgInvoicesRepoStub) SearchOrgInvoices(ctx context.Context, customerID string, rqlQuery *rql.Query) (SearchOrgInvoicesResult, error) {
	if s.searchFn != nil {
		return s.searchFn(ctx, customerID, rqlQuery)
	}
	return SearchOrgInvoicesResult{}, nil
}
func (s searchOrgInvoicesRepoStub) UpdateByID(context.Context, Invoice) (Invoice, error) {
	return Invoice{}, errors.New("not implemented")
}
func (s searchOrgInvoicesRepoStub) Delete(context.Context, string) error {
	return errors.New("not implemented")
}
func (s searchOrgInvoicesRepoStub) Search(context.Context, *rql.Query) ([]InvoiceWithOrganization, error) {
	return nil, errors.New("not implemented")
}

type noopCustomerService struct{}

func (noopCustomerService) GetByID(context.Context, string) (customer.Customer, error) {
	return customer.Customer{}, nil
}
func (noopCustomerService) List(context.Context, customer.Filter) ([]customer.Customer, error) {
	return nil, nil
}
func (noopCustomerService) GetDetails(context.Context, string) (customer.Details, error) {
	return customer.Details{}, nil
}

type noopCreditService struct{}

func (noopCreditService) Add(context.Context, credit.Credit) error { return nil }
func (noopCreditService) GetBalanceForRange(context.Context, string, time.Time, time.Time) (int64, error) {
	return 0, nil
}
func (noopCreditService) GetBalanceForRangeWithoutOverdraft(context.Context, string, time.Time, time.Time) (int64, error) {
	return 0, nil
}

type noopProductService struct{}

func (noopProductService) GetByID(context.Context, string) (product.Product, error) {
	return product.Product{}, nil
}

type noopLocker struct{}

func (noopLocker) TryLock(context.Context, string) (*db.Lock, error) { return nil, nil }

func TestService_SearchOrgInvoices_Validation(t *testing.T) {
	svc := &Service{
		repository:      searchOrgInvoicesRepoStub{},
		customerService: noopCustomerService{},
		creditService:   noopCreditService{},
		productService:  noopProductService{},
		locker:          noopLocker{},
	}

	_, err := svc.SearchOrgInvoices(context.Background(), "", &rql.Query{})
	assert.EqualError(t, err, "customer id not found")

	_, err = svc.SearchOrgInvoices(context.Background(), "cust-1", nil)
	assert.NoError(t, err)

	_, err = svc.SearchOrgInvoices(context.Background(), "cust-1", &rql.Query{
		GroupBy: []string{"state"},
	})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "group_by is not supported")
}

func TestService_SearchOrgInvoices_DelegatesToRepository(t *testing.T) {
	expected := SearchOrgInvoicesResult{
		Invoices: []Invoice{{ID: "inv-1"}},
		Pagination: SearchOrgInvoicesPagination{
			TotalCount: 1,
		},
	}
	var called bool

	svc := &Service{
		repository: searchOrgInvoicesRepoStub{
			searchFn: func(ctx context.Context, customerID string, rqlQuery *rql.Query) (SearchOrgInvoicesResult, error) {
				called = true
				assert.Equal(t, "cust-1", customerID)
				assert.Equal(t, 10, rqlQuery.Limit)
				return expected, nil
			},
		},
	}

	got, err := svc.SearchOrgInvoices(context.Background(), "cust-1", &rql.Query{Limit: 10})
	assert.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, expected, got)
}
