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
			log.Print("Responding to incoming req from: " + c.RemoteAddr().String())

			buf := make([]byte, MAX_LENGTH)

			c.Read(buf)

			go func() {
				var fileName string = saveFile(&buf)
				c.Write([]byte(fileName))
				c.Close()
			}()
		}(conn)
	}
}

func saveFile(buf *[]byte) string {
	h := fnv.New32a()
	h.Write(*buf)

	// Cheeky-breeky uint32->string
	sum32String := fmt.Sprint(h.Sum32())

	log.Print("Saving new paste with name: ", sum32String)

	f, err := os.OpenFile("pastes/"+sum32String, os.O_WRONLY|os.O_CREATE, 0666)

	if err != nil {
		log.Fatal(err)
	}

	defer f.Close()

	// Remove the null values from the file
	io.WriteString(f, string(bytes.Trim(*buf, "\x00")))

	return sum32String
}

// This will deal with the request to retrieve a paste
func retrieveHandler(res http.ResponseWriter, req *http.Request) {
	vars := mux.Vars(req)

	f, err := ioutil.ReadFile("pastes/" + vars["id"])

	if err != nil {
		res.WriteHeader(http.StatusInternalServerError)
		log.Print("Error encountered while trying to access paste")
		log.Fatal(err)
		return
	}

	log.Print("Found file: ", vars["id"], " and now sending to user")

	res.WriteHeader(http.StatusOK)

	res.Write(f)
}

func formUploadHandler(res http.ResponseWriter, req *http.Request) {
	// Convert from string->[]byte
	var code []byte = []byte(req.FormValue("code"))

	var fileName string = saveFile(&code)

	log.Print("Redirecting, created file ", fileName)
	http.Redirect(res, req, "/"+fileName, 301)
}

func retrieve() {
	router := mux.NewRouter().StrictSlash(true)

	// Retrieve paste
	router.HandleFunc("/{id:[0-9]+}", retrieveHandler).Methods("GET")

	// Form submit
	router.HandleFunc("/create", formUploadHandler).Methods("POST")

	// Homepage
	router.HandleFunc("/", func(res http.ResponseWriter, req *http.Request) {
		log.Print("Serving index.html")
		http.ServeFile(res, req, "./src/index.html")
	})

	log.Fatal(http.ListenAndServe(":8080", router))
}

func main() {
	go retrieve()
	create()
}
