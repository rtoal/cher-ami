// define(function(require, exports, module) {
// 	var marionette = require('marionette');
// 	var template = require('hbs!../templates/sidebar-view')

// 	var SidebarView = marionette.ItemView.extend({

// 	}

// 	exports.SidebarView = SidebarView;
// }


define(function(require, exports, module) {
    var marionette = require('marionette');
    var template = require('hbs!../templates/sidebar-view');

    var Sidebar = require('app/models/sidebar').Sidebar;

    var passCheck = require('app/utils/pass-check').PassCheck;

    var inputValidator = require('app/utils/input-validator').InputValidator;

    var SidebarView = marionette.ItemView.extend({
        template: template,

        //takes the div marionette creates and give it a class named mainContainer.
        tagName: "div",
        className: "mainContainer",
        ui: {
            handle: '#handle',
            email: '#input-email',
            pass: '#pass1',
            confirmPass: '#pass2',
            warning: '#confirmMessage',
            rememberMe: '#remember-me',
            signup: '#signup',
            inputForm: '#signupform'
        },

        events: {
            'click #remember-me': 'onRememberConfirm',
            'keyup #handle': 'inputValidate',
            'click #signup': 'onFormConfirm'
        },

        initialize: function(options) {

        },

        onRememberConfirm: function(options) {
            // Session-request method goes here
        },

        onFormConfirm: function(event) {
            event.preventDefault();
            var req = new Signup({
                handle: this.ui.handle.val(),
                email: this.ui.email.val(),
                password: this.ui.pass.val(),
                confirmpassword: this.ui.confirmPass.val()
            });
            console.log(req)
            req.save();
        },

        passwordMatch: function(event) {
            passCheck(this.ui.pass, this.ui.confirmPass, this.ui.warning)
        },

        inputValidate: function(event) {
             inputValidator(this.ui.inputForm)
        }

    });

    exports.SidebarView = SidebarView;
})