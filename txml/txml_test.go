package txml

import "testing"

func TestCorrectXMLUnescapesMarkup(t *testing.T) {
	raw := Result("&lt;user&gt;&lt;name&gt;alice&lt;/name&gt;&lt;/user&gt;")

	got := raw.CorrectXML()
	want := XMLString("<user><name>alice</name></user>")

	if got != want {
		t.Fatalf("CorrectXML() = %q, want %q", got, want)
	}
}

func TestCorrectXMLConvertsLatin1AndNormalizesHeader(t *testing.T) {
	raw := Result("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><user><name>caf\xe9</name></user>")

	got := raw.CorrectXML()
	want := XMLString("<?xml version=\"1.0\" encoding=\"utf-8\"?><user><name>caf\u00e9</name></user>")

	if got != want {
		t.Fatalf("CorrectXML() = %q, want %q", got, want)
	}
}

func TestXMLStringDecode(t *testing.T) {
	type user struct {
		Name string `xml:"name"`
	}

	var got user
	raw := XMLString("<user><name>alice</name></user>")

	if err := raw.Decode(&got); err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if got.Name != "alice" {
		t.Fatalf("Decode() = %+v, want name=alice", got)
	}
}

func TestXMLStringDecodeRejectsNilDestination(t *testing.T) {
	raw := XMLString("<user><name>alice</name></user>")

	if err := raw.Decode(nil); err == nil {
		t.Fatalf("Decode() expected nil destination error")
	}
}
