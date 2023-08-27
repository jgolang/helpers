package txml

import (
	"bytes"
	"encoding/xml"
	"strings"
)

// Result doc
type Result string

// CorrectXML encoding of the XML document provided by the CRM service
// Change the encoding header from ISO-8859-1 to utf-8
func (r Result) CorrectXML() XMLString {

	xmlStr := string(r)

	xmlStr = strings.Replace(xmlStr, "&lt;", "<", -1)
	xmlStr = strings.Replace(xmlStr, "&gt;", ">", -1)
	xmlStr = strings.Replace(xmlStr, "ISO-8859-1", "utf-8", -1)

	return XMLString(xmlStr)

} // end parse()

// XMLString doc
type XMLString string

// Decode doc
func (r XMLString) Decode(v interface{}) error {

	xmlStr := string(r)

	decoder := xml.NewDecoder(bytes.NewReader([]byte(xmlStr)))

	err := decoder.Decode(&v)
	if err != nil {
		return err
	}

	return nil

} // end decode()
