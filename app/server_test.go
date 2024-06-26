package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
)

var (
	client = &http.Client{Transport: &http.Transport{DisableCompression: true}}
)

func getBody(body io.ReadCloser) []byte {
	defer body.Close()
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		log.Fatalf("Failed to read res body: %v", err)
	}
	return bodyBytes
}

type Expected struct {
	status  int
	body    string
	headers map[string]string
}

func checkResponse(t *testing.T, res *http.Response, expected Expected) {
	if res.StatusCode != expected.status {
		t.Errorf("Expected status code %d, received %d", expected.status, res.StatusCode)
	}

	body := string(getBody(res.Body))
	if body != expected.body {
		t.Errorf(`Expected body to be "%s" but got "%s"`, expected.body, body)
	}

	log.Println("HEADERS", res.Header)
	for header, value := range expected.headers {
		if res.Header[header] == nil {
			t.Errorf(`Expected "%s" header to be present`, header)
		} else if actual := res.Header[header][0]; actual != value {
			t.Errorf(`Expected "%s" header to be "%s" but got "%s"`, header, value, actual)
		}
	}
}

func TestRoot(t *testing.T) {
	res, _ := http.Get("http://localhost:4221")
	checkResponse(t, res, Expected{status: 200})
}

func TestNotFound(t *testing.T) {
	res, _ := http.Get("http://localhost:4221/foo")
	checkResponse(t, res, Expected{status: 404})
}

func TestEcho(t *testing.T) {
	input := "abc"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:4221/echo/%s", input), nil)
	res, _ := client.Do(req)
	checkResponse(t, res, Expected{status: 200, body: input, headers: map[string]string{
		"Content-Length": strconv.Itoa(len(input)),
		"Content-Type":   "text/plain",
	}})
}

func TestEchoGzip(t *testing.T) {
	input := "abc"
	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:4221/echo/%s", input), nil)
	req.Header.Set("Accept-Encoding", "gzip")
	res, _ := client.Do(req)

	asGzip := string(gzipCompress([]byte(input)))
	checkResponse(t, res, Expected{status: 200, body: asGzip, headers: map[string]string{
		"Content-Length":   strconv.Itoa(len(asGzip)),
		"Content-Type":     "text/plain",
		"Content-Encoding": "gzip",
	}})
}

func TestUserAgent(t *testing.T) {
	input := "CodeCrafters/1.0"
	req, _ := http.NewRequest("GET", "http://localhost:4221/user-agent", nil)
	req.Header.Set("User-Agent", input)
	res, _ := client.Do(req)
	checkResponse(t, res, Expected{status: 200, body: input, headers: map[string]string{
		// "Content-Length": strconv.Itoa(len(input)),
		"Content-Type": "text/plain",
	}})
}
func TestUserAgentNoHeader(t *testing.T) {
	req, _ := http.NewRequest("GET", "http://localhost:4221/user-agent", nil)
	req.Header.Set("User-Agent", "")
	res, _ := client.Do(req)
	checkResponse(t, res, Expected{status: 400})
}

func TestReadFile(t *testing.T) {
	input := "Hello World"
	tmp, _ := os.CreateTemp("", "read.txt")
	defer os.Remove(tmp.Name())
	os.WriteFile(tmp.Name(), []byte(input), 0755)
	DIR = path.Dir(tmp.Name())

	req, _ := http.NewRequest("GET", fmt.Sprintf("http://localhost:4221/files/%s", path.Base(tmp.Name())), nil)
	res, _ := client.Do(req)
	checkResponse(t, res, Expected{status: 200, body: input, headers: map[string]string{
		"Content-Type":   "application/octet-stream",
		"Content-Length": strconv.Itoa(len(input)),
	}})
}

func TestWriteFile(t *testing.T) {
	input := "Hello World"
	tmp, _ := os.CreateTemp("", "write.txt")
	defer os.Remove(tmp.Name())
	DIR = path.Dir(tmp.Name())

	req, _ := http.NewRequest("POST", fmt.Sprintf("http://localhost:4221/files/%s", path.Base(tmp.Name())), strings.NewReader(input))
	req.Header.Set("Content-Type", "application/octet-stream")
	res, _ := client.Do(req)
	checkResponse(t, res, Expected{status: 201})

	contents, _ := os.ReadFile(tmp.Name())
	if string(contents) != input {
		t.Errorf("Content written to file does not match the input")
	}
}

func TestMain(m *testing.M) {
	fmt.Println("Starting server")
	l, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221")
		os.Exit(1)
	}

	go func() {
		for {
			conn, err := l.Accept()
			if err == nil {
				go handleConnection(conn, routes)
			}
		}
	}()

	code := m.Run()

	fmt.Println("Stopping server")
	l.Close()

	os.Exit(code)
}
