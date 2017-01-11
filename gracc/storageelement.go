package gracc

import (
	"encoding/json"
	"encoding/xml"
	"time"
)

// StorageElement is a flexible container for distributed storage element information.
type StorageElement struct {
	XMLName   xml.Name
	UniqueID  string    `xml:",omitempty"`
	Timestamp time.Time `xml:",omitempty"`
	Origin    origin    `xml:,omitempty"`
	Fields    []field   `xml:",any"`
	raw       []byte    `xml:"-"`
}

// ParseXML attempts to unmarshal the XML in xb into a StorageElement.
func (se *StorageElement) ParseXML(xb []byte) error {
	if err := xml.Unmarshal(xb, se); err != nil {
		return err
	}
	se.raw = append(se.raw, xb...) // copy contents
	return nil
}

// Id returns an identification string for the record.
func (se *StorageElement) Id() string {
	return se.UniqueID
}

// Type returns the type of the record.
func (se *StorageElement) Type() string {
	return se.XMLName.Local
}

// Raw returns the unaltered source of the record.
func (se *StorageElement) Raw() []byte {
	return se.raw
}

// ToJSON returns a JSON encoding of the Record, with certain elements
// transformed to fit the GRACC Raw Record schema.
// Indent specifies the string to use for each indentation level,
// if empty no indentation or pretty-printing is performed.
func (se *StorageElement) ToJSON(indent string) ([]byte, error) {
	var r = make(map[string]interface{})

	r["type"] = "StorageElement"

	r["UniqueID"] = se.UniqueID

	// Standard time instants
	if !se.Timestamp.IsZero() {
		r["Timestamp"] = se.Timestamp.Format(time.RFC3339)
	}

	// flatten other fields
	for _, f := range se.Fields {
		for k, v := range f.flatten() {
			r[k] = v
		}
	}

	// origin
	for k, v := range se.Origin.flatten() {
		r[k] = v
	}

	// add XML
	r["RawXML"] = string(se.Raw())

	if indent != "" {
		return json.MarshalIndent(r, "", indent)
	}
	return json.Marshal(r)
}
