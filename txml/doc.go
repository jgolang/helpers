// Package txml normalizes escaped XML payloads and decodes them into structs.
//
// [Result.CorrectXML] unescapes HTML entities across the whole document and
// rewrites a Latin-1 encoding declaration to utf-8, so use it only on fully
// escaped payloads; [XMLString.Decode] then decodes into the caller's value.
// A payload that mixes real markup with an escaped inner document is decoded
// in two steps without CorrectXML, as the README shows.
package txml
