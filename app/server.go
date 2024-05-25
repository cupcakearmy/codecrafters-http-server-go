package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func getFilepath(filename string) string {
	if len(os.Args) != 3 {
		log.Fatal("Not enough args")
	}
	dir, err := filepath.Abs(os.Args[2])
	if err != nil {
		log.Fatal(err)
	}
	return filepath.Join(dir, filename)
}

func handleConnection(conn net.Conn, routes Routes) {
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

	for _, route := range routes.stringRoutes {
		if request.Path == route.path && request.Method == route.method {
			Respond(conn, route.handler(request))
			return
		}
	}

	for _, route := range routes.regexpRoutes {
		if request.Method != route.method {
			continue
		}
		if matches := route.regex.FindStringSubmatch(request.Path); len(matches) > 0 {
			Respond(conn, route.handler(request, matches))
			return
		}
	}

	// if request.Path == "/" {
	// 	Respond(conn, Response{Version: request.Version, Code: OK})
	// 	return
	// }

	// if request.Path == "/user-agent" {
	// 	for _, header := range request.Headers {
	// 		if header.Name != "User-Agent" {
	// 			continue
	// 		}
	// 		Respond(conn, Response{
	// 			Version: request.Version,
	// 			Code:    OK,
	// 			Body:    header.Value,
	// 			Headers: []Header{{Name: "Content-Type", Value: "text/plain"}},
	// 		})
	// 		return
	// 	}
	// 	Respond(conn, Response{Version: request.Version, Code: BadRequest})
	// }

	// reFiles := regexp.MustCompile(`^/files/([A-Za-z0-9_\-.]+)$`)
	// if matches := reFiles.FindStringSubmatch(request.Path); len(matches) > 0 {
	// 	filename := getFilepath(matches[1])
	// 	file, err := os.ReadFile(filename)
	// 	if os.IsNotExist(err) {
	// 		Respond(conn, Response{Version: request.Version, Code: NotFound})
	// 		return
	// 	}
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}

	// 	Respond(conn, Response{
	// 		Version: request.Version,
	// 		Code:    OK,
	// 		BodyRaw: file,
	// 		Headers: []Header{{Name: "Content-Type", Value: "application/octet-stream"}},
	// 	})
	// 	return
	// }

	// reEcho := regexp.MustCompile(`^/echo/([A-Za-z]+)$`)
	// if matches := reEcho.FindStringSubmatch(request.Path); len(matches) > 0 {
	// 	Respond(conn, Response{
	// 		Version: request.Version,
	// 		Code:    OK,
	// 		Body:    matches[1],
	// 		Headers: []Header{{Name: "Content-Type", Value: "text/plain"}},
	// 	})
	// 	return
	// }

	Respond(conn, Response{Version: request.Version, Code: NotFound})
}

func main() {
	fmt.Println("Logs from your program will appear here!")

	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	routes := Routes{
		stringRoutes: []StringRoute{
			// ROOT
			{path: "/", method: "GET", handler: func(req Request) Response {
				return Response{Version: req.Version, Code: OK}
			}},

			// USER AGENT
			{path: "/user-agent", method: "GET", handler: func(req Request) Response {
				for _, header := range req.Headers {
					if header.Name != "User-Agent" {
						continue
					}
					return Response{
						Version: req.Version,
						Code:    OK,
						Body:    header.Value,
						Headers: []Header{{Name: "Content-Type", Value: "text/plain"}},
					}
				}
				return Response{Version: req.Version, Code: BadRequest}
			}},
		},

		regexpRoutes: []RegexRoute{

			// PATH PARAMETER
			{
				regex:  regexp.MustCompile(`^/echo/([A-Za-z]+)$`),
				method: "GET",
				handler: func(req Request, matches []string) Response {
					return Response{
						Version: req.Version,
						Code:    OK,
						Body:    matches[1],
						Headers: []Header{{Name: "Content-Type", Value: "text/plain"}},
					}
				}},
		},
	}

	for {
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting connection: ", err.Error())
			os.Exit(1)
		}
		go handleConnection(conn, routes)
	}

}
