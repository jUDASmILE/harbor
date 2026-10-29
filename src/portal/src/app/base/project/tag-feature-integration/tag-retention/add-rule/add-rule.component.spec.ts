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
import { ComponentFixture, TestBed } from '@angular/core/testing';
import { TranslateModule, TranslateService } from '@ngx-translate/core';
import { AddRuleComponent } from './add-rule.component';
import { CUSTOM_ELEMENTS_SCHEMA } from '@angular/core';
import {
    BrowserAnimationsModule,
    NoopAnimationsModule,
} from '@angular/platform-browser/animations';
import { ClarityModule } from '@clr/angular';
import { FormsModule } from '@angular/forms';
import { RouterTestingModule } from '@angular/router/testing';
import { provideHttpClientTesting } from '@angular/common/http/testing';
import { TagRetentionService } from '../tag-retention.service';
import { ErrorHandler } from '../../../../../shared/units/error-handler';
import { InlineAlertComponent } from '../../../../../shared/components/inline-alert/inline-alert.component';
import {
    provideHttpClient,
    withInterceptorsFromDi,
} from '@angular/common/http';
describe('AddRuleComponent', () => {
    let component: AddRuleComponent;
    let fixture: ComponentFixture<AddRuleComponent>;
    const mockTagRetentionService = {};

    beforeEach(() => {
        TestBed.configureTestingModule({
            schemas: [CUSTOM_ELEMENTS_SCHEMA],
            declarations: [AddRuleComponent, InlineAlertComponent],
            imports: [
                BrowserAnimationsModule,
                ClarityModule,
                TranslateModule.forRoot(),
                FormsModule,
                RouterTestingModule,
                NoopAnimationsModule,
            ],
            providers: [
                TranslateService,
                ErrorHandler,
                {
                    provide: TagRetentionService,
                    useValue: mockTagRetentionService,
                },
                provideHttpClient(withInterceptorsFromDi()),
                provideHttpClientTesting(),
            ],
        }).compileComponents();
    });

    beforeEach(() => {
        fixture = TestBed.createComponent(AddRuleComponent);
        component = fixture.componentInstance;
        fixture.detectChanges();
    });

    it('should create', () => {
        expect(component).toBeTruthy();
    });
    it('preserves regex patterns when editing and saving', () => {
        component.repoKind = 'regex';
        component.tagsKind = 'regex';
        component.repositories = 'library/(nginx|redis)';
        component.tagsInput = '\\Q release \\E[0-9]{2,3}';
        expect(component.tagsInput).toBe('\\Q release \\E[0-9]{2,3}');
        const emitted = spyOn(component.clickAdd, 'emit');
        component.add();
        expect(emitted).toHaveBeenCalledWith(component.rule);
        expect(component.rule.tag_selectors[0].pattern).toBe(
            '\\Q release \\E[0-9]{2,3}'
        );
    });

    it('treats otherwise identical patterns with different engines as distinct rules', () => {
        component.repositories = 'nginx';
        component.tagsInput = 'latest';
        const other = JSON.parse(JSON.stringify(component.rule));
        other.tag_selectors[0].kind = 'regex';
        expect(component.isSameRule(other)).toBeFalse();
    });
    it('preserves the stored pattern when switching engines', () => {
        component.tagsInput = 'a,b';
        expect(component.rule.tag_selectors[0].pattern).toBe('{a,b}');
        component.tagsKind = 'regex';
        expect(component.tagsInput).toBe('{a,b}');
        component.tagsInput = 'a{2,3}';
        expect(component.rule.tag_selectors[0].pattern).toBe('a{2,3}');
    });
});
