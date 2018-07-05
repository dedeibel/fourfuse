package fourfuse

import (
	"regexp"
	"strings"

	"bazil.org/fuse"
	fourc "github.com/moshee/go-4chan-api/api"
)

type HasDirent interface {
	GetDirent() fuse.Dirent
}

/*
 *
 * https://github.com/google/re2/wiki/Syntax
 *
 * Ll 	lowercase letter
 * Lm 	modifier letter
 * Lo 	other letter
 * Lt 	titlecase letter
 * Lu 	uppercase letter
 * N    number
 * Nd   decimal number
 *
 */
var fsNameInvalidChars *regexp.Regexp = regexp.MustCompile("[^ \\p{Ll}\\p{Lm}\\p{Lo}\\p{Lt}\\p{Lu}\\p{N}\\p{Nd}]+")

const fsNameMaxlen = 66

func sanitizePathSegment(str string) string {
	return strings.Replace(str, "/", "-", -1)
}

func sanitizedFileName(file *fourc.File) string {
	if file == nil {
		return "undefined"
	}

	return sanitizePathSegment(file.Name) + sanitizePathSegment(file.Ext)
}
