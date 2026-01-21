package response

import (
	"bytes"
	"fmt"
	"httpfromtcp/internal/headers"
	"strconv"
)

type StatusCode int
type WriterState string

const (
	StatusOk StatusCode = 200
	StatusBadRequest StatusCode = 400
	StatusInternalServerError StatusCode = 500

	StateReset WriterState = "Reset"
	StateStatusLineDone WriterState = "Status Line Done"
	StateHeadersDone WriterState = "Headers Completed"
	StateCompleted WriterState = "Completed"
)

type Writer struct {
	buffer bytes.Buffer
	state WriterState
}

func NewWriter() Writer {
	var buff bytes.Buffer
	writer := Writer {
		buffer: buff,
		state: StateReset,
	}
	return writer
}

func (w *Writer) ReadBuffer() string {
	return w.buffer.String()
}

func (w *Writer) Write(data []byte) (int, error) {
	n, err := w.buffer.Write(data)
	return n, err
}

func (w *Writer) WriteStatusLine(statusCode StatusCode) error {
	if w.state != StateReset {
		return fmt.Errorf("Cannot write status line - status is %s", w.state)
	}
	var statusLine string
	switch statusCode {
	case StatusOk:
		statusLine = "HTTP/1.1 200 OK\r\n"
	case StatusBadRequest:
		statusLine = "HTTP/1.1 400 Bad Request\r\n"
	case StatusInternalServerError:
		statusLine = "HTTP/1.1 500 Internal Server Error\r\n"
	default:
		break
	}

	_, err := w.Write([]byte(statusLine))
	if err != nil {
		return err
	}
	w.state = StateStatusLineDone
	return nil
}

func GetDefaultHeader(contentLen int) headers.Headers {
	headerList := headers.NewHeaders()
	headerList.Put("Content-Length", strconv.Itoa(contentLen))
	headerList.Put("Connection", "close")
	headerList.Put("Content-Type", "text/plain")
	return headerList
}

func (w *Writer) WriteHeaderValues(headers headers.Headers) error {
	if headers == nil {
		headers = GetDefaultHeader(0)
	}
	for key, value := range headers {
		_, err := fmt.Fprintf(w, "%s: %s\r\n", key, value)
		if err != nil {
			return err
		}
	}
	_, err := w.Write([]byte("\r\n"))
	if err != nil {
		return err
	}
	return nil
}

func (w *Writer) WriteHeaders(headers headers.Headers) error {
	if w.state != StateStatusLineDone {
		return fmt.Errorf("Cannot write headers - status is %s", w.state)
	}

	err := w.WriteHeaderValues(headers)
	if err != nil {
		return err
	}
	w.state = StateHeadersDone
	return nil
}

func (w *Writer) WriteBody(body string) (int, error) {
	if w.state != StateHeadersDone {
		return 0, fmt.Errorf("Cannot write response body - status is %s", w.state)
	}
	w.state = StateCompleted
	n, err := w.Write([]byte(body))
	if err != nil {
		return 0, err
	}

	w.state = StateCompleted
	return n, err
}

func (w *Writer) WriteChunkedBody(p []byte) (int, error) {
	if w.state != StateHeadersDone {
		return 0, fmt.Errorf("Cannot write response body - status is %s", w.state)
	}
	
	length := len(p)
	writeLen, err := fmt.Fprintf(w, "%X%s%s%s", length, headers.CRLF, p, headers.CRLF)
	return writeLen, err
}

func (w *Writer) WriteChunkedBodyDone() (int, error) {
	n, err := w.Write([]byte("0\r\n\r\n"))
	if err != nil {
		return 0, err
	}

	w.state = StateCompleted
	return n, nil
}

func (w *Writer) WriteResponse(statusCode StatusCode, msg string) {
	w.WriteStatusLine(statusCode)
	header := GetDefaultHeader(len(msg))
	w.WriteHeaders(header)
	w.WriteBody(msg)
}

func (w *Writer) WriteTrailers(h headers.Headers) error {
	if w.state != StateCompleted {
		return fmt.Errorf("Need to complete writing body before writing the trailers")
	}

	err := w.WriteHeaderValues(h)
	if err != nil {
		return err
	}

	return nil
}