package common

import "testing"

func TestGenKey(t *testing.T) {
	key, err := GenChacha20Key()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(KeyByteToString(key))
}
