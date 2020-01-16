package helpers

import (
	"math/rand"
	"time"
)

// The following is ispired by https://stackoverflow.com/a/31832326
func init() {
	rand.Seed(time.Now().UnixNano())
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

// RandomString generates an alphabetic string with uppercase and lowercase
// letters of size n with some chance of being unique
func RandomString(n int) string {
	return RandomStringRunes(n, letterRunes)
}

// RandomString generates an alphabetic string with uppercase and lowercase
// letters of size n with some chance of being unique
func RandomStringRunes(n int, runes []rune) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = runes[rand.Intn(len(runes))]
	}
	return string(b)
}
