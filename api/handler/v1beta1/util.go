package v1beta1

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// HTTP Codes defined here:
// https://github.com/grpc-ecosystem/grpc-gateway/blob/master/runtime/errors.go#L36

var (
	grpcInternalServerError = status.Errorf(codes.Internal, internalServerError.Error())
	grpcBadBodyError        = status.Error(codes.InvalidArgument, badRequestError.Error())
	grpcPermissionDenied    = status.Error(codes.PermissionDenied, permissionDeniedError.Error())
)

func mapOfStringValues(m map[string]interface{}) (map[string]string, error) {
	newMap := make(map[string]string)

	for key, value := range m {
		switch value := value.(type) {
		case string:
			newMap[key] = value
		default:
			return map[string]string{}, fmt.Errorf("value for %s key is not string", key)
		}
	}

	return newMap, nil
}

func mapOfInterfaceValues(m map[string]string) map[string]interface{} {
	newMap := make(map[string]interface{})

	for key, value := range m {
		newMap[key] = value
	}

	return newMap
}

func generateSlug(name string) string {
	preProcessed := strings.ReplaceAll(strings.TrimSpace(strings.TrimSpace(name)), "_", "-")
	return strings.Join(
		strings.Split(preProcessed, " "),
		"-",
	)
}
