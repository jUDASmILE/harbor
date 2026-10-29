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

package rule

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/selector"
	selectorindex "github.com/goharbor/harbor/src/lib/selector/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

func TestRegexReleaseSelectors(t *testing.T) {
	release := []*model.Selector{{Kind: "regex", Decoration: "matches", Pattern: `v?\d+\.\d+\.\d+`}}
	for _, tc := range []struct {
		tag       string
		immutable bool
	}{
		{"v1.9.0", true}, {"v1.10.3", true}, {"v2.0.0-rc.1", false}, {"v1.9.0-12-ge7e744a7", false}, {"latest", false},
	} {
		t.Run(tc.tag, func(t *testing.T) {
			matched, err := dimensionSelects(release, `{"untagged":false}`, &selector.Candidate{Tags: []string{tc.tag}}, selectorindex.Get)
			require.NoError(t, err)
			require.Equal(t, tc.immutable, matched)
		})
	}
	// Sharing a digest with a release tag does not make the latest tag immutable.
	matched, err := dimensionSelects(release, `{"untagged":false}`, &selector.Candidate{Tags: []string{"latest"}, Digest: "shared"}, selectorindex.Get)
	require.NoError(t, err)
	require.False(t, matched)
}
