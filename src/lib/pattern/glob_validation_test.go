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

package pattern

import (
	"testing"

	"github.com/bmatcuk/doublestar"
	"github.com/stretchr/testify/require"
)

func TestGlobValidationMatchesV1Grammar(t *testing.T) {
	for _, expression := range []string{`{[{],a}`, `{[}],a}`, `{[a,b],c}`, `[\\]`, `[a-]`, `[-a]`, `[]`, `{unclosed`, `trailing\`} {
		t.Run(expression, func(t *testing.T) { require.Error(t, Validate(expression, KindDoublestar)) })
	}
	for _, expression := range []string{`[^]`, `[{]`, `[}]`, `[\]]`, `[a\-]`, `[a-\]]`, `{[a,[b}]`, `[a-z]/{a,b}`, `foo}`, `{a,{b,c}}`} {
		t.Run(expression, func(t *testing.T) {
			m, err := Compile(expression, KindDoublestar)
			require.NoError(t, err)
			for _, value := range []string{"a", "b", "{", "}", "]", "-", "a/a", "foo}", "\\"} {
				want, err := doublestar.Match(expression, value)
				require.NoError(t, err)
				require.Equal(t, want, m.MatchString(value), "pattern %q value %q", expression, value)
			}
		})
	}
}
