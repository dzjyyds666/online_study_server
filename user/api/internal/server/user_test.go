package server

import "testing"

func TestGenerateRandomString(t *testing.T) {
	randomString := GenerateRandomString(10)
	t.Log(randomString)
}
