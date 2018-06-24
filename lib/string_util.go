package fourfuse

import (
	"regexp"
	"strings"
)

func trimWhitespace(str string) string {
	return strings.Trim(str, " \t\n")
}

func isEmptyString(str string) bool {
	return len(trimWhitespace(str)) == 0
}

func truncateString(str string, maxlen int) string {
	if len(str) > maxlen {
		return str[0:maxlen]
	} else {
		return str
	}
}

func replaceMultipleSpaceByOne(str string) string {
	return regexp.MustCompile(" {2,}").ReplaceAllString(str, " ")
}
