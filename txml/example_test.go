package txml_test

import (
	"fmt"

	"github.com/jgolang/helpers/txml"
)

// A mixed payload carries an escaped XML document inside a regular element.
// Decode it in two steps instead of calling CorrectXML on the whole payload.
func Example_mixedPayload() {
	type envelope struct {
		Body string `xml:"body"`
	}
	type user struct {
		Name string `xml:"name"`
	}
	payload := "<envelope><body>&lt;user&gt;&lt;name&gt;Tom &amp;amp; Jerry&lt;/name&gt;&lt;/user&gt;</body></envelope>"

	// Step 1: decode the outer element; the decoder unescapes the inner document once.
	var outer envelope
	if err := txml.XMLString(payload).Decode(&outer); err != nil {
		fmt.Println(err)
		return
	}

	// Step 2: decode the inner document.
	var inner user
	if err := txml.XMLString(outer.Body).Decode(&inner); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(inner.Name)
	// Output: Tom & Jerry
}
