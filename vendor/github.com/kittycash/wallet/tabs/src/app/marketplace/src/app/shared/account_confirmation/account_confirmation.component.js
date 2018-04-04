"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
var __metadata = (this && this.__metadata) || function (k, v) {
    if (typeof Reflect === "object" && typeof Reflect.metadata === "function") return Reflect.metadata(k, v);
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var forms_1 = require("@angular/forms");
var router_1 = require("@angular/router");
var authentication_service_1 = require("../../core/authentication/authentication.service");
var phone_validation_1 = require("./phone-validation");
var AccountConfirmationComponent = /** @class */ (function () {
    function AccountConfirmationComponent(router, formBuilder, authenticationService) {
        this.router = router;
        this.formBuilder = formBuilder;
        this.authenticationService = authenticationService;
        //Initialize an empty array with null values for the length of the code
        this.codeLength = 6;
        this.doConfirm = false;
        this.createSendCodeForm();
        this.createConfirmForm();
        //Create an empty array to loop the compontnets
        this.formComponents = Array(this.codeLength).fill(0).map(function (x, i) { return i; });
    }
    AccountConfirmationComponent.prototype.ngOnInit = function () {
    };
    AccountConfirmationComponent.prototype.sendCode = function () {
        //Send API call here to send the SMS
        this.doConfirm = true;
    };
    AccountConfirmationComponent.prototype.logout = function () {
        var _this = this;
        this.authenticationService.logout()
            .subscribe(function () { return _this.router.navigate(['/login'], { replaceUrl: true }); });
    };
    AccountConfirmationComponent.prototype.confirm = function () {
        var $this = this;
        var code = Object.keys(this.confirmCodeForm.value).map(function (key) {
            return $this.confirmCodeForm.value[key];
        }).join('');
        this.authenticationService.confirm({ code: code })
            .subscribe(function () {
            //Code is either confirmed or not from the api
        });
    };
    AccountConfirmationComponent.prototype.createSendCodeForm = function () {
        this.sendCodeForm = this.formBuilder.group({
            phoneNumber: ['', phone_validation_1.PhoneValidation.validate]
        });
    };
    AccountConfirmationComponent.prototype.createConfirmForm = function () {
        var form = {};
        for (var i = 0; i < this.codeLength; i++) {
            form['code' + i] = ['', forms_1.Validators.required];
        }
        this.confirmCodeForm = this.formBuilder.group(form);
    };
    AccountConfirmationComponent = __decorate([
        core_1.Component({
            selector: 'account-confirmation',
            templateUrl: './account_confirmation.component.html',
            styleUrls: ['./account_confirmation.component.scss']
        }),
        __metadata("design:paramtypes", [router_1.Router,
            forms_1.FormBuilder,
            authentication_service_1.AuthenticationService])
    ], AccountConfirmationComponent);
    return AccountConfirmationComponent;
}());
exports.AccountConfirmationComponent = AccountConfirmationComponent;
//# sourceMappingURL=account_confirmation.component.js.map