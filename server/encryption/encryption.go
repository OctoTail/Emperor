package encryption

import (
	"encoding/pem"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/rsa"
	"crypto/x509"
	"crypto/rand"
	"io"
	"io/ioutil"
	"errors"
)

const AES_LENGTH = 16
const HMAC_LENGTH = 32
var RsaPriv *rsa.PrivateKey
var RsaPub []byte

func adjustMsgLen(text []byte) []byte {
	var msg []byte
	if l := len(text); l < AES_LENGTH {
		msg = append(text, make([]byte, AES_LENGTH-l)...)
	} else {
		msg = text
	}
	return msg
}

func decryptAES(key, iv, ciphertext []byte) ([]byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
    	return nil, err
    }
    cfb := cipher.NewCFBDecrypter(block, iv)
    text := make([]byte,len(ciphertext))
    cfb.XORKeyStream(text, ciphertext)
    return text, nil
}

func encryptAES(key, text []byte) ([]byte, []byte, error) {
    block, err := aes.NewCipher(key)
    if err != nil {
    	return nil, nil, err
    }
    iv := make([]byte, AES_LENGTH)
    if _, err := io.ReadFull(rand.Reader, iv); err != nil {
    	return nil, nil, err
    }
    ciphertext := make([]byte, len(text))
    cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(ciphertext, text)
	return iv, ciphertext, nil
}

func DecryptRSA(ciphertext []byte) ([]byte, error) {
        return rsa.DecryptOAEP(sha256.New(), rand.Reader, RsaPriv, ciphertext, make([]byte, 0))
}

func checkMAC(text, textMAC, key []byte) bool {
	mac := hmac.New(sha256.New, key)
	mac.Write(text)
	expectedMAC := mac.Sum(nil)
	return hmac.Equal(textMAC, expectedMAC)
}

func makeHMAC(key, text []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(text)
	return mac.Sum(nil)
}

func LoadRSA(keyfile string) (*rsa.PrivateKey, []byte, error) {
	pemData, err := ioutil.ReadFile(keyfile)
	if err != nil {
		return nil, nil, err
	}
    block, _ := pem.Decode(pemData)
    if block == nil {
        return nil, nil, errors.New("Bad RSA private key file: not PEM encoded")
    }
    if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
        return nil, nil, errors.New("Unknown key type")
    }
    priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
    if err != nil {
    	return nil, nil, err
    }
    pub, err := x509.MarshalPKIXPublicKey(priv.Public())
    return priv, pub, err
}

func Decrypt(buf, key []byte) ([]byte, error) {
	iv := buf[2:AES_LENGTH + 2]
	textMAC := buf[AES_LENGTH + 2:AES_LENGTH + 2 + HMAC_LENGTH]
	ciphertext := buf[AES_LENGTH + 2 + HMAC_LENGTH:]
	text, err := decryptAES(key, iv, ciphertext)
	if err != nil {
		return nil, errors.New("sent an incorrectly encrypted message (AES)")
	}
	if !checkMAC(text, textMAC, key) {
		return nil, errors.New("sent an incorrect HMAC")
	}
	return text, nil
}

func Encrypt(text, key []byte) []byte {
		msg := adjustMsgLen(text)
	    iv, ciphertext, _ := encryptAES(key, msg)
	    mac := makeHMAC(key, msg)
	    res := append(make([]byte,0), iv...)
	    res = append(res, mac...)
	    res = append(res, ciphertext...)
		return res
}