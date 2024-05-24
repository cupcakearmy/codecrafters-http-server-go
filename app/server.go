package main

import (
	"fmt"
	"log"
	"strings"

	// Uncomment this block to pass the first stage
	"net"
	"os"
)

const (
	HTTPDelimiter = "\r\n"
)

type Header struct {
	Name  string
	Value string
}
type Request struct {
	Method  string
	Path    string
	Version string
	Body    string
	Headers []Header
}

type HttpResponse struct {
	Code    uint
	Message string
}

var (
	BadRequest = HttpResponse{Code: 400, Message: "Bad Response"}
	NotFound   = HttpResponse{Code: 404, Message: "Not Found"}
	OK         = HttpResponse{Code: 200, Message: "OK"}
)

// func BadRequest(conn net.Conn) {
// 	fmt.Fprintf(conn, "HTTP/1.1 400 Bad Request%s%s", HTTPDelimiter, HTTPDelimiter)
// }

func Respond(conn net.Conn, response HttpResponse) {
	fmt.Fprintf(conn, "HTTP/1.1 %d %s%s%s", response.Code, response.Message, HTTPDelimiter, HTTPDelimiter)
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	buffer := make([]byte, 1024)
	n, err := conn.Read(buffer)
	if err != nil {
		log.Fatal(err)
	}

	contents := string(buffer[:n])
	parts := strings.Split(contents, HTTPDelimiter)
	request := Request{}
	isBody := false
	for i, part := range parts {
		if i == 0 {
			head := strings.Split(part, " ")
			if len(head) != 3 {
				Respond(conn, BadRequest)
				return
			}
			request.Method = head[0]
			request.Path = head[1]
			request.Version = head[2]
			continue
		}

		if isBody {
			request.Body = part
			break
		}

		// Headers
		if part == "" {
			isBody = true
			continue
		}
		h := strings.SplitN(part, ": ", 2)
		header := Header{Name: h[0], Value: h[1]}
		request.Headers = append(request.Headers, header)
	}
	fmt.Println(request)

	if request.Path == "/" {
		Respond(conn, OK)
		return
	}
	Respond(conn, NotFound)
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	//
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn)
	}

}
