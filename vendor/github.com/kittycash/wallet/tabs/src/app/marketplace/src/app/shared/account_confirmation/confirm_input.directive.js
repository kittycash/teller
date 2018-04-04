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
var focus_service_1 = require("./focus.service");
var ConfirmInputDirective = /** @class */ (function () {
    function ConfirmInputDirective(el, focusService) {
        var _this = this;
        this.focusService = focusService;
        // Allow key codes for special events
        this.specialKeys = ['Backspace', 'Tab', 'End', 'Home', 'Delete'];
        this.el = el.nativeElement;
        this.subscription = this.focusService.getFocus().subscribe(function (response) {
            //Check if the focus is for the current Index
            if (response.focusIndex === parseInt(_this.confirmInput)) {
                _this.el.focus();
            }
        });
    }
    ConfirmInputDirective.prototype.ngOnDestroy = function () {
        // unsubscribe to ensure no memory leaks
        this.subscription.unsubscribe();
    };
    ConfirmInputDirective.prototype.onKeyDown = function (e) {
        //Allow special events
        if (this.specialKeys.indexOf(e.key) !== -1) {
            if (e.key == "Backspace") {
                this.doFocus(-1);
            }
            return;
        }
        //Only allow 0-9 to be input
        if (e.keyCode < 48 || e.keyCode > 57) {
            e.preventDefault();
            return;
        }
        //If more than 1 value, replace with the newest
        if (this.el.value && this.el.value.length > 0) {
            this.el.value = '';
        }
        this.doFocus(1);
    };
    ConfirmInputDirective.prototype.doFocus = function (direction) {
        //Timeout before changing focus to prevent refresh issues
        var $this = this;
        setTimeout(function () {
            var focusIndex = parseInt($this.confirmInput);
            var nextFocusIndex = focusIndex + direction;
            $this.focusService.sendFocus(nextFocusIndex);
        });
    };
    __decorate([
        core_1.Input(),
        __metadata("design:type", Object)
    ], ConfirmInputDirective.prototype, "confirmInput", void 0);
    __decorate([
        core_1.HostListener('keydown', ['$event']),
        __metadata("design:type", Function),
        __metadata("design:paramtypes", [Object]),
        __metadata("design:returntype", void 0)
    ], ConfirmInputDirective.prototype, "onKeyDown", null);
    ConfirmInputDirective = __decorate([
        core_1.Directive({
            selector: '[confirmInput]',
        }),
        __metadata("design:paramtypes", [core_1.ElementRef, focus_service_1.FocusService])
    ], ConfirmInputDirective);
    return ConfirmInputDirective;
}());
exports.ConfirmInputDirective = ConfirmInputDirective;
//# sourceMappingURL=confirm_input.directive.js.map