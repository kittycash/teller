"use strict";
Object.defineProperty(exports, "__esModule", { value: true });
var PhoneValidation = /** @class */ (function () {
    function PhoneValidation() {
    }
    PhoneValidation.validate = function (AC) {
        var phoneNumber = AC.value; // Get the input value
        //Regex to validate an international phone number
        var regex = /^\+(?:[0-9] ?){6,14}[0-9]$/;
        if (regex.test(phoneNumber)) {
            // Valid international phone number
            return null;
        }
        else {
            // Invalid international phone number
            return { phoneNumber: {
                    invalidNumber: phoneNumber
                } };
        }
    };
    return PhoneValidation;
}());
exports.PhoneValidation = PhoneValidation;
//# sourceMappingURL=phone-validation.js.map