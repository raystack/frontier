package customer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddress_HasMinimumRequiredAddress(t *testing.T) {
	tests := []struct {
		name     string
		address  Address
		expected bool
	}{
		{
			name: "complete address with postal code and country",
			address: Address{
				City:       "New York",
				Country:    "US",
				Line1:      "123 Main St",
				PostalCode: "10001",
				State:      "NY",
			},
			expected: true,
		},
		{
			name: "missing postal code",
			address: Address{
				City:    "New York",
				Country: "US",
				Line1:   "123 Main St",
				State:   "NY",
			},
			expected: false,
		},
		{
			name: "missing country",
			address: Address{
				City:       "New York",
				Line1:      "123 Main St",
				PostalCode: "10001",
				State:      "NY",
			},
			expected: false,
		},
		{
			name: "missing both postal code and country",
			address: Address{
				City:  "New York",
				Line1: "123 Main St",
				State: "NY",
			},
			expected: false,
		},
		{
			name: "only postal code and country present",
			address: Address{
				PostalCode: "10001",
				Country:    "US",
			},
			expected: true,
		},
		{
			name:     "empty address",
			address:  Address{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.address.HasMinimumRequiredAddress()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCustomer_HasMinimumRequiredAddress(t *testing.T) {
	tests := []struct {
		name     string
		customer Customer
		expected bool
	}{
		{
			name: "customer with complete address",
			customer: Customer{
				ID:    "1",
				Name:  "John Doe",
				Email: "john@example.com",
				Address: Address{
					City:       "New York",
					Country:    "US",
					Line1:      "123 Main St",
					PostalCode: "10001",
					State:      "NY",
				},
			},
			expected: true,
		},
		{
			name: "customer with incomplete address",
			customer: Customer{
				ID:    "1",
				Name:  "John Doe",
				Email: "john@example.com",
				Address: Address{
					City:  "New York",
					Line1: "123 Main St",
					State: "NY",
				},
			},
			expected: false,
		},
		{
			name: "customer with no address",
			customer: Customer{
				ID:    "1",
				Name:  "John Doe",
				Email: "john@example.com",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.customer.HasMinimumRequiredAddress()
			assert.Equal(t, tt.expected, result)
		})
	}
}