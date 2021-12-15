package utils

import (
	"crypto/sha256"
	"fmt"
)

func GenerateHash(password string) string {
	return fmt.Sprintf("%v", sha256.Sum256([]byte(password)))
}

func ComparePasswordHash(pass1, pass2 string) bool {
	passFromClient := GenerateHash(pass2)

	if pass1 == "pass" {
		adminpass := GenerateHash(pass1)
		return adminpass == passFromClient
	}
	return passFromClient == pass1
}
