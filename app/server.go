package main

import (
	"fmt"
	"net"
	"os"
	"regexp"
)

func handleConnection(conn net.Conn, routes Routes) {
	defer conn.Close()

	request, ok := parseRequest(conn)
	if !ok {
		Respond(conn, Response{Version: "HTTP/1.1", Code: BadRequest})
		return
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
				},
			},

			{
				regex:  regexp.MustCompile(`^/files/([A-Za-z0-9_\-.]+)`),
				method: "GET",
				handler: func(req Request, matches []string) Response {
					file, notFound := readFile(matches[1])
					if notFound {
						return Response{Version: req.Version, Code: NotFound}
					}
					return Response{
						Version: req.Version,
						Code:    OK,
						BodyRaw: file,
						Headers: []Header{{Name: "Content-Type", Value: "application/octet-stream"}},
					}
				},
			},

			{
				regex:  regexp.MustCompile(`^/files/([A-Za-z0-9_\-.]+)`),
				method: "POST",
				handler: func(req Request, matches []string) Response {
					writeFile(matches[1], []byte(req.Body))
					return Response{Version: req.Version, Code: Created}
				},
			},
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
