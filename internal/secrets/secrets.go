package secrets

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
)

// EncryptJSON encrypts a JSON-serializable value. If key is empty, stores plain base64 JSON (dev only).
func EncryptJSON(key string, v interface{}) (string, error) {
	raw, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	if key == "" {
		return "plain:" + base64.StdEncoding.EncodeToString(raw), nil
	}
	block, err := aes.NewCipher(deriveKey(key))
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, raw, nil)
	return "enc:" + base64.StdEncoding.EncodeToString(ciphertext), nil
}

// DecryptJSON decrypts into dest.
func DecryptJSON(key string, blob string, dest interface{}) error {
	if len(blob) > 5 && blob[:5] == "plain:" {
		raw, err := base64.StdEncoding.DecodeString(blob[5:])
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, dest)
	}
	if len(blob) > 4 && blob[:4] == "enc:" {
		if key == "" {
			return errors.New("BROKER_ENCRYPTION_KEY required to decrypt credentials")
		}
		data, err := base64.StdEncoding.DecodeString(blob[4:])
		if err != nil {
			return err
		}
		block, err := aes.NewCipher(deriveKey(key))
		if err != nil {
			return err
		}
		gcm, err := cipher.NewGCM(block)
		if err != nil {
			return err
		}
		nonceSize := gcm.NonceSize()
		if len(data) < nonceSize {
			return fmt.Errorf("ciphertext too short")
		}
		nonce, ciphertext := data[:nonceSize], data[nonceSize:]
		raw, err := gcm.Open(nil, nonce, ciphertext, nil)
		if err != nil {
			return err
		}
		return json.Unmarshal(raw, dest)
	}
	return fmt.Errorf("unknown credentials format")
}

func deriveKey(key string) []byte {
	sum := sha256.Sum256([]byte(key))
	return sum[:]
}
