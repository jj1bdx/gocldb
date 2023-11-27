# Development notes

## Aeronautical/Maritime Mobile rules

* AM/whatever -> AERONAUTICAL MOBILE
* whatever/AM -> AERONAUTICAL MOBILE
* whatever/MM -> MARITIME MOBILE
* whatever/MM[0-9] -> MARITIME MOBILE

## Prefix testing

* Choose the longest match with the prefix database
  - e.g., if N6BDX matches both with N and N6, then N6 is selected

## regexs

### Without (zero) Slash (/)

* Callsign: `([0-9]?[A-Z]+[0-9]+|[0-9][A-Z]+)([0-9A-Z]+)`
  - Grouping into prefix and (optional) suffix

### With One (1) slash

#### Cases

* callsign/callsign
* prefix/callsign
* callsign/designator
* not-even-a-prefix/not-even-a-prefix

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
* Valid 1-letter prefix:
  - /[FGIW]
* Ignore following designators
  - /[0-9A-Z] (anything 1-letter other than the listed prefix)
  - /QRP (or anything not-a-prefix)
* Others: test the designator as prefix

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

### With three (3) slashes or more

[TBD]
