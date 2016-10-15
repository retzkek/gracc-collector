package gracc

import (
	"fmt"
)

type Record interface {
	Id() string
	ToJSON(indent string) ([]byte, error)
	Raw() []byte
}

// ParseRecordXML will attempt to unmarshall the XML in buf into one of the
// record types, trying each in succession.
func ParseRecordXML(buf []byte) (Record, error) {
	var jur JobUsageRecord
	if err := jur.ParseXML(buf); err == nil {
		return &jur, nil
	}
	var se StorageElement
	if err := se.ParseXML(buf); err == nil {
		return &se, nil
	}
	var ser StorageElementRecord
	if err := ser.ParseXML(buf); err == nil {
		return &ser, nil
	}
	return nil, fmt.Errorf("unable to unmarshall XML into record")
}
