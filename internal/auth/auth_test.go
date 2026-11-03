package auth

import "testing"

func TestHashPassword(t *testing.T) {
	hash, err := HashPassword("secret")
	if err != nil {
		t.Fatal(err)
	}
	if !CheckPassword(hash, "secret") {
		t.Error("правильный пароль не прошёл проверку")
	}
	if CheckPassword(hash, "wrong") {
		t.Error("неправильный пароль прошёл проверку")
	}
}
