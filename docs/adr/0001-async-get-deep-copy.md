# 0001 — async.Container.Get returns a deep copy

- Status: accepted (2026-11-21)

## Context

The `async` package documents `Container` as safe for concurrent use, and `Container.Get` said it "copies the stored value". It assigned or converted the stored value shallowly, so slices, maps and pointed-to data were shared between the container and every goroutine that called `Get` for the key. Two readers writing what they received raced: silent corruption for slices, `fatal error: concurrent map writes` for maps. The owner chose to make `Get` return a deep copy rather than only documenting the sharing.

## Decision

- `Get` deep-copies the stored value with an unexported reflect-based copier before the assign and convert paths of `assignValue`. The slice-to-array length check still runs before `Convert`; the JSON fallback, key validation and error texts are unchanged. Standard library only, no new exported names, no `unsafe`.
- Kinds covered: pointers get a new pointee; slices a new backing array (`len == cap`); maps a new map with deep-copied values; arrays element-wise; structs have their exported fields deep-copied; interfaces have their dynamic value copied with its type preserved. Nil stays nil.
- A memo keyed by type and address (plus len and cap for slices) maps every pointer, map and slice already copied to its copy, so shared references stay shared within the copy and self-referencing values terminate.
- Fast path: a type that reaches no pointer, slice, map or interface through arrays and struct fields is returned unchanged.
- Limits, documented on `Get` and in the README:
  - unexported struct fields are copied shallowly, including exported fields promoted from an embedded unexported struct (the embedded field itself is unexported);
  - pointers to structs with unexported fields are shared as-is, since such types (`*strings.Builder`, `*bytes.Buffer`, `*sync.WaitGroup`, `*sql.DB`) usually own internal state that a copy would break or duplicate;
  - map keys, channels, functions and unsafe pointers are shared;
  - values neither assignable nor convertible to the destination go through the JSON round trip, as before.
- `Get` reads the value and its error under the read lock and copies outside it, so writers do not wait for in-flight copies. This is safe because the container never mutates a stored value; it only replaces map entries.

## Consequences

- Behavior change at v0: released as MINOR v0.2.0 and called out in the release notes. `Set`/`TrySet` document that a stored value must not be modified afterwards.
- `Get` no longer returns the identity of stored deep-copyable pointers, slices and maps; code that relied on observing later writes through them must change.
- Cost (`BenchmarkContainerGet`, before → after): `int` 94 → 117 ns/op and flat struct 101 → 208 ns/op with no extra allocation; `[]int` of 1000 98 ns → 5.6 µs, 8217 B, 3 allocs; map of 100 96 ns → 10.7 µs, 3576 B, 7 allocs; nested struct 98 ns → 1.6 µs, 488 B, 10 allocs.
- Copy recursion depth is proportional to the value's nesting depth or linked-list length.
- Interior pointers into copied arrays or slices (for example a pointer to an element of a stored slice) are not re-aliased to the copy; they get their own pointee.
