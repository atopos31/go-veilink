package common

import (
	"time"

	"github.com/sirupsen/logrus"
)

func InitLogrus(logLevel string) {
	level, err := logrus.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	logrus.SetLevel(level)
	logrus.SetFormatter(&logrus.TextFormatter{
		TimestampFormat: time.RFC3339,
		FullTimestamp:   true,
	})
}
