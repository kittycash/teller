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
var mykitties_component_1 = require("./mykitties.component");
var routes = core_2.Route.withAuthenticatedShell([
    { path: 'mine', component: mykitties_component_1.MyKittiesComponent, data: { title: core_2.extract('My Kitties') } }
]);
var MyKittiesRoutingModule = /** @class */ (function () {
    function MyKittiesRoutingModule() {
    }
    MyKittiesRoutingModule = __decorate([
        core_1.NgModule({
            imports: [router_1.RouterModule.forChild(routes)],
            exports: [router_1.RouterModule],
            providers: []
        })
    ], MyKittiesRoutingModule);
    return MyKittiesRoutingModule;
}());
exports.MyKittiesRoutingModule = MyKittiesRoutingModule;
//# sourceMappingURL=mykitties-routing.module.js.map