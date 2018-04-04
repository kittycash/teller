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
var router_1 = require("@angular/router");
var forms_1 = require("@angular/forms");
var operators_1 = require("rxjs/operators");
var environment_1 = require("@env/environment");
var core_2 = require("@app/core");
var log = new core_2.Logger('Login');
var LoginComponent = /** @class */ (function () {
    function LoginComponent(router, formBuilder, i18nService, authenticationService) {
        this.router = router;
        this.formBuilder = formBuilder;
        this.i18nService = i18nService;
        this.authenticationService = authenticationService;
        this.version = environment_1.environment.version;
        this.isLoading = false;
        this.createForm();
    }
    LoginComponent.prototype.ngOnInit = function () { };
    LoginComponent.prototype.login = function () {
        var _this = this;
        this.isLoading = true;
        this.authenticationService.login(this.loginForm.value)
            .pipe(operators_1.finalize(function () {
            _this.loginForm.markAsPristine();
            _this.isLoading = false;
        }))
            .subscribe(function (credentials) {
            log.debug(credentials.email + " successfully logged in");
            _this.router.navigate(['/'], { replaceUrl: true });
        }, function (error) {
            log.debug("Login error: " + error);
            _this.error = error;
        });
    };
    LoginComponent.prototype.setLanguage = function (language) {
        this.i18nService.language = language;
    };
    Object.defineProperty(LoginComponent.prototype, "currentLanguage", {
        get: function () {
            return this.i18nService.language;
        },
        enumerable: true,
        configurable: true
    });
    Object.defineProperty(LoginComponent.prototype, "languages", {
        get: function () {
            return this.i18nService.supportedLanguages;
        },
        enumerable: true,
        configurable: true
    });
    LoginComponent.prototype.createForm = function () {
        this.loginForm = this.formBuilder.group({
            email: ['', forms_1.Validators.email],
            password: ['', forms_1.Validators.required],
            remember: true
        });
    };
    LoginComponent = __decorate([
        core_1.Component({
            selector: 'app-login',
            templateUrl: './login.component.html',
            styleUrls: ['./login.component.scss']
        }),
        __metadata("design:paramtypes", [router_1.Router,
            forms_1.FormBuilder,
            core_2.I18nService,
            core_2.AuthenticationService])
    ], LoginComponent);
    return LoginComponent;
}());
exports.LoginComponent = LoginComponent;
//# sourceMappingURL=login.component.js.map