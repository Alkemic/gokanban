package helper

import (
	"fmt"
	"regexp"

	"github.com/russross/blackfriday"

	"gokanban/model"
)

const (
	commonHTMLFlags = 0 |
		blackfriday.HTML_USE_XHTML |
		blackfriday.HTML_USE_SMARTYPANTS |
		blackfriday.HTML_SMARTYPANTS_FRACTIONS |
		blackfriday.HTML_SMARTYPANTS_DASHES |
		blackfriday.HTML_SMARTYPANTS_LATEX_DASHES
	commonExtensions = 0 |
		blackfriday.EXTENSION_NO_INTRA_EMPHASIS |
		blackfriday.EXTENSION_TABLES |
		blackfriday.EXTENSION_FENCED_CODE |
		blackfriday.EXTENSION_AUTOLINK |
		blackfriday.EXTENSION_STRIKETHROUGH |
		blackfriday.EXTENSION_SPACE_HEADERS |
		blackfriday.EXTENSION_HEADER_IDS |
		blackfriday.EXTENSION_BACKSLASH_LINE_BREAK |
		blackfriday.EXTENSION_DEFINITION_LISTS |
		blackfriday.EXTENSION_HARD_LINE_BREAK

	checked  = `$1* <label ng-click="CheckToggle({params}, $$event)"><input type="checkbox" checked="checked" />$2</label>`
	normal   = `$1* <label ng-click="CheckToggle({params}, $$event)"><input type="checkbox" />$2</label>`
	disabled = `$1* ~~<input type="checkbox" disabled />$2`
)

var (
	emptyCheckboxRegexp    = regexp.MustCompile(`(?m:^(\s*)\*\s?\[ \](.*)$)`)
	checkedCheckboxRegexp  = regexp.MustCompile(`(?mi:^(\s*)\*\s?\[x\](.*)$)`)
	disabledCheckboxRegexp = regexp.MustCompile(`(?mi:^(\s*)\*\s?~~\[[ x]\](.*~~(.*)?)$)`)
	paramsRegexp           = regexp.MustCompile("({params})")
	checkboxLineRegexp     = regexp.MustCompile(`(?mi:^\s*\*\s?\[[ x]\](.*)$)`)
	checkboxRegexp         = regexp.MustCompile(`(?i:\[[ x]\])`)

	renderer = blackfriday.HtmlRenderer(commonHTMLFlags, "", "")
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

// RenderMarkdown returns rendered markdown
func RenderMarkdown(text string) (rendered string) {
	return string(blackfriday.Markdown(
		[]byte(text), renderer, commonExtensions,
	))
}

func toggleSingleCheckbox(t string) string {
	i := 0
	return checkboxRegexp.ReplaceAllStringFunc(t, func(s string) string {
		i++
		if i == 1 {
			if s == "[ ]" {
				return "[X]"
			}
			return "[ ]"
		}
		return s
	})
}

func ToggleCheckbox(t string, n int) string {
	i := 0
	return checkboxLineRegexp.ReplaceAllStringFunc(t, func(s string) string {
		i++
		if i == n {
			return toggleSingleCheckbox(s)
		}
		return s
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

func TaskToMap(task *model.Task) map[string]interface{} {
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

		"DescriptionRendered": RenderMarkdown(prepareCheckboxes(task.Description, task.ID)),
		"TaskProgress":        calculateTaskProgress(task.Description),
	}
}

func ColumnToMap(column *model.Column) map[string]interface{} {
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

func LoadTasksAsMap(tasks []*model.Task) []map[string]interface{} {
	tasksMap := make([]map[string]interface{}, 0)
	for _, task := range tasks {
		tasksMap = append(tasksMap, TaskToMap(task))
	}

	return tasksMap
}

func LoadColumnsAsMap(columns []*model.Column) []map[string]interface{} {
	columnsMap := make([]map[string]interface{}, 0)
	for _, column := range columns {
		columnsMap = append(columnsMap, ColumnToMap(column))
	}

	return columnsMap
}
