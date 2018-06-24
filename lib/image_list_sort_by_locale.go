package fourfuse

// #include <string.h>
// #include <locale.h>
import "C"

type FilenameByLocale []*listEntry

func (s FilenameByLocale) Len() int {
	return len(s)
}

func (s FilenameByLocale) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s FilenameByLocale) Less(i, j int) bool {
	return C.strcoll(
		C.CString(s[i].Slug()),
		C.CString(s[j].Slug())) < 0
}
