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
var core_2 = require("@ngx-translate/core");
var core_3 = require("@app/core");
var shared_1 = require("@app/shared");
var mykitties_routing_module_1 = require("./mykitties-routing.module");
var mykitties_component_1 = require("./mykitties.component");
var kitties_service_1 = require("../shared/kitties.service");
var ngx_pagination_1 = require("ngx-pagination");
var MyKittiesModule = /** @class */ (function () {
    function MyKittiesModule() {
    }
    MyKittiesModule = __decorate([
        core_1.NgModule({
            imports: [
                common_1.CommonModule,
                core_2.TranslateModule,
                core_3.CoreModule,
                shared_1.SharedModule,
                mykitties_routing_module_1.MyKittiesRoutingModule,
                ngx_pagination_1.NgxPaginationModule
            ],
            declarations: [
                mykitties_component_1.MyKittiesComponent
            ],
            providers: [
                kitties_service_1.KittiesService
            ]
        })
    ], MyKittiesModule);
    return MyKittiesModule;
}());
exports.MyKittiesModule = MyKittiesModule;
//# sourceMappingURL=mykitties.module.js.map