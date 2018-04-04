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
var __param = (this && this.__param) || function (paramIndex, decorator) {
    return function (target, key) { decorator(target, key, paramIndex); }
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var common_1 = require("@angular/common");
var router_1 = require("@angular/router");
var http_1 = require("@angular/http");
var core_2 = require("@ngx-translate/core");
var ng_bootstrap_1 = require("@ng-bootstrap/ng-bootstrap");
var shared_1 = require("@app/shared");
var shell_component_1 = require("./shell/shell.component");
var header_component_1 = require("./shell/header/header.component");
var route_reusable_strategy_1 = require("./route-reusable-strategy");
var authentication_service_1 = require("./authentication/authentication.service");
var authentication_guard_1 = require("./authentication/authentication.guard");
var i18n_service_1 = require("./i18n.service");
var http_service_1 = require("./http/http.service");
var http_cache_service_1 = require("./http/http-cache.service");
function createHttpService(backend, defaultOptions, httpCacheService) {
    return new http_service_1.HttpService(backend, defaultOptions, httpCacheService);
}
exports.createHttpService = createHttpService;
var CoreModule = /** @class */ (function () {
    function CoreModule(parentModule) {
        // Import guard
        if (parentModule) {
            throw new Error(parentModule + " has already been loaded. Import Core module in the AppModule only.");
        }
    }
    CoreModule = __decorate([
        core_1.NgModule({
            imports: [
                common_1.CommonModule,
                http_1.HttpModule,
                core_2.TranslateModule,
                ng_bootstrap_1.NgbModule,
                router_1.RouterModule,
                shared_1.SharedModule
            ],
            declarations: [
                header_component_1.HeaderComponent,
                shell_component_1.ShellComponent
            ],
            providers: [
                authentication_service_1.AuthenticationService,
                authentication_guard_1.AuthenticationGuard,
                i18n_service_1.I18nService,
                http_cache_service_1.HttpCacheService,
                {
                    provide: http_1.Http,
                    deps: [http_1.XHRBackend, http_1.RequestOptions, http_cache_service_1.HttpCacheService],
                    useFactory: createHttpService
                },
                {
                    provide: router_1.RouteReuseStrategy,
                    useClass: route_reusable_strategy_1.RouteReusableStrategy
                }
            ]
        }),
        __param(0, core_1.Optional()), __param(0, core_1.SkipSelf()),
        __metadata("design:paramtypes", [CoreModule])
    ], CoreModule);
    return CoreModule;
}());
exports.CoreModule = CoreModule;
//# sourceMappingURL=core.module.js.map