angular.module(
    "kanban",
    ["ngSanitize", "ui.bootstrap", "dndLists", "ui.bootstrap.popover", "kanban.templates"]
).directive("enterSubmit", () => ({
    restrict: "A",
    link: (scope, elem, attrs) => {
        elem.bind("keydown", (event) => {
            var code = event.keyCode || event.which

            if (code === 13 && event.ctrlKey) {
                event.preventDefault()
                scope.$apply(attrs.enterSubmit)
            }
        })
    }
})).directive("compileTemplate", ($compile, $parse) => ({
    link: (scope, element, attr) => {
        var parsed = $parse(attr.ngBindHtml)
        const getStringValue = () => {
            return (parsed(scope) || "").toString()
        }

        scope.$watch(getStringValue, () => {
            $compile(element, null, -9999)(scope)
        })
    }
})).filter("trustHtml", ($sce) => (text) => $sce.trustAsHtml(text))
.controller("KanbanCtrl", ($scope, $log, $uibModal, $http, $httpParamSerializer) => {
    $scope.colors = [
    "#ff0000", "#ff3300", "#ff6600", "#ff9900", "#ffcc00", "#ffff00",
    "#ccff00", "#99ff00", "#66ff00", "#33ff00", "#00ff00", "#00ff33",
    "#00ff66", "#00ff99", "#00ffcc", "#00ffff", "#00ccff", "#0099ff",
    "#0066ff", "#0033ff", "#0000ff", "#3300ff", "#6600ff", "#9900ff",
    "#cc00ff", "#ff00ff", "#ff00cc", "#ff0099", "#ff0066", "#ff0033", "#000"]

    $scope.hexToRgbA = (hex, opacity) => {
        opacity = opacity !== undefined ? opacity : 1.0
        var c
        if (/^#([A-Fa-f0-9]{3}){1,2}$/.test(hex)) {
            c = hex.substring(1).split("")
            if (c.length === 3) {
                c = [c[0], c[0], c[1], c[1], c[2], c[2]]
            }
            c = "0x"+c.join("")
            return `rgba(${[(c>>16)&255, (c>>8)&255, c&255].join(",")},${opacity})`
        }

        throw new Error("Bad Hex")
    }

    $scope.LoadColumns = () => {
        $scope.loading = true
        $http.get("/column/")
            .then((columns) => {
                $scope.columns = columns.data
                $scope.loading = false
            }, (data) => {
                console.error("Error loading data", data)
                $scope.loading = false
            })
    }
    $scope.LoadColumns()

    $scope.AddEditTask = (opts) => {
        opts = opts || {}
        var modalInstance = $uibModal.open({
            animation: $scope.animationsEnabled,
            templateUrl: "/frontend/templates/add_task.html",
            controller: "AddEditTaskCtrl",
            size: "lg",
            resolve: {
                task: () => opts.task || {},
                column: () => opts.column,
                parentScope: () => $scope
            }
        })

        modalInstance.result.then((selectedItem) => {
            $scope.selected = selectedItem
        })
    }
    $scope.log = () => {
        console.log.apply(console, arguments)
    }

    $scope.info = {SelectedTask: null}
    $scope.DndMoveToColumn = (column, index, element) => {
        column.Tasks.splice(index, 0, element)
        $scope.info.SelectedTask.Position = index + 1
        $scope.MoveToColumn($scope.info.SelectedTask, column)
        $scope.info.SelectedTask = null

        return true
    }

    $scope.DeleteTask = (task) => {
        var modalInstance = $uibModal.open({
            animation: $scope.animationsEnabled,
            templateUrl: "/frontend/templates/delete_task.html",
            controller: "DeleteTaskCtrl",
            size: "sm",
            resolve: {
                task: () => task || {},
                parentScope: () => $scope
            }
        })

        modalInstance.result.then((selectedItem) => {
            $scope.selected = selectedItem
        })
    }

    $scope.MoveToColumn = (task, column) => {
        $http({
            url: `/task/${task.ID}/`,
            method: "PUT",
            data: $httpParamSerializer({
                ColumnID: column.ID,
                Position: task.Position,
            }),
            headers: {"Content-Type": "application/x-www-form-urlencoded"}
        }).then((res) => {
            $scope.LoadColumns()
        }, () => {
            $scope.error = "Something went wrong"
        })
    }

    $scope.TaskOrder = (column) => {
        $uibModal.open({
            animation: $scope.animationsEnabled,
            templateUrl: "/frontend/templates/order_tasks.html",
            controller: "OrderTasksCtrl",
            size: "xlg",
            resolve: {
                column: column,
                parentScope: $scope,
            }
        })
    }

    $scope.CheckToggle = (taskId, checkId, event) => {
        event.preventDefault()
        $scope.loading = true
        $http({
            url: `/task/${taskId}/`,
            method: "PUT",
            data: $httpParamSerializer({
                checkId: checkId,
            }),
            headers: {"Content-Type": "application/x-www-form-urlencoded"}
        }).then((res) => {
            $scope.LoadColumns()
        }, () => {
            $scope.loading = false
            $scope.error = "Something went wrong"
        })
    }

    $scope.editUser = user => {
        $uibModal.open({
            templateUrl: "edit_user.html",
            controller: "EditUserCtrl",
            size: "small",
            resolve: {
                user: () => user,
                parentScope: () => $scope,
            }
        })
    }
}).controller("AddEditTaskCtrl", ($scope, $uibModalInstance, $http, $httpParamSerializer, task, column, parentScope) => {
    $scope.colors = parentScope.colors
    $scope.column = column
    $scope.task = task
    $scope.form = angular.copy(task)
    $scope.form.TagsString = (task.Tags || []).map((tag) => tag.Name).join(", ")

    $scope.save = () => {
        $scope.saving = true
        var data = {
            ID: $scope.form.ID,
            Title: $scope.form.Title,
            Description: $scope.form.Description,
            TagsString: $scope.form.TagsString,
            Color: $scope.form.Color,
        }
        if (column !== undefined) {
            data.ColumnID = column.ID
        }

        $http({
            url: `/task/${(task.ID === undefined ? "" : task.ID + "/")}`,
            method: (task.ID === undefined ? "POST" : "PUT"),
            data: $httpParamSerializer(data),
            headers: {"Content-Type": "application/x-www-form-urlencoded"}
        }).then((res) => {
            parentScope.LoadColumns()
            $uibModalInstance.close()
            $scope.saving = false
        }, () => {
            $scope.error = "Something went wrong"
            $scope.saving = false
        })
    }

    $scope.cancel = $modalInstance.dismiss
}).controller("DeleteTaskCtrl", ($scope, $uibModalInstance, $http, task, parentScope) => {
    $scope.task = task

    $scope.confirm = () => {
        $scope.deleting = true
        $http.delete(`/task/${task.ID}/`).then(
            (res) => {
                console.log(res)
                parentScope.LoadColumns()
                $uibModalInstance.close()
                $scope.deleting = false
            }, () => {
                $scope.error = "Something went wrong"
                $scope.deleting = false
            })
    }
}).controller("OrderTasksCtrl", ($scope, $uibModalInstance, $http, $timeout, column, parentScope) => {
    $scope.column = column
    $scope.tasks = column.Tasks
    $scope.parentScope = parentScope
    $scope.DndMoveToColumn = (column, index, element) => {
        parentScope.DndMoveToColumn(column, index, element)
        $timeout(() => {
            $scope.loading = true
            $http.get(`/column/${column.ID}/`)
                .then((data) => {
                    $scope.column = data.data
                    $scope.tasks = data.data.Tasks
                }, (data) => {
                    console.error("Error loading data", data)
                })
        }, 100)
    }
}).controller("EditUserCtrl", ($scope, $uibModalInstance, $http, $window, user) => {
    $scope.form = angular.copy(user)
    $scope.save = () => {
        $scope.saving = true
        let postData = `name=${encodeURIComponent($scope.form.name)}&email=${encodeURIComponent($scope.form.email)}&password=${$scope.form.password?encodeURIComponent($scope.form.password):''}`
        $http({
            method: 'POST',
            url: `/user/`,
            data: postData,
            headers: {'Content-Type': 'application/x-www-form-urlencoded'}
        }).then(() => {
            $window.location.reload()
            $uibModalInstance.close()
        }, () => {
            $scope.error = "Something went wrong"
            $scope.saving = false
        })
    }

    $scope.cancel = $uibModalInstance.dismiss
})
