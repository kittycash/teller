"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var common_1 = require("@angular/common");
var forms_1 = require("@angular/forms");
var core_2 = require("@ngx-translate/core");
var core_3 = require("@app/core");
var shared_1 = require("@app/shared");
var AccountConfirmationModule = /** @class */ (function () {
    function AccountConfirmationModule() {
    }
    AccountConfirmationModule = __decorate([
        core_1.NgModule({
            imports: [
                common_1.CommonModule,
                forms_1.ReactiveFormsModule,
                core_2.TranslateModule,
                core_3.CoreModule,
                shared_1.SharedModule
            ],
            declarations: [],
            providers: []
        })
    ], AccountConfirmationModule);
    return AccountConfirmationModule;
}());
exports.AccountConfirmationModule = AccountConfirmationModule;
//# sourceMappingURL=account_confirmation.module.js.map