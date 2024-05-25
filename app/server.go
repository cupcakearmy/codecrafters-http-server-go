package main

import (
	"fmt"
	"net"
	"os"
)

// type Handler = func(req Request, res Response)
// type Middleware = func(next Handler) Handler

// var m Middleware = func(next Handler) Handler {
// 	return func(req Request, res Response) {
// 		fmt.Println("Start")
// 		next(req, res)
// 		fmt.Println("End")
// 	}
// }

func handleConnection(conn net.Conn, routes Routes) {
	defer conn.Close()

	req, ok := parseRequest(conn)
	if !ok {
		Respond(conn, req, Response{Version: "HTTP/1.1", Code: BadRequest})
		return
	}

	fmt.Println(req)

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

	Respond(conn, req, Response{Version: req.Version, Code: NotFound})
}

func main() {
	fmt.Println("Logs from your program will appear here!")

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
