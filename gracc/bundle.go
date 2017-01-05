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

func (b *RecordBundle) AddRecord(rec Record) error {
	switch rec.(type) {
	case *JobUsageRecord:
		b.JobUsageRecords = append(b.JobUsageRecords, *rec.(*JobUsageRecord))
	case *StorageElement:
		b.StorageElements = append(b.StorageElements, *rec.(*StorageElement))
	case *StorageElementRecord:
		b.StorageElementRecords = append(b.StorageElementRecords, *rec.(*StorageElementRecord))
	default:
		b.OtherRecords = append(b.OtherRecords, XMLRecord{
			XMLName:  xml.Name{Local: rec.Type()},
			InnerXML: string(rec.Raw()),
		})
	}
	return nil
}

// XMLRecord is a generic structure for unmarshalling unknown XML data.
type XMLRecord struct {
	XMLName  xml.Name
	InnerXML string `xml:",innerxml"`
}
