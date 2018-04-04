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
var operators_1 = require("rxjs/operators");
var router_1 = require("@angular/router");
var i18n_service_1 = require("../../core/i18n.service");
var authentication_service_1 = require("../../core/authentication/authentication.service");
var kitties_service_1 = require("../kitties.service");
var CatBoxComponent = /** @class */ (function () {
    function CatBoxComponent(router, i18nService, authenticationService, kittiesService) {
        this.router = router;
        this.i18nService = i18nService;
        this.authenticationService = authenticationService;
        this.kittiesService = kittiesService;
        this.action = 'buy';
    }
    CatBoxComponent.prototype.ngOnInit = function () {
    };
    CatBoxComponent.prototype.showToDo = function () {
        alert("Still ToDo");
    };
    CatBoxComponent.prototype.breed = function () {
        var _this = this;
        if (this.authenticationService.isAuthenticated()) {
            alert("Breed Cat code here");
        }
        else {
            alert("You need to login to perform this action");
            this.authenticationService.logout()
                .subscribe(function () { return _this.router.navigate(['/login'], { replaceUrl: true }); });
        }
    };
    CatBoxComponent.prototype.buy = function () {
        var _this = this;
        if (this.authenticationService.isAuthenticated()) {
            alert("Buy Cat code here");
        }
        else {
            alert("You need to login to perform this action");
            this.authenticationService.logout()
                .subscribe(function () { return _this.router.navigate(['/login'], { replaceUrl: true }); });
        }
    };
    CatBoxComponent.prototype.reserve = function () {
        var _this = this;
        if (this.authenticationService.isAuthenticated()) {
            alert("Reserve box code here");
        }
        else {
            alert("You need to login to perform this action");
            this.authenticationService.logout()
                .subscribe(function () { return _this.router.navigate(['/login'], { replaceUrl: true }); });
        }
    };
    CatBoxComponent.prototype.removeDetails = function () {
        this.cat.details = false;
    };
    CatBoxComponent.prototype.getDetails = function (kitty_id) {
        var _this = this;
        this.kittiesService.getDetails({ kitty_id: kitty_id })
            .pipe(operators_1.finalize(function () { _this.isLoading = false; }))
            .subscribe(function (details) {
            _this.cat.details = details;
        });
    };
    __decorate([
        core_1.Input(),
        __metadata("design:type", Object)
    ], CatBoxComponent.prototype, "cat", void 0);
    __decorate([
        core_1.Input(),
        __metadata("design:type", String)
    ], CatBoxComponent.prototype, "action", void 0);
    CatBoxComponent = __decorate([
        core_1.Component({
            selector: 'cat-box',
            templateUrl: './catbox.component.html',
            styleUrls: ['./catbox.component.scss']
        }),
        __metadata("design:paramtypes", [router_1.Router,
            i18n_service_1.I18nService,
            authentication_service_1.AuthenticationService,
            kitties_service_1.KittiesService])
    ], CatBoxComponent);
    return CatBoxComponent;
}());
exports.CatBoxComponent = CatBoxComponent;
//# sourceMappingURL=catbox.component.js.map