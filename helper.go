package main

import (
	"fmt"
	"log"
	"net/http"
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
)

var renderer blackfriday.Renderer

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

func RenderMarkdown(text string) string {
	return string(blackfriday.Markdown(
		[]byte(text), renderer, commonExtensions,
	))
}
