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
var loader_component_1 = require("./loader/loader.component");
var catbox_component_1 = require("./catbox/catbox.component");
var account_confirmation_component_1 = require("./account_confirmation/account_confirmation.component");
var confirm_input_directive_1 = require("./account_confirmation/confirm_input.directive");
var focus_service_1 = require("./account_confirmation/focus.service");
var ng4_intl_phone_1 = require("ng4-intl-phone");
var SharedModule = /** @class */ (function () {
    function SharedModule() {
    }
    SharedModule = __decorate([
        core_1.NgModule({
            imports: [
                common_1.CommonModule,
                forms_1.FormsModule,
                core_2.TranslateModule,
                forms_1.ReactiveFormsModule,
                ng4_intl_phone_1.InternationalPhoneModule
            ],
            providers: [
                focus_service_1.FocusService
            ],
            declarations: [
                loader_component_1.LoaderComponent,
                catbox_component_1.CatBoxComponent,
                account_confirmation_component_1.AccountConfirmationComponent,
                confirm_input_directive_1.ConfirmInputDirective
            ],
            exports: [
                loader_component_1.LoaderComponent,
                catbox_component_1.CatBoxComponent,
                account_confirmation_component_1.AccountConfirmationComponent,
                confirm_input_directive_1.ConfirmInputDirective,
            ]
        })
    ], SharedModule);
    return SharedModule;
}());
exports.SharedModule = SharedModule;
//# sourceMappingURL=shared.module.js.map