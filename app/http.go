package main

import (
	"fmt"
	"log"
	"net"
	"regexp"
	"strconv"
	"strings"
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
}

func parseRequest(conn net.Conn) (Request, bool) {
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
				return Request{}, false
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

	return request, true
}
