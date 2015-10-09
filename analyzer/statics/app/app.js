'use strict';

var AM = angular.module('AM', ['ngRoute', 'leaflet-directive', 'ui.bootstrap', 'toastr'])
	.config(function ($routeProvider) {
    	$routeProvider
      		.when('/', {
        		templateUrl: 'app/views/main.html',
        		controller: 'UsersCtrl'
      		})
      		.otherwise({
        		redirectTo: '/'
      		});
    });
