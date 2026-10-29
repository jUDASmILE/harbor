// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package model

import (
	"github.com/goharbor/harbor/src/lib/errors"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
)

func TestFilterPatternValidation(t *testing.T) {
	tests := []struct {
		name   string
		filter Filter
		valid  bool
	}{
		{"legacy glob", Filter{Type: FilterTypeName, Value: "library/**"}, true},
		{"regex tag", Filter{Type: FilterTypeTag, Kind: "regex", Value: `v\d+\.\d+\.\d+`}, true},
		{"invalid regex", Filter{Type: FilterTypeTag, Kind: "regex", Value: `[`}, false},
		{"unsupported lookahead", Filter{Type: FilterTypeName, Kind: "regex", Value: `(?!dev).*`}, false},
		{"too long", Filter{Type: FilterTypeTag, Kind: "regex", Value: strings.Repeat("x", 513)}, false},
		{"unknown kind even if empty", Filter{Type: FilterTypeTag, Kind: "pcre", Value: ""}, false},
		{"regex label", Filter{Type: FilterTypeLabel, Kind: "regex", Value: []any{`release-\d+`}}, true},
		{"invalid regex label", Filter{Type: FilterTypeLabel, Kind: "regex", Value: []any{`[`}}, false},
		{"legacy exact label", Filter{Type: FilterTypeLabel, Value: []any{"[exact]"}}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.filter.Validate()
			if tt.valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.True(t, errors.IsErr(err, errors.BadRequestCode))
			}
		})
	}
}
