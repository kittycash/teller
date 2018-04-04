"use strict";
var __extends = (this && this.__extends) || (function () {
    var extendStatics = Object.setPrototypeOf ||
        ({ __proto__: [] } instanceof Array && function (d, b) { d.__proto__ = b; }) ||
        function (d, b) { for (var p in b) if (b.hasOwnProperty(p)) d[p] = b[p]; };
    return function (d, b) {
        extendStatics(d, b);
        function __() { this.constructor = d; }
        d.prototype = b === null ? Object.create(b) : (__.prototype = b.prototype, new __());
    };
})();
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
var Observable_1 = require("rxjs/Observable");
var throw_1 = require("rxjs/observable/throw");
var operators_1 = require("rxjs/operators");
var lodash_1 = require("lodash");
var environment_1 = require("@env/environment");
var logger_service_1 = require("../logger.service");
var http_cache_service_1 = require("./http-cache.service");
var request_options_args_1 = require("./request-options-args");
var log = new logger_service_1.Logger('HttpService');
/**
 * Provides a base framework for http service extension.
 * The default extension adds support for API prefixing, request caching and default error handler.
 */
var HttpService = /** @class */ (function (_super) {
    __extends(HttpService, _super);
    function HttpService(backend, defaultOptions, httpCacheService) {
        var _this = 
        // Customize default options here if needed
        _super.call(this, backend, defaultOptions) || this;
        _this.defaultOptions = defaultOptions;
        _this.httpCacheService = httpCacheService;
        return _this;
    }
    /**
     * Performs any type of http request.
     * You can customize this method with your own extended behavior.
     */
    HttpService.prototype.request = function (request, options) {
        var _this = this;
        var requestOptions = options || {};
        var url;
        if (typeof request === 'string') {
            url = request;
            request = this.getProperEndpoint(url); //environment.serverUrl + url;
        }
        else {
            url = request.url;
            request.url = this.getProperEndpoint(url); //environment.serverUrl + url;
        }
        if (!requestOptions.cache) {
            // Do not use cache
            return this.httpRequest(request, requestOptions);
        }
        else {
            return new Observable_1.Observable(function (subscriber) {
                var cachedData = requestOptions.cache === request_options_args_1.HttpCachePolicy.Update ?
                    null : _this.httpCacheService.getCacheData(url);
                if (cachedData !== null) {
                    // Create new response to avoid side-effects
                    subscriber.next(new http_1.Response(cachedData));
                    subscriber.complete();
                }
                else {
                    _this.httpRequest(request, requestOptions).subscribe(function (response) {
                        // Store the serializable version of the response
                        _this.httpCacheService.setCacheData(url, null, new http_1.ResponseOptions({
                            body: response.text(),
                            status: response.status,
                            headers: response.headers,
                            statusText: response.statusText,
                            type: response.type,
                            url: response.url
                        }));
                        subscriber.next(response);
                    }, function (error) { return subscriber.error(error); }, function () { return subscriber.complete(); });
                }
            });
        }
    };
    HttpService.prototype.get = function (url, options) {
        return this.request(url, lodash_1.extend({}, options, { method: http_1.RequestMethod.Get }));
    };
    HttpService.prototype.post = function (url, body, options) {
        return this.request(url, lodash_1.extend({}, options, {
            body: body,
            method: http_1.RequestMethod.Post
        }));
    };
    HttpService.prototype.put = function (url, body, options) {
        return this.request(url, lodash_1.extend({}, options, {
            body: body,
            method: http_1.RequestMethod.Put
        }));
    };
    HttpService.prototype.delete = function (url, options) {
        return this.request(url, lodash_1.extend({}, options, { method: http_1.RequestMethod.Delete }));
    };
    HttpService.prototype.patch = function (url, body, options) {
        return this.request(url, lodash_1.extend({}, options, {
            body: body,
            method: http_1.RequestMethod.Patch
        }));
    };
    HttpService.prototype.head = function (url, options) {
        return this.request(url, lodash_1.extend({}, options, { method: http_1.RequestMethod.Head }));
    };
    HttpService.prototype.options = function (url, options) {
        return this.request(url, lodash_1.extend({}, options, { method: http_1.RequestMethod.Options }));
    };
    // Customize the default behavior for all http requests here if needed
    HttpService.prototype.httpRequest = function (request, options) {
        var _this = this;
        var req = _super.prototype.request.call(this, request, options);
        if (!options.skipErrorHandler) {
            req = req.pipe(operators_1.catchError(function (error) { return _this.errorHandler(error); }));
        }
        return req;
    };
    // Determines if the endpoint should go to the real API or the fake one
    HttpService.prototype.getProperEndpoint = function (endpoint) {
        var validEndpoints = [
            "/iko/kitty/",
            "/iko/tx/",
            "/iko/tx_seq",
            "/iko/address"
        ];
        var valid = false;
        for (var i = 0; i < validEndpoints.length; i++) {
            if (endpoint.startsWith(validEndpoints[i])) {
                valid = true;
            }
        }
        if (valid) {
            return environment_1.environment.serverUrl + endpoint;
        }
        else {
            return environment_1.environment.fakeServerUrl + endpoint;
        }
    };
    // Customize the default error handler here if needed
    HttpService.prototype.errorHandler = function (response) {
        if (environment_1.environment.production) {
            // Avoid unchaught exceptions on production
            log.error('Request error', response);
            return throw_1._throw(response);
        }
        throw response;
    };
    HttpService = __decorate([
        core_1.Injectable(),
        __metadata("design:paramtypes", [http_1.ConnectionBackend,
            http_1.RequestOptions,
            http_cache_service_1.HttpCacheService])
    ], HttpService);
    return HttpService;
}(http_1.Http));
exports.HttpService = HttpService;
//# sourceMappingURL=http.service.js.map