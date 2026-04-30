# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Module setup

This repository has no `go.mod`. Before building, initialize the module:

```sh
go mod init github.com/jj1bdx/gocldb
go mod tidy
```

## Build and vet

```sh
go build .                    # verify library compiles
go build ./ctyxmldump/        # build ctyxmldump tool
go build ./dxcccl/            # build dxcccl CLI tool
go vet ./...                  # run static analysis
```

There are no tests in this repository.

## cty.xml runtime dependency

Both tools (and the library) require `cty.xml` from Club Log at startup. The file is searched in this order:

1. `/usr/local/share/dxcc/cty.xml`
2. Same directory as the executable

Use cty.xml version `2023-12-07T20:31:25+00:00` or later.

## Architecture

**Package `gocldb`** (root, two source files):

- `ctyxml.go` — XML parsing and database loading. Defines all XML struct types (`Clublog`, `EntitiesEntity`, etc.) and the six global lookup maps. `LoadCtyXml()` reads and unmarshals cty.xml, then populates those maps. Time fields with no value default to `minTime`/`maxTime` (year 0001 and 9999) so range checks always succeed.
- `checkcall.go` — Callsign lookup logic. Entry point is `CheckCallsign(call string, qsotime time.Time)`, which returns `CLDCheckResult`. The call must be pre-uppercased.

**Six global maps** (keyed as noted):

| Map | Key | Purpose |
|-----|-----|---------|
| `CLDMapEntity` | prefix string | DXCC entities by ITU prefix |
| `CLDMapEntityByAdif` | ADIF uint16 | Entity lookup by ADIF code |
| `CLDMapException` | callsign string | Per-callsign DXCC overrides |
| `CLDMapPrefix` | prefix string | Prefix-to-entity mapping (longest-match) |
| `CLDMapInvalid` | callsign string | Callsigns invalid for DXCC |
| `CLDMapZoneException` | callsign string | Per-callsign CQ Zone overrides |

Each map value is a `[]CLD*` slice because the same key can have multiple time-bounded records.

**`CheckCallsign` resolution order:**

1. Reject if malformed (non-`[0-9A-Z/]`, length > 16).
2. Return INVALID if callsign is in `CLDMapInvalid` for that time.
3. Detect aeronautical (`/AM`) and maritime mobile (`/MM[0-9]?` in non-first position).
4. Zero-slash path (`checkCallsignZeroSlash`): Exception map → longest prefix match.
5. One-slash path: Exception map check (both orderings), digit call-area rewrite, prefix map lookup.
6. Two-slash path: Identify which part holds the full callsign, apply special prefix rewrites (JD/M, JD/O, HK0/M, ZK1/S, E5/S), then prefix map lookup.
7. Post-processing (`postCheckCallsign`): apply `CLDMapZoneException`, then apply whitelist blocking.

**Special prefix rules** are scattered through `checkcall.go` as inline `SPECIAL RULE` comments. They cover: KG4, US call area prefixes, BS/7, Russian `/9` suffix, TK/2A-2B, 3D2, FO, FR, HK0, ZK1, E5, IS/IM0, KC4→CE9.

**Tools:**

- `ctyxmldump/ctyxmldump.go` — dumps all six maps to stdout; useful for inspecting loaded data.
- `dxcccl/dxcccl.go` — CLI tool: `dxcccl [-d] <callsign> [time]`. Accepts time as RFC3339, `"YYYY-MM-DD HH:MM:SS"`, or `YYYY-MM-DD`. Uses current UTC if omitted. The `-d` flag enables debug logging to stderr.

## Debug logging

Debug output is discarded by default. To enable:

```go
gocldb.SetDebugOutput(os.Stderr)
```

## Key constants

- `ClublogTimeLayout` — time format string for `time.Parse`/`Format` (`"2006-01-02T15:04:05-07:00"`)
- `MaxCtyXmlSize` — 50 MB read limit on cty.xml
- `CallsignMaxLength` — 16 characters
- Special ADIF codes: `AdifInvalid` (1000), `AdifMaritimeMobile` (999), `AdifAeronauticalMobile` (998)
