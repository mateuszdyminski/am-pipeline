'use strict';

AM.service('UsersService', ['$http', function($http) {
	this.findUsers = function(query) {
		return $http({
				    url: '/restapi/users', 
				    method: "GET",
				    params: query
				});
	};
}]);