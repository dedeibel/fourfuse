package fourfuse

// #include <string.h>
// #include <locale.h>
import "C"

func UseSystemLocale() {
	C.setlocale(C.LC_ALL, C.CString(""))
}
