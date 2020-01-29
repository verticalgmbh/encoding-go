package enc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriteValidXML(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewXMLWriter(&buffer)

	writer.BeginTag("root")
	writer.WriteAttribute("name", "karl")
	writer.WriteAttribute("other", "some")
	writer.WriteContent("Suffer")
	writer.CloseTag()

	assert.Equal(t, `<root name="karl" other="some">Suffer</root>`, buffer.String())
}
