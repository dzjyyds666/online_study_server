package core

import "crypto/rand"

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

func newPlateId(length int) string {
	return "PI_" + GenerateRandomString(length)
}

func newArticleId(length int) string {
	return "AI_" + GenerateRandomString(length)
}

func newCommentId(length int) string {
	return "CMI_" + GenerateRandomString(length)
}
