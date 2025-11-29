package request

import (
	"bytes"
	"fmt"
	"io"
	"strconv"

	"github.com/aabbcc333/httpfromtcp/internal/headers"
)

type RequestLine struct {
	HttpVersion   string
	RequestTarget string
	Method        string
	
}

type parserState string

const (
	StateInit    parserState = "init"
	StateHeaders parserState = "headers"
	StateBody    parserState = "body"
	StateDone    parserState = "done"
	StateError   parserState = "error"
)

type Request struct {
	RequestLine RequestLine
	Headers     *headers.Headers
	state       parserState
	Body string
}
func getInt(headers *headers.Headers, name string, defaultValue int) int{
	valueStr , exists := headers.Get(name)
	if !exists{
		return defaultValue
	}
	value , err := strconv.Atoi(valueStr)
	if err != nil {
		return defaultValue
	}
	return value 
}

func newRequest() *Request {
	return &Request{
		state:   StateInit,
		Headers: headers.NewHeaders(),
	}
}

var ERROR_MALFORMED_REQUEST_LINE = fmt.Errorf("malformed request line")
var ERROR_UNSUPPORTED_HTTP_VERSION = fmt.Errorf("unsuppored http version")
var ErrorRequestInErrorSTate = fmt.Errorf("error in request state")

var SEPRATOR = []byte("\r\n")

// parseRequestLine parses the HTTP request line from the given buffer.
// It returns the parsed RequestLine, the number of bytes consumed from b,
// and an error if the request line is malformed. If no full line is
// available yet (no CRLF found), it returns (nil, 0, nil).
func parseRequestLine(b []byte) (*RequestLine, int, error) {
	idx := bytes.Index(b, SEPRATOR)
	if idx == -1 {
		// no full line yet
		return nil, 0, nil
	}

	startLine := b[:idx]
	consumed := idx + len(SEPRATOR)

	parts := bytes.Split(startLine, []byte(" "))
	if len(parts) != 3 {
		return nil, consumed, ERROR_MALFORMED_REQUEST_LINE
	}

	httpParts := bytes.Split(parts[2], []byte("/"))
	if len(httpParts) != 2 || string(httpParts[0]) != "HTTP" {
		return nil, consumed, ERROR_MALFORMED_REQUEST_LINE
	}

	rl := &RequestLine{
		Method:        string(parts[0]),
		RequestTarget: string(parts[1]),
		HttpVersion:   string(httpParts[1]),
	}

	return rl, consumed, nil
}

func (r *Request) hasBody() bool{
	//Todo: update this for chunked enocding
	length := getInt(r.Headers, "content-length", 0)
	return length > 0 
}

// parse consumes data from the buffer according to the parser state.
func (r *Request) parse(data []byte) (int, error) {
	read := 0
	currentData := data[read:]
	 
	switch r.state {
	case StateError:
		return 0, ErrorRequestInErrorSTate

	case StateInit:
		rl, n, err := parseRequestLine(currentData)
		if err != nil {
			r.state = StateError
			return 0, err
		}
		if n == 0 {
			// need more data
			return 0, nil
		}
		r.RequestLine = *rl
		r.state = StateHeaders
		return n, nil

	case StateHeaders:
		n, done, err := r.Headers.Parse(currentData)
		if err != nil {
			r.state = StateError
			return 0, err
		}

		if n == 0 {
			return 0, nil
		}
		read += n
		if done {
			if r.hasBody(){
				r.state = StateBody
			}else{
				r.state = StateDone
			}
			
		}
	case StateBody:
		length := getInt(r.Headers, "content-length", 0)
		if length == 0{
			panic("chunked not implementedk")
		}
		remaining := min(length - len(r.Body) , len(currentData))
		r.Body += string(currentData[:remaining])

		read += remaining

		if len(r.Body) == length{
			r.state = StateDone
		}
	case StateDone:
		return 0, nil

	default:
		panic("bad code ")
	}
	return read, nil
}

func (r *Request) done() bool {
	return r.state == StateDone || r.state == StateError
}

// RequestFromReader reads from reader incrementally until it has parsed
// a full HTTP request (request line + headers), then returns a Request.
func RequestFromReader(reader io.Reader) (*Request, error) {
	req := newRequest()

	// buf accumulates everything we've read so far
	buf := make([]byte, 0, 1024)
	// tmp is a scratch buffer for each Read call
	tmp := make([]byte, 256)

	for !req.done() {
		n, err := reader.Read(tmp)
		if n > 0 {
			// append new bytes to buffer
			buf = append(buf, tmp[:n]...)

			// keep parsing as long as we consume data and are not done
			for !req.done() {
				consumed, perr := req.parse(buf)
				if perr != nil {
					return nil, perr
				}
				if consumed == 0 {
					// need more data
					break
				}
				// drop consumed bytes; remaining bytes may contain more data
				buf = buf[consumed:]
				if len(buf) == 0 {
					break
				}
			}
		}

		if err != nil {
			if err == io.EOF {
				// Reader ended: if we still haven't parsed a full request,
				// treat as error; otherwise it's fine.
				if !req.done() {
					return nil, fmt.Errorf("incomplete request before EOF")
				}
				break
			}
			return nil, err
		}
	}

	return req, nil
}
