// Package w2pcrypto provides easy-to-use functions for creating RSA key pairs,
// using them to sign messages and verify signatures
package w2pcrypto

import (
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/pem"
	"io/ioutil"
	"log"
	"os"
)

const RsaKeyBits int = 1024
const KeyFolder string = "./keys/"

// -----------
// - Structs -
// -----------

type PublicKey struct {
	*rsa.PublicKey
}
type PrivateKey struct {
	*rsa.PrivateKey
}

// -----------
// - Methods -
// -----------

// CheckError prints the error and exits if err is not nil
func CheckError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

// CreateKey is a wrapper for the rsa.GenerateKey function.
// It returns a *rsa.PrivateKey if no error is encountered.
func CreateKey() (*PrivateKey, *PublicKey) {
	rng := rand.Reader

	key, err := rsa.GenerateKey(rng, RsaKeyBits)
	CheckError(err)

	privkey := PrivateKey{key}
	pubkey := PublicKey{&key.PublicKey}
	return &privkey, &pubkey
}

// LoadPrivateKey loads the PEM encoded private key with the given filename
// It returns a *PrivateKey
func LoadPrivateKey(fileName string) *PrivateKey {
	k, err := ioutil.ReadFile(KeyFolder + fileName)
	CheckError(err)

	block, _ := pem.Decode(k)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("Failed to decode PEM block containing private key")
	}

	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	CheckError(err)

	return &PrivateKey{key}
}

// LoadPublicKey loads the PEM encoded public key with the given filename
// It returns a *PublicKey
func LoadPublicKey(fileName string) *PublicKey {
	k, err := ioutil.ReadFile(KeyFolder + fileName)
	CheckError(err)

	block, _ := pem.Decode(k)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		log.Fatal("Failed to decode PEM block containing public key")
	}

	general_key, err := x509.ParsePKIXPublicKey(block.Bytes)
	CheckError(err)
	key, ok := general_key.(*rsa.PublicKey)
	if !ok {
		log.Fatal("Failed to load public key")
	}
	return &PublicKey{key}
}

// StringToPublicKey converts an hex-encoded PEM public key into a PublicKey
func StringToPublicKey(str string) *PublicKey {
	pemdata, err := hex.DecodeString(str)
	CheckError(err)

	block, _ := pem.Decode(pemdata)
	if block == nil || block.Type != "RSA PUBLIC KEY" {
		log.Fatal("Failed to decode public key as string")
	}
	general_key, err := x509.ParsePKIXPublicKey(block.Bytes)
	CheckError(err)
	key, ok := general_key.(*rsa.PublicKey)
	if !ok {
		log.Fatal("Failed to cast into PublicKey")
	}

	return &PublicKey{key}
}

// StringToPublicKey converts an hex-encoded PEM private key into a PrivateKey
func StringToPrivateKey(str string) *PrivateKey {
	pemdata, err := hex.DecodeString(str)
	CheckError(err)

	block, _ := pem.Decode(pemdata)
	if block == nil || block.Type != "RSA PRIVATE KEY" {
		log.Fatal("Failed to decode public key as string")
	}
	key, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	CheckError(err)

	return &PrivateKey{key}
}

// PrivateKey

// String returns a string representation of the PrivateKey
func (key *PrivateKey) String() string {
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key.PrivateKey),
		},
	)
	return hex.EncodeToString(pemdata)
}

// Save stores the private in a PEM format onto the disk
func (key *PrivateKey) Save(fileName string) {
	outFile, err := os.Create(KeyFolder + fileName)
	CheckError(err)

	pem.Encode(
		outFile,
		&pem.Block{
			Type:  "RSA PRIVATE KEY",
			Bytes: x509.MarshalPKCS1PrivateKey(key.PrivateKey),
		},
	)
}

// SignMessage takes in a *PrivateKey pointer and a message as a string.
// It computes the SHA256 hash of the message and signs it.
// It returns the signature as a string if no error is encountered
func (key *PrivateKey) SignMessage(msg string) string {
	rng := rand.Reader
	message := []byte(msg)
	hashed := sha256.Sum256(message)

	signature, err := rsa.SignPKCS1v15(rng, key.PrivateKey, crypto.SHA256, hashed[:])
	CheckError(err)

	signature_str := hex.EncodeToString(signature)
	return signature_str
}

// PublicKey

// String returns a string representation of the PublicKey
func (key *PublicKey) String() string {
	k, err := x509.MarshalPKIXPublicKey(key.PublicKey)
	CheckError(err)
	pemdata := pem.EncodeToMemory(
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: k,
		},
	)
	return hex.EncodeToString(pemdata)
}

// Save stores the public key in PEM format onto disk
func (key *PublicKey) Save(fileName string) {
	k, err := x509.MarshalPKIXPublicKey(key.PublicKey)
	CheckError(err)

	outFile, err := os.Create(KeyFolder + fileName)
	CheckError(err)

	pem.Encode(
		outFile,
		&pem.Block{
			Type:  "RSA PUBLIC KEY",
			Bytes: k,
		},
	)
}

// VerifySignature takes in a *rsa.PrivateKey, a message and a signature.
// It verifies if the signature is valid using the function rsa.VerifyPKCS1v15
// It returns a boolean if a signature is valid or not
func (key *PublicKey) VerifySignature(pubkey *rsa.PublicKey, msg string, signature_str string) bool {
	message := []byte(msg)
	hashed := sha256.Sum256(message)
	signature, _ := hex.DecodeString(signature_str)
	err := rsa.VerifyPKCS1v15(key.PublicKey, crypto.SHA256, hashed[:], signature)
	return (err == nil)
}
