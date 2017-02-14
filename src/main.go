package main

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"io"
	"log"
	"net"
	"os"
)

const MAX_LENGTH = 3 << 20 // 3 MiB

func create() {
	l, err := net.Listen("tcp", ":2020")

	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Fatal(err)
		}

		// Respond
		go func(c net.Conn) {
			log.Print("Responding to incoming req from ", c.RemoteAddr())

			buf := make([]byte, MAX_LENGTH)

			c.Read(buf)

			go saveFileAndRespond(&buf, c)
		}(conn)
	}
}

func saveFileAndRespond(buf *[]byte, c net.Conn) {
	defer c.Close()

	h := fnv.New32a()
	h.Write(*buf)

	sum32String := fmt.Sprint(h.Sum32())

	log.Print("Saving new paste with name: ", sum32String)

	f, err := os.OpenFile("../pastes/"+sum32String, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Fatal(err)
		io.WriteString(c, err.Error())
	}

	defer f.Close()

	io.WriteString(f, string(bytes.Trim(*buf, "\x00")))

	io.WriteString(c, sum32String)
}

func main() {
	create()
}
