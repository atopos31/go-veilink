package pkg

import "testing"

func TestGenKey(t *testing.T) {
	key, err := GenChacha20Key()
	if err != nil {
		t.Fatal(err)
	}
	WriteKeyToFile("testkey", "test", KeyByteToString(key))
	t.Log(KeyByteToString(key))
}