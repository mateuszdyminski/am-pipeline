'use strict';

angular.module('AM').controller('MenuCtrl', function($scope, $location) {
    $scope.isNavCollapsed = true;
    $scope.isCollapsed = false;
    $scope.isCollapsedHorizontal = false;

    $scope.isActive = function(viewLocation) {
        return viewLocation === $location.path();
    };
});