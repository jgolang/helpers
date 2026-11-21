package txml

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"regexp"
	"strings"
	"unicode/utf8"
)

// Result represents a raw XML payload that may require normalization.
type Result string

// CorrectXML normalizes escaped markup and converts declared ISO-8859-1 payloads to UTF-8.
// It unescapes the whole document, so it is meant for fully escaped payloads;
// see the README for payloads that mix real and escaped markup.
func (r Result) CorrectXML() XMLString {
	xmlStr := string(r)
	xmlStr = html.UnescapeString(xmlStr)

	if declaresLatin1(xmlStr) {
		if !utf8.ValidString(xmlStr) {
			xmlStr = latin1ToUTF8(xmlStr)
		}

		xmlStr = rewriteDeclaredEncoding(xmlStr)
	}

	return XMLString(xmlStr)
} // end parse()

// XMLString is a normalized XML payload.
type XMLString string

// Decode unmarshals the XML payload into dst.
// dst must be a non-nil pointer.
func (r XMLString) Decode(v interface{}) error {
	if v == nil {
		return fmt.Errorf("txml: destination cannot be nil")
	}

	xmlStr := string(r)
	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlStr)))
	decoder.CharsetReader = charsetReader

	if err := decoder.Decode(v); err != nil {
		return err
	}

	return nil
} // end decode()

func declaresLatin1(xmlStr string) bool {
	lower := strings.ToLower(xmlStr)
	return strings.Contains(lower, "encoding=\"iso-8859-1\"") ||
		strings.Contains(lower, "encoding='iso-8859-1'")
}

// latin1Declaration matches the encoding attribute of the XML declaration only,
// so text content that mentions the charset name is left untouched.
var latin1Declaration = regexp.MustCompile(`(?i)(<\?xml[^>]*?\bencoding\s*=\s*["'])iso-8859-1(["'])`)

func rewriteDeclaredEncoding(xmlStr string) string {
	match := latin1Declaration.FindStringSubmatchIndex(xmlStr)
	if match == nil {
		return xmlStr
	}

	return xmlStr[:match[3]] + "utf-8" + xmlStr[match[4]:]
}

func latin1ToUTF8(value string) string {
	runes := make([]rune, 0, len(value))
	for i := 0; i < len(value); i++ {
		runes = append(runes, rune(value[i]))
	}
	return string(runes)
}

func charsetReader(charset string, input io.Reader) (io.Reader, error) {
	switch strings.ToLower(strings.TrimSpace(charset)) {
	case "iso-8859-1", "latin1", "latin-1":
		raw, err := io.ReadAll(input)
		if err != nil {
			return nil, err
		}
		return strings.NewReader(latin1ToUTF8(string(raw))), nil
	default:
		return nil, fmt.Errorf("txml: unsupported charset %q", charset)
	}
}
