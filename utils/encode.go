package utils

import (
	"crypto/sha256"
	"fmt"
)

func GenerateHash(password string) string {

	return fmt.Sprintf("%v", sha256.Sum256([]byte(password)))
}

func ComparePasswordHash(pass1, pass2 string) bool {
	passFromDB := GenerateHash(pass1)
	passFromClient := GenerateHash(pass2)
	return passFromClient == passFromDB
}
