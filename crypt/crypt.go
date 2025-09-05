package crypt

import (
	"crypto/aes"
	"crypto/cipher"
)

type CryptOpts struct {
	IV         string
	Passphrase string
	Salt       string
	Algorithm  string `default:"AES-256-CBC"`
	Digest     string `default:"sha1"`
	KeySize    int    `default:"256"`
	Iterations int    `default:"1000"`
}

type Crypt struct {
	key        []byte
	iv         string
	algorithm  string
	digest     string
	keySize    int
	iterations int
}

func New(opts CryptOpts) *Crypt {

	key, err := createKey(
		opts.Passphrase,
		opts.Salt,
		opts.Iterations,
		opts.KeySize,
		opts.Digest,
	)

	if err != nil {
		panic(err)
	}

	return &Crypt{
		iv:         opts.IV,
		algorithm:  opts.Algorithm,
		digest:     opts.Digest,
		keySize:    opts.KeySize,
		iterations: opts.Iterations,
		key:        key,
	}
}

func (c *Crypt) Encrypt(data []byte) ([]byte, error) {

	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	// Add PKCS7 padding
	paddedData := pkcs7Pad(data, aes.BlockSize)

	blockMode := cipher.NewCBCEncrypter(block, []byte(c.iv))
	encrypted := make([]byte, len(paddedData))
	blockMode.CryptBlocks(encrypted, paddedData)
	return encrypted, nil
}

func (c *Crypt) Decrypt(data []byte) ([]byte, error) {
	block, err := aes.NewCipher(c.key)
	if err != nil {
		return nil, err
	}

	blockMode := cipher.NewCBCDecrypter(block, []byte(c.iv))
	decrypted := make([]byte, len(data))
	blockMode.CryptBlocks(decrypted, data)

	// Remove PKCS7 padding
	unpadded, err := pkcs7Unpad(decrypted)
	if err != nil {
		return nil, err
	}

	return unpadded, nil
}
