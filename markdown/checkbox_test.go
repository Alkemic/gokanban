package markdown

import (
	"testing"
)

func TestToggleCheckbox(t *testing.T) {
	type args struct {
		t string
		n int
	}
	tests := []struct {
		name     string
		args     args
		expected string
	}{
		{
			name:     "should check first checkbox",
			args:     args{t: "*[ ] Foo bar\n*[x] Field 2", n: 1},
			expected: "*[X] Foo bar\n*[x] Field 2",
		}, {
			name:     "should uncheck second checkbox",
			args:     args{t: "*[ ] Foo bar\n*[X] Field 2", n: 2},
			expected: "*[ ] Foo bar\n*[ ] Field 2",
		}, {
			name:     "should uncheck second checkbox with small x",
			args:     args{t: "*[ ] Foo bar\n*[x] Field 2", n: 2},
			expected: "*[ ] Foo bar\n*[ ] Field 2",
		}, {
			name:     "should check first occurrence of square brackets",
			args:     args{t: "*[ ] [ ] Foo bar\n*[ ] Field 2", n: 1},
			expected: "*[X] [ ] Foo bar\n*[ ] Field 2",
		}, {
			name:     "should check first checkbox ignoring [S]",
			args:     args{t: "*[S] Foo bar\n*[x] Field 2", n: 1},
			expected: "*[S] Foo bar\n*[ ] Field 2",
		}, {
			name:     "should silently ignore out of range request",
			args:     args{t: "*[ ] Foo bar\n*[x] Field 2", n: 3},
			expected: "*[ ] Foo bar\n*[x] Field 2",
		}, {
			name:     "should silently ignore out of range request",
			args:     args{t: "*[ ] Foo bar\n*[x] Field 2", n: 3},
			expected: "*[ ] Foo bar\n*[x] Field 2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToggleCheckbox(tt.args.t, tt.args.n); got != tt.expected {
				t.Errorf("ToggleCheckbox() = %v, want %v", got, tt.expected)
			}
		})
	}
}
