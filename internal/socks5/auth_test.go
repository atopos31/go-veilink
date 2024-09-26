package socks5

import (
	"bytes"
	"reflect"
	"testing"
)

func TestAuth(t *testing.T) {
	t.Run("test auth", func(t *testing.T) {
		reader := bytes.NewReader([]byte{socks5Version, 2, NoAuth, GSSAPI})
		msg, err := NewClientAuthMessage(reader)
		if err != nil {
			t.Fatal(err)
		}

		if msg.Version != socks5Version {
			t.Fatal("version should be 0x05")
		}

		if msg.NMethod != 2 {
			t.Fatal("nmethod should be 0x01")
		}
		t.Log(msg.Methods)
		if !reflect.DeepEqual(msg.Methods, []byte{0x00, 0x01}) {
			t.Fatal("method should be 0x00")
		}
	})
}
