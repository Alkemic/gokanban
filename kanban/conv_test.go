package kanban

import (
	"reflect"
	"testing"
)

func Test_prepareCheckboxes(t *testing.T) {
	type args struct {
		t  string
		id uint
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name:     "",
			args:     args{t: "*[ ] Test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test</label>",
		}, {
			name:     "",
			args:     args{t: "* [ ] Test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test</label>",
		}, {
			name:     "",
			args:     args{t: "* [ ]Test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" />Test</label>",
		}, {
			name:     "",
			args:     args{t: "* [] Test", id: 13},
			expected: "* [] Test",
		}, {
			name:     "",
			args:     args{t: "* [x] Test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" checked=\"checked\" /> Test</label>",
		}, {
			name:     "",
			args:     args{t: "*[x]Test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" checked=\"checked\" />Test</label>",
		}, {
			name:     "",
			args:     args{t: "* ~~[x]Spam~~\n* ~~[x] Ham~~", id: 13},
			expected: "* ~~<input type=\"checkbox\" disabled />Spam~~\n* ~~<input type=\"checkbox\" disabled /> Ham~~",
		}, {
			name:     "",
			args:     args{t: "*[ ] Test*[x]Test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test*[x]Test</label>",
		}, {
			name: "It should render two checkboxes, second checked for task id 13",
			args: args{t: "*[ ] Foo bar\n*[x] Field 2", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Foo bar</label>\n" +
				"* <label ng-click=\"CheckToggle(13, 2, $event)\"><input type=\"checkbox\" checked=\"checked\" /> Field 2</label>",
		}, {
			name: "Should return list with intendation",
			args: args{t: "* [ ] Test\n  * [ ] Sub-test\n      * [ ] Sub-sub-test", id: 13},
			expected: "* <label ng-click=\"CheckToggle(13, 1, $event)\"><input type=\"checkbox\" /> Test</label>\n" +
				"  * <label ng-click=\"CheckToggle(13, 2, $event)\"><input type=\"checkbox\" /> Sub-test</label>\n" +
				"      * <label ng-click=\"CheckToggle(13, 3, $event)\"><input type=\"checkbox\" /> Sub-sub-test</label>",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotRend := prepareCheckboxes(tt.args.t, tt.args.id); gotRend != tt.expected {
				t.Errorf("prepareCheckboxes() = %v, want %v", gotRend, tt.expected)
			}
		})
	}
}

func Test_calculateTaskProgress(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		expected map[string]int
	}{
		{
			name:     "return nil when no checkboxes",
			text:     "*[S] Foo bar",
			expected: map[string]int(nil),
		}, {
			name:     "properly calculate when one unchecked",
			text:     "*[ ] Foo bar",
			expected: map[string]int{"Done": 0, "ToDo": 1},
		}, {
			name:     "properly calculate when one checked",
			text:     "*[X] Foo bar",
			expected: map[string]int{"Done": 1, "ToDo": 0},
		}, {
			name:     "properly calculate when two of four checked",
			text:     "*[X] Ham\n*[ ] Foo\n  *[X] Foo tar\n  *[ ] Foo rar",
			expected: map[string]int{"Done": 2, "ToDo": 2},
		}, {
			name:     "properly ignore when one striked task",
			text:     "*[X] Ham\n*[ ] Foo\n  *[X] Foo tar\n  *[ ] Foo rar\n  * ~~[ ] Foo lol~~",
			expected: map[string]int{"Done": 2, "ToDo": 2},
		}, {
			name:     "properly ignore wrongly wrote checkboxes markdown",
			text:     "*[X] Ham*[ ]Test\n*[ ] Foo\n  *[X] Foo tar*[ ]Test2\n  *[ ] Foo rar",
			expected: map[string]int{"Done": 2, "ToDo": 2},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := calculateTaskProgress(tt.text); !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("calculateTaskProgress() = %v, want %v", got, tt.expected)
			}
		})
	}
}
