"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var router_1 = require("@angular/router");
var core_2 = require("@app/core");
var forsale_component_1 = require("./forsale.component");
var routes = core_2.Route.withShell([
    { path: 'forsale', component: forsale_component_1.ForSaleComponent, data: { title: core_2.extract('Kitties For Sale') } }
]);
var ForSaleRoutingModule = /** @class */ (function () {
    function ForSaleRoutingModule() {
    }
    ForSaleRoutingModule = __decorate([
        core_1.NgModule({
            imports: [router_1.RouterModule.forChild(routes)],
            exports: [router_1.RouterModule],
            providers: []
        })
    ], ForSaleRoutingModule);
    return ForSaleRoutingModule;
}());
exports.ForSaleRoutingModule = ForSaleRoutingModule;
//# sourceMappingURL=forsale-routing.module.js.map