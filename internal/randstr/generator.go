// Package randstr предоставляет функции для генерации случайных строк.
// Используется для создания коротких идентификаторов URL.
package randstr

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateRandomStringURLSafe генерирует случайную строку, безопасную для использования в URL.
// Использует URL-safe base64 кодирование.
func GenerateRandomStringURLSafe(n int) (string, error) {
	b, err := GenerateRandomBytes(n)
	return base64.URLEncoding.EncodeToString(b), err
}

// GenerateRandomBytes генерирует случайные байты указанной длины.
// Использует криптографически безопасный генератор случайных чисел.
func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)

	if err != nil {
		return nil, err
	}

	return b, nil
}
