package core

import (
	"crypto/rand"
	"github.com/google/uuid"
)

func GenerateRandomString(length int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return ""
	}
	for i, b := range bytes {
		bytes[i] = letters[b%byte(len(letters))]
	}
	return string(bytes)
}

func NewClassId(length int) string {
	return "ci_" + GenerateRandomString(length)
}

func NewTaskId(length int) string {
	return "ti_" + GenerateRandomString(length)
}

func NewChapterId(length int) string {
	return "ch_" + GenerateRandomString(length)
}

func NewFid() string {
	u := uuid.New()
	return "fi_" + u.String()
}
