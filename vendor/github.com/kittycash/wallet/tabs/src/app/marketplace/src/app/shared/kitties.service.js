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
var http_1 = require("@angular/http");
var of_1 = require("rxjs/observable/of");
var operators_1 = require("rxjs/operators");
var i18n_service_1 = require("../core/i18n.service");
var routes = {
    explore: function () { return "/explore/all"; },
    boxes: function () { return "/boxes/all"; },
    forsale: function () { return "/forsale/all"; },
    forsire: function () { return "/forsire/all"; },
    mykitties: function (u) { return "/user/" + u.user_id + "/mine"; },
    details: function (k) { return "/iko/kitty/" + k.kitty_id; }
};
var KittiesService = /** @class */ (function () {
    function KittiesService(http, i18nService) {
        this.http = http;
        this.i18nService = i18nService;
    }
    KittiesService.prototype.getExplore = function () {
        return this.http.get(routes.explore(), this.getOptions())
            .pipe(operators_1.map(function (res) { return res.json(); }), operators_1.map(function (body) { return body.data; }), operators_1.catchError(function () { return of_1.of('Error, could not load boxes'); }));
    };
    KittiesService.prototype.getBoxes = function () {
        return this.http.get(routes.boxes(), this.getOptions())
            .pipe(operators_1.map(function (res) { return res.json(); }), operators_1.map(function (body) { return body.data; }), operators_1.catchError(function () { return of_1.of('Error, could not load boxes'); }));
    };
    KittiesService.prototype.getForSale = function () {
        return this.http.get(routes.forsale(), this.getOptions())
            .pipe(operators_1.map(function (res) { return res.json(); }), operators_1.map(function (body) { return body.data; }), operators_1.catchError(function () { return of_1.of('Error, could not load kittens'); }));
    };
    KittiesService.prototype.getForSire = function () {
        return this.http.get(routes.forsire(), this.getOptions())
            .pipe(operators_1.map(function (res) { return res.json(); }), operators_1.map(function (body) { return body.data; }), operators_1.catchError(function () { return of_1.of('Error, could not load kittens'); }));
    };
    KittiesService.prototype.getMyKitties = function (context) {
        return this.http.get(routes.mykitties(context), this.getOptions())
            .pipe(operators_1.map(function (res) { return res.json(); }), operators_1.map(function (body) { return body.data; }), operators_1.catchError(function () { return of_1.of('Error, could not load your kitties'); }));
    };
    KittiesService.prototype.getDetails = function (context) {
        return this.http.get(routes.details(context), this.getOptions())
            .pipe(operators_1.map(function (res) { return res.json(); }), operators_1.map(function (body) { return body.data; }), operators_1.catchError(function () { return of_1.of('Error, could not load your kitty details'); }));
    };
    KittiesService.prototype.getOptions = function () {
        var headers = new http_1.Headers();
        headers.append('Accept-Language', this.getLanguage());
        var opts = new http_1.RequestOptions({ cache: true });
        opts.headers = headers;
        return opts;
    };
    KittiesService.prototype.getLanguage = function () {
        var language = this.i18nService.language ? this.i18nService.language : 'en-Us';
        return language;
        //this.http.defaultOptions.headers.set("", language);
    };
    KittiesService = __decorate([
        core_1.Injectable(),
        __metadata("design:paramtypes", [http_1.Http, i18n_service_1.I18nService])
    ], KittiesService);
    return KittiesService;
}());
exports.KittiesService = KittiesService;
//# sourceMappingURL=kitties.service.js.map