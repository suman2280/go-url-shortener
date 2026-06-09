package shortener

import (
	"strings"
	"testing"
)

func TestGenerateCode_Length(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(code) != codeLength {
		t.Errorf("expected length %d, got %d", codeLength, len(code))
	}
}

func TestGenerateCode_Charset(t *testing.T) {
	code, err := GenerateCode()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, c := range code {
		if !strings.ContainsRune(alphabet, c) {
			t.Errorf("character %c not in allowed charset", c)
		}
	}
}

func TestGenerateCode_Unique(t *testing.T) {
	seen := make(map[string]bool)
	for i := 0; i < 100; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if seen[code] {
			t.Errorf("collision detected: %s", code)
		}
		seen[code] = true
	}
}

func TestGenerateCode_DeterministicRandomness(t *testing.T) {
	code1, _ := GenerateCode()
	code2, _ := GenerateCode()
	if code1 == code2 {
		t.Log("codes matched (extremely unlikely but possible)")
	}
}

func TestErrMaxRetries_Error(t *testing.T) {
	err := ErrMaxRetries{}
	if err.Error() == "" {
		t.Error("expected non-empty error message")
	}
}
