package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"net"
	"os"
)

const BUFSIZE int = 1024

func main() {

	log.SetFlags(0)

	if len(os.Args) != 4 {
		fmt.Println("Not enough arguments")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "-s":
		sendFile(os.Args[2], os.Args[3])
	case "-r":
		receiveFile(os.Args[2], os.Args[3])
	}
}

func sendFile(filename, address string) {

	connection, err := net.Dial("tcp", address)
	catch(err)
	defer connection.Close()

	log.Printf("Connected to %s\n", connection.RemoteAddr())

	file, err := os.Open(filename)
	catch(err)
	defer file.Close()

	stat, err := file.Stat()
	catch(err)

	size := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(size, stat.Size())
	connection.Write(size)

	buffer := make([]byte, BUFSIZE)
	hash := sha256.New()

	log.Printf("Sending %d bytes...\n", stat.Size())

	for {
		length, _ := file.Read(buffer)
		connection.Write(buffer[:length])
		hash.Write(buffer[:length])
		if length < BUFSIZE {
			break
		}
	}

	log.Printf("Sending hash: %x...\n", hash.Sum(nil))
	connection.Write(hash.Sum(nil))
}

func receiveFile(filename, address string) {

	listen, err := net.Listen("tcp", address)
	catch(err)
	defer listen.Close()

	log.Printf("Listening on %s\n", listen.Addr())

	connection, err := listen.Accept()
	catch(err)
	defer connection.Close()

	log.Printf("New incoming connection from %s\n", connection.RemoteAddr())

	file, err := os.Create(filename)
	catch(err)
	defer file.Close()

	sizebuf := make([]byte, binary.MaxVarintLen64)
	connection.Read(sizebuf)
	size, _ := binary.Varint(sizebuf)

	log.Printf("Reading %d bytes...\n", size)

	buffer := make([]byte, BUFSIZE)
	hash := sha256.New()

	for i := size / int64(BUFSIZE); i > 0; i-- {
		connection.Read(buffer)
		file.Write(buffer)
		hash.Write(buffer)
	}

	buffer = make([]byte, size%int64(BUFSIZE))
	connection.Read(buffer)
	file.Write(buffer)
	hash.Write(buffer)

	netHash := make([]byte, sha256.Size)
	connection.Read(netHash)

	if bytes.Equal(hash.Sum(nil), netHash) {
		log.Println("Hashes are equal")
	} else {
		log.Println("Hashes aren't equal")
	}

	log.Printf("%x\n%x\n", hash.Sum(nil), netHash)
}

func catch(err error) {
	if err != nil {
		panic(err)
	}
}
