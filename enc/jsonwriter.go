package enc

import (
	"fmt"
	"io"
	"log"
	"reflect"
	"strings"
	"unicode"
)

// JSONState current state of JSON writer
type JSONState int8

const (
	// JSONStateNone no structure written yet. Also used as item in compatibility matrix
	JSONStateNone JSONState = iota

	// JSONStateObject json structure currently in object state
	JSONStateObject

	// JSONStateArray json structure currently in array state
	JSONStateArray

	// JSONStateKey key of a json object
	JSONStateKey
)

var jsoncompatibility [16]bool = [...]bool{
	true, true, true, false,
	false, false, false, true,
	true, true, true, false,
	true, true, true, false}

// JSONWriter structure containing contextual data for a json stream
type JSONWriter struct {
	writer            io.Writer // writer used to write json data
	trailingseparator bool      // flag used to determine whether a separator is to be written before starting a new item
	structure         int64
	depth             int
}

// NewJSONWriter creates a new JSONWriter
//
// **Parameters**
//   - writer: writer to write json data to
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func NewJSONWriter(writer io.Writer) *JSONWriter {
	return &JSONWriter{
		writer:            writer,
		trailingseparator: false,
		structure:         0,
		depth:             0}
}

// JSONStateName converts a JSONState to a readable string
//
// **Parameters**
//   - jsontype: state type to convert
//   - isitem: determines whether the structure type to convert represents an item or current structure state
//
// **Returns**
//   - string: string representation of state
func JSONStateName(jsontype JSONState, isitem bool) string {
	switch jsontype {
	case JSONStateNone:
		if isitem {
			return "Item"
		}
		return "None"
	case JSONStateObject:
		return "Object"
	case JSONStateArray:
		return "Array"
	case JSONStateKey:
		return "Property Key"
	default:
		panic("Invalid json structure type")
	}
}

func (writer *JSONWriter) begin(jsontype JSONState) {
	if !jsoncompatibility[int(writer.structure&3)*4+int(jsontype)] {
		panic(fmt.Sprintf("Writing a %s is not valid when current state is %s", JSONStateName(jsontype, true), JSONStateName(JSONState(writer.structure&3), false)))
	}

	if writer.depth == -1 {
		panic("json structure already closed")
	}

	if writer.depth >= 31 {
		panic("This json writer can't be used to write json structures with a depth of more than 31")
	}

	if jsontype != JSONStateNone {
		writer.structure = (writer.structure << 2) | int64(jsontype)
		writer.depth++
	}

	if writer.trailingseparator {
		io.WriteString(writer.writer, ",")
		writer.trailingseparator = false
	}
}

func (writer *JSONWriter) end() {
	writer.structure = (writer.structure >> 2)
	writer.depth--

	if writer.depth > 0 && JSONState(writer.structure&3) == JSONStateKey {
		writer.structure = (writer.structure >> 2)
		writer.depth--
	}

	if writer.depth == 0 {
		writer.depth = -1
	}
}

func (writer *JSONWriter) escape(value string) string {
	var builder strings.Builder
	builder.WriteString("\"")
	for _, char := range value {
		switch char {
		default:
			if char < 0x20 {
				builder.WriteString(fmt.Sprintf("\\u%04d", char))
			} else {
				builder.WriteRune(char)
			}
		case '\\':
			builder.WriteString("\\\\")
		case '"':
			builder.WriteString("\\\"")
		}
	}
	builder.WriteString("\"")
	return builder.String()
}

// BeginObject starts writing a json object
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) BeginObject() *JSONWriter {
	writer.begin(JSONStateObject)
	io.WriteString(writer.writer, "{")
	return writer
}

// EndObject ends writing a json object
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) EndObject() *JSONWriter {
	writer.end()
	io.WriteString(writer.writer, "}")
	writer.trailingseparator = true
	return writer
}

// BeginArray starts writing an array
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) BeginArray() *JSONWriter {
	writer.begin(JSONStateArray)
	io.WriteString(writer.writer, "[")
	return writer
}

// EndArray terminates an open array
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) EndArray() *JSONWriter {
	writer.end()
	io.WriteString(writer.writer, "]")
	writer.trailingseparator = true
	return writer
}

// WriteKey writes a property key
//
// **Parameters**
//   - key: name of key to write
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) WriteKey(key string) *JSONWriter {
	writer.begin(JSONStateKey)
	io.WriteString(writer.writer, fmt.Sprintf("%s:", writer.escape(key)))
	return writer
}

// WriteProperty writes a property of a json object. Basically a wrapper for WriteKey and WriteItem
//
// **Parameters**
//   - name : name of the property
//   - value: property value
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) WriteProperty(name string, value interface{}) *JSONWriter {
	// don't write null values
	if value == nil {
		return writer
	}

	writer.WriteKey(name)
	writer.WriteItem(value)
	return writer
}

// Close closes all open structures of the json writer
func (writer *JSONWriter) Close() {
	for writer.depth > 0 {
		switch JSONState(writer.structure & 3) {
		case JSONStateObject:
			writer.EndObject()
		case JSONStateArray:
			writer.EndArray()
		case JSONStateKey:
			writer.WriteItem(nil)
		default:
			panic("Invalid json state")
		}
	}
}

func (writer *JSONWriter) toCamelCase(data string) string {
	if len(data) == 0 {
		return data
	}

	var firstrune rune
	for _, firstrune = range data {
		break
	}

	if unicode.IsLower(firstrune) {
		return data
	}

	return string(unicode.ToLower(firstrune)) + data[1:]
}

func (writer *JSONWriter) writeObject(item interface{}) {
	writer.BeginObject()
	typeinfo := reflect.TypeOf(item)
	if typeinfo.Kind() == reflect.Ptr {
		typeinfo = typeinfo.Elem()
	}

	typevalue := reflect.ValueOf(item)
	if typevalue.Kind() == reflect.Ptr {
		typevalue = typevalue.Elem()
	}

	for i := 0; i < typeinfo.NumField(); i++ {
		fieldvalue := typevalue.Field(i)
		if fieldvalue.Kind() == reflect.Ptr && fieldvalue.IsNil() {
			continue
		}

		field := typeinfo.Field(i)
		writer.WriteProperty(writer.toCamelCase(field.Name), fieldvalue.Interface())
	}
	writer.EndObject()
}

// WriteItem writes an arbitrary item
//
// **Parameters**
//   - item: item to write
//
// **Returns**
//   - *JSONWriter: created JSONWriter
func (writer *JSONWriter) WriteItem(item interface{}) *JSONWriter {
	writer.begin(JSONStateNone)

	if item == nil {
		io.WriteString(writer.writer, "null")
	} else {
		switch reflect.TypeOf(item).Kind() {
		case reflect.Slice, reflect.Array:
			writer.BeginArray()
			value := reflect.ValueOf(item)
			for i := 0; i < value.Len(); i++ {
				writer.WriteItem(value.Index(i).Interface())
			}
			writer.EndArray()
		case reflect.Map, reflect.Chan:
			log.Printf("Maps and Channels are not supported for now")
		default:
			switch item.(type) {
			default:
				writer.writeObject(item)
			case bool:
				if item.(bool) {
					io.WriteString(writer.writer, "true")
				} else {
					io.WriteString(writer.writer, "false")
				}
			case string:
				io.WriteString(writer.writer, writer.escape(item.(string)))
			case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
				io.WriteString(writer.writer, fmt.Sprintf("%d", item))
			case float32, float64:
				io.WriteString(writer.writer, fmt.Sprintf("%f", item))
			case complex64, complex128:
				log.Printf("Complex number items not supported")
			}
			writer.trailingseparator = true
		}
	}

	if JSONState(writer.structure&3) == JSONStateKey {
		writer.end()
	}
	return writer
}
