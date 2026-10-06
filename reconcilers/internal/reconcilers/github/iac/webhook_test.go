package iac

import (
	"encoding/hex"
	"testing"
)

func TestGenerateSignature(t *testing.T) {
	// https://docs.github.com/en/enterprise-cloud@latest/webhooks/using-webhooks/validating-webhook-deliveries
	expected := "757107ea0eb2509fc211221cce984b8a37570b6d7586c22c46f4379c8b043e17"
	resultByte := generateSignature([]byte("It's a Secret to Everybody"), []byte("Hello, World!"))
	result := hex.EncodeToString(resultByte)

	if result != expected {
		t.Errorf("expected %s, but got %s", expected, result)
	}
}
