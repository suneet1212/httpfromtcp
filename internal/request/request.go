package request

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"httpfromtcp/internal/headers"
)

type ParserState int

const (
	Initialized ParserState = iota
	ParsingHeaders
	ParsingBody
	Done
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
}

type Request struct {
	RequestLine RequestLine
	state ParserState
	Headers headers.Headers
	Body []byte
}

func newRequest() *Request {
	return &Request{
		state: Initialized,
		Headers: headers.NewHeaders(),
	}
}

const CRLF = "\r\n"

func RequestFromReader(reader io.Reader) (*Request, error) {
	request := newRequest()
	var data []byte
	var unparsed []byte
	var parsedData []byte
	lengthRead := 0
	lengthParsed := 0

	for !(request.state == Done) {
		bytesRead := make([]byte, 1024)
		len, err := reader.Read(bytesRead)
		if err != nil && !(err == io.EOF && request.state == ParsingBody) {
			return nil, err
		}
		lengthRead += len

		// _, err = data.Write(bytesRead[0:len])
		data = append(data, bytesRead[0:len]...)
		unparsed = append(unparsed, bytesRead[0:len]...)

		if(request.state == ParsingBody && err == io.EOF) || request.state != ParsingBody {
			parsedLength, parseErr := request.parse(unparsed)
			parsedData = append(parsedData, unparsed[:parsedLength]...)
			if parseErr != nil {
				return nil, parseErr
			}
			if err == io.EOF && request.state != Done {
				return nil, fmt.Errorf("Received Body of length smaller than content length")
			}
			unparsed = unparsed[parsedLength:]

			lengthParsed += parsedLength
		}

		if request.state == Done {
			break
		}
	}

	return request, nil
}

func parseRequestLine(request *Request, data string) (int, error) {
	reqLine, _, found := strings.Cut(data, CRLF)
	if !found {
		return 0, nil
	}

	reqLineElements := strings.Split(reqLine, " ")
	if len(reqLineElements) != 3 {
		return 0, fmt.Errorf("Start line of length: %d has too few or too many strings", len(reqLineElements))
	}

	if !validateMethod(reqLineElements[0]) {
		return 0, fmt.Errorf("Request Method %s is not uppercase", request.RequestLine.Method)
	}

	verString := strings.Split(reqLineElements[2], "/")
	if !validateVersion(verString) {
		return 0, fmt.Errorf("Request Version %s is not valid", request.RequestLine.HttpVersion)
	}

	request.RequestLine.Method = reqLineElements[0]
	request.RequestLine.RequestTarget = reqLineElements[1]
	request.RequestLine.HttpVersion = verString[1]
	return len([]byte(reqLine)) + len(CRLF), nil
}

func validateMethod(method string) bool {
	return strings.ToUpper(method) == method
}

func validateVersion(version []string) bool {
	if len(version) == 2 && version[0] == "HTTP" && version[1] == "1.1" {
		return true
	}
	return false
}

func (r *Request) parse(data []byte) (int, error) {
	parsedLen := 0
outer:
	for {
		switch r.state {
		case Initialized:
			parsedLength, err := parseRequestLine(r, string(data))
			if err != nil {
				return parsedLength, err
			}
			if parsedLength == 0 {
				break outer
			}
			r.state = ParsingHeaders
			parsedLen += parsedLength
			data = data[parsedLength:]

		case ParsingHeaders:
			len, done, err := r.Headers.Parse(data)
			if err != nil {
				return parsedLen, err
			}

			if len == 0 {
				break outer
			}

			data = data[len:]
			parsedLen += len
			if done {
				r.state = ParsingBody
			}

		case ParsingBody:
			cl, isPresent := r.Headers.Get("Content-Length")
			if !isPresent {
				r.state = Done
				continue
			}

			length, err := checkContentLength(cl)
			if err != nil {
				return parsedLen, err
			}
			if length == 0 {
				r.state = Done
				continue
			}

			r.Body = append(r.Body, data...)
			parsedLen += len(data)
			data = data[len(data):]

			if len(r.Body) > length {
				return parsedLen, fmt.Errorf("Received Body of length greater than content length")
			} else if len(r.Body) == length {
				r.state = Done
			} else {
				break outer
			}

		case Done:
			cl, _ := r.Headers.Get("Content-Length")
			length, err := checkContentLength(cl)
			if err != nil {
				return parsedLen, err
			}
			if length > 0 && len(data) > 0 {
				return parsedLen, fmt.Errorf("Received Body of length greater than content length")
			}
			break outer

		default:
			return 0, fmt.Errorf("Err: Unknown state")

		}
	}
	return parsedLen, nil
}

func checkContentLength(cl string) (int, error) {
	if len(cl) == 0 {
		return 0, nil
	}

	len, err := strconv.Atoi(cl)
	if err != nil {
		return 0, err
	} else {
		return len, nil
	}
}