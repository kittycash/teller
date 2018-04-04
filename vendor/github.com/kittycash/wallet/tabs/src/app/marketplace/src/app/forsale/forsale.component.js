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
var kitties_service_1 = require("../shared/kitties.service");
var ForSaleComponent = /** @class */ (function () {
    function ForSaleComponent(kittiesService) {
        this.kittiesService = kittiesService;
        this.page = 1;
        this.key = 'kitty_id';
        this.reverse = false;
    }
    ForSaleComponent.prototype.ngOnInit = function () {
        var _this = this;
        this.isLoading = true;
        this.kittiesService.getForSale()
            .pipe(operators_1.finalize(function () { _this.isLoading = false; }))
            .subscribe(function (kitties) {
            _this.kitties = kitties;
        });
    };
    ForSaleComponent.prototype.showToDo = function () {
        alert("Still ToDo");
    };
    //set default sorting
    // this.key = 'kitty_id'; 
    // this.reverse = false;
    ForSaleComponent.prototype.sort = function (key) {
        this.key = key;
        this.reverse = !this.reverse;
    };
    ForSaleComponent = __decorate([
        core_1.Component({
            selector: 'app-forsale',
            templateUrl: './forsale.component.html',
            styleUrls: ['./forsale.component.scss']
        }),
        __metadata("design:paramtypes", [kitties_service_1.KittiesService])
    ], ForSaleComponent);
    return ForSaleComponent;
}());
exports.ForSaleComponent = ForSaleComponent;
//# sourceMappingURL=forsale.component.js.map