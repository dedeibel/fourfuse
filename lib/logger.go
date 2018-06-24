package fourfuse

import (
	"log"
	"os"
)

const traceEnabled = false
const debugEnabled = false

var (
	Log *log.Logger
)

func InitializeLogger() {
	Log = log.New(os.Stderr, "fourfuse: ", log.LstdFlags)
}

func LogTrace(v ...interface{}) {
	if traceEnabled {
		Log.Print(v...)
	}
}

func LogTracef(format string, v ...interface{}) {
	if traceEnabled {
		Log.Printf(format, v...)
	}
}

func LogDebug(v ...interface{}) {
	if debugEnabled {
		Log.Print(v...)
	}
}

func LogDebugf(format string, v ...interface{}) {
	if debugEnabled {
		Log.Printf(format, v...)
	}
}

func LogInfo(v ...interface{}) {
	Log.Print(v...)
}

func LogInfof(format string, v ...interface{}) {
	Log.Printf(format, v...)
}

func LogError(v ...interface{}) {
	Log.Print(v...)
}

func LogErrorf(format string, v ...interface{}) {
	Log.Printf(format, v...)
}
