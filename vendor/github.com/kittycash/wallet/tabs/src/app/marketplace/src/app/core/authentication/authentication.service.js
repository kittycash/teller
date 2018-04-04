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
var of_1 = require("rxjs/observable/of");
var credentialsKey = 'credentials';
/**
 * Provides a base for authentication workflow.
 * The Credentials interface as well as login/logout methods should be replaced with proper implementation.
 */
var AuthenticationService = /** @class */ (function () {
    function AuthenticationService() {
        var savedCredentials = sessionStorage.getItem(credentialsKey) || localStorage.getItem(credentialsKey);
        if (savedCredentials) {
            this._credentials = JSON.parse(savedCredentials);
        }
    }
    /**
     * Authenticates the user.
     * @param {LoginContext} context The login parameters.
     * @return {Observable<Credentials>} The user credentials.
     */
    AuthenticationService.prototype.login = function (context) {
        // Replace by proper registration call
        var data = {
            email: context.email,
            token: '123456',
            user_id: 100,
            confirmed: true
        };
        this.setCredentials(data, context.remember);
        return of_1.of(data);
    };
    /**
     * Registers the user.
     * @param {RegisterContext} context The registration parameters.
     * @return {Observable<Credentials>} The user credentials.
     */
    AuthenticationService.prototype.register = function (context) {
        // Replace by proper authentication call
        var data = {
            email: context.email,
            token: '123456',
            user_id: 100,
            confirmed: false
        };
        this.setCredentials(data, false);
        return of_1.of(data);
    };
    /**
     * Logs out the user and clear credentials.
     * @return {Observable<boolean>} True if the user was logged out successfully.
     */
    AuthenticationService.prototype.logout = function () {
        // Customize credentials invalidation here
        this.setCredentials();
        return of_1.of(true);
    };
    /**
     * Confirms the authentication code is successful.
     * @return {Observable<boolean>} True if the user was logged out successfully.
     */
    AuthenticationService.prototype.confirm = function (context) {
        // Get the credentials
        var data = this.credentials;
        data.confirmed = true;
        this.setCredentials(data, false);
        return of_1.of(data);
    };
    /**
     * Checks is the user is authenticated.
     * @return {boolean} True if the user is authenticated.
     */
    AuthenticationService.prototype.isAuthenticated = function () {
        return !!this.credentials;
    };
    /**
     * Checks is the user has confirmed their account.
     * @return {boolean} True if the user's account has been confirmed.
     */
    AuthenticationService.prototype.isConfirmed = function () {
        return !!this.credentials.confirmed;
    };
    Object.defineProperty(AuthenticationService.prototype, "credentials", {
        /**
         * Gets the user credentials.
         * @return {Credentials} The user credentials or null if the user is not authenticated.
         */
        get: function () {
            return this._credentials;
        },
        enumerable: true,
        configurable: true
    });
    /**
     * Sets the user credentials.
     * The credentials may be persisted across sessions by setting the `remember` parameter to true.
     * Otherwise, the credentials are only persisted for the current session.
     * @param {Credentials=} credentials The user credentials.
     * @param {boolean=} remember True to remember credentials across sessions.
     */
    AuthenticationService.prototype.setCredentials = function (credentials, remember) {
        this._credentials = credentials || null;
        if (credentials) {
            var storage = remember ? localStorage : sessionStorage;
            storage.setItem(credentialsKey, JSON.stringify(credentials));
        }
        else {
            sessionStorage.removeItem(credentialsKey);
            localStorage.removeItem(credentialsKey);
        }
    };
    AuthenticationService = __decorate([
        core_1.Injectable(),
        __metadata("design:paramtypes", [])
    ], AuthenticationService);
    return AuthenticationService;
}());
exports.AuthenticationService = AuthenticationService;
//# sourceMappingURL=authentication.service.js.map