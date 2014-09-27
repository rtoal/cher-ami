define(function(require, exports, module) {

    var backbone = require('backbone');
    var marionette = require('marionette');
    var app = require('app/app');

    var HeaderView = require('app/views/header-view').HeaderView;
    var SignupView = require('app/views/signup-view').SignupView;
    var FooterView = require('app/views/footer-view').FooterView

    var AppController = marionette.Controller.extend({

        initialize: function(options) {
            this.app = app;

            // Initialization of views will go here.
            this.app.headerRegion.show(new HeaderView());
            this.app.mainRegion.show(new SignupView());
            this.app.footerRegion.show(new FooterView());
        },

        // Needed for AppRouter to initialize index route.
        index: function() {

        }

    });

    exports.AppController = AppController;

});
