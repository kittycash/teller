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
var authentication_service_1 = require("../core/authentication/authentication.service");
var kitties_service_1 = require("../shared/kitties.service");
var MyKittiesComponent = /** @class */ (function () {
    function MyKittiesComponent(authenticationService, kittiesService) {
        this.authenticationService = authenticationService;
        this.kittiesService = kittiesService;
        this.page = 1;
    }
    MyKittiesComponent.prototype.ngOnInit = function () {
        var _this = this;
        var user_id = this.authenticationService.credentials.user_id;
        this.isLoading = true;
        this.kittiesService.getMyKitties({ user_id: user_id })
            .pipe(operators_1.finalize(function () { _this.isLoading = false; }))
            .subscribe(function (kitties) {
            _this.kitties = kitties;
        });
    };
    MyKittiesComponent.prototype.showToDo = function () {
        alert("Still ToDo");
    };
    MyKittiesComponent = __decorate([
        core_1.Component({
            selector: 'app-mykitties',
            templateUrl: './mykitties.component.html',
            styleUrls: ['./mykitties.component.scss']
        }),
        __metadata("design:paramtypes", [authentication_service_1.AuthenticationService, kitties_service_1.KittiesService])
    ], MyKittiesComponent);
    return MyKittiesComponent;
}());
exports.MyKittiesComponent = MyKittiesComponent;
//# sourceMappingURL=mykitties.component.js.map