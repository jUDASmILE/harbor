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
	"regexp"
	"regexp/syntax"
	"strings"
	"sync"
	"unicode/utf8"

	"github.com/bmatcuk/doublestar"

	"github.com/goharbor/harbor/src/lib/errors"
)

const (
	// KindRegex indicates regular expression pattern matching
	KindRegex = "regex"
	// KindDoublestar indicates doublestar (glob) pattern matching
	KindDoublestar = "doublestar"
)

// MaxRegexLength bounds the cost of compiling a user-supplied regular expression.
const MaxRegexLength = 512

// Matcher is a reusable, immutable pattern compiled when a policy is loaded.
// A Matcher may be used concurrently by multiple goroutines.
type Matcher struct {
	glob  string
	regex *regexp.Regexp
	all   bool
}

// Keep repeated policy loads from recompiling identical expressions. Bounding the
// cache also bounds memory across edits and across projects.
var regexCache = struct {
	sync.Mutex
	matchers map[string]*Matcher
}{matchers: make(map[string]*Matcher)}

// Compile prepares a full-string matcher. An absent kind means doublestar.
// Pattern text is preserved verbatim; an empty pattern matches every value.
func Compile(filterPattern, kind string) (*Matcher, error) {
	if err := ValidateKind(kind); err != nil {
		return nil, err
	}
	m := &Matcher{glob: filterPattern, all: filterPattern == ""}
	if kind == KindRegex {
		if utf8.RuneCountInString(filterPattern) > MaxRegexLength {
			return nil, errors.Errorf("regular expression exceeds the maximum length of %d characters", MaxRegexLength)
		}
		// Parse the original expression before anchoring. Building the wrapper
		// as syntax also preserves valid unclosed \Q quotes and inline flags,
		// which can swallow or weaken anchors in a concatenated string wrapper.
		regexCache.Lock()
		defer regexCache.Unlock()
		if cached, ok := regexCache.matchers[filterPattern]; ok {
			return cached, nil
		}
		expr, err := syntax.Parse(filterPattern, syntax.Perl)
		if err != nil {
			return nil, err
		}
		anchored := &syntax.Regexp{Op: syntax.OpConcat, Sub: []*syntax.Regexp{
			{Op: syntax.OpBeginText}, expr, {Op: syntax.OpEndText},
		}}
		m.regex, err = regexp.Compile(anchored.String())
		if err != nil {
			return nil, err
		}
		if len(regexCache.matchers) >= 1024 {
			regexCache.matchers = make(map[string]*Matcher)
		}
		regexCache.matchers[filterPattern] = m
	} else if !validateDoublestarPattern(filterPattern) {
		return nil, doublestar.ErrBadPattern
	}
	return m, nil
}

// MatchString reports whether the entire value matches the compiled pattern.
func (m *Matcher) MatchString(value string) bool {
	if m.all {
		return true
	}
	if m.regex != nil {
		return m.regex.MatchString(value)
	}
	matched, _ := doublestar.Match(m.glob, value) // syntax checked at construction
	return matched
}

// Validate checks the kind and pattern without changing the pattern text.
func Validate(filterPattern, kind string) error {
	_, err := Compile(filterPattern, kind)
	return err
}

// ValidateRepositoryFilter preserves the proxy-cache filter's whitespace handling.
func ValidateRepositoryFilter(filterPattern, kind string) error {
	return Validate(strings.TrimSpace(filterPattern), kind)
}

// ValidateKind validates the pattern engine. Empty means the doublestar default.
func ValidateKind(kind string) error {
	switch kind {
	case "", KindRegex, KindDoublestar:
		return nil
	default:
		return errors.Errorf("unsupported pattern kind %q, must be %q, %q, or empty (defaults to %q)", kind, KindDoublestar, KindRegex, KindDoublestar)
	}
}

// Match preserves the proxy-cache filter's whitespace handling. Policy filters
// should use Compile once at construction and reuse the resulting matcher.
func Match(value, filterPattern, kind string) (bool, error) {
	m, err := Compile(strings.TrimSpace(filterPattern), kind)
	if err != nil {
		return false, err
	}
	return m.MatchString(value), nil
}

// validateDoublestarPattern follows doublestar v1's parser independently of a
// candidate. Its alternative splitter deliberately counts braces and commas
// inside character classes, and its class terminator treats a bracket preceded
// by any backslash as escaped. Preserve those quirks until the engine is upgraded.
func validateDoublestarPattern(s string) bool {
	return validateGlob(s, make(map[string]bool))
}

func validateGlob(s string, memo map[string]bool) (valid bool) {
	if result, ok := memo[s]; ok {
		return result
	}
	defer func() { memo[s] = valid }()
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		i += size
		switch r {
		case '\\':
			if i >= len(s) {
				return false
			}
			escaped, width := utf8.DecodeRuneInString(s[i:])
			if escaped == utf8.RuneError {
				return false
			}
			i += width
		case '[':
			end := globClassEnd(s, i)
			if end < 0 || !validateGlobClass([]rune(s[i:end])) {
				return false
			}
			i = end + 1
		case '{':
			options, end := globAlternatives(s, i)
			if end < 0 {
				return false
			}
			// v1 concatenates each alternative with the remaining pattern. This
			// also permits a class opened in an alternative to close in the suffix.
			for _, option := range options {
				if !validateGlob(option+s[end:], memo) {
					return false
				}
			}
			return true
		}
	}
	return true
}

func globClassEnd(s string, start int) int {
	for i := start; i < len(s); i++ {
		if s[i] == ']' && (i == start || s[i-1] != '\\') {
			return i
		}
	}
	return -1
}

func validateGlobClass(class []rune) bool {
	if len(class) == 0 {
		return false
	}
	i := 0
	if class[0] == '^' {
		i++ // v1 accepts [^] as a negated empty class.
	}
	for i < len(class) {
		if class[i] == '-' {
			return false
		}
		low := class[i]
		i++
		if low == '\\' {
			if i >= len(class) {
				return false
			}
			i++
		}
		if i < len(class) && class[i] == '-' {
			i++
			if i >= len(class) || class[i] == '-' {
				return false
			}
			high := class[i]
			i++
			if high == '\\' {
				if i >= len(class) {
					return false
				}
				i++
			}
		}
	}
	return true
}

func globAlternatives(s string, start int) ([]string, int) {
	var options []string
	depth, escaped, optionStart := 1, false, start
	for i := start; i < len(s); {
		r, width := utf8.DecodeRuneInString(s[i:])
		if escaped {
			escaped = false
		} else {
			switch r {
			case '\\':
				escaped = true
			case '{':
				depth++
			case '}':
				depth--
				if depth == 0 {
					return append(options, s[optionStart:i]), i + width
				}
			case ',':
				if depth == 1 {
					options = append(options, s[optionStart:i])
					optionStart = i + width
				}
			}
		}
		i += width
	}
	return nil, -1
}
