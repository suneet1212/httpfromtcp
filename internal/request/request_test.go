package request

import (
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	inputDataStrs := [...]string{
		"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		"GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		"POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n -d '{\"flavor\":\"dark mode\"}'",
		"POST / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n -d '{\"flavor\":\"dark mode\"}'",
		"/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		"get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		"GET /coffee HTTP/2.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
	}
	var maxSize int;
	for _, str := range inputDataStrs {
		maxSize = max(maxSize, len(str))
	}

	for byteSize := 1; byteSize < maxSize+5; byteSize += 5 {
		t.Logf("Starting to test for byteSize %d", byteSize)
		// Test: Good GET Request line
		reader := &chunkReader{
			data: "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

		// Test: Good GET Request line with path
		reader = &chunkReader{
			data: "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "GET", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

		// Test: Good POST Request line with path
		reader = &chunkReader{
			data:"POST /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n -d '{\"flavor\":\"dark mode\"}'",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

		// Test: Good POST Request line 
		reader = &chunkReader{
			data:"POST / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n -d '{\"flavor\":\"dark mode\"}'",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "POST", r.RequestLine.Method)
		assert.Equal(t, "/", r.RequestLine.RequestTarget)
		assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

		// Test: Invalid number of parts in request line
		reader = &chunkReader{
			data:"/coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)

		// Test: Invalid Method
		reader = &chunkReader{
			data:"get /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)

		// Test: Invalid Version
		reader = &chunkReader{
			data:"GET /coffee HTTP/2.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)
	}
}

func TestHeaders(t *testing.T) {
	inputDataStrs := [...]string{
		"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nUser-Agent: Suneet\r\n\r\n",
		"GET / HTTP/1.1\r\n\r\n",
		"GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nuser-agent: Suneet\r\n",
		"GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
	}
	var maxSize int;
	for _, str := range inputDataStrs {
		maxSize = max(maxSize, len(str))
	}

	for byteSize := 1; byteSize < maxSize+5; byteSize += 5 {
		// Test: Standard Headers
		reader := &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])

		// Test: Empty Header
		reader = &chunkReader{
			data:            "GET / HTTP/1.1\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, 0, len(r.Headers))

		// Test: Duplicate Headers
		reader = &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nUser-Agent: Suneet\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0, Suneet", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])

		// Test: Case Insensitive Header
		reader = &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nuser-agent: Suneet\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "localhost:42069", r.Headers["host"])
		assert.Equal(t, "curl/7.81.0, Suneet", r.Headers["user-agent"])
		assert.Equal(t, "*/*", r.Headers["accept"])

		// Test: Malformed Header
		reader = &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost localhost:42069\r\n\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.Error(t, err)

		// Test: Missing End of Headers
		reader = &chunkReader{
			data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\nuser-agent: Suneet\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.Error(t, err)
	}
}

func TestBody(t *testing.T) {
	inputDataStrs := [...]string{
		"POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
		"POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
		"POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content sdfghjkuytrertyuioiasjhd",
	}
	var maxSize int;
	for _, str := range inputDataStrs {
		maxSize = max(maxSize, len(str))
	}

	for byteSize := 1; byteSize < maxSize+5; byteSize += 5 {
		// Test: Standard Body
		reader := &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 13\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: byteSize,
		}
		r, err := RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "hello world!\n", string(r.Body))

		// Test: Empty Body, 0 reported content length
		reader = &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 0\r\n" +
				"\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))

		// Test: Empty Body, No reported content length
		reader = &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))

		// // Test: Body shorter than reported content length
		reader = &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.Error(t, err)

		// Test: No content length but body exists
		reader = &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"\r\n" +
				"hello world!\n",
			numBytesPerRead: byteSize,
		}
		r, err = RequestFromReader(reader)
		require.NoError(t, err)
		require.NotNil(t, r)
		assert.Equal(t, "", string(r.Body))

		// Test: Body longer then reported length
		reader = &chunkReader{
			data: "POST /submit HTTP/1.1\r\n" +
				"Host: localhost:42069\r\n" +
				"Content-Length: 20\r\n" +
				"\r\n" +
				"partial content sdfghjkuytrertyuioiasjhd",
			numBytesPerRead: byteSize,
		}
		_, err = RequestFromReader(reader)
		require.Error(t, err)
	}

}