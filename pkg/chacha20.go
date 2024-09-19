package pkg

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"os"
	"strings"
	"time"

	"golang.org/x/crypto/chacha20"
)

const DefaultKeyPath = "./key"

type Chacha20Stream struct {
	key     []byte
	encoder *chacha20.Cipher
	decoder *chacha20.Cipher
	conn    VeilConn
}

func NewChacha20Stream(key []byte, conn VeilConn) (*Chacha20Stream, error) {
	s := &Chacha20Stream{
		key:  key, // should be exactly 32 bytes
		conn: conn,
	}

	var err error
	nonce := make([]byte, chacha20.NonceSizeX)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	s.encoder, err = chacha20.NewUnauthenticatedCipher(s.key, nonce)
	if err != nil {
		return nil, err
	}

	if n, err := s.conn.Write(nonce); err != nil || n != len(nonce) {
		return nil, errors.New("write nonce failed: " + err.Error())
	}
	return s, nil
}

func (s *Chacha20Stream) Read(p []byte) (int, error) {
	if s.decoder == nil {
		nonce := make([]byte, chacha20.NonceSizeX)
		if n, err := io.ReadAtLeast(s.conn, nonce, len(nonce)); err != nil || n != len(nonce) {
			return n, errors.New("can't read nonce from stream: " + err.Error())
		}
		decoder, err := chacha20.NewUnauthenticatedCipher(s.key, nonce)
		if err != nil {
			return 0, errors.New("generate decoder failed: " + err.Error())
		}
		s.decoder = decoder
	}

	n, err := s.conn.Read(p)
	if err != nil || n == 0 {
		return n, err
	}
	s.decoder.XORKeyStream(p[:n], p[:n])
	return n, nil
}

func (s *Chacha20Stream) Write(p []byte) (int, error) {
	dst := make([]byte, len(p))
	s.encoder.XORKeyStream(dst, p)
	return s.conn.Write(dst)
}

func (s *Chacha20Stream) Close() error {
	return s.conn.Close()
}

func (s *Chacha20Stream) SetWriteDeadline(t time.Time) error {
	return s.conn.SetWriteDeadline(t)
}

func GenChacha20Key() ([]byte, error) {
	key := make([]byte, chacha20.KeySize)
	if _, err := rand.Read(key); err != nil {
		return nil, err
	}
	return key, nil
}

func KeyStringToByte(key string) ([]byte,error) {
	return base64.StdEncoding.DecodeString(key)
}

func KeyByteToString(key []byte) string {
	return base64.StdEncoding.EncodeToString(key)
}

func WriteKeyToFile(keyPath string, clientID string, key string) error {
	keyFilePath := strings.Builder{}
	keyFilePath.WriteString(keyPath)
	keyFilePath.WriteRune(os.PathSeparator)
	keyFilePath.WriteString(clientID)
	keyFilePath.WriteString(".key")

	err := os.Mkdir(keyPath,os.ModeDir)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(keyFilePath.String(), os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		return err
	}

	defer file.Close()
	if _, err := file.WriteString(key); err != nil {
		return err
	}
	return nil
}
