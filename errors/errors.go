package errors

import (
	"errors"
	limrun "github.com/limrun-inc/go-sdk"
	"strings"
)

// IsUnauthenticated returns whether the API error means
// unauthenticated.
func IsUnauthenticated(err error) bool {
	var apiErr *limrun.Error
	ok := errors.As(err, &apiErr)
	if !ok {
		return false
	}
	b := string(apiErr.DumpResponse(true))
	return strings.Contains(b, `{"message":"unauthenticated:`)
}
