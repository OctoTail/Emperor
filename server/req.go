package main

import (
	"net"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/rand"
	"io"
	"fmt"
)

const AES_LENGTH=16
const HMAC_LENGTH=32
const DIAL_PORT=":8888"
const PID_LENGTH = 2

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

func makeHMAC(key, text []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(text)
	return mac.Sum(nil)
}

func adjustMsgLen(text []byte) []byte {
	var msg []byte
	if l := len(text); l < AES_LENGTH {
		msg = append(text, make([]byte, AES_LENGTH-l)...)
	} else {
		msg = text
	}
	return msg
}

func main() {
	AES_KEY := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16}
	MSG := []byte{22,23,24,25,26}
	PID := []byte{0,1}
	ServerAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8888")
    LocalAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    Conn, _ := net.DialUDP("udp", LocalAddr, ServerAddr)
    defer Conn.Close()
    msg := adjustMsgLen(MSG)
    iv, ciphertext, _ := encryptAES(AES_KEY, msg)
    mac := makeHMAC(AES_KEY, msg)
    fmt.Printf("%v\n",mac)
    buf := make([]byte,0)
    buf = append(buf, PID...)
    buf = append(buf, iv...)
    buf = append(buf, mac...)
    buf = append(buf, ciphertext...)
    Conn.Write(buf)
    p :=  make([]byte, 2048)
    n,_ := Conn.Read(p)
    fmt.Printf("%v\n", p[:n])

	iv = p[:AES_LENGTH]
	//textMAC := p[AES_LENGTH:AES_LENGTH + HMAC_LENGTH]
	ciphertext = p[AES_LENGTH + HMAC_LENGTH:n]
	text, _ := decryptAES(AES_KEY, iv, ciphertext)
	fmt.Printf("%v\n",text)
}