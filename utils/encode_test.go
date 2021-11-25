package utils

import "testing"

func TestGenerateHash(t *testing.T) {
	pass := "Asd123@"
	hash := GenerateHash(pass)
	if hash != "" {
		t.Errorf("want: new hash, got: %v", hash)
	}
}

func TestComparePasswordHash(t *testing.T) {
	p :=  "Asd123@"
	pp :=  "Asd123@1"

	res := ComparePasswordHash(p, pp)

	if !res {
		t.Errorf("want: true, got: %v", res)
	} 
}