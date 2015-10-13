'use strict';

AM.service('UsersService', ['$http',
    function($http) {
        this.findUsers = function(query) {
            return $http({
                url: '/restapi/users',
                method: "GET",
                params: query
            });
        };
        this.autocompleteNick = function(query) {
            return $http({
                url: '/restapi/autocomplete',
                method: "GET",
                params: query
            });
        };
        this.aggregations = function(query) {
            return $http({
                url: '/restapi/aggregations',
                method: "GET",
                params: query
            });
        };
    }
]);