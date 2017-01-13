package gracc

import (
	"encoding/xml"
)

// RecordBundle is a structure for unmarshalling a bundle of records, as would
// be sent from a probe.
type RecordBundle struct {
	XMLName               xml.Name               `xml:"RecordEnvelope"`
	UsageRecords          []JobUsageRecord       `xml:"UsageRecord,omitempty"`
	JobUsageRecords       []JobUsageRecord       `xml:"JobUsageRecord,omitempty"`
	StorageElements       []StorageElement       `xml:"StorageElement,omitempty"`
	StorageElementRecords []StorageElementRecord `xml:"StorageElementRecord,omitempty"`
	OtherRecords          []XMLRecord            `xml:",omitempty,any"`
}

// RecordCount returns the total number of records in the bundle.
func (b *RecordBundle) RecordCount() int {
	return len(b.UsageRecords) +
		len(b.JobUsageRecords) +
		len(b.StorageElements) +
		len(b.StorageElementRecords) +
		len(b.OtherRecords)
}

// Records returns all recognized records. Records that did not match
// a known type (i.e. those that are in OtherRecords) are not included!
func (b *RecordBundle) Records() chan Record {
	recs := make(chan Record, b.RecordCount())
	defer close(recs)
	for i, _ := range b.UsageRecords {
		recs <- &b.UsageRecords[i]
	}
	for i, _ := range b.JobUsageRecords {
		recs <- &b.JobUsageRecords[i]
	}
	for i, _ := range b.StorageElements {
		recs <- &b.StorageElements[i]
	}
	for i, _ := range b.StorageElementRecords {
		recs <- &b.StorageElementRecords[i]
	}
	return recs
}

// AddRecord attempts to coerce rec into a known type and add it to the
// appropriate list. Otherwise, it is added to OtherRecords.
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
