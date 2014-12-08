define(function(require, exports, module) {

    var backbone = require('backbone');

    var Message = Backbone.Model.extend({
        url: 'api/messages',
        defaults: {
            content: null,
            author: null,
            id: null
        },

        initialize: function() {
        },

        update: function() {
        }
    });

    exports.Message = Message;

});
