"use strict";
var __decorate = (this && this.__decorate) || function (decorators, target, key, desc) {
    var c = arguments.length, r = c < 3 ? target : desc === null ? desc = Object.getOwnPropertyDescriptor(target, key) : desc, d;
    if (typeof Reflect === "object" && typeof Reflect.decorate === "function") r = Reflect.decorate(decorators, target, key, desc);
    else for (var i = decorators.length - 1; i >= 0; i--) if (d = decorators[i]) r = (c < 3 ? d(r) : c > 3 ? d(target, key, r) : d(target, key)) || r;
    return c > 3 && r && Object.defineProperty(target, key, r), r;
};
Object.defineProperty(exports, "__esModule", { value: true });
var core_1 = require("@angular/core");
var Subject_1 = require("rxjs/Subject");
var FocusService = /** @class */ (function () {
    function FocusService() {
        this.subject = new Subject_1.Subject();
    }
    FocusService.prototype.sendFocus = function (focusIndex) {
        this.subject.next({ focusIndex: focusIndex });
    };
    FocusService.prototype.getFocus = function () {
        return this.subject.asObservable();
    };
    FocusService = __decorate([
        core_1.Injectable()
    ], FocusService);
    return FocusService;
}());
exports.FocusService = FocusService;
//# sourceMappingURL=focus.service.js.map