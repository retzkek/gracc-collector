package gracc

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
)

type SERTest struct {
	SourceXMLFile string
	RefJSONFile   string
}

var SERTests = []SERTest{
	{"test_data/StorageElementRecord01.xml", "test_data/StorageElementRecord01.json"},
	{"test_data/StorageElementRecord02.xml", "test_data/StorageElementRecord02.json"},
}

func TestSERUnmarshal(t *testing.T) {
	for _, jt := range SERTests {
		// read source XML and parse into StorageElement
		f, err := os.Open(jt.SourceXMLFile)
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(f); err != nil {
			t.Error(err)
		}
		var v StorageElementRecord
		if err := v.ParseXML(buf.Bytes()); err != nil {
			t.Error(err)
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
