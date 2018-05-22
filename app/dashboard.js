angular
	.module('daily', ['ngRoute'])
	.config(function($routeProvider, $locationProvider) {
		$locationProvider.hashPrefix('');
		$routeProvider
			.when('/dashboard', {templateUrl: 'static/dashboard.html', controller: DashboardCtrl})
			.otherwise({redirectTo: 'dashboard'});
	});

function DashboardCtrl($scope, $http) {
    $http.get('api/v1/builds').then(function(response) {
        $scope.builds = response.data;

        $scope.activeBuild = $scope.builds[0];

        $scope.renderStatus($scope.activeBuild);
    });

    $scope.renderStatus = function (build) {
        $http.get('api/v1/status/' + build).then(function(response) {
            $scope.statuses = response.data;
        });
    };
}
