package main

import (
	"regexp"
)

var routes = Routes{
	stringRoutes: []StringRoute{
		// ROOT
		{path: "/", method: "GET", handler: func(req Request) Response {
			return Response{Version: req.Version, Code: OK}
		}},

		// USER AGENT
		{path: "/user-agent", method: "GET", handler: func(req Request) Response {
			userAgent := req.Headers["User-Agent"]
			if userAgent == "" {
				return Response{Version: req.Version, Code: BadRequest}
			}

			return Response{
				Version: req.Version,
				Code:    OK,
				Body:    userAgent,
				Headers: map[string]string{"Content-Type": "text/plain"},
			}
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
					Headers: map[string]string{"Content-Type": "text/plain"},
				}
			},
		},

		// Read file
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
					Headers: map[string]string{"Content-Type": "application/octet-stream"},
				}
			},
		},

		// Write file
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
