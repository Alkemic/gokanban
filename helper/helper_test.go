package helper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrepareCheckboxes(t *testing.T) {
	assert.Equal(
		t, prepareCheckboxes("*[ ] Test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test</label>",
	)
	assert.Equal(
		t, prepareCheckboxes("* [ ] Test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test</label>",
	)
	assert.Equal(
		t, prepareCheckboxes("* [ ]Test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" />Test</label>",
	)
	assert.Equal(
		t, prepareCheckboxes("* [] Test", 13),
		"* [] Test",
	)

	assert.Equal(
		t, prepareCheckboxes("* [x] Test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" checked=\"checked\" /> Test</label>",
	)
	assert.Equal(
		t, prepareCheckboxes("*[x]Test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" checked=\"checked\" />Test</label>",
	)

	assert.Equal(
		t, prepareCheckboxes("* ~~[x]Spam~~\n* ~~[x] Ham~~", 13),
		"* ~~<input type=\"checkbox\" disabled />Spam~~\n* ~~<input type=\"checkbox\" disabled /> Ham~~",
	)

	assert.Equal(
		t, prepareCheckboxes("*[ ] Test*[x]Test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test*[x]Test</label>",
	)

	assert.Equal(
		t, prepareCheckboxes("*[ ] Foo bar\n*[x] Field 2", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Foo bar</label>\n"+
			"* <label ng-click=\"CheckToggle(13, 2, $event)\"><input type=\"checkbox\" checked=\"checked\" /> Field 2</label>",
		"It should render two checkboxes, second checked for task id 13",
	)

	assert.Equal(
		t, prepareCheckboxes("* [ ] Test\n  * [ ] Sub-test\n      * [ ] Sub-sub-test", 13),
		"* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test</label>\n"+
			"  * <label ng-click=\"CheckToggle(13, 2, $event)\"><input type=\"checkbox\" /> Sub-test</label>\n"+
			"      * <label ng-click=\"CheckToggle(13, 3, $event)\"><input type=\"checkbox\" /> Sub-sub-test</label>",
		"Should return list with intendation",
	)
}

func TestRenderMarkdown(t *testing.T) {
	assert.Equal(
		t, RenderMarkdown("*List*\n\n* Foo\n* Bar"),
		"<p><em>List</em></p>\n\n"+
			"<ul>\n"+
			"<li>Foo<br /></li>\n"+
			"<li>Bar<br /></li>\n"+
			"</ul>\n",
		"It should render markdown",
	)
}

func TestToggleCheckbox(t *testing.T) {
	assert.Equal(
		t, ToggleCheckbox("*[ ] Foo bar\n*[x] Field 2", 1),
		"*[X] Foo bar\n*[x] Field 2",
	)
	assert.Equal(
		t, ToggleCheckbox("*[ ] Foo bar\n*[X] Field 2", 2),
		"*[ ] Foo bar\n*[ ] Field 2",
	)

	assert.Equal(
		t, ToggleCheckbox("*[S] Foo bar\n*[x] Field 2", 1),
		"*[S] Foo bar\n*[ ] Field 2",
		"It should ignore improperlly setted up marked checkbox",
	)

	assert.Equal(
		t, ToggleCheckbox("*[ ] Foo bar\n*[x] Field 2", 3),
		"*[ ] Foo bar\n*[x] Field 2",
		"It should quietly ignore out of range request",
	)

	assert.Equal(
		t, ToggleCheckbox("*[ ] Foo bar\n*[x] Field 2", 3),
		"*[ ] Foo bar\n*[x] Field 2",
		"It should quietly ignore out of range request",
	)
}

func TestToggleSingleCheckbox(t *testing.T) {
	assert.Equal(
		t, toggleSingleCheckbox("*[ ] Foo bar"),
		"*[X] Foo bar",
	)
	assert.Equal(
		t, toggleSingleCheckbox("*[X] Foo bar"),
		"*[ ] Foo bar",
	)
	assert.Equal(
		t, toggleSingleCheckbox("*[x] Foo bar"),
		"*[ ] Foo bar",
	)

	assert.Equal(
		t, toggleSingleCheckbox("* [x] Foo bar"),
		"* [ ] Foo bar",
	)
	assert.Equal(
		t, toggleSingleCheckbox("*  [x] Foo bar"),
		"*  [ ] Foo bar",
	)

	assert.Equal(
		t, toggleSingleCheckbox("*[ ] [ ] Foo bar"),
		"*[X] [ ] Foo bar",
		"It should only change first occurence",
	)
}

func TestCalculateTaskProgress(t *testing.T) {
	assert.Equal(
		t, calculateTaskProgress("*[S] Foo bar"),
		map[string]int(nil),
	)

	assert.Equal(
		t, calculateTaskProgress("*[ ] Foo bar"),
		map[string]int{"Done": 0, "ToDo": 1},
	)
	assert.Equal(
		t, calculateTaskProgress("*[X] Foo bar"),
		map[string]int{"Done": 1, "ToDo": 0},
	)
	assert.Equal(
		t, calculateTaskProgress("*[X] Ham\n*[ ] Foo\n  *[X] Foo tar\n  *[ ] Foo rar"),
		map[string]int{"Done": 2, "ToDo": 2},
	)
	assert.Equal(
		t, calculateTaskProgress("*[X] Ham\n*[ ] Foo\n  *[X] Foo tar\n  *[ ] Foo rar\n  * ~~[ ] Foo lol~~"),
		map[string]int{"Done": 2, "ToDo": 2},
	)
	assert.Equal(
		t, calculateTaskProgress("*[X] Ham*[ ]Test\n*[ ] Foo\n  *[X] Foo tar*[ ]Test2\n  *[ ] Foo rar"),
		map[string]int{"Done": 2, "ToDo": 2},
	)
}
