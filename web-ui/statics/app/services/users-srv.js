'use strict';

AM.service('UsersService', ['$http',
    function($http) {
        this.findUsers = function(query) {
            return $http({
                url: 'https://API_SERVER/api/users',
                method: "GET",
                params: query
            });
        };
        this.autocompleteNick = function(query) {
            return $http({
                url: 'https://API_SERVER/api/autocomplete',
                method: "GET",
                params: query
            });
        };
        this.aggregations = function(query) {
            return $http({
                url: 'https://API_SERVER/api/aggregations',
                method: "GET",
                params: query
            });
        };
    }
]);