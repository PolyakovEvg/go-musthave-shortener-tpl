package randstr

import (
	"testing"
)

func TestGenerateRandomStringURLSafe(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{
			name: "generate 8 bytes",
			n:    8,
		},
		{
			name: "generate 16 bytes",
			n:    16,
		},
		{
			name: "generate 32 bytes",
			n:    32,
		},
		{
			name: "generate 0 bytes",
			n:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomStringURLSafe(tt.n)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == "" && tt.n > 0 {
				t.Error("expected non-empty result")
			}
		})
	}
}

func TestGenerateRandomStringURLSafe_Uniqueness(t *testing.T) {
	results := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result, err := GenerateRandomStringURLSafe(16)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if results[result] {
			t.Errorf("duplicate random string generated: %s", result)
		}
		results[result] = true
	}

	if len(results) != 100 {
		t.Errorf("expected 100 unique strings, got %d", len(results))
	}
}

func TestGenerateRandomBytes(t *testing.T) {
	tests := []struct {
		name string
		n    int
	}{
		{
			name: "generate 8 bytes",
			n:    8,
		},
		{
			name: "generate 16 bytes",
			n:    16,
		},
		{
			name: "generate 32 bytes",
			n:    32,
		},
		{
			name: "generate 0 bytes",
			n:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GenerateRandomBytes(tt.n)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if len(result) != tt.n {
				t.Errorf("expected %d bytes, got %d", tt.n, len(result))
			}
		})
	}
}

func TestGenerateRandomBytes_Uniqueness(t *testing.T) {
	results := make(map[string]bool)

	for i := 0; i < 100; i++ {
		result, err := GenerateRandomBytes(16)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		key := string(result)
		if results[key] {
			t.Errorf("duplicate random bytes generated")
		}
		results[key] = true
	}

	if len(results) != 100 {
		t.Errorf("expected 100 unique byte slices, got %d", len(results))
	}
}
