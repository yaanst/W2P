// Package w2pcrypto provides easy-to-use functions for creating RSA key pairs,
// using them to sign messages and verify signatures
package w2pcrypto

import (
    "os"
    "log"
    "crypto"
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha256"
    "encoding/gob"
    "encoding/hex"
)

const RSA_KEY_BITS int = 512
const KEY_DIR string = "./keys/"

// WebsiteKeyMap is a mapping from a website ID to its associated *rsa.PublicKey
// used to verify signatures
type WebsiteKeyMap map[string]*rsa.PublicKey

// Set adds or updates an entry of the WebsiteKeyMap with the *rsa.PrivateKey
// given in parameter.
func (wkm *WebsiteKeyMap) Set(website string, key *rsa.PublicKey) {
    (*wkm)[website] = key
}
// Get returns the *rsa.PrivateKey corresponding to the website or nil if it
// does not exist
func (wkm *WebsiteKeyMap) Get(website string) *rsa.PublicKey {
    return (*wkm)[website]
}

// CheckError checks if there was an error.
// If there was, it logs it and exits
func CheckError(err error) {
    if err != nil {
        log.Fatal(err)
    }
}

/*****************************
            Crypto
*****************************/

// CreateKeyPair is a wrapper for the rsa.GenerateKey function.
// It returns a *rsa.PrivateKey if no error is encountered.
func CreateKey() *rsa.PrivateKey {
	rng := rand.Reader

    privkey, err := rsa.GenerateKey(rng, RSA_KEY_BITS)
    CheckError(err)
    return privkey
}

// SaveKey that saves the given private key to disk
// It uses Gob encoding (https://blog.golang.org/gobs-of-data)
func SaveKey(fileName string, key *rsa.PrivateKey) {
    outFile, err := os.Create(KEY_DIR + fileName)
    CheckError(err)

    encoder := gob.NewEncoder(outFile)
    err = encoder.Encode(key)
    CheckError(err)
    outFile.Close()
}

// LoadKey loads the key with the given name
// It returns a *rsa.PrivateKey (from which the public key can be retrieved)
func LoadKey(fileName string) *rsa.PrivateKey {
    var key rsa.PrivateKey

    inFile, err := os.Open(KEY_DIR + fileName)
    CheckError(err)

    decoder := gob.NewDecoder(inFile)
    err = decoder.Decode(&key)
    CheckError(err)
    inFile.Close()

    return &key
}

// SignMessage takes in a *rsa.PrivateKey pointer and a message as a string.
// It computes the SHA256 hash of the message and signs it.
// It returns the signature as a string if no error is encountered
func SignMessage(privkey *rsa.PrivateKey, msg string) string {
	rng := rand.Reader
	message := []byte(msg)
	hashed := sha256.Sum256(message)

	signature, err := rsa.SignPKCS1v15(rng, privkey, crypto.SHA256, hashed[:])
    CheckError(err)

    signature_str := hex.EncodeToString(signature)
    return signature_str
}

// VerifySignature takes in a *rsa.PrivateKey, a message and a signature.
// It verifies if the signature is valid using the function rsa.VerifyPKCS1v15
// It returns a boolean if a signature is valid or not
func VerifySignature(pubkey *rsa.PublicKey, msg string, signature_str string) bool {
	message := []byte(msg)
	hashed := sha256.Sum256(message)
	signature, _ := hex.DecodeString(signature_str)

	err := rsa.VerifyPKCS1v15(pubkey, crypto.SHA256, hashed[:], signature)
    return (err == nil)
}
