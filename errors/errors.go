/*
Copyright 2025 Limrun, Inc.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

	http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

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
