package utils

import (
	"fmt"
	"time"

	"github.com/dlshle/gommon/errors"
)

func Retry(action func() error, retryInterval time.Duration, maxRetry int) error {
	var err error
	for i := 0; i < maxRetry; i++ {
		err = withPanicRecovery(action)
		if err == nil {
			return nil
		}
		time.Sleep(retryInterval)
	}
	return err
}

func RetryWithoutWait(action func() error, maxRetry int) error {
	return Retry(action, 0, maxRetry)
}

func withPanicRecovery(cb func() error) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = errors.Error(fmt.Sprintf("panic recovered: %v", r))
		}
	}()
	return cb()
}
