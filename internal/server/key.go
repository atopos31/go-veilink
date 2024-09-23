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
		k.Kmap.Store(clientID, key)
		strKey := pkg.KeyByteToString(key)
		if err := pkg.WriteKeyToFile(pkg.DefaultKeyPath, clientID, strKey); err != nil {
			logrus.Debugf("write %s key to file failed: %s, err:%s", clientID, pkg.DefaultKeyPath,err.Error())
		}
		logrus.Debugf("write %s key to file success: %s", clientID, pkg.DefaultKeyPath)
	} else {
		key = value.([]byte)
	}
	logrus.Debugf("%s key: %s", clientID, pkg.KeyByteToString(key))
	return key, nil
}
