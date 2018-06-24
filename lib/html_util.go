package fourfuse

import (
	"html"
	"regexp"
	"strings"
)

func replaceHtmlMarkup(str string) string {
	return regexp.MustCompile("</?(p|b|br|span|tr|table|td|a)(\\s+|\\s+[^>]*)?>").ReplaceAllString(str, " ")
}

func replaceHtmlMarkupConvertNewlines(str string) string {
	withAdditionalNewlines := regexp.MustCompile("</(tr|table|p)(\\s+|\\s+[^>]*)?>").ReplaceAllString(str, "\n")
	withHtmlRemoved := regexp.MustCompile("</?(p|b|span|tr|table|td|a)(\\s+|\\s+[^>]*)?>").ReplaceAllString(withAdditionalNewlines, " ")
	return strings.Replace(withHtmlRemoved, "<br>", "\n", -1)
}

func htmlToText(htmlContent string) string {
	text := html.UnescapeString(htmlContent)
	text = replaceHtmlMarkupConvertNewlines(text)
	text = replaceMultipleSpaceByOne(text)
	return text
}
