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

package immutable

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/goharbor/harbor/src/lib/errors"
	immutablepkg "github.com/goharbor/harbor/src/pkg/immutable"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

type validationManager struct {
	immutablepkg.Manager
	stored *model.Metadata
	writes int
}

func (m *validationManager) GetImmutableRule(context.Context, int64) (*model.Metadata, error) {
	return m.stored, nil
}
func (m *validationManager) CreateImmutableRule(context.Context, *model.Metadata) (int64, error) {
	m.writes++
	return 1, nil
}
func (m *validationManager) UpdateImmutableRule(context.Context, int64, *model.Metadata) error {
	m.writes++
	return nil
}
func (m *validationManager) EnableImmutableRule(context.Context, int64, bool) error {
	m.writes++
	return nil
}

func TestPatternValidationBeforeWrite(t *testing.T) {
	invalid := &model.Metadata{ID: 1, TagSelectors: []*model.Selector{{Kind: "regex", Decoration: "matches", Pattern: "["}}}
	manager := &validationManager{stored: &model.Metadata{ID: 1}}
	ctl := NewAPIController(manager)
	_, err := ctl.CreateImmutableRule(context.Background(), invalid)
	require.True(t, errors.IsErr(err, errors.BadRequestCode))
	err = ctl.UpdateImmutableRule(context.Background(), 1, invalid)
	require.True(t, errors.IsErr(err, errors.BadRequestCode))
	require.Zero(t, manager.writes)
	invalid.TagSelectors[0].Pattern = `v?\d+\.\d+\.\d+`
	_, err = ctl.CreateImmutableRule(context.Background(), invalid)
	require.NoError(t, err)
	require.Equal(t, 1, manager.writes)
}

func TestEnablingValidatesPersistedSelectors(t *testing.T) {
	manager := &validationManager{stored: &model.Metadata{ID: 1, Disabled: true, TagSelectors: []*model.Selector{{Kind: "regex", Decoration: "matches", Pattern: "["}}}}
	ctl := NewAPIController(manager)
	err := ctl.UpdateImmutableRule(context.Background(), 1, &model.Metadata{ID: 1, Disabled: false})
	require.True(t, errors.IsErr(err, errors.BadRequestCode))
	require.Zero(t, manager.writes)
}
