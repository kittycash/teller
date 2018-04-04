"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var PasswordValidation = /** @class */ (function () {
    function PasswordValidation() {
    }
    PasswordValidation.MatchPassword = function (AC) {
        var password = AC.get('password').value; // Get the password input value
        var passwordConfirm = AC.get('passwordConfirm').value; // get the passwordConfirm input value
        if (password != passwordConfirm) {
            AC.get('passwordConfirm').setErrors({ MatchPassword: true });
        }
        else {
            return null;
        }
    };
    return PasswordValidation;
}());
exports.PasswordValidation = PasswordValidation;
//# sourceMappingURL=password-validation.js.map