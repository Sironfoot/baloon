package baloon

import (
	"math/rand"
	"time"
)

var characters = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

func randomCharacters(n int) string {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	b := make([]rune, n)
	for i := range b {
		b[i] = characters[r.Intn(len(characters))]
	}
	return string(b)
}

func truncate(text string, maxLength int, affix string) string {
	if len(text) <= maxLength {
		return text
	}

	return text[:maxLength] + affix
}
