# Callsign parsing notes

## Software rules of Club Log

* [Mapping of portable callsigns](https://clublog.freshdesk.com/support/solutions/articles/3000065656-mapping-of-portable-callsigns)
* [Mapping of KG4 Calls](https://clublog.freshdesk.com/support/solutions/articles/3000065658-mapping-of-kg4-calls)
* [Batch lookups of DXCCs](https://clublog.freshdesk.com/support/solutions/articles/167890-batch-lookups-of-dxccs)
* [List of other rules](https://clublog.freshdesk.com/support/solutions/folders/3000012296)

## Rule application sequence

(From up to down)

### Primary/root rules

* Input: a full callsign and the contact time
* Check maximum length (accept only 16 or less) and letters in a full callsign (`[0-9A-Z/]` only, no space or others)
  - Reject as an invalid callsign if failed
* Check if the callsign in in the DXCC-Invalid (CLDMapInvalid) table
  - If the callsign entry exists, check the contact time
    - If the time is matched, return as DXCC Invalid entity
    - If no match is found, do nothing
* Parse and split parts with slashes (/)
  - Reject if the parts are four (5) or more (i.e., three (4) or more slashes)
* If callsign has no slash, invoke zero-slash processing and return
* If callsign has null part at the top or the end of callsign, reject it 

### Zero-slash processing

* Check if the callsign is in the Exception (CLDMapException) table
  - If the callsign entry exists, check the contact time
    - If the time is matched, use and set the DXCC/CQZ info, as whitelisted
    - If the time is not matched, repeat checking all time entries
    - If no match is found, do nothing
* Check if the callsign is in the ZoneException (CLDMapZoneException) table
 - If the callsign entry exists, check the contact time
    - If the time is matched, use the info and update the DXCC/CQZ info
    - If the time is not matched, repeat checking all time entries
    - If no match is found, do nothing
* If DXCC/CQZ info is set, check whitelisting and exit
* If the remaining callsign contains zero (0) slash
  - Check matching with `([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)`
    - So that a callsign will be split into prefix and suffix
    - If failed, the callsign is invalid
  - Check prefix for determining DXCC/CQZ info
    - Use the longest prefix match in the Prefix (CLDPrefix) table
    - If the prefix match exists, check the contact time
      - If the time is matched, use and set the DXCC/CQZ info, then check whitelisting and exit
      - If the time is not matched, repeat checking all time entries
      - If no match is found, do nothing
    - If no match is found, use a shorter prefix match, then repeat until no match
    - If no prefix in the prefix table is matched, the callsign is invalid

### One-or-more-slash callsign processing

* (Here the remaining callsign contains at least one (1) slash)
* Check Aeronautical/Maritime Mobile prefix/symbol
  - If found, set the result and exit
* Apply special prefix rules here (TBD)
  - e.g., FO/M, 3D2/R, 3D2/C, etc.
  - FO/M (Marquesas) examples:
    - FO/M/JJ1BDX and FO/JJ1BDX/M are both Marquesas in Club Log
    - OTOH, JJ1BDX/FO/M is French Polynesia in Club Log (as JJ1BDX/FO)
      - JJ1BDX/FO/M should also be treated as Marquesas
  - 3D2/C (Conway Reef) examples:
    - 3D2BDX/C is Conway Reef
    - 3D2/C/JJ1BDX and 3D2/JJ1BDX/C are Conway Reef in Club Log
    - OTOH, JJ1BDX/3D2/C is Fiji in Club Log (as JJ1BDX/3D2)
      - JJ1BDX/3D2/C should also be treated as Conway Reef
  - If matched and resolved, set the result, check whitelisting, and exit
* If the remaining callsign contains one (1) slash
  - Remove designators to ignore (apply designator removal rules)
    - Reject if the result string length is zero
  - Apply the 1-slash callsign rules then exit (TBD)
* If the remaining callsign contains two (2) slashes
  - Remove designators to ignore (apply designator removal rules)
    - Reject if the result string length is zero
  - Apply the 2-slash callsign rules then exit (TBD)

### Whitelisting check required for DXCC/CQZ info check

* Check whitelisting status on the Entity (CLDMapEntity) table
  - Whitelisted callsigns only reside in the Exception (CLDMapException) table

## Designator removal rules

* Designator: the last slash-split part of a callsign
* Applied recursively to remove multiple non-necessary designators
* Ignore following designators
  - /[A-EHJ-VX-Z] (anything 1-letter alphabet other than the listed prefix)
    - Applied recursively to remove the following examples:
      - /P, /M/P, /P/M, /N, /A/M
  - / and two-letter designators which are invalid prefixes
    - /2K, /AE, /AG, /EO, /FF. /GA, /GP, /HQ, /KT, /LH, /LT 
    - /PM, /RP, /SJ, /SK, /XA, /XB, /XP
  - /[A-Z]{3,}: /QRP or anything 3 or more letters that only contain alphabets
    - including /LGT
  - / and other longer strings
    - /QRP1W, /QRP5W, /Y2K 

## Aeronautical/Maritime Mobile rules

* AM/whatever -> AERONAUTICAL MOBILE
* whatever/AM -> AERONAUTICAL MOBILE
* MM/whatever -> Scotland (valid prefix)
* MM[0-9]/whatever -> Scotland (valid prefix)
* whatever/MM -> MARITIME MOBILE
* whatever/MM[0-9] -> MARITIME MOBILE

## Letters allowed in a full callsign

[0-9A-Z/] only

* Uppercase alphabets (A to Z)
  - No uppercasing within gocldb functions
* Number digits (0 to 9)
* and Slash (/)

## Maximum length of a full callsign

The maximum length of a full callsign is sixteen (16).

## Maximum slashes (/) allowed in a full callsign

* Practically two (2)
  - Allowing A, A/B, A/B/C only
  - Malformed and discarded if each of A, B, C is a zero-length string
* Reference: Club Log has no upper limit

## Prefix testing for non-exception callsigns

* Choose the longest match with the prefix database
  - e.g., if N6BDX matches both with N and N6, then N6 is selected

## regexs for non-exception callsigns

### Without (zero) Slash (/)

* Should be: `([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)`
  - No need to allow callsigns without "Call Area" numbers
  - Grouping info prefix and (optional) suffix
* Reference: Club Log's callsign: `([0-9]?[A-Z]+[0-9]+|[0-9][A-Z]+)([0-9A-Z]+)`
  - Grouping into prefix and (optional) suffix

### With One (1) slash

#### Cases

* callsign/callsign
* prefix/callsign
* callsign/designator or a prefix
* not-even-a-prefix/not-even-a-prefix: invalid

#### Length rule for two full callsigns

For two callsigns Part1/Part2 (e.g, JJ1BDX/N6BDX -> Part1: JJ1BDX, Part2: N6BDX)

* If length(Part1) <= length(part2)
  - then use Part1 for prefix testing
  - else (if length of Part1 is longer) use Part2 for prefix testing

#### For prefix/callsign

* Test prefix
* If valid, use the prefix
* If not valid, test the prefix of the callsign

#### For callsign/designator

* The designator should be tested as a prefix
* Valid 1-letter prefix in a designator:
  - /[FGIW]
* 1 or more number-digits is parsed as the call area digits
  - /[0-9]+
  - /1, /11, /130, etc.
* Others: test the designator as a prefix

### With two (2) slashes

Still investigating...

#### Cases for non-split prefixes

* callsign/whatever/whatever -> prefix of 1st callsign
* prefix/whatever/whatever -> prefix
* not-even-a-prefix/not-even-a-prefix/not-even-a-prefix -> invalid

#### Special case of split prefix (prefix with one slash, e.g., FO/M)

* split-prefix: exact match for each part only
* split-prefix-1st/whatever/split-prefix-2nd -> split-prefix
* split-prefix-1st/split-prefix-2nd/whatever -> split-prefix
* whatever/split-prefix-1st/split-prefix-2nd -> split-prefix-1st

[TBD]
