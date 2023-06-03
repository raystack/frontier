package health

import (
	"context"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/status"
)

func TestHealthCheck(t *testing.T) {
	type testCase struct {
		Description  string
		Request      *grpc_health_v1.HealthCheckRequest
		ExpectStatus codes.Code
		PostCheck    func(resp *grpc_health_v1.HealthCheckResponse) error
	}

	var testCases = []testCase{
		{
			Description:  `should return OK if server is serving`,
			ExpectStatus: codes.OK,
			Request:      &grpc_health_v1.HealthCheckRequest{Service: "compass"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			ctx := context.Background()

			healthHandler := &HealthHandler{}

			got, err := healthHandler.Check(ctx, tc.Request)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
				return
			}
			if tc.PostCheck != nil {
				if err := tc.PostCheck(got); err != nil {
					t.Error(err)
					return
				}
			}
		})
	}
}

func TestHealthWatch(t *testing.T) {
	type testCase struct {
		Description  string
		Request      *grpc_health_v1.HealthCheckRequest
		ExpectStatus codes.Code
		PostCheck    func(wc grpc_health_v1.Health_WatchClient) error
	}

	var testCases = []testCase{
		{
			Description:  `should return unimplemented`,
			ExpectStatus: codes.Unimplemented,
			Request:      &grpc_health_v1.HealthCheckRequest{Service: "compass"},
		},
	}
	for _, tc := range testCases {
		t.Run(tc.Description, func(t *testing.T) {
			healthHandler := NewHandler()

			err := healthHandler.Watch(tc.Request, nil)
			code := status.Code(err)
			if code != tc.ExpectStatus {
				t.Errorf("expected handler to return Code %s, returned Code %s instead", tc.ExpectStatus.String(), code.String())
				return
			}
		})
	}
}
