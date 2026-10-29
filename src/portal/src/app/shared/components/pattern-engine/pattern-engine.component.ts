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
    EventEmitter,
    Input,
    Output,
    OnChanges,
    SimpleChanges,
} from '@angular/core';
import { FormsModule } from '@angular/forms';
import { TranslateModule } from '@ngx-translate/core';
import { PATTERN_KINDS, validatePattern } from '../../units/pattern.util';

@Component({
    selector: 'hbr-pattern-engine',
    imports: [FormsModule, TranslateModule],
    templateUrl: './pattern-engine.component.html',
    styleUrls: ['./pattern-engine.component.scss'],
})
export class PatternEngineComponent implements OnChanges {
    @Input() kind: string;
    @Input() pattern: string;
    @Input() serverError: string;
    fieldError = '';
    @Input() engines: string[] = PATTERN_KINDS;
    @Output() kindChange = new EventEmitter<string>();

    ngOnChanges(changes: SimpleChanges) {
        if (changes.serverError) {
            this.fieldError = /regexp|regex|pattern/i.test(
                this.serverError || ''
            )
                ? this.serverError
                : '';
        } else if (changes.pattern || changes.kind) {
            this.fieldError = '';
        }
    }

    get error(): string | null {
        return validatePattern(this.pattern, this.kind);
    }
}
