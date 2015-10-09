'use strict';

angular.module('AM').controller('UsersCtrl', function($scope, UsersService, toastr) {
    $scope.itemsPerPage = 100;
    $scope.currentPage = 1;
    $scope.pageCount = 0;
    $scope.total = 0;
    $scope.query = {};

	$scope.findUsers = function() {
		$scope.query.s =  ($scope.currentPage - 1) * $scope.itemsPerPage;
		$scope.query.l = $scope.itemsPerPage;

		UsersService.findUsers($scope.query)
		.success(function(response, status, headers) {
			$scope.users = response.users;
			$scope.total = response.total;

			if ($scope.users === undefined || $scope.users.length === 0) {
				$scope.markers = undefined;
				toastr.error("It seems that we don't have any user which meets your criteria!");
			} else {
				if ($scope.currentPage === 1) {
					toastr.info("We found " + response.total + " users which meet your criteria");
				}
				$scope.markers = response.users.map(function(user) {
			        return {
			            lng: parseFloat(user.longitude),
			            lat: parseFloat(user.latitude),
			            message:  JSON.stringify(user, null, 2),
			            focus: true
			        };
			    });
			}
		})
		.error(function(response) {
            
        });
	};

	$scope.firstRun = true;
	$scope.$watch('currentPage', function() {
        if ($scope.firstRun) {
        	$scope.firstRun = false;
        	return;
        }
        $scope.findUsers();
    });
});