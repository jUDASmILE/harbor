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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestValidateSelectors(t *testing.T) {
	tests := []struct {
		name       string
		kind       string
		decoration string
		pattern    string
		invalid    bool
	}{
		{"legacy default", "", "matches", "v*.*.*", false},
		{"empty regex", "regex", "matches", "", false},
		{"final releases", "regex", "matches", `v?\d+\.\d+\.\d+`, false},
		{"excluding prereleases", "regex", "excludes", `.*-(alpha|beta|rc)\.?\d*`, false},
		{"regex quantifier", "regex", "matches", `v[0-9]{2,3}`, false},
		{"invalid regex", "regex", "matches", `v[`, true},
		{"lookahead", "regex", "matches", `v(?!-rc)`, true},
		{"too long", "regex", "matches", strings.Repeat("a", 513), true},
		{"unknown kind", "pcre", "matches", "", true},
		{"wrong tag decoration", "regex", "repoMatches", ".*", true},
		{"malformed glob", "doublestar", "matches", "{release,dev", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Metadata{
				TagSelectors:   []*Selector{{Kind: tt.kind, Decoration: tt.decoration, Pattern: tt.pattern}},
				ScopeSelectors: map[string][]*Selector{"repository": {{Kind: "doublestar", Decoration: "repoMatches", Pattern: "library/**"}}},
			}
			if tt.invalid {
				require.Error(t, m.ValidateSelectors())
			} else {
				require.NoError(t, m.ValidateSelectors())
			}
		})
	}
	t.Run("validate all repository selectors", func(t *testing.T) {
		m := &Metadata{ScopeSelectors: map[string][]*Selector{"repository": {
			{Kind: "doublestar", Decoration: "repoMatches", Pattern: "**"},
			{Kind: "regex", Decoration: "repoExcludes", Pattern: "["},
		}}}
		require.Error(t, m.ValidateSelectors())
	})
	t.Run("reject null selector", func(t *testing.T) {
		require.Error(t, (&Metadata{TagSelectors: []*Selector{nil}}).ValidateSelectors())
	})
}
