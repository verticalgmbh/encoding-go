package enc

import (
	"fmt"
	"io"
	"log"
)

// XMLWriter writes xml data to a writer
type XMLWriter struct {
	writer      io.Writer // writer used to write xml data to
	trailingtag bool      // indicator whether a start tag is still open
	tags        []string  // collection of open tags
}

// NewXMLWriter creates a new XMLWriter
//
// **Parameters**
//   - writer: writer used to write data
//
// **Returns**
//   - *XMLWriter: created xml writer
func NewXMLWriter(writer io.Writer) *XMLWriter {
	return &XMLWriter{
		writer:      writer,
		trailingtag: false,
		tags:        make([]string, 8)}
}

func (writer *XMLWriter) checkTrailingTag() {
	if writer.trailingtag {
		io.WriteString(writer.writer, ">")
		writer.trailingtag = false
	}
}

// BeginTag starts a new xml tag
//
// **Parameters**
//   - name: name of xml tag
//
// **Returns**
//   - *XMLWriter: this writer for fluent behavior
func (writer *XMLWriter) BeginTag(name string) *XMLWriter {
	writer.checkTrailingTag()

	io.WriteString(writer.writer, fmt.Sprintf("<%s", name))
	writer.tags = append(writer.tags, name)
	writer.trailingtag = true
	return writer
}

// CloseTag closes the currently open tag
//
// **Returns**
//   - *XMLWriter: this writer for fluent behavior
func (writer *XMLWriter) CloseTag() *XMLWriter {
	length := len(writer.tags)
	if length == 0 {
		log.Print("Tried to close a tag for XmlWriter with no open tag left")
		return writer
	}

	if writer.trailingtag {
		io.WriteString(writer.writer, "/>")
		writer.trailingtag = false
	} else {
		tagname := writer.tags[length-1]
		io.WriteString(writer.writer, fmt.Sprintf("</%s>", tagname))
	}

	writer.tags = writer.tags[:length-1]
	return writer
}

// WriteAttribute writes an attribute to an open tag
//
// **Parameters**
//   - key:	  attribute name
//   - value: attribute value
//
// **Returns**
//   - *XMLWriter: this writer for fluent behavior
func (writer *XMLWriter) WriteAttribute(key string, value string) *XMLWriter {
	if !writer.trailingtag {
		log.Print(fmt.Sprintf("ERR: tried to write attribute '%s=%s' to a closed tag", key, value))
		return writer
	}

	io.WriteString(writer.writer, fmt.Sprintf(" %s=\"%s\"", key, value))
	return writer
}

// WriteContent writes content data of a tag
//
// **Parameters**
//   - content: content to write
//
// **Returns**
//   - *XMLWriter: this writer for fluent behavior
func (writer *XMLWriter) WriteContent(content string) *XMLWriter {
	writer.checkTrailingTag()

	io.WriteString(writer.writer, content)
	return writer
}
