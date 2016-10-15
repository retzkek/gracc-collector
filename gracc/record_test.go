package gracc

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

type RecordTest struct {
	SourceXMLFile string
	RefJSONFile   string
}

var Tests = []RecordTest{
	{"test_data/JobUsageRecord01.xml", "test_data/JobUsageRecord01.json"},
	{"test_data/JobUsageRecord02.xml", "test_data/JobUsageRecord02.json"},
	{"test_data/JobUsageRecord03.xml", "test_data/JobUsageRecord03.json"},
	{"test_data/JobUsageRecord04.xml", "test_data/JobUsageRecord04.json"},
	{"test_data/JobUsageRecord05.xml", "test_data/JobUsageRecord05.json"},
	{"test_data/JobUsageRecord06.xml", "test_data/JobUsageRecord06.json"},
	{"test_data/StorageElement01.xml", "test_data/StorageElement01.json"},
	{"test_data/StorageElement02.xml", "test_data/StorageElement02.json"},
	{"test_data/StorageElementRecord01.xml", "test_data/StorageElementRecord01.json"},
	{"test_data/StorageElementRecord02.xml", "test_data/StorageElementRecord02.json"},
}

func TestUnmarshal(t *testing.T) {
	for _, jt := range Tests {
		// read source XML and parse into Record
		f, err := os.Open(jt.SourceXMLFile)
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(f); err != nil {
			t.Error(err)
		}
		v, err := ParseRecordXML(buf.Bytes())
		if err != nil {
			t.Fatal(err)
		}

		// read reference JSON
		g, err := os.Open(jt.RefJSONFile)
		if err != nil {
			t.Error(err)
		}
		defer g.Close()
		buf.Reset()
		if _, err := buf.ReadFrom(g); err != nil {
			t.Error(err)
		}
		var rref map[string]interface{}
		if err := json.Unmarshal(buf.Bytes(), &rref); err != nil {
			t.Error(err)
		}

		t.Logf("=== %s ===\n", jt.SourceXMLFile)
		if j, err := v.ToJSON("    "); err != nil {
			t.Error(err)
		} else {
			//t.Logf("%s", j)
			// Compare
			var r map[string]interface{}
			if err := json.Unmarshal(j, &r); err != nil {
				t.Error(err)
			}
			delete(r, "RawXML")
			for k, v := range r {
				if v != rref[k] {
					t.Logf("'%s' Expected: '%v' Got '%v'", k, rref[k], v)
					t.Fail()
				}
			}
		}
	}
}
