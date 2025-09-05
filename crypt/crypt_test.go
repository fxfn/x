package crypt

import (
	"fmt"
	"testing"
)

func TestEncrypt(t *testing.T) {
	crypt := New(CryptOpts{
		Passphrase: "password",
		Salt:       "salt",
		IV:         "1234567890123456", // 16 bytes for AES
		Algorithm:  "AES-256-CBC",
		Digest:     "sha1",
		KeySize:    256,
		Iterations: 1000,
	})

	data := []byte("hello, world")
	encrypted, err := crypt.Encrypt(data)
	fmt.Printf("%v\n", encrypted)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	if len(encrypted) == 0 {
		t.Fatalf("encrypted data is empty")
	}

	// Verify encrypted data is different from original
	if string(encrypted) == string(data) {
		t.Fatalf("encrypted data should be different from original")
	}

	// Verify encrypted data has expected length (original + padding, rounded to block size)
	expectedMinLength := len(data)
	if len(encrypted) < expectedMinLength {
		t.Fatalf("encrypted data length %d is less than expected minimum %d", len(encrypted), expectedMinLength)
	}
}

func TestDecrypt(t *testing.T) {
	crypt := New(CryptOpts{
		Passphrase: "password",
		Salt:       "salt",
		IV:         "1234567890123456", // 16 bytes for AES
		Algorithm:  "AES-256-CBC",
		Digest:     "sha1",
		KeySize:    256,
		Iterations: 1000,
	})

	encrypted := []byte{226, 5, 59, 105, 68, 252, 157, 134, 127, 213, 111, 183, 20, 175, 23, 172}

	decrypted, err := crypt.Decrypt(encrypted)
	fmt.Printf("%s\n", string(decrypted))
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	crypt := New(CryptOpts{
		Passphrase: "password",
		Salt:       "salt",
		IV:         "1234567890123456", // 16 bytes for AES
		Algorithm:  "AES-256-CBC",
		Digest:     "sha1",
		KeySize:    256,
		Iterations: 1000,
	})

	data := []byte("hello, world")
	encrypted, err := crypt.Encrypt(data)
	if err != nil {
		t.Fatalf("failed to encrypt: %v", err)
	}

	decrypted, err := crypt.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("failed to decrypt: %v", err)
	}

	if string(decrypted) != string(data) {
		t.Fatalf("decrypted data should be the same as original")
	}
}
