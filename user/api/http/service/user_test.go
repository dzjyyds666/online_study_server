package userHttpService

import (
	"github.com/dzjyyds666/opensource/sdk"
	"testing"
)

func TestGenerateRandomString(t *testing.T) {
	randomString := GenerateRandomString(10)
	t.Log(randomString)
}

func TestParseToken(t *testing.T) {
	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJkYXRhIjoiZXlKMWFXUWlPaUpNV0Y4NWVEVjZSRGhpZDNCRUlpd2ljbTlzWlNJNk1YMD0ifQ.zj8tOSQwwbA5atX2hYh-1iZd4xkmpQ1MRk3H605GVjQ"
	jwtToken, err := sdk.ParseJwtToken("aaron519", token)
	if err != nil {
		t.Log(err)
	}

	t.Log(jwtToken)
}

func TestNewUser(t *testing.T) {

}
