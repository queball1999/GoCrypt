package encryption

import (
	"testing"
)

// TestDeriveKey tests the key derivation from a password
func TestDeriveKey(t *testing.T) {
	password := "testpassword"
	salt, err := GenerateSalt()
	if len(salt) != 16 {
		t.Fatalf("Expected salt length of 16, but got %d", len(salt))
	} else if err != nil {
		t.Fatalf("There was an error while generating salt: %v", err)
	}
	
	key := DeriveKey(password, salt)
	if key == nil {
		t.Fatalf("Failed to derive key: %v", key)
	}

	if len(key) != 32 {
		t.Fatalf("Expected key length to be 32, but got %d", len(key))
	}
}

// TestDeriveKeyError tests error handling in key derivation
func TestDeriveKeyError(t *testing.T) {
	key := DeriveKey("", nil) // Invalid password and salt
	if key == nil {
		t.Fatalf("Expected error for invalid key derivation input")
	}
}

// TestGenerateSalt tests salt generation
func TestGenerateSalt(t *testing.T) {
	salt, err := GenerateSalt()
	if len(salt) != 16 {
		t.Fatalf("Expected salt length of 16, but got %d", len(salt))
	} else if err != nil {
		t.Fatalf("There was an error while generating salt: %v", err)
	}
}