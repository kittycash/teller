"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var platform_browser_1 = require("@angular/platform-browser");
var core_1 = require("@angular/core");
var forms_1 = require("@angular/forms");
var http_1 = require("@angular/http");
var service_worker_1 = require("@angular/service-worker");
var core_2 = require("@ngx-translate/core");
var ng_bootstrap_1 = require("@ng-bootstrap/ng-bootstrap");
var environment_1 = require("@env/environment");
var core_3 = require("@app/core");
var shared_1 = require("@app/shared");
var explore_module_1 = require("./explore/explore.module");
var boxes_module_1 = require("./boxes/boxes.module");
var about_module_1 = require("./about/about.module");
var mykitties_module_1 = require("./mykitties/mykitties.module");
var forsale_module_1 = require("./forsale/forsale.module");
var den_module_1 = require("./den/den.module");
var login_module_1 = require("./login/login.module");
var register_module_1 = require("./register/register.module");
var app_component_1 = require("./app.component");
var app_routing_module_1 = require("./app-routing.module");
var AppModule = /** @class */ (function () {
    function AppModule() {
    }
    AppModule = __decorate([
        core_1.NgModule({
            imports: [
                platform_browser_1.BrowserModule,
                service_worker_1.ServiceWorkerModule.register('/ngsw-worker.js', { enabled: environment_1.environment.production }),
                forms_1.FormsModule,
                http_1.HttpModule,
                core_2.TranslateModule.forRoot(),
                ng_bootstrap_1.NgbModule.forRoot(),
                core_3.CoreModule,
                shared_1.SharedModule,
                explore_module_1.ExploreModule,
                boxes_module_1.BoxesModule,
                about_module_1.AboutModule,
                mykitties_module_1.MyKittiesModule,
                forsale_module_1.ForSaleModule,
                den_module_1.DenModule,
                login_module_1.LoginModule,
                register_module_1.RegisterModule,
                app_routing_module_1.AppRoutingModule
            ],
            declarations: [
                app_component_1.AppComponent
            ],
            providers: [],
            entryComponents: [],
            bootstrap: [app_component_1.AppComponent]
        })
    ], AppModule);
    return AppModule;
}());
exports.AppModule = AppModule;
//# sourceMappingURL=app.module.js.map