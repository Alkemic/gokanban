package kanban

import (
	"fmt"
	"regexp"

	"gokanban/markdown"
	"gokanban/model"
)

const (
	checked  = `$1* <label ng-click="CheckToggle({params}, $$event)"><input type="checkbox" checked="checked" />$2</label>`
	normal   = `$1* <label ng-click="CheckToggle({params}, $$event)"><input type="checkbox" />$2</label>`
	disabled = `$1* ~~<input type="checkbox" disabled />$2`
)

var (
	checkedCheckboxRegexp  = regexp.MustCompile(`(?mi:^(\s*)\*\s?\[x\](.*)$)`)
	emptyCheckboxRegexp    = regexp.MustCompile(`(?m:^(\s*)\*\s?\[ \](.*)$)`)
	disabledCheckboxRegexp = regexp.MustCompile(`(?mi:^(\s*)\*\s?~~\[[ x]\](.*~~(.*)?)$)`)
	paramsRegexp           = regexp.MustCompile("({params})")
)

func prepareCheckboxes(t string, id uint) (rend string) {
	rend = disabledCheckboxRegexp.ReplaceAllString(
		checkedCheckboxRegexp.ReplaceAllString(
			emptyCheckboxRegexp.ReplaceAllString(t, normal),
			checked,
		),
		disabled,
	)

	i := 0

	return paramsRegexp.ReplaceAllStringFunc(rend, func(_ string) string {
		i++
		return fmt.Sprintf("%d, %d", id, i)
	})
}

func calculateTaskProgress(t string) map[string]int {
	done, todo := 0, 0
	checkedCheckboxRegexp.ReplaceAllStringFunc(
		emptyCheckboxRegexp.ReplaceAllStringFunc(t, func(s string) string { todo++; return s }), func(s string) string { done++; return s },
	)

	if done+todo > 0 {
		result := map[string]int{}
		result["Done"] = done
		result["ToDo"] = todo
		return result
	}

	return nil
}

func taskToMap(task *model.Task) map[string]interface{} {
	return map[string]interface{}{
		"ID":          task.ID,
		"Title":       task.Title,
		"Description": task.Description,
		"Tags":        task.Tags,
		"Column":      task.Column,
		"ColumnID":    task.ColumnID,
		"Position":    task.Position,
		"Color":       task.Color,

		"CreatedAt": task.CreatedAt,
		"DeletedAt": task.DeletedAt,
		"UpdatedAt": task.UpdatedAt,

		"DescriptionRendered": markdown.RenderMarkdown(prepareCheckboxes(task.Description, task.ID)),
		"TaskProgress":        calculateTaskProgress(task.Description),
	}
}

func columnToMap(column *model.Column) map[string]interface{} {
	return map[string]interface{}{
		"ID":        column.ID,
		"Name":      column.Name,
		"Limit":     column.Limit,
		"Position":  column.Position,
		"CreatedAt": column.CreatedAt,
		"DeletedAt": column.DeletedAt,
		"UpdatedAt": column.UpdatedAt,
	}
}

func tasksToMap(tasks []*model.Task) []map[string]interface{} {
	tasksMap := make([]map[string]interface{}, 0)
	for _, task := range tasks {
		tasksMap = append(tasksMap, taskToMap(task))
	}

	return tasksMap
}

func columnsToMap(columns []*model.Column) []map[string]interface{} {
	columnsMap := make([]map[string]interface{}, 0)
	for _, column := range columns {
		columnsMap = append(columnsMap, columnToMap(column))
	}

	return columnsMap
}
