package gracc

import (
	"encoding/xml"
)

// RecordBundle is a structure for unmarshalling a bundle of records, as would
// be sent from a probe.
type RecordBundle struct {
	XMLName               xml.Name               `xml:"RecordEnvelope"`
	JobUsageRecords       []JobUsageRecord       `xml:"JobUsageRecord,omitempty"`
	StorageElements       []StorageElement       `xml:"StorageElement,omitempty"`
	StorageElementRecords []StorageElementRecord `xml:"StorageElementRecord,omitempty"`
	OtherRecords          []XMLRecord            `xml:",omitempty,any"`
}

func (b *RecordBundle) RecordCount() int {
	return len(b.JobUsageRecords) +
		len(b.StorageElements) +
		len(b.StorageElementRecords) +
		len(b.OtherRecords)
}

func (b *RecordBundle) Records() []Record {
	recs := make([]Record, 0, b.RecordCount())
	for _, r := range b.JobUsageRecords {
		recs = append(recs, &r)
	}
	for _, r := range b.StorageElements {
		recs = append(recs, &r)
	}
	for _, r := range b.StorageElementRecords {
		recs = append(recs, &r)
	}
	return recs
}

// XMLRecord is a generic structure for unmarshalling unknown XML data.
type XMLRecord struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
}
