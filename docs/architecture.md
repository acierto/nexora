# Initial architecture decisions

Nexora runs as a standalone Go binary. `cmd/nexora` only handles process startup;
`internal/cli` coordinates configuration, loading, registry construction, and
resolution. `internal/knowledge` and `internal/resolver` have no YAML or I/O
dependencies. There is no public SDK yet.

Configuration and metadata share `internal/yamlutil`. The sole production
dependency is the YAML organization's maintained `go.yaml.in/yaml/v3`, pinned to
v3.0.5. Its [Node API](https://pkg.go.dev/go.yaml.in/yaml/v3#Node) lets us validate
types and duplicate/unknown keys before decoding, avoiding YAML-to-string
coercion and partially accepted metadata. We check all nested fields, reject
nulls and extra documents, and preserve explicit false/zero values.

Standard-library command parsing keeps the command layer small. Markdown is
never rendered or executed. Every invocation rebuilds an in-memory snapshot;
persistent indexing and extension interfaces await actual consumers and measured
needs.

See the [initial verification record](verification.md) for tested behavior and
platform-build limitations.
