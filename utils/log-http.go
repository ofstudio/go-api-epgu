package utils

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"regexp"
	"strings"
)

type Logger interface {
	Print(...any)
}

func LogReq(req *http.Request, logger Logger) {
	if logger == nil {
		return
	}
	dump, _ := httputil.DumpRequestOut(req, true)
	logger.Print(">>> Request to ", req.URL.String(), "\n", sanitize(string(dump)), "\n\n")
}

func LogRes(res *http.Response, logger Logger) {
	if logger == nil {
		return
	}
	dump, _ := httputil.DumpResponse(res, true)
	logger.Print("<<< Response from ", res.Request.URL.String(), "\n", sanitize(string(dump)), "\n")
}

func sanitize(dump string) string {
	dump = sanitizeMultipartFile(dump)
	return dump
}

var reMultipartBinary = regexp.MustCompile(`Content-Type: application/octet-stream\r\n\r\n([\s\S]*?)\r\n--`)

func sanitizeMultipartFile(dump string) string {
	matches := reMultipartBinary.FindStringSubmatch(dump)
	if len(matches) == 2 {
		replace := fmt.Sprintf("[ %d bytes of binary data... ]", len(matches[1]))
		dump = strings.Replace(dump, matches[1], replace, 1)
	}
	return dump
}
