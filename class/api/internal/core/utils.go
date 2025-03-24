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
	return "ci_" + GenerateRandomString(8)
}

func NewChapterId(length int) string {
	return "ch_" + GenerateRandomString(8)
}

func NewStudyClass(length int) string {
	return "sc_" + GenerateRandomString(8)
}

func NewFid() string {
	u := uuid.New()
	return "fi_" + u.String()
}
