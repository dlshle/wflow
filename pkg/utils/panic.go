package utils

import (
	"fmt"
	"runtime"
	"strconv"

	"github.com/dlshle/gommon/errors"
)

func HandlePanic(existingErr error) error {
	if r := recover(); r != nil {
		return errors.Error(fmt.Sprintf("panic: %v \n call stack trace:\n %s", r, getCallStackTrace(10)))
	}
	return existingErr
}

func getCallStackTrace(maxSkip int) string {
	callTraceStr := ""
	for i := 1; i < maxSkip; i++ {
		_, file, line, ok := runtime.Caller(i)
		if ok {
			callTraceStr += (file + ":" + strconv.Itoa(line) + "\n")
		}
	}
	return callTraceStr
}
