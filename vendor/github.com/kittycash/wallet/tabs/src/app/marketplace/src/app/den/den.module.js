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
var den_routing_module_1 = require("./den-routing.module");
var den_component_1 = require("./den.component");
var kitties_service_1 = require("../shared/kitties.service");
var ngx_pagination_1 = require("ngx-pagination");
var ng2_order_pipe_1 = require("ng2-order-pipe");
var DenModule = /** @class */ (function () {
    function DenModule() {
    }
    DenModule = __decorate([
        core_1.NgModule({
            imports: [
                common_1.CommonModule,
                forms_1.FormsModule,
                core_2.TranslateModule,
                core_3.CoreModule,
                shared_1.SharedModule,
                den_routing_module_1.DenRoutingModule,
                ngx_pagination_1.NgxPaginationModule,
                ng2_order_pipe_1.Ng2OrderModule
            ],
            declarations: [
                den_component_1.DenComponent
            ],
            providers: [
                kitties_service_1.KittiesService
            ]
        })
    ], DenModule);
    return DenModule;
}());
exports.DenModule = DenModule;
//# sourceMappingURL=den.module.js.map