package main

import (
	"net"
	"fmt"
    "./encryption"
    "encoding/binary"
    "math"
)

const DIAL_PORT=":8888"

func main() {
	AES_KEY := []byte{1,2,3,4,5,6,7,8,9,10,11,12,13,14,15,16}
	PID := []byte{0,1}

	ServerAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:8888")
    LocalAddr, _ := net.ResolveUDPAddr("udp", "127.0.0.1:0")
    Conn, _ := net.DialUDP("udp", LocalAddr, ServerAddr)
    defer Conn.Close()

    go func() {
        for {
            p :=  make([]byte, 2048)
            n,_ := Conn.Read(p)

            msg, err := encryption.Decrypt(p[:n], AES_KEY)
            if err!= nil {
                fmt.Printf("%s\n",err)
                return
            }
            switch msg[0] {
            case 0:
                fmt.Printf("Unit[")
                id := binary.BigEndian.Uint32(msg[1:5])
                fmt.Printf("%d].%d := %x\n", id, msg[5], msg[6:])
            case 1:
                fmt.Printf("Timer[")
                id := binary.BigEndian.Uint32(msg[1:5])
                start := math.Float64frombits(binary.BigEndian.Uint64(msg[5:13]))
                delta := math.Float64frombits(binary.BigEndian.Uint64(msg[13:21]))
                fmt.Printf("%d] := %f->%f\n", id, start, delta)
                break
            default:
                fmt.Printf("%d[",msg[0])
                id := binary.BigEndian.Uint32(msg[1:5])
                fmt.Printf("%d].%d := %x\n", id, msg[5], msg[6:])
            }
        }} ()
    
    for {
        input := make([]byte,0)
        for {
            var x byte
            _, err := fmt.Scanf("%x",&x)
            if err != nil {
                break
            }
            input = append(input, x)
        }
        buf := encryption.Encrypt(input, AES_KEY)
        buf = append(PID,buf...)
        Conn.Write(buf)
    }
}
