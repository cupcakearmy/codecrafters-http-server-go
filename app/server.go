package main

import (
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strconv"
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

type HttpCode struct {
	Code    uint
	Message string
}

var (
	BadRequest = HttpCode{Code: 400, Message: "Bad Response"}
	NotFound   = HttpCode{Code: 404, Message: "Not Found"}
	OK         = HttpCode{Code: 200, Message: "OK"}
)

type Response struct {
	Code    HttpCode
	Version string
	Body    string
	BodyRaw []byte
	Headers []Header
}

func Respond(conn net.Conn, response Response) {
	fmt.Fprintf(conn, "%s %d %s%s", response.Version, response.Code.Code, response.Code.Message, HTTPDelimiter)
	bodySize := 0
	if response.Body != "" {
		bodySize = len(response.Body)
	} else {
		bodySize = len(response.BodyRaw)
	}
	if bodySize > 0 {
		response.Headers = append(response.Headers, Header{Name: "Content-Length", Value: strconv.Itoa(bodySize)})
	}
	for _, header := range response.Headers {
		fmt.Fprintf(conn, "%s: %s%s", header.Name, header.Value, HTTPDelimiter)
	}

	fmt.Fprint(conn, HTTPDelimiter)
	if bodySize > 0 {
		if response.Body != "" {
			fmt.Fprint(conn, response.Body)
		} else {
			conn.Write(response.BodyRaw)
		}
	}

	// fmt.Fprintf(conn, "HTTP/1.1 %d %s%s%s", response.Code, response.Message, HTTPDelimiter, HTTPDelimiter)
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
				Respond(conn, Response{Version: "HTTP/1.1", Code: BadRequest})
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
		Respond(conn, Response{Version: request.Version, Code: OK})
		return
	}

	if request.Path == "/user-agent" {
		for _, header := range request.Headers {
			if header.Name != "User-Agent" {
				continue
			}
			Respond(conn, Response{
				Version: request.Version,
				Code:    OK,
				Body:    header.Value,
				Headers: []Header{{Name: "Content-Type", Value: "text/plain"}},
			})
			return
		}
		Respond(conn, Response{Version: request.Version, Code: BadRequest})
	}

	reFiles := regexp.MustCompile(`^/files/([A-Za-z0-9_\-.]+)$`)
	if matches := reFiles.FindStringSubmatch(request.Path); len(matches) > 0 {

		if len(os.Args) != 3 {
			log.Fatal("Not enough args")
		}
		dir, err := filepath.Abs(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
		filename := filepath.Join(dir, matches[1])

		// _, err:=os.Stat(filename)
		// fmt.Println(file)
		file, err := os.ReadFile(filename)
		if os.IsNotExist(err) {
			Respond(conn, Response{Version: request.Version, Code: NotFound})
			return
		}
		if err != nil {
			log.Fatal(err)
		}

		Respond(conn, Response{
			Version: request.Version,
			Code:    OK,
			BodyRaw: file,
			Headers: []Header{{Name: "Content-Type", Value: "application/octet-stream"}},
		})
		return
	}

	reEcho := regexp.MustCompile(`^/echo/([A-Za-z]+)$`)
	if matches := reEcho.FindStringSubmatch(request.Path); len(matches) > 0 {
		Respond(conn, Response{
			Version: request.Version,
			Code:    OK,
			Body:    matches[1],
			Headers: []Header{{Name: "Content-Type", Value: "text/plain"}},
		})
		return
	}

	Respond(conn, Response{Version: request.Version, Code: NotFound})
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
