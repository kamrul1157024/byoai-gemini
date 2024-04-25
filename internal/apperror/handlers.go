package apperror

import (
	"github.com/kamrul1157024/byoai-gemini/internal/loggers"
)

func CheckAndPanic(e error) {
	if e != nil {
		panic(e)
	}
}

func CheckAndLog(e error, log string) {
	if e != nil {
		loggers.AppLogger.Error(log, e)
	}
}
