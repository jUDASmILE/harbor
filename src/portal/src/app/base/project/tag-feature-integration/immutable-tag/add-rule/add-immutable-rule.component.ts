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
import {
    ImmutableRetentionRule,
    RuleMetadate,
} from '../../tag-retention/retention';
import { ImmutableTagService } from '../immutable-tag.service';
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
@Component({
    selector: 'app-add-immutable-rule',
    templateUrl: './add-immutable-rule.component.html',
    styleUrls: ['./add-immutable-rule.component.scss'],
    standalone: false,
})
export class AddImmutableRuleComponent {
    addRuleOpened: boolean = false;
    @Output() clickAdd = new EventEmitter<ImmutableRetentionRule>();
    @Input() rules: ImmutableRetentionRule[];
    @Input() projectId: number;
    metadata: RuleMetadate = new RuleMetadate();
    rule: ImmutableRetentionRule;
    isAdd: boolean = true;
    editRuleOrigin: ImmutableRetentionRule;
    onGoing: boolean = false;
    @ViewChild(InlineAlertComponent) inlineAlert: InlineAlertComponent;
    constructor(private immutableTagService: ImmutableTagService) {
        this.rule = new ImmutableRetentionRule(this.projectId);
    }

    get repoSelect() {
        if (
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository[0]
        ) {
            return this.rule.scope_selectors.repository[0].decoration;
        }
        return '';
    }

    set repoSelect(repoSelect) {
        if (
            this.rule &&
            this.rule.scope_selectors &&
            this.rule.scope_selectors.repository[0]
        ) {
            this.rule.scope_selectors.repository[0].decoration = repoSelect;
        }
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
        if (
            this.rule &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0]
        ) {
            return this.rule.tag_selectors[0].decoration;
        }
        return '';
    }

    set tagsSelect(tagsSelect) {
        if (
            this.rule &&
            this.rule.tag_selectors &&
            this.rule.tag_selectors[0]
        ) {
            this.rule.tag_selectors[0].decoration = tagsSelect;
        }
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
        return false;
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
        if (this.rules && this.rules.length > 0) {
            for (let i = 0; i < this.rules.length; i++) {
                if (this.isSameRule(this.rules[i])) {
                    return true;
                }
            }
        }
        return false;
    }
    isSameRule(rule: ImmutableRetentionRule): boolean {
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

        if (
            this.rule.tag_selectors[0].decoration !==
            rule.tag_selectors[0].decoration
        ) {
            return false;
        }
        return (
            this.rule.tag_selectors[0].pattern === rule.tag_selectors[0].pattern
        );
    }

    getI18nKey(str: string) {
        return this.immutableTagService.getI18nKey(str);
    }
}
