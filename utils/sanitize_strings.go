package utils

import (
	"html"
	"regexp"
	"sort"
	"strings"

	"github.com/kennygrant/sanitize"
	"github.com/microcosm-cc/bluemonday"
	"github.com/navidrome/navidrome/conf"
)

var quotesRegex = regexp.MustCompile("[“”‘’'\"\\[\\(\\{\\]\\)\\}]")

func SanitizeStrings(text ...string) string {
	sanitizedText := strings.Builder{}
	for _, txt := range text {
		sanitizedText.WriteString(strings.TrimSpace(sanitize.Accents(strings.ToLower(txt))) + " ")
	}
	words := make(map[string]struct{})
	for _, w := range strings.Fields(sanitizedText.String()) {
		words[w] = struct{}{}
	}
	var fullText []string
	for w := range words {
		w = quotesRegex.ReplaceAllString(w, "")
		if w != "" {
			fullText = append(fullText, w)
		}
	}
	sort.Strings(fullText)
	return strings.Join(fullText, " ")
}

func SplitAndJoinStrings(text string) string {
	return strings.Join(regexp.MustCompile(conf.Server.ArtistsSeparator).Split(text, -1), " ")
}

var policy = bluemonday.UGCPolicy()

func SanitizeText(text string) string {
	s := policy.Sanitize(text)
	return html.UnescapeString(s)
}
