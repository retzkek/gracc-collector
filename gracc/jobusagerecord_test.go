package gracc

import (
	"bytes"
	"os"
	"testing"
)

type JURTest struct {
	SourceXMLFile string
	RefJSONFile   string
}

var JURTests = []JURTest{
	{"test_data/JobUsageRecord01.xml", "test_data/JobUsageRecord01.json"},
	{"test_data/JobUsageRecord02.xml", "test_data/JobUsageRecord02.json"},
	{"test_data/JobUsageRecord03.xml", "test_data/JobUsageRecord03.json"},
}

func TestJURUnmarshal(t *testing.T) {
	for _, jt := range JURTests {
		// read source XML and parse into JobUsageRecord
		f, err := os.Open(jt.SourceXMLFile)
		if err != nil {
			t.Error(err)
		}
		defer f.Close()
		var buf bytes.Buffer
		if _, err := buf.ReadFrom(f); err != nil {
			t.Error(err)
		}
		var v JobUsageRecord
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

		// Output for comparison (TODO: compare!)
		t.Logf("=== %s ===", jt.SourceXMLFile)
		t.Logf("%s", v.Raw())
		t.Logf("\n---\n")
		t.Logf("%s", buf.Bytes())
		t.Logf("\n---\n")

		if j, err := v.ToJSON("    "); err != nil {
			t.Error(err)
		} else {
			t.Logf("%s", j)
		}

		t.Logf("\n\n")
	}
}
