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
import { displayPattern, encodePattern, validatePattern } from './pattern.util';

describe('pattern engine input handling', () => {
    it('preserves regex commas, quantifiers, braces and whitespace', () => {
        for (const value of [
            'v[0-9]{2,3}',
            '(release,final|latest)',
            '{literal}',
            '\\Q literal value \\E',
        ]) {
            expect(encodePattern(value, 'regex')).toBe(value);
            expect(displayPattern(value, 'regex')).toBe(value);
        }
    });

    it('preserves the legacy glob comma-list format', () => {
        expect(encodePattern('a,b')).toBe('{a,b}');
        expect(encodePattern('a{1,2}', 'doublestar')).toBe('a{1,2}');
        expect(displayPattern('{a,b}')).toBe('a,b');
    });

    it('accepts RE2 features outside JavaScript syntax', () => {
        for (const value of [
            '(?i)latest',
            '(?i:latest)',
            '(?P<x>a)(?P<x>b)',
            '(?<x>a)(?<x>b)',
            '(?P<1>a)',
            '[(?=]',
            '[(?=foo)]',
            '[😀-🙏]',
            '[\\x{61}-\\x{7a}]',
            '[](?=]',
            '[[:alnum:]]+',
            '\\p{L}+',
            '\\123',
            '\\Afoo\\z',
            '\\Qtest[',
        ]) {
            expect(validatePattern(value, 'regex')).toBeNull();
        }
    });

    it('identifies common syntax errors and unsupported PCRE operators', () => {
        for (const value of [
            '[',
            'foo(',
            '(?=foo)',
            '(?<=foo)',
            '(?P=name)',
            '(a)\\1',
        ]) {
            expect(validatePattern(value, 'regex')).toBe(
                'PATTERN.REGEX_INVALID'
            );
        }
    });

    it('counts the regex limit in Unicode characters', () => {
        expect(validatePattern('猫'.repeat(512), 'regex')).toBeNull();
        expect(validatePattern('猫'.repeat(513), 'regex')).toBe(
            'PATTERN.TOO_LONG'
        );
        expect(validatePattern('x'.repeat(513), 'doublestar')).toBeNull();
    });
});
