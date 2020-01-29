package enc

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestData struct {
	Name   string
	Number int
	Data   interface{}
}

func assertPanic(t *testing.T, f func()) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	f()
}

func TestWriteValidJson(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginObject()
	writer.WriteProperty("firstname", "Diether")
	writer.WriteProperty("lastname", "Boffel")
	writer.WriteProperty("age", 35)
	writer.WriteKey("contact")
	writer.BeginObject()
	writer.WriteProperty("email", "d.boffel@fims.it")
	writer.EndObject()
	writer.EndObject()

	assert.Equal(t, `{"firstname":"Diether","lastname":"Boffel","age":35,"contact":{"email":"d.boffel@fims.it"}}`, buffer.String())
}

func TestCloseOpenStructure(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginArray()
	writer.BeginObject()
	writer.WriteKey("key")
	writer.Close()

	assert.Equal(t, `[{"key":null}]`, buffer.String())
}

func TestCloseClosedWriter(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginArray()
	writer.EndArray()
	writer.Close()
}

func TestKeyAtEmptyFails(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	assertPanic(t, func() { writer.WriteKey("anything") })
}

func TestKeyAtArrayFails(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)
	writer.BeginArray()
	assertPanic(t, func() { writer.WriteKey("anything") })
}

func TestKeyAtKeyFails(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginObject()
	writer.WriteKey("something")
	assertPanic(t, func() { writer.WriteKey("anything") })
}

func TestObjectItemClosesParentKey(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginObject()
	writer.WriteKey("something")
	writer.BeginObject()
	writer.EndObject()
	writer.WriteKey("another")
	writer.BeginObject()
	writer.EndObject()
	writer.EndObject()

	assert.Equal(t, `{"something":{},"another":{}}`, buffer.String())
}

func TestWriteObjectItem(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginObject()
	writer.WriteKey("test")
	writer.WriteItem(&TestData{
		Name:   "Test",
		Number: 8})
	writer.EndObject()

	assert.Equal(t, `{"test":{"name":"Test","number":8}}`, buffer.String())
}

func TestWriteObjectItemWithSubObjects(t *testing.T) {
	var buffer bytes.Buffer
	writer := NewJSONWriter(&buffer)

	writer.BeginObject()
	writer.WriteKey("test")
	writer.WriteItem(&TestData{
		Name:   "Test",
		Number: 8,
		Data: &TestData{
			Name:   "SubTest",
			Number: 12}})
	writer.EndObject()

	assert.Equal(t, `{"test":{"name":"Test","number":8,"data":{"name":"SubTest","number":12}}}`, buffer.String())
}
