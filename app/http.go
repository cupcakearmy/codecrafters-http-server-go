package main

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
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
	BodyRaw []byte
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
	Created    = HttpCode{Code: 201, Message: "Created"}
)

type Response struct {
	Code    HttpCode
	Version string
	Body    string
	BodyRaw []byte
	Headers []Header
}

type StringRoute struct {
	path    string
	method  string
	handler func(req Request) Response
}

type RegexRoute struct {
	regex   *regexp.Regexp
	method  string
	handler func(req Request, matches []string) Response
}

type Routes struct {
	stringRoutes []StringRoute
	regexpRoutes []RegexRoute
}

// func createHandler(routes Routes) func(conn net.Conn) {
// 	return func(conn net.Conn) {

// 		for _, route := range(routes.stringRoutes){

// 		}

// 	}
// }

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
