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

package handler

import (
	"context"
	"testing"

	repctlmodel "github.com/goharbor/harbor/src/controller/replication/model"
	"github.com/goharbor/harbor/src/pkg/reg/model"
	operation "github.com/goharbor/harbor/src/server/v2.0/restapi/operations/retention"
	"github.com/stretchr/testify/require"
)

func TestRegexSelectorMetadata(t *testing.T) {
	api := &retentionAPI{}
	response := api.GetRentenitionMetadata(context.Background(), operation.GetRentenitionMetadataParams{}).(*operation.GetRentenitionMetadataOK)
	tags, repositories := make(map[string][]string), make(map[string][]string)
	for _, item := range response.Payload.TagSelectors {
		tags[item.Kind] = item.Decorations
	}
	for _, item := range response.Payload.ScopeSelectors {
		repositories[item.Kind] = item.Decorations
	}
	for _, kind := range []string{"doublestar", "regex"} {
		require.Equal(t, []string{"matches", "excludes"}, tags[kind])
		require.Equal(t, []string{"repoMatches", "repoExcludes"}, repositories[kind])
	}
	again := api.GetRentenitionMetadata(context.Background(), operation.GetRentenitionMetadataParams{}).(*operation.GetRentenitionMetadataOK)
	require.Equal(t, response.Payload, again.Payload, "metadata requests must not mutate the shared template")
}

func TestRegexReplicationKindResponse(t *testing.T) {
	policy := &repctlmodel.Policy{Filters: []*model.Filter{
		{Type: model.FilterTypeTag, Kind: "regex", Value: `v\d+\.\d+\.\d+`},
		{Type: model.FilterTypeName, Value: "library/**"},
	}}
	response := convertReplicationPolicy(policy)
	require.Equal(t, "regex", *response.Filters[0].Kind)
	require.Equal(t, policy.Filters[0].Value, response.Filters[0].Value)
	require.Nil(t, response.Filters[1].Kind, "existing API responses must keep absent kinds absent")
}
