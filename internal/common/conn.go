package common

import "time"

type VeilConn interface {
	Read(p []byte) (int, error)
	Write(p []byte) (int, error)
	Close() error
	SetWriteDeadline(t time.Time) error
}

type ConnWithClose interface {
	Close() error
}
