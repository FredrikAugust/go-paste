package main

import (
	"bytes"
	"fmt"
	"github.com/gorilla/mux"
	"hash/fnv"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
)

// Max file size
const MAX_LENGTH = 3 << 20 // 3 MiB

func TCPServer() {
	l, err := net.Listen("tcp", ":2020")

	defer l.Close()

	if err != nil {
		log.Print(err)
		return
	}

	for {
		conn, err := l.Accept()

		if err != nil {
			log.Print(err)
			return
		}

		// Respond
		go func(c net.Conn) {
			log.Print("Responding to incoming req from: " + c.RemoteAddr().String())

			buf := make([]byte, MAX_LENGTH)

			_, err := c.Read(buf)

			if err != nil {
				log.Print(err)
				return
			}

			var fileName string = SaveFile(&buf)
			c.Write([]byte(fileName))
			c.Close()
		}(conn)
	}
}

func SaveFile(buf *[]byte) string {
	h := fnv.New32a()
	h.Write(*buf)

	// Cheeky-breeky uint32->string
	sum32String := fmt.Sprint(h.Sum32())

	log.Print("Saving new paste with name: ", sum32String)

	f, err := os.OpenFile("pastes/"+sum32String, os.O_WRONLY|os.O_CREATE, 0666)

	defer f.Close()

	if err != nil {
		log.Print(err)
		return ""
	}

	// Remove the null values from the file
	io.WriteString(f, string(bytes.Trim(*buf, "\x00")))

	return sum32String
}

// This will deal with the request to retrieve a paste
func RetrieveHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	_, err := os.Stat("pastes/" + vars["id"])

	if err != nil {
		log.Print(err.Error())
		res.WriteHeader(http.StatusNotFound)
		return
	}

	f, err := ioutil.ReadFile("pastes/" + vars["id"])

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Print("Error encountered while trying to access paste")
		log.Print(err)
		return
	}

	log.Print("Found file: ", vars["id"], " and now sending to user")

	res.WriteHeader(http.StatusOK)

	res.Write(f)
}

func FormUploadHandler(res http.ResponseWriter, req *http.Request) {
	// Convert from string->[]byte
	code := []byte(req.FormValue("code"))

	var fileName string = SaveFile(&code)

	log.Print("Redirecting, created file ", fileName)
	http.Redirect(res, req, "/"+fileName, http.StatusMovedPermanently)
}

func HttpServer() {
	router := mux.NewRouter().StrictSlash(true)

	// Retrieve paste
	router.HandleFunc("/{id:[0-9]+}", RetrieveHandler).Methods("GET")

	// Form submit
	router.HandleFunc("/create", FormUploadHandler).Methods("POST")

	// Homepage
	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log.Print("Serving index.html")
		http.ServeFile(res, req, "./src/index.html")
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	go HttpServer()
	TCPServer()
}
