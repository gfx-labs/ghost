# AGENTS.md

## What this is

Go library for EVM ABI encoding/decoding. Module path: `gfx.cafe/open/ghost`. Hosted on self-hosted GitLab at `gfx.cafe`.

## Packages

- `ghost` (root) — deprecated `Client` interface mirroring `ethclient.Client`. Don't extend.
- `abi/` — low-level ABI encode/decode. Builder pattern for encoding, `Decoder` for decoding. All EVM types defined in `types.go`.
- `abir/` — reflection-based ABI encode/decode mapping Go structs via `abi:"typename"` struct tags.
- `abipath/` — path-based navigation of ABI-encoded data using a lexer from `gfx.cafe/util/go/lexer`.
- `utility/multicall/` — Multicall3 calldata builder using `abi.Builder`.

## Commands

```sh
go test ./...       # run all tests (or: make test)
make coverage       # generates coverage.out + coverage.html
```

No linter, formatter, or CI config exists — only `go vet`/`go test` apply.

## Testing

- Uses `testify` (`assert`/`require`). Tests are white-box (same package) except root which is black-box (`ghost_test`).
- Table-driven tests throughout. ABI tests compare hex output via `PrettyHex()`.
- `abir` roundtrip tests use generics to encode then decode structs.
- No external services or fixtures needed.

## Conventions

- `abir` encode/decode wraps panics with `recover()` — errors surface as return values, not panics.
- No codegen or generated files.
- Commit style is informal/terse (no conventional commits).
