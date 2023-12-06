// gocldb callsign checker

package gocldb

import (
	"errors"
	"fmt" // for debug only
	"regexp"
	// "strconv"
	"strings"
	"time"
	"unicode"
)

const (
	CallsignMaxLength = 16
)

// See "Batch lookups of DXCCs"
// https://clublog.freshdesk.com/support/solutions/articles/167890-batch-lookups-of-dxccs
const (
	AdifInternetRepeater    = 997
	AdifAeronauticalMobile  = 998
	AdifMaritimeMobile      = 999
	AdifInvalid             = 1000
	NameSatInternetRepeater = "SATELLITE, INTERNET OR REPEATER"
	NameAeronauticalMobile  = "AERONAUTICAL MOBILE"
	NameMaritimeMobile      = "MARITIME MOBILE"
	NameInvalid             = "INVALID"
)

// CheckCallsign result
type CLDCheckResult struct {
	// DXCC Entity Code
	Adif uint16
	// Entity Name
	Name string
	// Entity prefix
	Prefix string
	// CQ Zone Number
	Cqz uint8
	// Continent (ADIF Field CONT)
	Cont string
	// Longitude
	Long float64
	// Latitude
	Lat float64
	// True if a deleted DXCC entity
	Deleted bool
	// True if blocked by whitelisting
	BlockedByWhitelist bool
	// True if DXCC-invalid QSO
	Invalid bool
	// CLDException info if applicable
	HasRecordException bool
	RecordException    CLDException
	// CLDZoneException info if applicable
	HasRecordZoneException bool
	RecordZoneException    CLDZoneException
	// CLDInvalid info if applicable
	HasRecordInvalid bool
	RecordInvalid    CLDInvalid
}

// Returns initial state of CLDCheckResult
func InitCLDCheckResult() CLDCheckResult {
	var v CLDCheckResult
	v.Adif = 0
	v.Name = ""
	v.Prefix = ""
	v.Cqz = 0
	v.Cont = ""
	v.Long = 0.0
	v.Lat = 0.0
	v.Deleted = false
	v.BlockedByWhitelist = false
	v.Invalid = false
	v.HasRecordException = false
	v.RecordException = CLDException{}
	v.HasRecordZoneException = false
	v.RecordZoneException = CLDZoneException{}
	v.HasRecordInvalid = false
	v.RecordInvalid = CLDInvalid{}

	return v
}

// Errors
var ErrMalformedCallsign = errors.New("Malformed callsign")
var ErrNotReached = errors.New("Jumped into unreachable code")

// Check if a given time is in the time range
// between lower and upper (inclusive)
func TimeInRange(t time.Time, lower time.Time, upper time.Time) bool {
	return (t.Compare(lower) >= 0) && (t.Compare(upper) <= 0)
}

// Check if a callsign and a given time is in CLDMapException
// Returns CLDMapException and bool
// If bool is true, the match exists; if false, did not matched
func InExceptionMap(call string, t time.Time) (CLDException, bool) {
	exceptions, refexists := CLDMapException[call]
	if !refexists {
		return CLDException{}, false
	}
	// Scan the result slice to find out whether the matching period exists
	// Return the first matched result
	for _, s := range exceptions {
		if TimeInRange(t, s.Start, s.End) {
			return s, true
		}
	}
	// If not found, return so
	return CLDException{}, false
}

// Check if a callsign and a given time is in CLDZoneException
// Returns CLDZoneException and bool
// If bool is true, the match exists; if false, did not matched
func InZoneExceptionMap(call string, t time.Time) (CLDZoneException, bool) {
	exceptions, refexists := CLDMapZoneException[call]
	if !refexists {
		return CLDZoneException{}, false
	}
	// Scan the result slice to find out whether the matching period exists
	// Return the first matched result
	for _, s := range exceptions {
		if TimeInRange(t, s.Start, s.End) {
			return s, true
		}
	}
	// If not found, return so
	return CLDZoneException{}, false
}

// Check if a callsign and a given time is in CLDMapInvalid
// Returns CLDInvalid and bool
// If bool is true, the match exists; if false, did not matched
func InInvalidMap(call string, t time.Time) (CLDInvalid, bool) {
	exceptions, refexists := CLDMapInvalid[call]
	if !refexists {
		return CLDInvalid{}, false
	}
	// Scan the result slice to find out whether the matching period exists
	// Return the first matched result
	for _, s := range exceptions {
		if TimeInRange(t, s.Start, s.End) {
			return s, true
		}
	}
	// If not found, return so
	return CLDInvalid{}, false
}

// Check the longest prefix match
// of a given callsign in CLDMapPrefixNoSlash
// Returns the matched prefix, corresponding CLDPrefix, and bool
// If bool is true, the match exists; if false, did not matched
// How to search:
// You need to scan and list all the possible prefixes
// and look them up from the longer to the shorter ones
// to find the longest matched prefix with the time range matching
func InPrefixMap(call string, t time.Time) (string, CLDPrefix, bool) {
	matched := make(map[int]string, 4)
	ml := 0
	// Search all map entries for matched prefixes
	for p := range CLDMapPrefix {
		if strings.HasPrefix(call, p) {
			pl := len(p)
			matched[pl] = p
			if ml < pl {
				ml = pl
			}
		}
	}
	fmt.Printf("InPrefixMap matched: %#v\n", matched)
	// Sort matched prefixes into longest to shortset order
	prefixes := make([]string, 0, 8)
	for i := ml; i > 0; i-- {
		p, exists := matched[i]
		if exists {
			prefixes = append(prefixes, p)
		}
	}
	fmt.Printf("InPrefixMap prefixes: %#v\n", prefixes)
	// Search if a matched time entry exists in a prefix
	// and if exists return the result
	for _, p := range prefixes {
		entry := CLDMapPrefix[p]
		for _, s := range entry {
			if TimeInRange(t, s.Start, s.End) {
				fmt.Printf("InPrefixMap s: %#v\n", s)
				return p, s, true
			}
		}
	}
	fmt.Printf("InPrefixMap unable to match prefix\n")
	return "", CLDPrefix{}, false
}

// SPECIAL RULE: E5/N check function for North Cook Islands
// InPrefixMap-style function but dedicated for E5/N prefix only
// TODO: extend CLDPrefix map for E5/N
func E5NPrefixMap(t time.Time) (string, CLDPrefix, bool) {
	// Use E5 (South Cook Islands)
	zk1NTime, _ := time.Parse(time.DateTime, "2006-01-01")
	_, e5Entry, e5Ok := InPrefixMap("E5", t)
	_, zk1NEntry, zk1NOk := InPrefixMap("ZK1/N", zk1NTime)
	if !zk1NOk {
		fmt.Printf("E5NPrefixMap failed to retrieve ZK1/N data\n")
		return "", CLDPrefix{}, false
	}
	if e5Ok {
		// MANUALLY rewrite for E5/N...
		// Use ZK1/N as the base Entry
		e5nEntry := zk1NEntry
		// Use Start and End members of E5
		e5nEntry.Start = e5Entry.Start
		e5nEntry.End = e5Entry.End
		fmt.Printf("E5NPrefixMap s: %#v\n", e5nEntry)
		return "E5/N", e5nEntry, true
	}

	fmt.Printf("E5NPrefixMap unable to match prefix\n")
	return "", CLDPrefix{}, false
}

var DistractionSuffixes = map[string]bool{
	"P":  true,
	"2K": true, "AE": true, "AG": true, "EO": true,
	"FF": true, "GA": true, "GP": true, "HQ": true,
	"KT": true, "LH": true, "LT": true, "PM": true,
	"RP": true, "SJ": true, "SK": true, "XA": true,
	"XB": true, "XP": true,
	"QRP1W": true, "QRP5W": true, "Y2K": true,
}

// Remove unnecessary distraction suffix
func RemoveDistractionSuffix(callparts []string) ([]string, bool) {
	l := len(callparts)
	if l < 2 {
		return callparts, false
	}
	p := l - 1
	s := callparts[p]
	fmt.Printf("RemoveDistractionSuffix: p: %d, s: %s, ", p, s)

	// Remove single suffix in the list
	if DistractionSuffixes[s] {
		callparts2 := callparts[:p]
		fmt.Printf("callparts: %#v\n", callparts2)
		return callparts2, true
	}
	// Remove three or more alphabet-only letter suffix
	threealphas := regexp.MustCompile(`^[A-Z]{3,}$`)
	// If not, return with malformed callsign error
	if threealphas.MatchString(s) {
		callparts2 := callparts[:p]
		fmt.Printf("callparts: %#v\n", callparts2)
		return callparts2, true
	}
	// Remove two or more digit-only letter suffix
	twodigits := regexp.MustCompile(`^[0-9]{2,}$`)
	if twodigits.MatchString(s) {
		callparts2 := callparts[:p]
		fmt.Printf("callparts: %#v\n", callparts2)
		return callparts2, true
	}
	// Remove "/M/P", "/P/M", "/A/M"
	if l >= 3 {
		p2 := l - 2
		s2 := callparts[p2]
		fmt.Printf("RemoveDistractionSuffix: p2: %d, s2: %s, ", p2, s2)
		if ((s == "M") && (s2 == "P")) ||
			((s == "P") && (s2 == "M")) ||
			((s == "A") && (s2 == "M")) {
			callparts2 := callparts[:p2]
			fmt.Printf("callparts: %#v\n", callparts2)
			return callparts2, true
		}
	}
	// No removal
	fmt.Printf("no removal, callparts: %#v\n", callparts)
	return callparts, false
}

// Remove unnecessary distraction suffix recursively
func RemoveDistractionSuffixes(callparts []string) []string {
	for {
		callparts2, f := RemoveDistractionSuffix(callparts)
		fmt.Printf("RemoveDistractionSuffixes: removed: %t, partlength: %d, callparts: %s\n", f, len(callparts), callparts)
		if !f {
			return callparts2
		} else {
			callparts = callparts2
		}
	}
}

// Split prefix and suffix from a callsign-like string
// Return prefix and suffix
func SplitCallsign(call string) (string, string) {
	// Find prefix or prefix + suffix
	prefixsuffix := regexp.MustCompile(`^([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)$`)
	matches := prefixsuffix.FindStringSubmatch(call)
	l := len(matches)
	if l == 3 {
		return matches[1], matches[2]
	} else {
		return "", ""
	}
}

func CheckException(call string, qsotime time.Time, oldresult CLDCheckResult) (CLDCheckResult, bool) { // Result value
	result := oldresult

	// Check CLDMapException here
	er, exists := InExceptionMap(call, qsotime)
	// If exists, return the result in the database
	if exists {
		result.Adif = er.Adif
		result.Name = er.Entity
		result.Prefix = CLDMapEntityByAdif[er.Adif].Prefix
		result.Cqz = er.Cqz
		result.Cont = er.Cont
		result.Long = er.Long
		result.Lat = er.Lat
		result.Deleted = CLDMapEntityByAdif[er.Adif].Deleted
		result.HasRecordException = true
		result.RecordException = er
	} else {
		result.HasRecordException = false
	}

	return result, exists
}

func CheckZoneException(call string, qsotime time.Time, oldresult CLDCheckResult) (CLDCheckResult, bool) {
	// Result value
	result := oldresult

	// Check CLDZoneException here
	zer, exists := InZoneExceptionMap(call, qsotime)
	if exists {
		result.Cqz = zer.Zone
		result.HasRecordZoneException = true
		result.RecordZoneException = zer
	} else {
		result.HasRecordZoneException = false
	}

	return result, exists
}

// Parse a callsign and time
// with given callsign and contact/QSO time
// Note well: callsign must be uppercased
func CheckCallsign(call string, qsotime time.Time) (CLDCheckResult, error) {
	// Result value
	result1 := InitCLDCheckResult()

	// Check if callsign consists of
	// digits, capital letters, and slashes only
	// from length 1 to 16 characters
	regcallcheck := regexp.MustCompile(`^[0-9|A-Z|\/]{1,16}$`)
	// If not, return with malformed callsign error
	if !(regcallcheck.MatchString(call)) {
		return result1, ErrMalformedCallsign
	}

	// Check CLDMapInvalid here
	ir, exists := InInvalidMap(call, qsotime)
	// If exists, return as an DXCC-invalid callsign
	if exists {
		result1.Adif = 0
		result1.Name = NameInvalid
		result1.Invalid = true
		result1.HasRecordInvalid = true
		result1.RecordInvalid = ir

		return result1, nil
	}

	// Split callsign separated by "/" into parts
	callparts := strings.Split(call, "/")
	// Check how many parts in the callparts
	partlength := len(callparts)

	fmt.Printf("partlength: %d, callparts: %#v\n", partlength, callparts)

	// Check Aeronautical Mobile
	// If any part in the callparts contains "AM"
	if partlength > 1 {
		for _, s := range callparts {
			if s == "AM" {
				// Aeronautical Mobile Callsign
				result1.Adif = 0
				result1.Name = NameAeronauticalMobile
				result1.Invalid = true
				result1.HasRecordInvalid = false
				return result1, nil
			}
		}
	}

	// Check Maritime Mobile
	// (If second or later part in the callparts contains "MM[0-9]?")
	// exception: if the first part contains "MM[0-9]?", that is Scotland
	mmcheck := regexp.MustCompile(`^MM[0-9]?$`)
	for i := 1; i < partlength; i++ {
		s := callparts[i]
		if mmcheck.MatchString(s) {
			// Maritime Mobile Callsign
			result1.Adif = 0
			result1.Name = NameMaritimeMobile
			result1.Invalid = true
			result1.HasRecordInvalid = false
			return result1, nil
		}
	}

	// If the callsign does not contain slashes
	// Use the processing function for zero-slash callsign
	if partlength == 1 {
		return CheckCallsign0(call, qsotime)
	}

	// If a zero-length string in a split part of a callsign is found,
	// treat it as malformed and exit
	for _, s := range callparts {
		if len(s) == 0 {
			return result1, ErrMalformedCallsign
		}
	}

	// CLDMapException check
	result2, found2 := CheckException(call, qsotime, result1)
	if found2 {
		return PostCheckCallsign(call, qsotime, result2)
	}
	// If KL7/JJ1BDX form, also check with JJ1BDX/KL7
	// for CLDMapException and CLDMapZoneException
	if partlength == 2 {
		callswapped := callparts[1] + "/" + callparts[0]
		result3, found3 := CheckException(callswapped, qsotime, result2)
		if found3 {
			return PostCheckCallsign(call, qsotime, result3)
		}
	}

	// 3-part-split callsign test
	// valid cases (check in this respective sequence)
	//   full-callsign/prefix-part1/prefix-part2
	//   prefix-part1/full-callsign/prefix-part2
	//   prefix-part1/prefix-part2/full-callsign
	if partlength == 3 {
		rp := ""
		prefix, suffix := SplitCallsign(callparts[0])
		if suffix != "" {
			//   full-callsign/prefix-part1/prefix-part2
			rp = callparts[1] + "/" + callparts[2]
		} else {
			prefix, suffix = SplitCallsign(callparts[1])
			if suffix != "" {
				rp = callparts[0] + "/" + callparts[2]
			} else {
				prefix, suffix = SplitCallsign(callparts[2])
				if suffix != "" {
					rp = callparts[0] + "/" + callparts[1]
				} else {
					return result2, ErrMalformedCallsign
				}
			}
		}
		fmt.Printf("rp = %s, prefix = %s, suffix = %s\n", rp, prefix, suffix)

		// special rules for 3D2, FO, FR are covered with InPrefixMap

		// SPECIAL RULE: JD/M and JD/O
		// Minami Torishima
		if rp == "JD/M" {
			rp = "JD1M"
		}
		// Ogasawara
		if rp == "JD/O" {
			rp = "JD1"
		}
		// SPECIAL RULE: HK0/M for Malpelo
		if rp == "HK0/M" {
			rp = "HK0M"
		}
		// SPECIAL RULE: ZK1/S
		if rp == "ZK1/S" {
			rp = "ZK1"
		}
		// SPECIAL RULE: E5/S
		if rp == "E5/S" {
			rp = "E5"
		}

		fmt.Printf("rp after rewrite: %s\n", rp)
		var mp string
		var mpm CLDPrefix
		var found bool
		// SPECIAL RULE: E5/N
		if rp == "E5/N" {
			mp, mpm, found = E5NPrefixMap(qsotime)
		} else {
			// Normal prefix lookup
			mp, mpm, found = InPrefixMap(rp, qsotime)
		}
		fmt.Printf("mp: %s, mpm: %#v, found: %t\n", mp, mpm, found)

		adif := mpm.Adif
		result2.Adif = adif
		result2.Name = mpm.Entity
		result2.Prefix = mp
		result2.Cqz = mpm.Cqz
		result2.Cont = mpm.Cont
		result2.Long = mpm.Long
		result2.Lat = mpm.Lat
		result2.Deleted = CLDMapEntityByAdif[adif].Deleted

		return PostCheckCallsign(call, qsotime, result2)
	}

	// Remove Distraction Suffixes
	callparts2 := RemoveDistractionSuffixes(callparts)
	partlength2 := len(callparts2)
	fmt.Printf("truncated callparts: partlength: %d, callparts: %s\n", partlength2, callparts2)

	// Rebuild reduced callsign from callparts
	if partlength2 == 0 {
		return result1, ErrMalformedCallsign
	}
	call2 := ""
	for i := 0; i < (partlength2 - 1); i++ {
		call2 = call2 + callparts2[i] + "/"
	}
	call2 = call2 + callparts2[partlength2-1]
	fmt.Printf("rebuilt callsign: %s\n", call2)

	// CLDMapException check for the rebuilt callsign again
	result3, found3 := CheckException(call2, qsotime, result1)
	if found3 {
		return PostCheckCallsign(call2, qsotime, result3)
	}
	// If KL7/JJ1BDX form, also check with JJ1BDX/KL7
	// for CLDMapException and CLDMapZoneException
	if partlength2 == 2 {
		callswapped2 := callparts2[1] + "/" + callparts2[0]
		result3, found3 := CheckException(callswapped2, qsotime, result1)
		if found3 {
			return PostCheckCallsign(call2, qsotime, result3)
		}
	}

	// If the last part of the slash-split callsign
	// contains only a single digit,
	// use the digit to replace the call area part of the callsign
	if partlength2 == 2 {
		ls := callparts2[1]
		rd := ""
		if (len(ls) == 1) && unicode.IsDigit(rune(ls[0])) {
			rd = ls
			// Assume the first part is a full callsign
			prefixnumsuffix := regexp.MustCompile(`^([0-9]?[A-Z]+)([0-9]+)([0-9A-Z]+)$`)
			matches := prefixnumsuffix.FindStringSubmatch(callparts2[0])
			if len(matches) < 4 {
				return result1, ErrMalformedCallsign
			}
			newprefix := matches[1]
			newcallarea := rd
			newsuffix := matches[3]

			// SPECIAL RULE: US prefix rules
			usprefix := regexp.MustCompile(`^[KNW][A-Z]{0,1}$|^A[A-L]$`)
			if usprefix.MatchString(newprefix) {
				newprefix = "K"
			}

			// SPECIAL RULE: BS/7 -> BS0 (CHINA), not BS7
			if (newprefix == "BS") && (newcallarea == "7") {
				newcallarea = "0"
			}

			// SPECIAL RULE: Russian prefix/9:
			// add "V" to the top of the suffix
			// so that UA9AA/9 -> UA9VAA, RU9I/9 -> RU9VI
			// (to Zone 18)
			if ((newprefix[0] == 'R') || (newprefix[0] == 'U')) &&
				(newcallarea == "9") {
				newsuffix = "V" + newsuffix
			}

			newcall := newprefix + newcallarea + newsuffix
			return CheckCallsign0(newcall, qsotime)
		}
	}

	// If the callsign does not contain slashes
	// Use the processing function for zero-slash callsign
	if partlength2 == 1 {
		return CheckCallsign0(call2, qsotime)
	}

	// Use the first two parts of split callsign
	// to determine the result prefix

	// rp: reference prefix for InPrefixMap
	rp := ""

	prefix1, suffix1 := SplitCallsign(callparts2[0])
	fmt.Printf("prefix1: %s, suffix1: %s\n", prefix1, suffix1)
	prefix2, suffix2 := SplitCallsign(callparts2[1])
	fmt.Printf("prefix2: %s, suffix2: %s\n", prefix2, suffix2)

	// prefix-only (true) or full callsign (false)
	isprefix1 := len(suffix1) == 0
	isprefix2 := len(suffix2) == 0
	if isprefix1 && isprefix2 {
		// BS7H/KL7 -> KL7, KL7/BS7H -> KL7, JJ1/KL7 -> JJ1
		if len(prefix1) <= len(prefix2) {
			rp = callparts2[0]
		} else {
			rp = callparts2[1]
		}
	} else if isprefix1 {
		// KL7/JJ1BDX
		rp = callparts2[0]
	} else if isprefix2 {
		// JJ1BDX/KL7
		rp = callparts2[1]
		// SPECIAL RULE: Ignore /M or /N suffixes: use first part
		if (rp == "M") || (rp == "N") {
			rp = callparts2[0]
		}
	} else {
		// JJ1BDX/N6BDX
		if len(callparts2[0]) <= len(callparts2[1]) {
			rp = callparts2[0]
		} else {
			rp = callparts2[1]
		}
	}
	fmt.Printf("rp: %s\n", rp)

	// SPECIAL RULE: TK/2A and TK/2B is CORSICA
	if strings.HasPrefix(prefix1, "TK") &&
		((callparts2[1] == "2A") || (callparts2[1] == "2B")) {
		rp = "TK"
	}

	// SPECIAL RULE: 3D2 with /C or /S
	if strings.HasPrefix(prefix1, "3D2") {
		rp = "3D2/" + callparts2[1]
	}

	// SPECIAL RULE: FO with /A, /C, /M
	if strings.HasPrefix(prefix1, "FO") {
		rp = "FO/" + callparts2[1]
	}

	// SPECIAL RULE: FR with /E, /G, /J, /T
	if strings.HasPrefix(prefix1, "FR") {
		rp = "FR/" + callparts2[1]
	}

	// SPECIAL RULE: HK0/M -> HK0M
	if strings.HasPrefix(prefix1, "HK0") {
		rp = "HK0" + callparts2[1]
	}

	// SPECIAL RULE: ZK1/N and ZK1/S
	if strings.HasPrefix(prefix1, "ZK1") {
		if callparts2[1] == "N" {
			// North Cook Islands ZK1/N
			rp = "ZK1/N"
		} else {
			// South Cook Islands ZK1
			rp = "ZK1"
		}
	}
	// SPECIAL RULE: E5/N and E5/S
	if strings.HasPrefix(prefix1, "E5") {
		if callparts2[1] == "N" {
			// North Cook Islands E5/N
			// Note: this is NOT in the CLDMapPrefix
			// so really special treatment has to be applied
			rp = "E5/N"
		} else {
			// South Cook Islands E5
			rp = "E5"
		}
	}

	// SPECIAL RULE:
	// Sardinia:
	// IS -> IS0
	// IM -> IM0
	if rp == "IS" {
		rp = "IS0"
	}
	if rp == "IM" {
		rp = "IM0"
	}

	// SPECIAL RULE:
	// Antarctica: KC4 -> CE9
	if rp == "KC4" {
		rp = "CE9"
	}

	fmt.Printf("rp after rewrite: %s\n", rp)

	var mp string
	var mpm CLDPrefix
	var found bool
	// SPECIAL RULE for E5/N
	if rp == "E5/N" {
		mp, mpm, found = E5NPrefixMap(qsotime)
	} else {
		// Normal prefix lookup
		mp, mpm, found = InPrefixMap(rp, qsotime)
	}
	fmt.Printf("mp: %s, mpm: %#v, found: %t\n", mp, mpm, found)

	adif := mpm.Adif
	result1.Adif = adif
	result1.Name = mpm.Entity
	result1.Prefix = mp
	result1.Cqz = mpm.Cqz
	result1.Cont = mpm.Cont
	result1.Long = mpm.Long
	result1.Lat = mpm.Lat
	result1.Deleted = CLDMapEntityByAdif[adif].Deleted

	return PostCheckCallsign(call2, qsotime, result1)
}

// Parse a callsign (assuming without slash) and time
// with given callsign and contact/QSO time
// Note well: callsign must be uppercased
func CheckCallsign0(call string, qsotime time.Time) (CLDCheckResult, error) {
	// Result value
	result1 := InitCLDCheckResult()

	// Check Exception database and if found use it
	result2, found2 := CheckException(call, qsotime, result1)
	if found2 {
		return PostCheckCallsign(call, qsotime, result2)
	}

	// Extract prefix from a callsign
	prefix, suffix := SplitCallsign(call)
	fmt.Printf("call: %s, prefix: %s, suffix: %s\n", call, prefix, suffix)

	// Find a longest valid prefix in the CLDMapPrefixNoSlash
	mp, mpm, found := InPrefixMap(call, qsotime)
	fmt.Printf("mp: %s, mpm: %#v, found: %t\n", mp, mpm, found)

	// SPECIAL RULE:
	// For KG4 prefix
	// if suffix is 2-letter, then it remains Gitmo
	// else, it's USA
	if (mp == "KG4") && (len(suffix) != 2) {
		mp, mpm, found = InPrefixMap("K", qsotime)
		fmt.Printf("KG4 prefix rewrite\n")
	}

	fmt.Printf("After rewrite: mp: %s, mpm: %#v, found: %t\n", mp, mpm, found)

	adif := mpm.Adif
	result1.Adif = adif
	result1.Name = mpm.Entity
	result1.Prefix = mp
	result1.Cqz = mpm.Cqz
	result1.Cont = mpm.Cont
	result1.Long = mpm.Long
	result1.Lat = mpm.Lat
	result1.Deleted = CLDMapEntityByAdif[adif].Deleted

	return PostCheckCallsign(call, qsotime, result1)
}

// Post-process Callsign check
func PostCheckCallsign(call string, qsotime time.Time, oldresult CLDCheckResult) (CLDCheckResult, error) {

	// CLDMapException check
	result2, found2 := CheckZoneException(call, qsotime, oldresult)

	var result3 CLDCheckResult
	if found2 {
		result3 = result2
	} else {
		result3 = oldresult
	}

	me := CLDMapEntityByAdif[result3.Adif]
	// If whitelisted and within the time range of whitelist
	// and if not in the Exception database,
	// then the callsign is BLOCKED and invalidated by the whitelist
	if me.Whitelist &&
		TimeInRange(qsotime, me.WhitelistStart, me.WhitelistEnd) &&
		!result3.HasRecordException {
		result3.Adif = 0
		result3.Name = NameInvalid
		result3.BlockedByWhitelist = true
		result3.Invalid = true
	}

	return result3, nil
}
