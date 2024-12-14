package common

type ConnWithClose interface {
	Close() error
}
