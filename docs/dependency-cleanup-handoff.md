# Dependency cleanup handoff

## Deferred outbound work

The dependency cleanup merged in gg removes its direct and build-time use of
`github.com/gorilla/websocket` and `github.com/json-iterator/go`. Both module
paths still appear in `go list -m all` because
`github.com/daeuniverse/outbound` uses them in packages that gg does not build:

- `github.com/daeuniverse/outbound/transport/ws` imports Gorilla WebSocket.
- `github.com/daeuniverse/outbound/dialer/v2ray` imports json-iterator.

This was rechecked against gg `master` with outbound
`v0.0.0-20260623082230-cfd9e39fd5e0`. The root `go.mod` and `go.sum` do not
contain these transitive requirements because the module graph is pruned, but
the selected build list and outbound's graph edges are reproducible with:

```text
$ go list -m all | rg 'github.com/(gorilla/websocket|json-iterator/go)'
github.com/gorilla/websocket v1.5.0
github.com/json-iterator/go v1.1.12

$ go mod graph | rg '^github.com/daeuniverse/outbound.* (github.com/gorilla/websocket|github.com/json-iterator/go)'
github.com/daeuniverse/outbound@v0.0.0-20260623082230-cfd9e39fd5e0 github.com/gorilla/websocket@v1.5.0
github.com/daeuniverse/outbound@v0.0.0-20260623082230-cfd9e39fd5e0 github.com/json-iterator/go@v1.1.12
```

`go mod why -m` confirms that gg does not need either module, but removing the
remaining module-graph edges correctly requires changes in outbound. A local
empty-module replacement or a long-lived fork dependency would only hide the
upstream problem and is not the intended solution.

This upstream work was prototyped and integration-tested, but publishing a fork
or upstream pull requests was explicitly deferred for this run. The current gg
cleanup therefore stops at this boundary: these module-graph edges remain until
outbound adopts equivalent changes and gg updates to that revision.

## Proposed upstream pull requests

Keep these as two pull requests because each removes one dependency and they
can be reviewed independently. Both change outbound's `go.mod` and `go.sum`, so
the second branch must be rebased after the first merge.

### Replace Gorilla WebSocket

Replace `github.com/gorilla/websocket` with
`github.com/coder/websocket v1.8.14` in outbound's `transport/ws` package.

The migration must preserve:

- context-aware dialing through `netproxy.Dialer`;
- the configured WebSocket Host header and path, including query-encoded paths;
- TLS server name, ALPN, insecure-skip-verify, and TLS fragmentation;
- binary stream semantics through `websocket.NetConn`;
- UDP passthrough without wrapping or changing the packet connection.

Add tests for a real WebSocket handshake and binary tunnel round trip, Host and
path forwarding, TLS option construction, and UDP passthrough. Remove the old
Gorilla connection adapter after switching to `websocket.NetConn`.

The implementation should update `transport/ws/ws.go`, delete the obsolete
Gorilla adapter in `transport/ws/conn.go`, add the handshake and transport tests
in `transport/ws/ws_test.go`, and update `go.mod` and `go.sum`. Construct a
`websocket.DialOptions` with an `http.Client` whose transport delegates
`DialContext` to the supplied `netproxy.Dialer`; use `websocket.NetConn` with
`websocket.MessageBinary` for the returned stream.

### Replace json-iterator

Replace `github.com/json-iterator/go` with `encoding/json` in outbound's
`dialer/v2ray` package.

Plain `encoding/json` is not behaviorally equivalent for common VMess links:
fields such as `v`, `port`, and `aid` are often encoded as JSON numbers even
though the public structure stores strings. Introduce narrow fuzzy decoders
that accept strings, numbers, and null for string fields, and boolean,
zero/one, and recognized string forms for `allowInsecure`. Invalid values must
return errors rather than silently becoming false.

The decoded-field mapping must cover every serializable `V2Ray` field,
including Reality fields (`fp`, `pbk`, `sid`, and `spx`). Add tests for numeric
and null compatibility, invalid fuzzy values, and a full export/parse round
trip. The migration should also correct the malformed `Fingerprint` JSON tag.

The implementation should update `dialer/v2ray/v2ray.go`, extend
`dialer/v2ray/v2ray_test.go`, and remove json-iterator from `go.mod` and
`go.sum`. Decode into an internal mirror whose fields use narrow `UnmarshalJSON`
types, then explicitly map every field into the public `V2Ray` value. Continue
using `encoding/json` for export and use `json.RawMessage` for isolated legacy
fields instead of introducing another permissive JSON dependency.

## Integration evidence

A combined prototype of the two changes was tested from the latest gg `master`
through a temporary local module replacement. The prototype itself is not a
durable artifact; the file-level design and compatibility requirements above
are the handoff source of truth. The combined state passed:

- outbound targeted race tests for `transport/ws` and `dialer/v2ray`;
- outbound targeted `go vet`, `go mod verify`, and `go mod tidy -diff`;
- outbound full-package compilation with tests disabled;
- gg's complete Linux arm64 test suite at 18.3% coverage;
- gg's complete Linux amd64 test suite at 18.8% coverage;
- `go list -m all` with all originally targeted dependency paths absent, except
  the intentional local `gopkg.in/yaml.v3` compatibility identity.

Outbound's full-package compile currently needs `-vet=off` because its existing
`protocol/vless/key.go` has an unrelated non-constant `fmt.Errorf` vet finding.

## Completion steps

When the upstream work is resumed:

1. Implement the two file-level changes described above on separate outbound
   branches and run their targeted tests.
2. Publish two draft pull requests without manually requesting Copilot or
   CodeRabbit reviews.
3. Rebase the second pull request after the first merges, resolving the module
   files by retaining Coder WebSocket and removing both legacy modules.
4. After both changes merge, update gg to the resulting outbound revision in a
   separate dependency pull request.
5. Run `go mod tidy`, full Linux arm64/amd64 tests with coverage, both `ko`
   builds and container smoke tests.
6. Verify the original dependency list against source, `go.mod`, `go.sum`,
   `go list -deps ./...`, `go list -m all`, and `go mod graph`.
