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

func TestCorrectXMLRewritesSingleQuotedDeclaration(t *testing.T) {
	raw := Result("<?xml version='1.0' encoding='iso-8859-1'?><user><name>alice</name></user>")

	got := raw.CorrectXML()
	want := XMLString("<?xml version='1.0' encoding='utf-8'?><user><name>alice</name></user>")

	if got != want {
		t.Fatalf("CorrectXML() = %q, want %q", got, want)
	}
}

func TestCorrectXMLPreservesEncodingNameInContent(t *testing.T) {
	raw := Result("<?xml version=\"1.0\" encoding=\"ISO-8859-1\"?><doc><note>Sent as ISO-8859-1 or iso-8859-1</note></doc>")

	got := raw.CorrectXML()
	want := XMLString("<?xml version=\"1.0\" encoding=\"utf-8\"?><doc><note>Sent as ISO-8859-1 or iso-8859-1</note></doc>")

	if got != want {
		t.Fatalf("CorrectXML() = %q, want %q", got, want)
	}
}

// CorrectXML unescapes the whole document, so it is meant for fully escaped
// payloads. The two tests below pin what happens to entities that belong to
// the outer level of a mixed payload.

func TestCorrectXMLOuterAmpersandEntityBreaksDecode(t *testing.T) {
	type user struct {
		Name string `xml:"name"`
	}
	raw := Result("<user><name>Tom &amp; Jerry</name></user>")
	var got user

	err := raw.CorrectXML().Decode(&got)

	if err == nil {
		t.Fatalf("Decode() expected syntax error for a bare '&', got %+v", got)
	}
}

func TestCorrectXMLEscapedMarkupInTextBecomesElements(t *testing.T) {
	raw := Result("<user><name>&lt;b&gt;bold&lt;/b&gt;</name></user>")

	got := raw.CorrectXML()
	want := XMLString("<user><name><b>bold</b></name></user>")

	if got != want {
		t.Fatalf("CorrectXML() = %q, want %q", got, want)
	}
}
