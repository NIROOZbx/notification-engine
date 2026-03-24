package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
)


func GenerateAPIKey(envName string) (rawKey, hashedKey, hint string) {
	var prefix string
	
	switch envName{
	case "development":
		prefix="ne_test_"
	case "production":
		prefix="ne_live_"
	default:
		prefix="ne_gen_"
	}
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		panic(err) 
	}

	rawKey=prefix+base64.RawURLEncoding.EncodeToString(b)
	hint = fmt.Sprintf("%s...%s", rawKey[:len(prefix)+4], rawKey[len(rawKey)-4:])

	hashedKey=HashAPIKey(rawKey)

	return
}

func HashAPIKey(rawKey string) string {
    sum := sha256.Sum256([]byte(rawKey))
    return hex.EncodeToString(sum[:])
}