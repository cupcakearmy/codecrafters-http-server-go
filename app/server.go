package main

import (
	"fmt"
	"net"
	"os"
)

func handleConnection(conn net.Conn, routes Routes) {
	defer conn.Close()

	req, ok := parseRequest(conn)
	if !ok {
		Respond(conn, req, Response{Version: "HTTP/1.1", Code: BadRequest})
		return
	}

	fmt.Println(req)

	// Loop over the available routes. First string, then regexp
	for _, route := range routes.stringRoutes {
		if req.Path == route.path && req.Method == route.method {
			Respond(conn, req, route.handler(req))
			return
		}
	}

	for _, route := range routes.regexpRoutes {
		if req.Method != route.method {
			continue
		}
		if matches := route.regex.FindStringSubmatch(req.Path); len(matches) > 0 {
			Respond(conn, req, route.handler(req, matches))
			return
		}
	}

	// Catch all 404
	Respond(conn, req, Response{Version: req.Version, Code: NotFound})
}

func main() {
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
		go handleConnection(conn, routes)
	}

}
