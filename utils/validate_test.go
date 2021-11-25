package utils

import (
	"fmt"
	"testing"
)

func TestCheckIIN(t *testing.T) {
	// iin := "940217450216"
	iin := "990824351277"

	res := checkIIN(iin)
	if res != nil {
		t.Errorf("want: nil, got: %v", res)
	}
}

func TestCheckTwelve(t *testing.T) {
	iin := []int{9, 4, 0, 2, 1, 7, 4, 5, 0, 2, 1, 6}
	res := checkTwelve(iin, 1)
	if res != 6 {
		t.Errorf("want: 6, got: %v", res)
	}
}

func TestUsername(t *testing.T) {
	names := []string{"Albina", "medi", "hahaha1", "qwe!@#"}
	results := []error{fmt.Errorf("username must contain only lowercase letters and digits"), nil, nil, fmt.Errorf("username must contain only lowercase letters and digits")}

	for i := 0; i < 4; i++ {
		res := checkUsername(names[i])
		if res != results[i] {
			t.Errorf("want: %v, got: %v", results[i], res)
		}
	}

}

func TestCheckPassword(t *testing.T) {
	password := []string{"asdasad", "Asd123@", "Qwe123!@!", "123456", "1234"}
	result := []error{fmt.Errorf("password must contain at least 1 digit, 1 uppercase and 1 lowercase letter"),
		nil, nil, fmt.Errorf("password must contain at least 1 digit, 1 uppercase and 1 lowercase letter"),
		fmt.Errorf("password must be at least 6 characters in length")}
	for i := 0; i < 5; i++ {
		res := checkPassword(password[i])
		if res != result[i] {
			t.Errorf("want: %v, got: %v", result[i], res)
		}
	}

}

func TestIsSpecialChar(t *testing.T) {

	// word := "Asd123@"
	word := "Qwe123!@!"
	for i, w := range word {
		res := isSpecialChar(w)
		if !res {
			t.Errorf("want: false %v, got: %v", i, res)
		}
	}

}

func TestIsAlpha(t *testing.T) {

	// word := "Asd123@"
	word := "Qwe123!@!"
	for i, w := range word {
		res := isAlpha(w)
		if !res {
			t.Errorf("want: false %v, got: %v", i, res)
		}
	}
}
