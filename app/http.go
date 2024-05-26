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

type Request struct {
	Method  string
	Path    string
	Version string
	Body    string
	BodyRaw []byte
	Headers map[string]string
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
	Headers map[string]string
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

func Respond(conn net.Conn, req Request, res Response) {
	// Create headers if not existent
	if res.Headers == nil {
		res.Headers = make(map[string]string)
	}

	// Base response line
	fmt.Fprintf(conn, "%s %d %s%s", res.Version, res.Code.Code, res.Code.Message, HTTPDelimiter)

	var body []byte
	if res.Body != "" {
		body = []byte(res.Body)
	} else {
		body = res.BodyRaw
	}
	bodySize := len(body)

	// Check if gzip is accepted and encode the body
	isGzip := strings.Contains(req.Headers["Accept-Encoding"], "gzip")
	if isGzip && bodySize > 0 {
		body = gzipCompress(body)
		bodySize = len(body)
		res.Headers["Content-Encoding"] = "gzip"
	}

	// Set the size of the content
	if bodySize > 0 {
		res.Headers["Content-Length"] = strconv.Itoa(bodySize)
	}

	// Write headers
	for header, value := range res.Headers {
		fmt.Fprintf(conn, "%s: %s%s", header, value, HTTPDelimiter)
	}

	// Delimiter for the body, always present
	fmt.Fprint(conn, HTTPDelimiter)

	// Write body if available
	if bodySize > 0 {
		conn.Write(body)
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
	request := Request{Headers: map[string]string{}}
	isBody := false
	for i, part := range parts {
		// The first part is always the basic information
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

		// Body
		if isBody {
			request.Body = part
			break
		}

		// Headers
		if part == "" {
			// If the header is "empty" it means that we have arrived at the body, and we'll skip to it
			isBody = true
			continue
		}
		h := strings.SplitN(part, ": ", 2)
		request.Headers[h[0]] = h[1]
	}

	return request, true
}
