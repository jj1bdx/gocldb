# Callsign parsing notes

This document describes the callsign resolution algorithm actually implemented
in `checkcall.go`. The entry point is:

```go
func CheckCallsign(call string, qsotime time.Time) (CLDCheckResult, error)
```

The callsign must be pre-uppercased; gocldb performs no case conversion.

## Software rules of Club Log (references)

* [Mapping of portable callsigns](https://clublog.freshdesk.com/support/solutions/articles/3000065656-mapping-of-portable-callsigns)
* [Mapping of KG4 Calls](https://clublog.freshdesk.com/support/solutions/articles/3000065658-mapping-of-kg4-calls)
* [Mapping of SWL calls](https://clublog.freshdesk.com/support/solutions/articles/3000065657-mapping-of-swl-calls)
  (not implemented in gocldb)
* [Related articles](https://clublog.freshdesk.com/support/solutions/articles/3000065659-related-articles)
* [Batch lookups of DXCCs](https://clublog.freshdesk.com/support/solutions/articles/167890-batch-lookups-of-dxccs)
* [List of other rules (folder)](https://clublog.freshdesk.com/support/solutions/folders/3000012296)

## Input constraints

* Letters allowed in a full callsign: `[0-9A-Z/]` only
  (uppercase alphabet, digits, and slash; no space or other characters)
* Maximum length of a full callsign: sixteen (16) characters
  (`CallsignMaxLength`), minimum one (1)
* Validation regex: `^[0-9A-Z/]{1,16}$` — failure returns
  `ErrMalformedCallsign`

## Special result names and ADIF codes

The following constants are defined after the Club Log batch-lookup API:

| Constant | Value | Name constant |
|----------|-------|---------------|
| `AdifInternetRepeater` | 997 | `NameSatInternetRepeater` |
| `AdifAeronauticalMobile` | 998 | `NameAeronauticalMobile` |
| `AdifMaritimeMobile` | 999 | `NameMaritimeMobile` |
| `AdifInvalid` | 1000 | `NameInvalid` |

**Note:** the `Adif*` numeric constants are currently *not* stored in results.
Every special/invalid outcome (DXCC-invalid, aeronautical mobile, maritime
mobile, prefix-not-found, whitelist-blocked) sets `Adif = 0` and only the
`Name` string distinguishes the case, together with `Invalid = true`.
There is no internet/repeater (997) detection logic at all;
`AdifInternetRepeater` and `NameSatInternetRepeater` are unused.

## Rule application sequence (top-level `CheckCallsign`)

Executed strictly in this order:

1. **Syntax check.** Reject with `ErrMalformedCallsign` unless the whole
   callsign matches `^[0-9A-Z/]{1,16}$`.
2. **DXCC-invalid check.** Look up the full callsign in `CLDMapInvalid`.
   Scan all time-bounded records for the callsign; if the contact time is
   within a record's range, return immediately with `Name = "INVALID"`,
   `Invalid = true`, `Adif = 0`. This return *bypasses* post-processing
   (no zone-exception or whitelist handling).
3. **Slash split.** Split the callsign on `/` into parts.
4. **Aeronautical mobile.** If there are two or more parts and *any* part
   (including the first) is exactly `AM`, return immediately with
   `Name = "AERONAUTICAL MOBILE"`, `Invalid = true`, `Adif = 0`
   (bypasses post-processing).
5. **Maritime mobile.** If any part *other than the first* matches
   `^MM[0-9]?$`, return immediately with `Name = "MARITIME MOBILE"`,
   `Invalid = true`, `Adif = 0` (bypasses post-processing).
   The first part is deliberately excluded: `MM/...` and `MM0.../...`
   are Scotland, resolved by normal prefix matching.
   Note the asymmetry: the AM test is an exact string match on `AM`,
   while the MM test allows an optional trailing digit (`MM0`–`MM9`).
6. **Zero-slash dispatch.** If there is exactly one part (no slash),
   process with the zero-slash procedure (see below) and return.
7. **Empty-part rejection.** If any split part is a zero-length string
   (leading, trailing, or doubled slash: `/A`, `A/`, `A//B`), reject with
   `ErrMalformedCallsign`.
8. **Exception check, full callsign.** Look up the full callsign (with
   slashes) in `CLDMapException` with time matching. On a hit, take the
   DXCC/CQZ data from the exception record and go to post-processing.
9. **Exception check, swapped form.** If there are exactly two parts,
   also look up `part2/part1` (e.g. `KL7/JJ1BDX` is also checked as
   `JJ1BDX/KL7`). This implements Club Log's bidirectional exception rule.
   On a hit, go to post-processing (keyed on the *original* callsign).
10. **Two-slash (3-part) path.** If there are exactly three parts, use the
    dedicated 3-part procedure (see below) and return.
11. **Distraction-suffix removal.** Strip ignorable designators from the
    end of the part list (see "Distraction suffix removal rules"), then
    rebuild the reduced callsign by re-joining the surviving parts with `/`.
12. **Exception re-check on the rebuilt callsign** (and, if the rebuilt
    callsign has exactly two parts, its swapped form as well), same as
    steps 8–9. On a hit, go to post-processing keyed on the rebuilt callsign.
13. **Single-digit call-area rewrite.** If the rebuilt callsign has exactly
    two parts and the last part is a single digit, rewrite the call area
    (see below) and re-enter the zero-slash procedure with the result.
14. **Zero-slash dispatch of the reduced callsign.** If suffix removal left
    only one part, process it with the zero-slash procedure.
15. **General two-part resolution.** Otherwise choose a reference prefix
    from the first two parts, apply the special prefix rules, look it up in
    the prefix map, and go to post-processing (see below).

## Zero-slash processing (`checkCallsignZeroSlash`)

Applied to callsigns without a slash (either originally, or after
reduction/rewrite by the top-level flow):

1. **Exception check.** Look up the callsign in `CLDMapException` with time
   matching; on a hit, use the record's DXCC/CQZ data and go to
   post-processing.
2. **Prefix/suffix split.** Split the callsign with
   `^([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)$` into prefix (through the call-area
   digits) and suffix. Only the suffix is used, and only by the KG4 rule.
   (Callsigns without a call-area digit or without a suffix simply produce
   empty prefix/suffix here; prefix lookup still proceeds.)
3. **Longest-prefix match.** Look the whole callsign up in `CLDMapPrefix`
   using the longest-match algorithm (see below).
4. **SPECIAL RULE: KG4 (Guantanamo Bay).** If the matched prefix is `KG4`
   and the suffix from step 2 is *not* exactly two characters, redo the
   lookup with the literal prefix `K`: `KG4AA` stays Guantanamo Bay,
   `KG4A`/`KG4AAA` become USA.
5. **No match.** If no prefix matched, set `Name = "INVALID"`,
   `Invalid = true`, `Adif = 0`.
6. Go to post-processing.

## Longest-prefix matching (`inPrefixMap`)

* Scan every key of `CLDMapPrefix` and collect those that are a prefix of
  the (possibly rewritten) callsign string.
* Try the collected prefixes from longest to shortest. For each prefix,
  scan its time-bounded records in load order and return the first record
  whose `[Start, End]` range contains the contact time.
* If the longest matching prefix has no time-valid record, *fall back* to
  the next-shorter matching prefix, and so on. Only when no matching prefix
  has a time-valid record does the lookup fail.
* Caveat: if two distinct map keys of the same length both match, only one
  survives candidate collection (Go map iteration order decides which), so
  same-length ties are resolved non-deterministically. In practice cty.xml
  prefixes of equal length that both match the same callsign are rare.

## Two-slash (3-part) callsign processing

Applied when the *original* callsign has exactly three parts (this runs
before distraction-suffix removal, and the exception maps have already
been checked for the full string).

1. **Identify the full-callsign part.** Try, in order, part 1, part 2, then
   part 3 against `^([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)$` (prefix + call-area
   digits + non-empty suffix). The first part that matches is taken as the
   operator's full callsign; the *other two parts, joined with `/` in their
   original order*, become the reference prefix `rp`:
   * `full-callsign/prefix-part1/prefix-part2` → `rp = part2/part3`
   * `prefix-part1/full-callsign/prefix-part2` → `rp = part1/part3`
   * `prefix-part1/prefix-part2/full-callsign` → `rp = part1/part2`
   * No part matches → `ErrMalformedCallsign`.

   Examples: `FO/M/JJ1BDX`, `FO/JJ1BDX/M`, and `JJ1BDX/FO/M` all yield
   `rp = "FO/M"` (Marquesas); `3D2/C/JJ1BDX`, `3D2/JJ1BDX/C`, and
   `JJ1BDX/3D2/C` all yield `rp = "3D2/C"` (Conway Reef).
2. **Special rewrites of `rp`** (exact string matches, applied in order):
   * `JD/M` → `JD1M` (Minami Torishima)
   * `JD/O` → `JD1` (Ogasawara)
   * `HK0/M` → `HK0M` (Malpelo)
   * `ZK1/S` → `ZK1` (South Cook Islands)
   * `E5/S` → `E5` (South Cook Islands)

   The `3D2/*`, `FO/*`, and `FR/*` combined prefixes need no rewrite:
   they exist as literal keys in the prefix map (e.g. `3D2/C`, `FO/A`,
   `FR/G`), so `inPrefixMap` resolves them directly.
3. **Prefix lookup.** Look `rp` up with `inPrefixMap`. On a miss, set
   `Name = "INVALID"`, `Invalid = true`, `Adif = 0`.
4. Go to post-processing, keyed on the original 3-part callsign.

## Distraction suffix removal rules

Implemented by `removeDistractionSuffixes`, an iterative loop that removes
ignorable trailing parts one step at a time until nothing more matches.
Each step examines the *last* slash-separated part `s` (and for the pair
rule, the second-to-last part) and removes it when any of the following
holds:

1. `s` is in the fixed suffix list (exact match):
   * Single letters: `P`, `X` — **only these two**; there is no general
     "any single letter" removal rule
   * Two-character: `2K`, `AE`, `AG`, `EO`, `FF`, `GA`, `GP`, `HQ`, `KT`,
     `LH`, `LT`, `PM`, `RP`, `SJ`, `SK`, `XA`, `XB`, `XP`
   * Longer, containing digits: `QRP1W`, `QRP5W`, `Y2K`
2. `s` matches `^[A-Z]{3,}$` — three or more letters, alphabet only
   (covers `/QRP`, `/QRO`, `/LGT`, `/MOBILE`, ...)
3. `s` matches `^[0-9]{2,}$` — two or more digits (e.g. `/11`, `/130`);
   a *single* digit is deliberately kept for the call-area rewrite
4. Trailing pair removal (only when at least three parts remain): the last
   two parts together are `/P/M`, `/M/P`, or `/M/A`, in which case both are
   removed at once.

Note that `M` and `N` as single trailing letters are **not** distraction
suffixes; they survive removal and are handled later by the two-part
resolution (`/M`, `/N` ignore rule) and by the split-prefix special rules
(`FO/M`, `ZK1/N`, ...).

After removal, the surviving parts are re-joined with `/` into the reduced
callsign used by all subsequent steps (including exception re-check and
zone-exception lookup).

## Single-digit call-area rewrite (two parts, `/0`–`/9`)

When the reduced callsign is `fullcall/d` with a single digit `d`:

1. Split the first part with `^([0-9]?[A-Z]+)([0-9]+)([0-9A-Z]+)$` into
   letter prefix, call-area digits, and suffix. If it does not match,
   reject with `ErrMalformedCallsign`.
2. Replace the call-area digits entirely with `d`
   (`W3ABC/6` → `W6ABC`, `JJ1BDX/2` → `JJ2BDX`).
3. **SPECIAL RULE: US prefixes.** If the letter prefix matches
   `^[KNW][A-Z]{0,1}$|^A[A-L]$`, normalize it to `K`
   (`N6BDX/7` → `K7BDX`).
4. **SPECIAL RULE: BS/7 (Scarborough Reef vs China).** `BS.../7` rewrites
   the call area to `0`, i.e. `BS0...` (China), not `BS7` (Scarborough
   Reef).
5. **SPECIAL RULE: Russian `/9` (Zone 18).** If the letter prefix starts
   with `R` or `U` and the new call area is `9`, prepend `V` to the
   suffix: `UA9AA/9` → `UA9VAA`, `RU9I/9` → `RU9VI`. This lands in the
   `UA9V` prefix block (CQ Zone 18).
6. Feed the rewritten callsign to the zero-slash procedure. (Subsequent
   zone-exception lookups are keyed on the rewritten callsign.)

## General two-part resolution

Applied to what remains after distraction-suffix removal when none of the
earlier branches took the call. Only the **first two parts** are used; if
three or more parts survive removal, the extra parts are silently ignored
(see "Edge cases" below).

Each of the two parts is classified with
`^([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)$`:
a part that matches (has a suffix after the call-area digits) is a *full
callsign*; a part that does not is a *prefix-like* part. The reference
prefix `rp` is chosen as:

* **prefix / prefix** (e.g. `JJ1/KL7`): the part whose regex-prefix is
  shorter wins (`BS7H/KL7` → `KL7`, `JJ1/KL7` → `JJ1`).
* **prefix / callsign** (e.g. `KL7/JJ1BDX`): the first part.
* **callsign / prefix** (e.g. `JJ1BDX/KL7`): the second part —
  *except* when the second part is exactly `M` or `N`
  (**SPECIAL RULE: ignore `/M`, `/N`**), in which case the whole first
  part is used for prefix matching instead.
* **callsign / callsign** (e.g. `JJ1BDX/N6BDX`): the shorter whole part
  wins; on equal length, the first part.

Then the special prefix rules below are applied to `rp`, and `rp` is looked
up with `inPrefixMap`. On a miss the result is `Name = "INVALID"`,
`Invalid = true`, `Adif = 0`. Either way the flow ends in post-processing,
keyed on the reduced callsign.

### Special prefix rules (two-part form)

Applied in this order as independent rewrites; a later rule can override an
earlier one. Rules 1–7 test the *regex-prefix of the first part*
(`strings.HasPrefix`), so they require the first part to be a full callsign
(prefix + digits + suffix); rules 8–10 test the resulting `rp` string
itself.

1. **TK (Corsica):** first-part prefix starts with `TK` and the second part
   is `2A` or `2B` → `rp = TK`. (Corsican calls signed `/2A` or `/2B`
   would otherwise be misread as France.)
2. **3D2 (Fiji / Conway Reef / Rotuma):** first-part prefix starts with
   `3D2` → `rp = 3D2/<part2>` (e.g. `3D2AG/R` → `3D2/R` Rotuma,
   `/C` → Conway Reef; resolved by literal prefix-map keys).
3. **FO (French Polynesia and dependencies):** first-part prefix starts
   with `FO` → `rp = FO/<part2>` (`/A` Austral Is., `/C` Clipperton,
   `/M` Marquesas).
4. **FR (Réunion and dependencies):** first-part prefix starts with `FR`
   → `rp = FR/<part2>` (`/E` Europa, `/G` Glorioso, `/J` Juan de Nova,
   `/T` Tromelin — deleted-entity keys included).
5. **HK0 (San Andres / Malpelo):** first-part prefix starts with `HK0`
   → `rp = HK0<part2>` concatenated without slash (`HK0X/M` → `HK0M`
   Malpelo; `/A` → `HK0A` San Andres).
6. **ZK1 (Cook Islands):** first-part prefix starts with `ZK1` →
   `rp = ZK1/N` (North Cook Is.) if the second part is `N`, else
   `rp = ZK1` (South Cook Is.).
7. **E5 (Cook Islands):** first-part prefix starts with `E5` →
   `rp = E5/N` if the second part is `N`, else `rp = E5`.
8. **IS → IS0, IM → IM0 (Sardinia):** if `rp` is exactly `IS` or `IM`,
   append `0` so the call maps to Sardinia rather than Italy.
9. **KC4 → CE9 (Antarctica):** if `rp` is exactly `KC4`, replace it with
   `CE9` so it maps to Antarctica.

## Post-processing (`postCheckCallsign`)

Every resolution path ends here **except** the DXCC-invalid, aeronautical
mobile, and maritime mobile early returns:

1. **Zone exception.** Look the callsign up in `CLDMapZoneException` with
   time matching; on a hit, overwrite the CQ zone (`Cqz`) with the
   exception value. The lookup key varies by path: the original callsign
   for exception-map hits and the 3-part path; the reduced (rebuilt)
   callsign after distraction-suffix removal; the rewritten callsign after
   the single-digit call-area rewrite.
2. **Whitelist blocking.** If the resolved entity (by ADIF code) is marked
   `Whitelist` in `CLDMapEntity`, the contact time falls within the
   entity's whitelist period, *and* the result did **not** come from an
   exception-map record, then the callsign is blocked: `Adif = 0`,
   `Name = "INVALID"`, `Invalid = true`, `BlockedByWhitelist = true`.
   Whitelisted entities accept only callsigns present in
   `CLDMapException`; exception-sourced results are exempt from blocking.
   (The other result fields — `Prefix`, `Cqz`, coordinates — retain the
   values of the blocked entity.)

## Edge cases and implementation notes

* **Three or more slashes are not rejected.** Contrary to earlier versions
  of this document, there is no explicit part-count limit. A callsign with
  four or more parts skips the 3-part path (which requires exactly three)
  and goes through distraction-suffix removal; if three or more parts
  still remain, the general two-part resolution uses only the first two
  parts and silently discards the rest. Example: `JJ1BDX/KL7/M/P` →
  suffix removal strips `/P`, leaving `JJ1BDX/KL7/M`; the trailing `M`
  cannot be removed (a lone trailing `M` is not a distraction suffix and
  the pair rule does not apply), so resolution proceeds on `JJ1BDX` and
  `KL7` with `M` ignored.
* **Bare `AM` or `MM` (no slash)** is not caught by the aeronautical/
  maritime mobile checks (which require ≥ 2 parts / a non-first part) and
  is resolved as an ordinary callsign by prefix matching.
* **Early returns bypass post-processing.** DXCC-invalid, `/AM`, and
  `/MM` results never receive zone-exception or whitelist processing,
  whereas "prefix not found" invalid results do pass through
  `postCheckCallsign` (a harmless asymmetry, since `Adif = 0` matches no
  entity there).
* **`Adif` is 0 in all special results.** See "Special result names and
  ADIF codes" above; callers must dispatch on `Name`/`Invalid`/
  `BlockedByWhitelist`, not on ADIF codes 997–1000.
* **Time-range handling.** All map records are time-bounded; records with
  no start/end in cty.xml are loaded as year 0001 / 9999 (`minTime` /
  `maxTime`), so range checks always succeed for unbounded records. When a
  key has multiple records, the first record (in cty.xml order) whose
  range contains the contact time wins.
* **Result field caveats.** `hasRecordException` is the only private
  tracking field that affects behavior (it exempts exception results from
  whitelist blocking); `hasRecordInvalid` and `hasRecordZoneException` are
  write-only. On whitelist blocking, `Prefix`/`Cqz`/`Cont`/`Long`/`Lat`/
  `Deleted` keep the blocked entity's values while `Adif`/`Name` report
  INVALID.
