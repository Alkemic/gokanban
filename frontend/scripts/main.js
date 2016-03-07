'use strict';
var App = angular.module(
    'kanban',
    ['ngSanitize', 'ui.bootstrap', 'dndLists', 'mdMarkdownIt']
);

App.config(function(markdownItConverterProvider) {
    markdownItConverterProvider.config('commonmark', {
        breaks: true,
        html: true
    });
});

App.controller('KanbanCtrl',
function($scope, $log, $uibModal, $http, $httpParamSerializer) {
    $scope.LoadColumns = function() {
        $scope.loading = true;
        $http.get('/column/')
            .then(function(columns) {
                $scope.columns = columns.data;
                $scope.loading = false;
            }, function(data) {
                console.error('Error loading data', data);
                $scope.loading = false;
            });
    };
    $scope.LoadColumns();

    $scope.AddEditTask = function(opts) {
        opts = opts || {};
        var modalInstance = $uibModal.open({
            animation: $scope.animationsEnabled,
            templateUrl: '/frontend/templates/add_task.html',
            controller: 'AddEditTaskCtrl',
            size: 'lg',
            resolve: {
                task: function() { return opts.task || {}; },
                column: function() { return opts.column; },
                parentScope: function() { return $scope; }
            }
        });

        modalInstance.result.then(function (selectedItem) {
            $scope.selected = selectedItem;
        }, function () {
            $log.info('Modal dismissed at: ' + new Date());
        });
    };
    $scope.log = function() {
        console.log.apply(console, arguments);
    }

    $scope.info ={SelectedTask: null};
    $scope.DndMoveToColumn = function(column, index, element) {
        column.Tasks.splice(index, 0, element);
        $scope.info.SelectedTask.Position = index + 1;
        $scope.MoveToColumn($scope.info.SelectedTask, column);
        $scope.info.SelectedTask = null;

        return true;
    }

    $scope.DeleteTask = function(task) {
        var modalInstance = $uibModal.open({
            animation: $scope.animationsEnabled,
            templateUrl: '/frontend/templates/delete_task.html',
            controller: 'DeleteTaskCtrl',
            size: 'sm',
            resolve: {
                task: function() { return task || {}; },
                parentScope: function() { return $scope; }
            }
        });

        modalInstance.result.then(function (selectedItem) {
            $scope.selected = selectedItem;
        }, function () {
            $log.info('Modal dismissed at: ' + new Date());
        });
    };

    $scope.MoveToColumn = function(task, column) {
        $http({
            url: '/task/' + task.ID + '/',
            method: 'PUT',
            data: $httpParamSerializer({
                ColumnID: column.ID,
                Position: task.Position,
            }),
            headers: {'Content-Type': 'application/x-www-form-urlencoded'}
        }).then(function(res) {
            $scope.LoadColumns();
        }, function() {
            $scope.error = 'Something went wrong';
        });
    };

    $scope.TaskOrder = function(column) {
            $scope.orderingInColumn
            var modalInstance = $uibModal.open({
            animation: $scope.animationsEnabled,
            templateUrl: '/frontend/templates/order_tasks.html',
            controller: 'OrderTasksCtrl',
            size: 'lg',
            resolve: {
                column: column,
                parentScope: $scope,
            }
        });
    };
});

App.controller('AddEditTaskCtrl', function(
    $scope, $uibModalInstance, $http, $httpParamSerializer,
    task, column, parentScope
) {
    $scope.column = column;
    $scope.task = task;
    $scope.form = angular.copy(task);

    if (task && task.Tags){
        $scope.form.TagsString = '';
        for (var i = task.Tags.length - 1; i >= 0; i--) {
            $scope.form.TagsString = task.Tags[i].Name +
                (task.Tags.length - 1 > i ? ', ':'') +
                $scope.form.TagsString;
        }
    }

    $scope.save = function() {
        var data = {
            ID: $scope.form.ID,
            Title: $scope.form.Title,
            Description: $scope.form.Description,
            TagsString: $scope.form.TagsString,
        };
        if (column !== undefined) {
            data.ColumnID = column.ID;
        }

        $http({
            url: '/task/' + (task.ID === undefined ? '' : task.ID + '/'),
            method: (task.ID === undefined ? 'POST' : 'PUT'),
            data: $httpParamSerializer(data),
            headers: {'Content-Type': 'application/x-www-form-urlencoded'}
        }).then(function(res) {
            parentScope.LoadColumns();
            $uibModalInstance.close();
        }, function() {
            $scope.error = 'Something went wrong';
        });
    };

    $scope.cancel = function() {
        $modalInstance.dismiss();
    };
});

App.controller('DeleteTaskCtrl', function(
    $scope, $uibModalInstance, $http,
    task, parentScope
) {
    $scope.task = task;

    $scope.confirm = function() {
        $http.delete('/task/' + task.ID + '/').then(
            function(res) {
                console.log(res)
                parentScope.LoadColumns();
                $uibModalInstance.close();
            }, function() {
                $scope.error = 'Something went wrong';
            });
    };
});

App.controller('OrderTasksCtrl', function(
    $scope, $uibModalInstance, $http, $timeout,
    column, parentScope
) {
    $scope.column = column;
    $scope.tasks = column.Tasks;
    $scope.parentScope = parentScope;
    $scope.DndMoveToColumn = function(column, index, element) {
        parentScope.DndMoveToColumn(column, index, element);
        $timeout(function() {
            $scope.loading = true;
            $http.get('/column/' + column.ID + '/')
                .then(function(data) {
                    $scope.column = data.data;
                    $scope.tasks = data.data.Tasks;
                }, function(data) {
                    console.error('Error loading data', data);
                });
        }, 100);
    };
});
