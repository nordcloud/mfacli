package codec

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"fmt"
)

var (
	ErrInvalidPassword = fmt.Errorf("Invalid password")
)

func BuildEncKey(password string) []byte {
	sum := sha256.Sum256([]byte(password))
	result := make([]byte, sha256.Size)
	for i := 0; i < sha256.Size; i++ {
		result[i] = sum[i]
	}
	return result
}

func Decrypt(encrypted []byte, key []byte) (map[string]string, error) {
	if len(encrypted) <= aes.BlockSize {
		return nil, fmt.Errorf("Ciphertext is too short")
	}

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext, iv := encrypted[aes.BlockSize:], encrypted[:aes.BlockSize]
	decrypted := make([]byte, len(ciphertext))
	decrypter := cipher.NewCBCDecrypter(c, iv)
	decrypter.CryptBlocks(decrypted, ciphertext)

	decryptedKey, decrypted := decrypted[:sha256.Size], decrypted[sha256.Size:]
	if bytes.Compare(decryptedKey, key) != 0 {
		return nil, ErrInvalidPassword
	}

	decrypted = unpad(decrypted)

	var secrets map[string]string
	err = json.Unmarshal(decrypted, &secrets)
	if err != nil {
		return nil, err
	}
	return secrets, nil
}

func Encrypt(secrets map[string]string, key []byte) ([]byte, error) {
	iv := make([]byte, aes.BlockSize)
	_, err := rand.Read(iv)
	if err != nil {
		return nil, err
	}

	src, err := json.Marshal(secrets)
	if err != nil {
		return nil, err
	}
	src = pad(src)
	src = append(key, src...)

	c, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	ciphertext := make([]byte, len(src))
	encrypter := cipher.NewCBCEncrypter(c, iv)
	encrypter.CryptBlocks(ciphertext, src)

	return append(iv, ciphertext...), nil
}

func pad(src []byte) []byte {
	length := len(src)
	padValue := aes.BlockSize - length%aes.BlockSize

	dst := make([]byte, length+padValue)
	for i := 0; i < length; i++ {
		dst[i] = src[i]
	}

	for i := 0; i < padValue; i++ {
		dst[length+i] = byte(padValue)
	}

	return dst
}

func unpad(src []byte) []byte {
	length := len(src)
	return src[:length-int(src[length-1])]
}
