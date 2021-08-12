package formats

import (
	"bytes"
	"fmt"
)

func QuoteString(writer *bytes.Buffer, input []byte) {
	sanitized := bytes.ReplaceAll(input, []byte("\n"), []byte("\\n"))
	sanitized = bytes.ReplaceAll(sanitized, []byte("\r"), []byte("\\r"))
	writer.WriteByte(quotes)
	var occurrenceIndex = bytes.IndexByte(sanitized, quotes)
	var remaining = sanitized
	for occurrenceIndex != -1 {
		if occurrenceIndex > 0 {
			writer.Write(remaining[:occurrenceIndex])
		}
		writer.WriteByte(quotes)
		writer.WriteByte(quotes)
		remaining = remaining[occurrenceIndex+1:]
		occurrenceIndex = bytes.IndexByte(remaining, quotes)
	}
	writer.Write(remaining)
	writer.WriteByte(quotes)
}
func AppendWithQuote(buffer *bytes.Buffer, value interface{}) {
	switch v := value.(type) {
	case []byte:
		if len(v) > 0 {
			QuoteString(buffer, v)
		}
	case string:
		if len(v) > 0 {
			QuoteString(buffer, []byte(v))
		}
	case Json:
		toBytes := v.ToBytes()
		if len(toBytes) > 2 {
			QuoteString(buffer, toBytes)
		}
	default:
		buffer.WriteString(fmt.Sprintf("%v", v))
	}
}

var openPar = []byte(`{`)
var openWithComma = []byte(`,"`)
var keyVal = []byte(`":`)
var quotes byte = '"'
var endPar = []byte(`}`)

type Json struct {
	started bool
	closed  bool
	buffer  *bytes.Buffer
}

func NewJson() Json {
	return Json{
		buffer: &bytes.Buffer{},
	}
}
func (js *Json) HasContent() bool {
	return js.started
}
func (js *Json) Close() {
	if js.started && !js.closed {
		js.buffer.Write(endPar)
		js.closed = true
	}
}

func (js *Json) ToString() string {
	js.Close()
	return js.buffer.String()
}

func (js *Json) ToBytes() []byte {
	js.Close()
	return js.buffer.Bytes()
}

func (js *Json) Add(key []byte, val interface{}) {
	if val != nil {
		js.prepareForKey()
		js.buffer.Write(key)
		js.buffer.Write(keyVal)
		AppendWithQuote(js.buffer, val)
	}
}

func (js *Json) prepareForKey() {
	if js.started {
		js.buffer.Write(openWithComma)
	} else {
		js.started = true
		js.buffer.Write(openPar)
		js.buffer.WriteByte(quotes)
	}
}
