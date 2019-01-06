'use strict';
var App = angular.module(
    'kanban',
    ['ngSanitize', 'ui.bootstrap', 'dndLists', 'ui.bootstrap.popover']
);

App.directive('enterSubmit', function () {
    return {
        restrict: 'A',
        link: function (scope, elem, attrs) {
            elem.bind('keydown', function(event) {
                var code = event.keyCode || event.which;

                if (code === 13 && event.ctrlKey) {
                    event.preventDefault();
                    scope.$apply(attrs.enterSubmit);
                }
            });
        }
    }
});

App.directive('compileTemplate', function($compile, $parse){
    return {
        link: function(scope, element, attr){
            var parsed = $parse(attr.ngBindHtml);
            function getStringValue() {
                return (parsed(scope) || '').toString();
            }

            scope.$watch(getStringValue, function() {
                $compile(element, null, -9999)(scope);
            });
        }
    }
});

App.filter('trustHtml', function($sce) {
    return function(text) {
        return $sce.trustAsHtml(text);
    }
});

App.controller('KanbanCtrl',
function($scope, $log, $uibModal, $http, $httpParamSerializer) {
    $scope.colors = [
    '#ff0000', '#ff3300', '#ff6600', '#ff9900', '#ffcc00', '#ffff00',
    '#ccff00', '#99ff00', '#66ff00', '#33ff00', '#00ff00', '#00ff33',
    '#00ff66', '#00ff99', '#00ffcc', '#00ffff', '#00ccff', '#0099ff',
    '#0066ff', '#0033ff', '#0000ff', '#3300ff', '#6600ff', '#9900ff',
    '#cc00ff', '#ff00ff', '#ff00cc', '#ff0099', '#ff0066', '#ff0033', '#000'];

    $scope.hexToRgbA = function(hex, opacity) {
        opacity = typeof opacity !== 'undefined' ? opacity : 1.0;
        var c;
        if (/^#([A-Fa-f0-9]{3}){1,2}$/.test(hex)) {
            c = hex.substring(1).split('');
            if (c.length== 3) {
                c = [c[0], c[0], c[1], c[1], c[2], c[2]];
            }
            c = '0x'+c.join('');
            return 'rgba(' + [(c>>16)&255, (c>>8)&255, c&255].join(',') + ', ' + opacity + ')';
        }

        throw new Error('Bad Hex');
    }

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
            size: 'xlg',
            resolve: {
                column: column,
                parentScope: $scope,
            }
        });
    };

    $scope.CheckToggle = function(taskId, checkId, event) {
        event.preventDefault();
        $scope.loading = true;
        $http({
            url: '/task/' + taskId + '/',
            method: 'PUT',
            data: $httpParamSerializer({
                checkId: checkId,
            }),
            headers: {'Content-Type': 'application/x-www-form-urlencoded'}
        }).then(function(res) {
            $scope.LoadColumns();
        }, function() {
            $scope.loading = false;
            $scope.error = 'Something went wrong';
        });
    };
});

App.controller('AddEditTaskCtrl', function(
    $scope, $uibModalInstance, $http, $httpParamSerializer,
    task, column, parentScope
) {
    $scope.colors = parentScope.colors;
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
        $scope.saving = true;
        var data = {
            ID: $scope.form.ID,
            Title: $scope.form.Title,
            Description: $scope.form.Description,
            TagsString: $scope.form.TagsString,
            Color: $scope.form.Color,
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
            $scope.saving = false;
        }, function() {
            $scope.error = 'Something went wrong';
            $scope.saving = false;
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
        $scope.deleting = true;
        $http.delete('/task/' + task.ID + '/').then(
            function(res) {
                console.log(res);
                parentScope.LoadColumns();
                $uibModalInstance.close();
                $scope.deleting = false;
            }, function() {
                $scope.error = 'Something went wrong';
                $scope.deleting = false;
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
