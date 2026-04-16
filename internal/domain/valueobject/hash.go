package valueobject

import "crypto/rand"

const base62Chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// NewHash generates a random Base62 string of the specified length.
func NewHash(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	for i, b := range bytes {
		bytes[i] = base62Chars[b%62]
	}

	return string(bytes), nil
}

// ValidateHash checks if a given string is a valid base62 hash and returns length validity
func ValidateHash(hash string) bool {
	if len(hash) < 5 || len(hash) > 10 {
		return false
	}
	for _, char := range hash {
		valid := false
		for _, b62Char := range base62Chars {
			if char == b62Char {
				valid = true
				break
			}
		}
		if !valid {
			return false
		}
	}
	return true
}
