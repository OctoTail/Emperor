package main

import (
	"net"
	"log"
	"encoding/binary"
	"./encryption"
	"./data"
)

const LISTEN_PORT=":8888"

func checkFatal(err error){
	if err != nil {
		log.Fatal(err)
	}
}

func checkWarn(err error) error{
	if err != nil {
		log.Println(err)
		return err
	} else {
		return nil
	}
}

func handleReq(n int, addr net.Addr, buf []byte, sCon *net.UDPConn){
	if n < 2 {
		log.Printf("%v Sent less than 2 bytes\n", addr)
		return
	}
	pId := binary.BigEndian.Uint16(buf[:2])
	switch {
	case pId == 0 && n == 2: //Public key request
		if _,err := sCon.WriteTo(encryption.RsaPub, addr); err != nil {
			log.Printf("Failed to reply to %v : %v", addr, err)
		}
	case pId == 0 && n > 2: //Registration
		ciphertext := buf[2:n]
		text, err := encryption.DecryptRSA(ciphertext)
		if err != nil{
			log.Printf("%v Incorrectly encrypted message (RSA)\n")
		}
		log.Println(text)
	case n >= 2*encryption.AES_LENGTH + 2 + encryption.HMAC_LENGTH: //Request
		pId-- //Switch to real pId
		if int(pId) >= len(data.Players) {
			print(pId)
			log.Printf("%v Sent an invalid player id\n", addr)
			return
		}
		key := data.Players[pId].Key
		text, err := encryption.Decrypt(buf[2:n], key)
		if err != nil {
			log.Printf("%v %s\n", addr, err)
			return
		}
		log.Printf("%v %d\n", addr, text)
		data.Players[pId].Addr = addr
		switch text[0] {
		case 1: //SET2
			if err := data.ReqSET2(text[1:], pId); err != nil {
				log.Printf("%v %s\n", addr, err)
			}
		default:
			log.Printf("%v Unknown request: %x", addr, text[0])
		}

		//Reply
		/*res := encryption.Encrypt([]byte{1,2,3,5,8},key)
		_, err = sCon.WriteTo(res, addr)
		checkWarn(err)*/

	default:
		log.Printf("%v Incorrectly encrypted message\n", addr)
	}
}


func main() {
	encryption.LoadRSA("private.pem")
	sAddr,err := net.ResolveUDPAddr("udp", LISTEN_PORT)
	checkFatal(err)
	sCon,err := net.ListenUDP("udp", sAddr)
	checkFatal(err)
	data.Init(sCon)
	defer sCon.Close()
	buf := make([]byte,1024)
	for {
		n, addr, err := sCon.ReadFrom(buf)
		if checkWarn(err) == nil {
			go handleReq(n, addr, buf, sCon)
		}
	}
}
