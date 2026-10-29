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
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	artifactctl "github.com/goharbor/harbor/src/controller/artifact"
	"github.com/goharbor/harbor/src/lib"
	"github.com/goharbor/harbor/src/lib/errors"
	artifacttesting "github.com/goharbor/harbor/src/testing/controller/artifact"
)

func TestPushPropagatesImmutabilityEvaluationErrors(t *testing.T) {
	original := artifactctl.Ctl
	t.Cleanup(func() { artifactctl.Ctl = original })
	for _, tc := range []struct {
		name    string
		failure error
		allowed bool
	}{
		{"first push", errors.NotFoundError(nil), true},
		{"matcher failed", errors.New("invalid persisted regex"), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctl := &artifacttesting.Controller{}
			ctl.On("GetByReference", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil, tc.failure)
			artifactctl.Ctl = ctl
			req := httptest.NewRequest("PUT", "/v2/library/hello/manifests/v1.2.3", nil)
			req = req.WithContext(lib.WithArtifactInfo(context.Background(), lib.ArtifactInfo{Repository: "library/hello", Tag: "v1.2.3"}))
			err := handlePush(req)
			if tc.allowed {
				require.NoError(t, err)
			} else {
				require.ErrorIs(t, err, tc.failure)
			}
		})
	}
}
