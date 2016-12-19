package gracc

import (
	"encoding/xml"
)

// RecordBundle is a structure for unmarshalling a bundle of records, as would
// be sent from a probe.
type RecordBundle struct {
	XMLName               xml.Name               `xml:"RecordEnvelope"`
	JobUsageRecords       []JobUsageRecord       `xml:",omitempty"`
	StorageElements       []StorageElement       `xml:",omitempty"`
	StorageElementRecords []StorageElementRecord `xml:",omitempty"`
	OtherRecords          []XMLRecord            `xml:",omitempty"`
}

// XMLRecord is a generic structure for unmarshalling unknown XML data.
type XMLRecord struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
}
