// Package hasher — bcrypt cho mật khẩu (KHÁC repo mẫu learn_go dùng md5 — design §1).
package hasher

import "golang.org/x/crypto/bcrypt"

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	return string(b), err
}

func ComparePassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}
