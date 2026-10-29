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

package regexp

import (
	iselector "github.com/goharbor/harbor/src/lib/selector"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSelectTagsAndUntagged(t *testing.T) {
	candidates := []*iselector.Candidate{
		{Tags: []string{"v1.9.0", "latest"}, Digest: "shared"},
		{Tags: []string{"v2.0.0-rc.1"}},
		{Tags: []string{"v1.9.0-12-ge7e744a7"}},
		{Tags: []string{"latest"}},
		{},
	}
	for _, tc := range []struct {
		decoration, extras string
		indices            []int
	}{
		{Matches, `{"untagged":false}`, []int{0}},
		{Matches, `{"untagged":true}`, []int{0, 4}},
		{Excludes, `{"untagged":false}`, []int{0, 1, 2, 3, 4}},
		{Excludes, `{"untagged":true}`, []int{0, 1, 2, 3}},
	} {
		s, err := New(tc.decoration, `v?\d+\.\d+\.\d+`, tc.extras)
		require.NoError(t, err)
		actual, err := s.Select(candidates)
		require.NoError(t, err)
		var expected []*iselector.Candidate
		for _, i := range tc.indices {
			expected = append(expected, candidates[i])
		}
		require.Equal(t, expected, actual)
	}
	// Even an expression that matches an empty string cannot select an untagged
	// candidate unless the explicit flag allows it.
	s, err := New(Matches, `.*`, `{"untagged":false}`)
	require.NoError(t, err)
	actual, err := s.Select(candidates[4:])
	require.NoError(t, err)
	require.Empty(t, actual)
	_, err = New(Matches, `[`, "")
	require.Error(t, err)
}

func TestSelectRepositoriesAndNamespaces(t *testing.T) {
	candidates := []*iselector.Candidate{
		{Repository: "library/nginx", Namespace: "prod"},
		{Repository: "other/library/nginx", Namespace: "staging"},
	}
	for _, tc := range []struct {
		decoration, expression string
		expected               int
	}{
		{RepoMatches, `library/.*`, 0}, {RepoExcludes, `library/.*`, 1},
		{NSMatches, `prod`, 0}, {NSExcludes, `prod`, 1},
	} {
		s, err := New(tc.decoration, tc.expression, "")
		require.NoError(t, err)
		actual, err := s.Select(candidates)
		require.NoError(t, err)
		require.Equal(t, []*iselector.Candidate{candidates[tc.expected]}, actual)
	}
}
