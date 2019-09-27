'use strict';

var AM = angular.module('AM', ['ngRoute', 'leaflet-directive', 'ui.bootstrap', 'toastr'])
    .config(function($routeProvider) {
        $routeProvider
            .when('/full', {
                templateUrl: 'app/views/full.html',
                controller: 'UsersCtrl'
            })
            .when('/auto', {
                templateUrl: 'app/views/autocomplete.html',
                controller: 'UsersCtrl'
            })
            .when('/facets', {
                templateUrl: 'app/views/facets.html',
                controller: 'UsersCtrl'
            })
            .when('/geo', {
                templateUrl: 'app/views/geo.html',
                controller: 'UsersCtrl'
            })
            .otherwise({
                redirectTo: '/full'
            });
    });