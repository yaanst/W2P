package w2pcrypto

import (
    "os"
    "fmt"
    "crypto"
    "crypto/rsa"
    "crypto/rand"
    "crypto/sha256"
    "encoding/hex"
)


type WebsiteKeyMap map[string]*rsa.PublicKey

func (wkm *WebsiteKeyMap) Set(website string, key *rsa.PublicKey) {
    (*wkm)[website] = key
}

func CheckError(err error, msg string) bool {
    if err != nil {
        fmt.Fprintln(os.Stderr, msg, err)
        return true
    }
    return false
}

func SignMessage(rsaPrivateKey *rsa.PrivateKey, msg string) (string,error) {
	rng := rand.Reader
	message := []byte(msg)
	hashed := sha256.Sum256(message)
	signature, err := rsa.SignPKCS1v15(rng, rsaPrivateKey, crypto.SHA256, hashed[:])

    signature_str := ""
    if CheckError(err, "Error from signing: %s\n") {
        return signature_str, err
    }
    signature_str = hex.EncodeToString(signature)
    return signature_str, err
}

func VerifySignature(rsaPrivateKey *rsa.PrivateKey, msg string, sig string) bool {
	message := []byte(msg)
	signature, _ := hex.DecodeString(sig)
	hashed := sha256.Sum256(message)

	err := rsa.VerifyPKCS1v15(&rsaPrivateKey.PublicKey, crypto.SHA256, hashed[:], signature)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error from verification: %s\n", err)
		return false
	}
    return true
}
