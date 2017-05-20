package mi_session

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base64"
	"fmt"
	"strings"
)

const (
	signatureSeparator = "."
	salt               = "mi-session"
)

type Session struct {
	SecretKey string
}

func urlSafeB64Encode(str string) []byte {
	r := strings.NewReplacer("+", "-", "/", "_")
	str = r.Replace(str)

	encoded := make([]byte, base64.RawStdEncoding.EncodedLen(len(str)))
	base64.RawStdEncoding.Encode(encoded, []byte(str))

	return encoded
}

func urlSafeB64Decode(str string) ([]byte, error) {
	r := strings.NewReplacer("-", "+", "_", "/")
	str = r.Replace(str)

	decoded, err := base64.RawStdEncoding.DecodeString(str)
	if err != nil {
		return nil, fmt.Errorf("URLSafe Base64 Decode failed, %v", err)
	}

	return decoded, nil
}

func (s *Session) getSignature(key, value string) []byte {
	h := hmac.New(sha1.New, []byte(key))
	h.Write([]byte(value))
	return h.Sum(nil)
}

func (s *Session) getDeriveKey() []byte {
	h := hmac.New(sha1.New, []byte(s.SecretKey))
	h.Write([]byte(salt))
	return h.Sum(nil)
}

func (s *Session) verifySignature(value, signature string) error {
	decodedSignature, err := urlSafeB64Decode(signature)
	if err != nil {
		return fmt.Errorf("Decode signature failed, %v", err)
	}

	key := s.getDeriveKey()

	if bytes.Compare(decodedSignature, s.getSignature(string(key), value)) != 0 {
		return fmt.Errorf("Invalid signature")
	}

	return nil
}

func (s *Session) Unsign(signedValue string) (string, error) {
	if !strings.Contains(signedValue, signatureSeparator) {
		return "", fmt.Errorf("No %s found in value", signatureSeparator)
	}

	index := strings.LastIndex(signedValue, signatureSeparator)
	value, signature := signedValue[:index], signedValue[index+1:]

	err := s.verifySignature(value, signature)
	if err != nil {
		return "", err
	}

	return value, nil
}
