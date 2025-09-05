package crypt

import (
	"crypto/pbkdf2"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"hash"
)

func createKey(passphrase, salt string, iterations, keySize int, digest string) ([]byte, error) {
	var hasher func() hash.Hash
	switch digest {
	case "sha1":
		hasher = sha1.New
	case "sha256":
		hasher = sha256.New
	case "sha512":
		hasher = sha512.New
	}

	return pbkdf2.Key(hasher, passphrase, []byte(salt), iterations, keySize/8)
}
