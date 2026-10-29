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
import {
    Component,
    Output,
    EventEmitter,
    ViewChild,
    Input,
} from '@angular/core';
import { Retention, Rule, RuleMetadate } from '../retention';
import { TagRetentionService } from '../tag-retention.service';
import { compareValue } from '../../../../../shared/units/utils';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import {
    encodePattern,
    displayPattern,
    validatePattern,
    PATTERN_KINDS,
} from '../../../../../shared/units/pattern.util';

const EXISTING_RULE = 'TAG_RETENTION.EXISTING_RULE';
const INVALID_RULE = 'TAG_RETENTION.INVALID_RULE';
const MAX = 2100000000;
@Component({
    selector: 'add-rule',
    templateUrl: './add-rule.component.html',
    styleUrls: ['./add-rule.component.scss'],
    standalone: false,
})
export class AddRuleComponent {
    addRuleOpened: boolean = false;
    @Output() clickAdd = new EventEmitter<Rule>();
    @Input() retention: Retention;
    metadata: RuleMetadate = new RuleMetadate();
    rule: Rule = new Rule();
    isAdd: boolean = true;
    editRuleOrigin: Rule;
    onGoing: boolean = false;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(private tagRetentionService: TagRetentionService) {}

    set template(template) {
        this.rule.template = template;
    }

    get template() {
        return this.rule.template;
    }

    get unit(): string {
        let str = '';
        this.metadata.templates.forEach(t => {
            if (t.rule_template === this.rule.template) {
                str = t.params[0].unit;
            }
        });
        return str;
    }

    get num() {
        return this.rule.params[this.template];
    }

    set num(num) {
        if (num) {
            num = num.trim();
        }
        if (parseInt(num, 10) > 0) {
            num = parseInt(num, 10);
        }
        this.rule.params[this.template] = num;
    }

    get repoSelect() {
        return this.rule.scope_selectors.repository[0].decoration;
    }

    set repoSelect(repoSelect) {
        this.rule.scope_selectors.repository[0].decoration = repoSelect;
    }

    get repoKind(): string {
        return this.rule.scope_selectors.repository[0].kind || 'doublestar';
    }

    set repoKind(kind: string) {
        this.rule.scope_selectors.repository[0].kind = kind;
    }

    get tagsKind(): string {
        return this.rule.tag_selectors[0].kind || 'doublestar';
    }

    set tagsKind(kind: string) {
        this.rule.tag_selectors[0].kind = kind;
    }

    get scopeEngines(): string[] {
        const engines = this.metadata.scope_selectors
            .map(s => s.kind)
            .filter(Boolean);
        return engines.length ? engines : PATTERN_KINDS;
    }

    get tagEngines(): string[] {
        const engines = this.metadata.tag_selectors
            .map(s => s.kind)
            .filter(Boolean);
        return engines.length ? engines : PATTERN_KINDS;
    }

    decorations(scope: boolean): string[] {
        const metadata = scope
            ? this.metadata.scope_selectors
            : this.metadata.tag_selectors;
        const kind = scope ? this.repoKind : this.tagsKind;
        return (
            (metadata.find(s => s.kind === kind) || metadata[0])?.decorations ||
            []
        );
    }

    set repositories(value: string) {
        this.rule.scope_selectors.repository[0].pattern = encodePattern(
            value,
            this.repoKind
        );
    }

    get repositories(): string {
        return displayPattern(
            this.rule.scope_selectors.repository[0].pattern,
            this.repoKind
        );
    }

    get tagsSelect() {
        return this.rule.tag_selectors[0].decoration;
    }

    set tagsSelect(tagsSelect) {
        this.rule.tag_selectors[0].decoration = tagsSelect;
    }

    set tagsInput(value: string) {
        this.rule.tag_selectors[0].pattern = encodePattern(
            value,
            this.tagsKind
        );
    }

    get tagsInput(): string {
        return displayPattern(
            this.rule.tag_selectors[0].pattern,
            this.tagsKind
        );
    }

    set untagged(untagged) {
        let extras = JSON.parse(this.rule.tag_selectors[0].extras);
        extras.untagged = untagged;
        this.rule.tag_selectors[0].extras = JSON.stringify(extras);
    }

    get untagged() {
        if (this.rule.tag_selectors[0] && this.rule.tag_selectors[0].extras) {
            let extras = JSON.parse(this.rule.tag_selectors[0].extras);
            if (extras.untagged !== undefined) {
                return extras.untagged;
            }
            return false;
        } else {
            return false;
        }
    }

    get labelsSelect() {
        return this.rule.tag_selectors[1].decoration;
    }

    set labelsSelect(labelsSelect) {
        this.rule.tag_selectors[1].decoration = labelsSelect;
    }

    set labelsInput(labelsInput) {
        this.rule.tag_selectors[1].pattern = labelsInput;
    }

    get labelsInput() {
        return this.rule.tag_selectors[1].pattern;
    }

    canNotAdd(): boolean {
        if (
            this.onGoing ||
            (!this.isAdd && compareValue(this.editRuleOrigin, this.rule))
        ) {
            return true;
        }
        if (
            validatePattern(this.repositories, this.repoKind) ||
            validatePattern(this.tagsInput, this.tagsKind)
        ) {
            return true;
        }
        if (
            (this.repoKind !== 'regex' && !this.repositories) ||
            (this.tagsKind !== 'regex' && !this.tagsInput)
        ) {
            return true;
        }
        return (
            !this.rule.template ||
            (this.hasParam() &&
                (this.rule.params[this.template] === undefined ||
                    this.rule.params[this.template] === '' ||
                    !(
                        Number(this.rule.params[this.template]) >= 0 &&
                        Number(this.rule.params[this.template]) < MAX
                    )))
        );
    }

    open() {
        this.addRuleOpened = true;
        this.inlineAlert.alertClose = true;
        this.onGoing = false;
    }

    close() {
        this.addRuleOpened = false;
    }

    cancel() {
        this.close();
    }

    add() {
        // convert string "0" to number 0
        if (this.rule.params[this.template] === '0') {
            this.rule.params[this.template] = 0;
        }
        // Preserve whitespace in regex literals and quoted expressions.
        for (const selector of [
            this.rule.scope_selectors.repository[0],
            this.rule.tag_selectors[0],
        ]) {
            if (selector.kind !== 'regex') {
                selector.pattern = selector.pattern.replace(/\s+/g, '');
            }
        }
        if (
            this.repoKind !== 'regex' &&
            this.rule.scope_selectors.repository[0].decoration !==
                'repoMatches' &&
            this.rule.scope_selectors.repository[0].pattern
        ) {
            let str = this.rule.scope_selectors.repository[0].pattern;
            str = str.replace(/[{}]/g, '');
            const arr = str.split(',');
            for (let i = 0; i < arr.length; i++) {
                if (arr[i] && arr[i].trim() && arr[i] === '**') {
                    this.inlineAlert.showInlineError(INVALID_RULE);
                    return;
                }
            }
        }
        if (this.isExistingRule()) {
            this.inlineAlert.showInlineError(EXISTING_RULE);
            return;
        }
        this.clickAdd.emit(this.rule);
    }
    isExistingRule(): boolean {
        if (
            this.retention &&
            this.retention.rules &&
            this.retention.rules.length > 0
        ) {
            for (let i = 0; i < this.retention.rules.length; i++) {
                if (this.isSameRule(this.retention.rules[i])) {
                    return true;
                }
            }
        }
        return false;
    }
    isSameRule(rule: Rule): boolean {
        if (
            this.repoKind !==
                (rule.scope_selectors.repository[0].kind || 'doublestar') ||
            this.tagsKind !== (rule.tag_selectors[0].kind || 'doublestar')
        ) {
            return false;
        }
        if (
            this.rule.scope_selectors.repository[0].decoration !==
            rule.scope_selectors.repository[0].decoration
        ) {
            return false;
        }
        if (
            this.rule.scope_selectors.repository[0].pattern !==
            rule.scope_selectors.repository[0].pattern
        ) {
            return false;
        }
        if (this.rule.template !== rule.template) {
            return false;
        }
        if (
            this.hasParam() &&
            JSON.stringify(this.rule.params) !== JSON.stringify(rule.params)
        ) {
            return false;
        }
        if (
            this.rule.tag_selectors[0].decoration !==
            rule.tag_selectors[0].decoration
        ) {
            return false;
        }
        if (
            this.rule.tag_selectors[0].extras !== rule.tag_selectors[0].extras
        ) {
            return false;
        }
        return (
            this.rule.tag_selectors[0].pattern === rule.tag_selectors[0].pattern
        );
    }

    getI18nKey(str: string) {
        return this.tagRetentionService.getI18nKey(str);
    }
    hasParam(): boolean {
        if (this.metadata && this.metadata.templates) {
            let flag: boolean = false;
            this.metadata.templates.forEach(t => {
                if (t.rule_template === this.template) {
                    if (t.params && t.params.length > 0) {
                        flag = true;
                    }
                }
            });
            return flag;
        }
        return false;
    }
}
