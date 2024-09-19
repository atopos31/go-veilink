package server

import (
	"sync"

	"github.com/atopos31/go-veilink/pkg"
	"github.com/sirupsen/logrus"
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
	var err error
	value, ok := k.Kmap.Load(clientID)
	if !ok {
		key, err = pkg.GenChacha20Key()
		if err != nil {
			return nil, err
		}
		strKey := pkg.KeyByteToString(key)
		pkg.WriteKeyToFile(pkg.DefaultKeyPath, clientID, strKey)
		logrus.Debugf("write %s key to file success: %s",clientID, pkg.DefaultKeyPath)
	}
	key = value.([]byte)
	logrus.Debugf("%s key: %s", clientID, key)
	return key, nil
}
