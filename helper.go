package main

import (
	"fmt"
	"log"
	"net/http"
	"regexp"
	"time"

	"github.com/russross/blackfriday"
)

const (
	commonHtmlFlags = 0 |
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

	checked = "* <label ng-click=\"CheckToggle({params})\"><input type=\"checkbox\" " +
		"checked=\"checked\" />$1</label>"
	normal = "* <label ng-click=\"CheckToggle({params})\"><input type=\"checkbox\" " +
		"/>$1</label>"
)

var (
	renderer blackfriday.Renderer

	emptyCheckboxRegexp   = regexp.MustCompile(`(?m:^\s*\*\s?\[ \](.*)$)`)
	checkedCheckboxRegexp = regexp.MustCompile(`(?mi:^\s*\*\s?\[x\](.*)$)`)
	paramsRegexp          = regexp.MustCompile("({params})")
	checkboxLineRegexp    = regexp.MustCompile(`(?mi:^\s*\*\s?\[[ |x]\](.*)$)`)
	checkboxRegexp        = regexp.MustCompile(`(?i:\[[ |x]\])`)
)

func init() {
	renderer = blackfriday.HtmlRenderer(commonHtmlFlags, "", "")
}

func TimeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func TimeTrackDecorator(f http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		defer TimeTrack(
			time.Now(),
			fmt.Sprintf(
				"%4s %s",
				r.Method,
				r.RequestURI,
			),
		)
		f(w, r)
	}
}

func prepareCheckboxes(t string, id uint) (rend string) {
	rend = checkedCheckboxRegexp.ReplaceAllString(
		emptyCheckboxRegexp.ReplaceAllString(t, normal), checked,
	)

	i := 0

	return paramsRegexp.ReplaceAllStringFunc(rend, func(_ string) string {
		i++
		return fmt.Sprintf("%d, %d", id, i)
	})
}

func RenderMarkdown(text string) (rendered string) {
	return string(blackfriday.Markdown(
		[]byte(text), renderer, commonExtensions,
	))
}

func logTask(id, cId int, a string) {
	db.Save(&TaskLog{TaskID: id, OldColumnId: cId, Action: a})
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

func toggleCheckbox(t string, n int) string {
	i := 0
	return checkboxLineRegexp.ReplaceAllStringFunc(t, func(s string) string {
		i++
		if i == n {
			return toggleSingleCheckbox(s)
		}
		return s
	})
}
