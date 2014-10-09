define(function(require, exports, module) {

    var backbone = require('backbone');

    var Login = Backbone.Model.extend({
        url: '/api/login',
        defaults: {
            handle: null,
            password: null
        },

        initialize: function() {

        }
    });

    exports.Login = Login;

});