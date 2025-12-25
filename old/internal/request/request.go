package request

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"
)
type Request struct {
	RequestLine RequestLine
}

type RequestLine struct {
	HttpVersion string
	RequestTarget string
	Method string
}

func RequestFromReader(reader io.Reader) (*Request, error) {
    req, err := io.ReadAll(reader)
    if err != nil {
        fmt.Println("Failed to read request")
        return nil, errors.New("failed to read request")
    }

    var lines []string = strings.Split(string(req), "\r\n")
    reqLine := lines[0]

    var request Request
    requestLine, err := parseRequestLine(reqLine)
    if err != nil {
        return nil, err
    }
    request.RequestLine = *requestLine
    fmt.Println(request.RequestLine.Method)
    fmt.Println(request.RequestLine.RequestTarget)
    fmt.Println(request.RequestLine.HttpVersion)
    return &request, nil
}

func parseRequestLine(reqLine string) (*RequestLine, error) {
    reqLineSplit := strings.Split(reqLine, " ")

    if len(reqLineSplit) != 3 {
        fmt.Println("Incorrect request line format")
        return nil, errors.New("incorrect request line format")
    }
    var requestLine RequestLine
    
    if strings.ToUpper(reqLineSplit[0]) != reqLineSplit[0] && !isCapital(reqLineSplit[0]) {
        fmt.Println("Invalid http method name")
        return nil, errors.New("invalid http method name")
    }
    requestLine.Method = reqLineSplit[0]

    if !isCapital(reqLineSplit[0]) {
        fmt.Println("Invalid request target")
        return nil, errors.New("invalid request target")
    }
    requestLine.RequestTarget = reqLineSplit[1]
    
    if reqLineSplit[2] != "HTTP/1.1" {
        fmt.Println("Invalid http version")
        return nil, errors.New("invalid http version")
    }
    requestLine.HttpVersion = strings.Split(reqLineSplit[2], "/")[1]

    return &requestLine, nil
}

func isCapital(s string) (bool) {
    return !strings.ContainsFunc(s, func(r rune) bool {
        return !unicode.IsUpper(r)
    })
}