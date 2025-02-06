package common

import (
	"io"
	"sync"

	"github.com/sirupsen/logrus"
)

// Join two connections together and return the number of bytes transferred
func Join(c1 io.ReadWriteCloser, c2 io.ReadWriteCloser) (inCount int64, outCount int64) {
	var wait sync.WaitGroup
	pipe := func(to io.ReadWriteCloser, from io.ReadWriteCloser, count *int64) {
		defer wait.Done()

		var err error
		*count, err = io.Copy(to, from)
		if err != nil {
			logrus.Errorf("Join conns error: %v", err)
		}
	}

	wait.Add(2)
	go pipe(c1, c2, &inCount)
	go pipe(c2, c1, &outCount)
	wait.Wait()
	return
}
