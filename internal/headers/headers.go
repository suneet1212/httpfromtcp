package headers

import (
	"bytes"
	"fmt"
	"strings"
	"unicode"
)

const CRLF = "\r\n"
const VALID_HEADER_KEY_SPECIAL_CHARS = "!#$%&'*+-.^_`|~"

type Headers map[string]string

func NewHeaders() Headers {
	var h Headers = make(Headers)
	return h
}

func (h Headers) Get(key string) (string, bool) {
	key = strings.ToLower(key)
	str, isPresent := h[key]
	return str, isPresent
}

func (h Headers) Put(key, value string) {
	finalKey := strings.ToLower(key)
	_, keyExists := h[finalKey]
	if keyExists {
		h[finalKey] = h[finalKey] + ", " + value
	} else {
		h[finalKey] = value
	}
}

func (h Headers) Replace(key, value string) {
	h[key] = value
}

func (h Headers) Remove(key string) {
	key = strings.ToLower(key)
	delete(h, key)
}

func (h Headers) Parse(data []byte) (n int, done bool, err error) {
	n = 0
	done = false
	err = nil

	header, _, found := bytes.Cut(data, []byte(CRLF))
	if !found {
		return
	}

	if len(header) == 0 {
		// found end of header
		done = true
		n += len(CRLF)
		return
	}

	headerArray := bytes.SplitN(header, []byte(":"), 2)
	if len(headerArray) != 2 {
		err = fmt.Errorf("Invalid Header")
		return
	}

	key := bytes.TrimLeftFunc(headerArray[0], unicode.IsSpace)
	value := bytes.TrimFunc(headerArray[1], unicode.IsSpace) 

	if len(key) == 0 {
		err = fmt.Errorf("Header Key is of 0 length")
		return
	}

	if unicode.IsSpace(rune(key[len(key)-1])) {
		err = fmt.Errorf("Header Key cannot be followed by a whitespace character")
		return
	}

	isInvalid := bytes.ContainsFunc(key, isInvalidHeaderKeyRune)
	if isInvalid {
		err = fmt.Errorf("Header Key contains an invalid character")
		return
	}

	h.Put(string(key), string(value))
	
	n = len(header) + len(CRLF)
	return
}

func isInvalidHeaderKeyRune(r rune) bool {
	return !unicode.IsLetter(r) && !unicode.IsNumber(r) && !strings.ContainsRune(VALID_HEADER_KEY_SPECIAL_CHARS, r)
}