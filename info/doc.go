// Package info provides [Info], a dynamic metadata map for values whose keys
// are not known at compile time.
//
// Prefer the Lookup methods, which report whether the key exists and holds
// the requested type, over the typed Get methods, which are kept for
// compatibility and return a zero value in both cases.
package info
