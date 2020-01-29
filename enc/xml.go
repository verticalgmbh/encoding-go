package enc

import "encoding/xml"

// GetXMLAttribute retrieves an attribute by name from an attribute array
//
// **Parameters**
//   - attributes: collection of attributes of which to find a named attribute
//   - name      : name of attribute to find
//
// **Returns**
//   - *xml.Attr: Attribute with the specified name if any was found, nil otherwise
func GetXMLAttribute(attributes []xml.Attr, name string) *xml.Attr {
	for _, attribute := range attributes {
		if attribute.Name.Local == name {
			return &attribute
		}
	}

	return nil
}
