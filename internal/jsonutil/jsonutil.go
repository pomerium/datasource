package jsonutil

import (
	"bufio"
	"encoding/json"
	"io"
	"iter"
)

// A JSONArrayStream represents a streaming array of JSON objects.
type JSONArrayStream struct {
	buf               *bufio.Writer
	closed, wroteOnce bool
}

// NewJSONArrayStream creates a new JSONArrayStream.
func NewJSONArrayStream(dst io.Writer) *JSONArrayStream {
	return &JSONArrayStream{
		buf: bufio.NewWriter(dst),
	}
}

// Close writes the ']' and flushes the stream.
func (stream *JSONArrayStream) Close() error {
	if stream.closed {
		return nil
	}
	stream.closed = true

	if stream.wroteOnce {
		_, err := stream.buf.Write([]byte{'\n', ']'})
		if err != nil {
			return err
		}
	}
	return stream.buf.Flush()
}

// Encode adds an object to the stream.
func (stream *JSONArrayStream) Encode(obj any) error {
	bs, err := json.Marshal(obj)
	if err != nil {
		return err
	}

	if !stream.wroteOnce {
		stream.wroteOnce = true
		_, err := stream.buf.Write([]byte{'[', '\n'})
		if err != nil {
			return err
		}
	} else {
		_, err := stream.buf.Write([]byte{',', '\n'})
		if err != nil {
			return err
		}
	}

	_, err = stream.buf.Write(bs)
	return err
}

func StreamWriteArray[T any](w io.Writer, src iter.Seq2[T, error]) error {
	stream := NewJSONArrayStream(w)
	defer stream.Close()

	for v, err := range src {
		if err != nil {
			return err
		}
		err = stream.Encode(v)
		if err != nil {
			return err
		}
	}
	return nil
}
