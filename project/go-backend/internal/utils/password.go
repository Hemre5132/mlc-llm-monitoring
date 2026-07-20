package utils

import "golang.org/x/crypto/bcrypt"

// HashPassword, düz metin şifreyi bcrypt ile hashler.
func HashPassword(plain string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(bytes), err
}

// CheckPassword, düz metin şifreyi kayıtlı hash ile karşılaştırır.
func CheckPassword(plain, hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
