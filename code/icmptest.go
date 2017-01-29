package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"os"
)

/**
   0                   1                   2                   3
   0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1 2 3 4 5 6 7 8 9 0 1
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |     Type      |     Code      |          Checksum             |
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |           Identifier          |        Sequence Number        |
  +-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
  |     Data ...
  +-+-+-+-+-
*/

func main() {
	if len(os.Args) != 2 {
		fmt.Println("USAGE: ", os.Args[0], "host")
		os.Exit(1)
	}
	service := os.Args[1]

	conn, err := net.Dial("ip4:icmp", service)
	checkError(err)

	var msg [512]byte
	msg[0] = 8  //echo
	msg[1] = 0  //code 0
	msg[2] = 0  //checksum
	msg[3] = 0  //checksum
	msg[4] = 0  //identifier[0]
	msg[5] = 13 //identifier[1]
	msg[6] = 0  //sequence[0]
	msg[7] = 37 //sequence[1]
	msg[8] = 99
	len := 9

	check := checkSum(msg[0:len])
	msg[2] = byte(check >> 8)
	msg[3] = byte(check & 255)

	_, err = conn.Write(msg[0:len])
	checkError(err)

	_, err = conn.Read(msg[0:])
	checkError(err)
	fmt.Println(msg[0 : 20+len])

	fmt.Println("Got response")
	if msg[20+5] == 13 {
		fmt.Println("Identifier matches")
	}
	if msg[20+7] == 37 {
		fmt.Println("Sequence matches")
	}
	if msg[20+8] == 99 {
		fmt.Println("Custom data matches")
	}

	os.Exit(0)
}

func checkSum(msg []byte) uint16 {
	sum := 0

	len := len(msg)
	for i := 0; i < len-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if len%2 == 1 {
		sum += int(msg[len-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error : %s", err.Error())
		os.Exit(0)
	}
}

func readFully(conn net.Conn) ([]byte, error) {
	defer conn.Close()

	result := bytes.NewBuffer(nil)
	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		result.Write(buf[0:n])
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
	}
	return result.Bytes(), nil
}
