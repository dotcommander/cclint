# Change cclint

## Purpose

Implement cclint changes safely and verify behavior before review.

## Prerequisites

- Go toolchain configured
- Project cloned locally

## Main workflow

1. Run tests before edits:

```bash
go test ./...
```

2. Implement change in the relevant package under `internal/`.

3. Re-run build and tests:

```bash
go build ./...
go test ./...
```

4. Validate docs or rules changed by your implementation:

```bash
cclint --scores
```

## Verification

- `go build ./...` succeeds
- `go test ./...` succeeds
- cclint output reflects expected behavior on changed fixtures/examples

## Related docs

- Rules source map: `docs/rules/README.md`
- Scoring behavior: `docs/scoring/README.md`
- Programmatic APIs: `docs/guides/programmatic-usage.md`
