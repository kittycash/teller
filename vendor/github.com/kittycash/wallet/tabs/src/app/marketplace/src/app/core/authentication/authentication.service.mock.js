"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var of_1 = require("rxjs/observable/of");
var MockAuthenticationService = /** @class */ (function () {
    function MockAuthenticationService() {
        this.credentials = {
            email: 'test@test.com',
            token: '123',
            user_id: 100,
            confirmed: true
        };
    }
    MockAuthenticationService.prototype.login = function (context) {
        return of_1.of({
            email: context.email,
            token: '123456',
            user_id: 100,
            confirmed: true
        });
    };
    MockAuthenticationService.prototype.logout = function () {
        this.credentials = null;
        return of_1.of(true);
    };
    MockAuthenticationService.prototype.isAuthenticated = function () {
        return !!this.credentials;
    };
    return MockAuthenticationService;
}());
exports.MockAuthenticationService = MockAuthenticationService;
//# sourceMappingURL=authentication.service.mock.js.map