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

package util

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/goharbor/harbor/src/lib/pattern"
	"github.com/goharbor/harbor/src/pkg/reg/model"
)

// Match returns whether str matches the pattern. An omitted kind preserves glob matching.
func Match(expression, str string, kinds ...string) (bool, error) {
	kind := ""
	if len(kinds) > 0 {
		kind = kinds[0]
	}
	matcher, err := pattern.Compile(expression, kind)
	if err != nil {
		return false, err
	}
	return matcher.MatchString(str), nil
}

// Compile prepares a reusable matcher before an adapter enumerates resources.
func Compile(expression, kind string) (*pattern.Matcher, error) {
	return pattern.Compile(expression, kind)
}

// FilterKind returns the engine selected for the specified filter.
func FilterKind(filters []*model.Filter, filterType string) string {
	for _, filter := range filters {
		if filter.Type == filterType {
			return filter.Kind
		}
	}
	return ""
}

// ListingPattern returns a glob-safe namespace prefix for repository discovery.
// Regex syntax must never be interpreted as a glob or an exact remote name.
func ListingPattern(expression, kind string) string {
	if kind != pattern.KindRegex {
		return expression
	}
	re, err := regexp.Compile(expression)
	if err != nil {
		return ""
	}
	prefix, _ := re.LiteralPrefix()
	if index := strings.Index(prefix, "/"); index >= 0 && !strings.ContainsAny(prefix[:index], "*?[]{}\\,") {
		return prefix[:index] + "/**"
	}
	return ""
}

// SpecificPaths applies the exact-repository optimization only when safe for the engine.
func SpecificPaths(expression, kind string) ([]string, bool) {
	if kind != pattern.KindRegex {
		return IsSpecificPath(expression)
	}
	re, err := regexp.Compile(expression)
	if err != nil {
		return nil, false
	}
	prefix, complete := re.LiteralPrefix()
	if complete && prefix != "" {
		return []string{prefix}, true
	}
	return nil, false
}

// IsSpecificPath checks whether the input path is a specified string
// If it is, the function returns a string array that parsed from the input path
// A specified string means we can get a specific string array after parsing it
// "library/hello-world" is a specified string as it only matches "library/hello-world"
// "library/**" isn't a specified string as it can match all string that starts with "library/"
// "library/{test,busybox}" is a specified string as it only matches "library/hello-world" and "library/busybox"
func IsSpecificPath(path string) ([]string, bool) {
	if len(path) == 0 {
		return nil, false
	}
	components := [][]string{}
	for component := range strings.SplitSeq(path, "/") {
		strs, ok := IsSpecificPathComponent(component)
		if !ok {
			return nil, false
		}
		components = append(components, strs)
	}

	result := []string{}
	for _, component := range components {
		result = combinationPathComponents(result, component)
	}
	return result, true
}

func combinationPathComponents(components1, components2 []string) []string {
	if len(components1) == 0 {
		return components2
	}
	if len(components2) == 0 {
		return components1
	}
	components := []string{}
	for _, component1 := range components1 {
		for _, component2 := range components2 {
			components = append(components, fmt.Sprintf("%s/%s", component1, component2))
		}
	}
	return components
}

// IsSpecificPathComponent checks whether the input path component is a specified string
// If it is, the function returns a string array that parsed from the input component
// A specified string means we can get a specific string array after parsing it
// "library" is a specified string as it only matches "library"
// "library*" isn't a specified string as it can match all string that starts with "library"
// "{library, test}" is a specified string as it only matches "library" and "test"
// Note: the function doesn't support the component that contains more than one "{"
// such as "a{b{c,d}e}f"
func IsSpecificPathComponent(component string) ([]string, bool) {
	if len(component) == 0 {
		return nil, false
	}
	// contains any of *?[\\]^
	if strings.ContainsAny(component, "*?[\\]^/") {
		return nil, false
	}
	// doesn't contain {},
	if !strings.ContainsAny(component, "{},") {
		return []string{component}, true
	}
	// support only one pair of {} currently
	n := strings.Count(component, "{")
	if n > 1 {
		return nil, false
	}
	i := strings.Index(component, "{")
	if i == -1 {
		return nil, false
	}
	j := strings.LastIndex(component, "}")
	if j == -1 {
		return nil, false
	}
	if i > j {
		return nil, false
	}
	prefix := component[:i]
	suffix := ""
	if j+1 < len(component) {
		suffix = component[j+1:]
	}
	components := []string{}
	for str := range strings.SplitSeq(component[i+1:j], ",") {
		components = append(components, prefix+str+suffix)
	}
	return components, true
}
