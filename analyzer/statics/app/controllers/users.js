'use strict';

angular.module('AM').controller('UsersCtrl', function($scope, $location, UsersService, toastr) {
    $scope.itemsPerPage = 100;
    $scope.currentPage = 1;
    $scope.pageCount = 0;
    $scope.total = 0;
    $scope.query = {};

    $scope.findUsers = function() {
        if ($scope.query.query) {
            $scope.query.query = $scope.query.query.toLowerCase();
        }
        $scope.query.s = ($scope.currentPage - 1) * $scope.itemsPerPage;
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
                            lng: user.location.lon,
                            lat: user.location.lat,
                            message: JSON.stringify(user, null, 2),
                            focus: true
                        };
                    });
                }
            })
            .error(function(response) {

            });
    };

    $scope.autocomplete = function(val) {
        var keywords = [];
        keywords.push(val);

        return UsersService.autocompleteNick({
                nick: val
            })
            .then(function(response) {
                for (var i in response.data) {
                    keywords.push(response.data[i]);
                }
                return keywords;
            });
    }

    $scope.aggregations = function() {
        UsersService.aggregations($scope.query)
            .success(function(response, status, headers) {
                $scope.buckets = response;
            })
            .error(function(response) {

            });
    }

    $scope.$on('leafletDirectiveMap.click', function(event, args) {
        $scope.query.lat = args.leafletEvent.latlng.lat;
        $scope.query.lon = args.leafletEvent.latlng.lng;
    });

    $scope.firstRun = true;
    $scope.$watch('currentPage', function() {
        if ($scope.firstRun) {
            $scope.firstRun = false;
            return;
        }

        $scope.findUsers();
    });
});