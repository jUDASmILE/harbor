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
	"context"
	"sync"

	"github.com/goharbor/harbor/src/controller/immutable"
	"github.com/goharbor/harbor/src/lib/q"
	iselector "github.com/goharbor/harbor/src/lib/selector"
	"github.com/goharbor/harbor/src/lib/selector/selectors/doublestar"
	"github.com/goharbor/harbor/src/lib/selector/selectors/index"
	"github.com/goharbor/harbor/src/pkg/immutable/match"
	"github.com/goharbor/harbor/src/pkg/immutable/model"
)

// Matcher ...
type Matcher struct {
	mutex     sync.Mutex
	selectors map[selectorKey]iselector.Selector
}

// Match ...
func (rm *Matcher) Match(ctx context.Context, pid int64, c iselector.Candidate) (bool, error) {
	rules, err := immutable.Ctr.ListImmutableRules(ctx, q.New(q.KeyWords{"ProjectID": pid}))
	if err != nil {
		return false, err
	}

	for _, r := range rules {
		if r.Disabled {
			continue
		}

		// Every selector of a dimension is evaluated; checking only the first
		// one let candidates covered by the others escape the rule.
		repositorySelectors := nonNil(r.ScopeSelectors["repository"])
		if len(repositorySelectors) < 1 {
			continue
		}
		matched, err := dimensionSelects(repositorySelectors, "", &c, rm.getSelector)
		if err != nil {
			return false, err
		}
		if !matched {
			continue
		}

		// match tag according to the tag selectors.
		// for immutable policy, should not keep untagged artifacts by default.
		tagSelectors := nonNil(r.TagSelectors)
		if len(tagSelectors) < 1 {
			continue
		}
		matched, err = dimensionSelects(tagSelectors, "{\"untagged\": false}", &c, rm.getSelector)
		if err != nil {
			return false, err
		}
		if !matched {
			continue
		}

		return true, nil
	}
	return false, nil
}

// dimensionSelects reports whether a rule dimension (repository or tag) puts
// the candidate in scope. The selectors must behave like the portal's single
// `{a,b}` pattern: inclusion selectors are alternatives (any may match) and an
// exclusion selector removes whatever it matches, so every exclusion selector
// has to select the candidate. A dimension holding only exclusions starts from
// everything.
//
// A multi-tag candidate is evaluated per tag because a doublestar selector
// selects an artifact when any single tag passes; combining selector results on
// the whole artifact would let tags excluded by different selectors cancel out.
func dimensionSelects(selectors []*model.Selector, extras string, c *iselector.Candidate, get selectorGetter) (bool, error) {
	if len(c.Tags) <= 1 {
		return selectsOne(selectors, extras, c, get)
	}
	for _, tag := range c.Tags {
		single := *c
		single.Tags = []string{tag}
		ok, err := selectsOne(selectors, extras, &single, get)
		if err != nil || ok {
			return ok, err
		}
	}
	return false, nil
}

func selectsOne(selectors []*model.Selector, extras string, c *iselector.Candidate, get selectorGetter) (bool, error) {
	hasInclusion, included := false, false
	for _, sel := range selectors {
		s, err := get(sel.Kind, sel.Decoration, sel.Pattern, extras)
		if err != nil {
			return false, err
		}
		selected, err := s.Select([]*iselector.Candidate{c})
		if err != nil {
			return false, err
		}
		if isExclusion(sel.Decoration) {
			if len(selected) == 0 {
				return false, nil
			}
			continue
		}
		hasInclusion = true
		included = included || len(selected) > 0
	}
	return !hasInclusion || included, nil
}

// nonNil tolerates null selector entries persisted by older Harbor versions.
func nonNil(selectors []*model.Selector) []*model.Selector {
	out := make([]*model.Selector, 0, len(selectors))
	for _, s := range selectors {
		if s != nil {
			out = append(out, s)
		}
	}
	return out
}

func isExclusion(decoration string) bool {
	switch decoration {
	case doublestar.Excludes, doublestar.RepoExcludes, doublestar.NSExcludes:
		return true
	}
	return false
}

type selectorGetter func(kind, decoration, pattern, extras string) (iselector.Selector, error)

type selectorKey struct {
	kind, decoration, pattern, extras string
}

// Cache immutable selector instances across artifact evaluations. The current
// policy is read for each request, so policy edits take effect immediately.
func (rm *Matcher) getSelector(kind, decoration, pattern, extras string) (iselector.Selector, error) {
	key := selectorKey{kind, decoration, pattern, extras}
	rm.mutex.Lock()
	defer rm.mutex.Unlock()
	if selected, ok := rm.selectors[key]; ok {
		return selected, nil
	}
	selected, err := index.Get(kind, decoration, pattern, extras)
	if err != nil {
		return nil, err
	}
	// Bound memory when policies change repeatedly or many projects are used.
	if rm.selectors == nil || len(rm.selectors) >= 1024 {
		rm.selectors = make(map[selectorKey]iselector.Selector)
	}
	rm.selectors[key] = selected
	return selected, nil
}

// NewRuleMatcher ...
func NewRuleMatcher() match.ImmutableTagMatcher {
	return &Matcher{}
}
