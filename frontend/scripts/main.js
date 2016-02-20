'use strict';
var App = angular.module(
    'kanban',
    ['ui.bootstrap']
);

App.controller('KanbanCtrl',
function($scope, $uibModal, $http) {
    $http.get('/column/')
        .then(function(columns) {
            $scope.columns = columns.data;
            console.log($scope.columns.length)
        });
})