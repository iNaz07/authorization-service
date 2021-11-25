package utils

import (
	"fmt"
	"strconv"
)

func ValidateCreds(username, password, iin string) error {
	if err := checkUsername(username); err != nil {
		return err
	}
	if err := checkPassword(password); err != nil {
		return err
	}
	if err := checkIIN(iin); err != nil {
		return err
	}
	return nil
}

func checkUsername(name string) error {
	for _, letter := range name {
		if !isNumeric(letter) && !isAlpha(letter) {
			return fmt.Errorf("username must contain only lowercase letters and digits")
		}
	}
	return nil
}

func checkPassword(pass string) error {
	var countDigit, countLower, countUpper int
	if len(pass) < 6 {
		return fmt.Errorf("password must be at least 6 characters in length")
	}

	for _, letter := range pass {
		if isNumeric(letter) {
			countDigit++
			continue
		}
		if isAlpha(letter) {
			countLower++
			continue
		}
		if letter >= 65 && letter <= 90 {
			countUpper++
			continue
		}
		if !isSpecialChar(letter) {
			return fmt.Errorf("password must contain at least 1 digit, 1 uppercase and 1 lowercase letter and special characters")
		}
	}
	if countDigit < 1 || countLower < 1 || countUpper < 1 {
		return fmt.Errorf("password must contain at least 1 digit, 1 uppercase and 1 lowercase letter")
	}
	return nil
}

func checkIIN(iin string) error {
	if len(iin) != 12 {
		return fmt.Errorf("invalid IIN: length is not 12")
	}
	iinDigits, err := atoi(iin)

	if iinDigits[6] < 3 || iinDigits[6] > 6 {
		return fmt.Errorf("invalid IIN: 7 digit incorrect")
	}

	if err != nil {
		return fmt.Errorf("convert IIN error")
	}
	res := checkTwelve(iinDigits, 1)
	if res == 10 {
		res = checkTwelve(iinDigits, 3)
	}
	if iinDigits[11] != res {
		return fmt.Errorf("invalid IIN: 12 digit incorrect")
	}
	return nil
}

func isNumeric(letter rune) bool {
	return letter >= 48 && letter <= 57
}

func isAlpha(letter rune) bool {
	return letter >= 97 && letter <= 122
}

func isSpecialChar(letter rune) bool {
	special := "~!@#$%^&*_-+=`|\\(){}[]:;\"'<>,.?/"
	for _, char := range special {
		if char == letter {
			return true
		}
	}
	return false
}

func atoi(iin string) ([]int, error) {
	res := []int{}
	for _, i := range iin {
		digit, err := strconv.Atoi(string(i))
		if err != nil {
			return nil, err
		}
		res = append(res, digit)
	}
	return res, nil
}

func checkTwelve(iin []int, start int) int {
	res := 0
	for i := 0; i < 10; i++ {
		res += start * int(iin[i])
		start++
	}

	return res % 11
}
