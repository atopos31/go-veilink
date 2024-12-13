package server

import (
	"errors"
	"sync"
)

type keymap struct {
	Kmap sync.Map
}

func NewKeyMap() *keymap {
	return &keymap{
		Kmap: sync.Map{},
	}
}

func (k *keymap) Get(clientID string) ([]byte, error) {
	var key []byte
	value, ok := k.Kmap.Load(clientID)
	if !ok {
		return nil, errors.New("key not found")
	} else {
		key = value.([]byte)
	}

	return key, nil
}

func (k *keymap) Set(clientID string, key []byte) {
	k.Kmap.Store(clientID, key)
}
