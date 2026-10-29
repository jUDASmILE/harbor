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
export const PATTERN_KINDS = ['doublestar', 'regex'];

// Comma lists and brace removal belong only to the legacy glob input format.
export function encodePattern(pattern: string, kind?: string): string {
    if (kind === 'regex' || !pattern) {
        return pattern;
    }
    return pattern.includes(',') && !/[{}]/.test(pattern)
        ? '{' + pattern + '}'
        : pattern;
}

export function displayPattern(pattern: string, kind?: string): string {
    return kind !== 'regex' && /^{\S+}$/.test(pattern)
        ? pattern.slice(1, -1)
        : pattern;
}

export function validatePattern(pattern: string, kind?: string): string | null {
    if (kind !== 'regex' || !pattern) {
        return null;
    }
    if (Array.from(pattern).length > 512) {
        return 'PATTERN.TOO_LONG';
    }
    // Go regexp is authoritative at save time. Normalize its supported syntax
    // before using the browser parser for immediate feedback.
    const quoted = pattern.replace(/\\Q([\s\S]*?)(?:\\E|$)/g, (_, text) =>
        text.replace(/[.*+?^$()|[\]\\{}]/g, '\\$&')
    );
    const operators = quoted.replace(/\[\^?\]?(?:\\.|[^\]\\])*\]/g, '');
    if (
        /(^|[^\\])(?:\\\\)*(?:\(\?(?:[=!]|<[=!])|\\[1-9](?![0-7]{2})|\(\?P=)/.test(
            operators
        )
    ) {
        return 'PATTERN.REGEX_INVALID';
    }
    // The browser cannot validate POSIX and Go Unicode character classes.
    if (/\[\[:|\[\^?\]|\\[pP]|\\x\{|[^\x00-\x7f]/.test(quoted)) {
        return null;
    }
    const browserPattern = quoted
        .replace(/\(\?(?:P?<)[A-Za-z_0-9]+>/g, '(?:')
        .replace(/\\[Az]/g, '')
        .replace(/\(\?[imsU-]+:/g, '(?:')
        .replace(/\(\?[imsU-]+\)/g, '');
    try {
        new RegExp(browserPattern);
    } catch {
        return 'PATTERN.REGEX_INVALID';
    }
    return null;
}
