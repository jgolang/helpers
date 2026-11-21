# helpers

Small collection of Go utilities for dynamic metadata, concurrent value sharing, and XML normalization.

## Packages

- `info`: dynamic `Info` map with compatibility getters and safer lookups.
- `async`: concurrent container for sharing one value and one error per key.
- `txml`: escaped XML normalization and struct decoding helpers.

## Recommended usage

### `info`

Use `Lookup*` when you need to distinguish between "missing key" and "zero value".

```go
var meta info.Info
meta.Set("enabled", true)

enabled, ok := meta.LookupBool("enabled")
```

`Get*` is kept for compatibility and returns zero values when the key is missing or the type does not match.

### `async`

If the key comes from untrusted input, use `TrySet`.

```go
var c async.Container
if err := c.TrySet("job", result, nil); err != nil {
    return err
}
```

`Set` panics when the key is `nil` or not comparable, similar to how an invalid insertion into a regular `map` would fail at runtime.

`Get` returns a deep copy of the stored value, so goroutines can modify what they receive without racing. Do not modify a value after storing it. Limits of the copy:

- unexported struct fields are copied shallowly, including exported fields promoted from an embedded unexported struct (the embedded field itself is unexported);
- pointers to structs with unexported fields are shared as-is, since such types usually own internal state (locks, buffers) that a copy would break;
- map keys, channels, functions and unsafe pointers are shared;
- when the stored value is neither assignable nor convertible to the destination type, it is decoded through a JSON round trip instead.

### `txml`

Normalize first when the payload is escaped or declares a Latin-1 header:

```go
raw := txml.Result(payload)
clean := raw.CorrectXML()

var out Response
if err := clean.Decode(&out); err != nil {
    return err
}
```

`CorrectXML` unescapes HTML entities across the whole document, so use it only for fully escaped payloads. When a Latin-1 declaration is present, only the `encoding` value of the `<?xml …?>` declaration is rewritten to `utf-8`; text content is left as is.

On a mixed payload (real markup with an escaped document inside an element), entities that belong to the outer level are unescaped too: `&amp;` becomes a bare `&` and `Decode` fails with a syntax error, and escaped markup in text such as `&lt;b&gt;` becomes real elements. Decode such payloads in two steps instead, without `CorrectXML`:

```go
type envelope struct {
    Body string `xml:"body"` // holds the escaped inner document
}

var outer envelope
if err := txml.XMLString(payload).Decode(&outer); err != nil {
    return err
}

var out Response
if err := txml.XMLString(outer.Body).Decode(&out); err != nil {
    return err
}
```

## Compatibility

Recent changes to consider if the library already has consumers:

- `info.Info.Get` returns `nil` when the map itself is `nil`.
- `info.Info` now provides `Lookup*` helpers for explicit contracts.
- `async.Container` now exposes `TrySet` to avoid panics on invalid keys.
- `async.Container.Get` returns a deep copy of the stored value (v0.2.0); before, slices, maps and pointers were shared between the container and every caller.
- `async.Container.Get` quotes the key in its not-found error (`key ("missing")`), so control characters in untrusted keys are escaped.
- `txml.XMLString.Decode` now decodes directly into the caller-provided destination pointer.

## License

Released under the [MIT License](LICENSE).
