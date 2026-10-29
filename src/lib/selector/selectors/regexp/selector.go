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
	"encoding/json"

	"github.com/goharbor/harbor/src/lib/pattern"
	iselector "github.com/goharbor/harbor/src/lib/selector"
)

const (
	// Kind ...
	Kind = "regex"
	// Matches [pattern] for tag (default)
	Matches = "matches"
	// Excludes [pattern] for tag (default)
	Excludes = "excludes"
	// RepoMatches represents repository matches [pattern]
	RepoMatches = "repoMatches"
	// RepoExcludes represents repository excludes [pattern]
	RepoExcludes = "repoExcludes"
	// NSMatches represents namespace matches [pattern]
	NSMatches = "nsMatches"
	// NSExcludes represents namespace excludes [pattern]
	NSExcludes = "nsExcludes"
)

// selector for regular expression
type selector struct {
	// Pre defined pattern declarator
	// "matches", "excludes", "repoMatches" or "repoExcludes"
	decoration string
	// The pattern expression
	matcher *pattern.Matcher
	// whether match untagged
	untagged bool
}

// Select candidates by regular expressions
func (s *selector) Select(artifacts []*iselector.Candidate) (selected []*iselector.Candidate, err error) {
	value := ""
	excludes := false

	for _, art := range artifacts {
		switch s.decoration {
		case Matches:
			s, err := s.tagSelectMatch(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case Excludes:
			s, err := s.tagSelectExclude(art)
			if err != nil {
				return nil, err
			}
			if s {
				selected = append(selected, art)
			}
		case RepoMatches:
			value = art.Repository
		case RepoExcludes:
			value = art.Repository
			excludes = true
		case NSMatches:
			value = art.Namespace
		case NSExcludes:
			value = art.Namespace
			excludes = true
		}

		if len(value) > 0 {
			matched := s.matcher.MatchString(value)

			if (matched && !excludes) || (!matched && excludes) {
				selected = append(selected, art)
			}
		}
	}

	return selected, nil
}

func (s *selector) tagSelectMatch(artifact *iselector.Candidate) (selected bool, err error) {
	if len(artifact.Tags) > 0 {
		for _, t := range artifact.Tags {
			matched := s.matcher.MatchString(t)
			if matched {
				return true, nil
			}
		}
		return false, nil
	}
	return s.untagged, nil
}

func (s *selector) tagSelectExclude(artifact *iselector.Candidate) (selected bool, err error) {
	if len(artifact.Tags) > 0 {
		for _, t := range artifact.Tags {
			matched := s.matcher.MatchString(t)
			if !matched {
				return true, nil
			}
		}
		return false, nil
	}
	return !s.untagged, nil
}

// New compiles the expression once and constructs a regex selector.
func New(decoration string, value any, extras string) (iselector.Selector, error) {
	untagged := true // default behavior for upgrade, active keep the untagged images
	if decoration == Excludes {
		untagged = false
	}
	if extras != "" {
		var extraObj struct {
			Untagged bool `json:"untagged"`
		}
		if err := json.Unmarshal([]byte(extras), &extraObj); err == nil {
			untagged = extraObj.Untagged
		}
	}

	var p string
	if value != nil {
		p, _ = value.(string)
	}
	matcher, err := pattern.Compile(p, Kind)
	if err != nil {
		return nil, err
	}

	return &selector{
		decoration: decoration,
		matcher:    matcher,
		untagged:   untagged,
	}, nil
}
