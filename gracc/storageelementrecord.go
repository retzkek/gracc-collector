package gracc

import (
	"encoding/json"
	"encoding/xml"
	"time"
)

type StorageElementRecord struct {
	XMLName   xml.Name  `xml:"StorageElementRecord"`
	UniqueID  string    `xml:",omitempty"`
	Timestamp time.Time `xml:",omitempty"`
	Origin    origin    `xml:,omitempty"`
	Fields    []field   `xml:",any"`
	raw       []byte    `xml:"-"`
}

func (ser *StorageElementRecord) ParseXML(xb []byte) error {
	if err := xml.Unmarshal(xb, ser); err != nil {
		return err
	}
	ser.raw = append(ser.raw, xb...) // copy contents
	return nil
}

// Id returns an identification string for the record.
func (ser *StorageElementRecord) Id() string {
	return ser.UniqueID
}

// Raw returns the unaltered source of the record.
func (ser *StorageElementRecord) Raw() []byte {
	return ser.raw
}

// ToJSON returns a JSON encoding of the Record, with certain elements
// transformed to fit the GRACC Raw Record schema.
// Indent specifies the string to use for each indentation level,
// if empty no indentation or pretty-printing is performed.
func (ser *StorageElementRecord) ToJSON(indent string) ([]byte, error) {
	var r = make(map[string]interface{})

	r["UniqueID"] = ser.UniqueID

	// Standard time instants
	if !ser.Timestamp.IsZero() {
		r["Timestamp"] = ser.Timestamp.Format(time.RFC3339)
	}

	// flatten other fields
	for _, f := range ser.Fields {
		for k, v := range f.flatten() {
			r[k] = v
		}
	}

	// origin
	for k, v := range ser.Origin.flatten() {
		r[k] = v
	}

	// add XML
	r["RawXML"] = string(ser.Raw())

	if indent != "" {
		return json.MarshalIndent(r, "", indent)
	}
	return json.Marshal(r)
}
