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
// Returns CLDMapInvalid and bool
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

var DistractionSuffixes = map[string]bool{
	"P": true, "M": true, "N": true, "A": true,
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
	fmt.Printf("RemoveDistractionSuffix: p: %d, s: %s ", p, s)

	if DistractionSuffixes[s] {
		callparts = callparts[:p]
		fmt.Printf("callparts: %#v\n", callparts)
		return callparts, true
	}
	// Remove three or more alphabet-only letter suffix
	if (len(s) >= 3) &&
		unicode.IsUpper([]rune(s)[0]) &&
		unicode.IsUpper([]rune(s)[1]) &&
		unicode.IsUpper([]rune(s)[2]) {
		callparts = callparts[:p]
		return callparts, true
	}
	// No removal
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
// Return prefix and (optional) suffix
func SplitCallsign(call string) (string, string) {
	prefixsuffix := regexp.MustCompile(`^([0-9]?[A-Z]+[0-9]+)([0-9A-Z]+)$`)
	matches := prefixsuffix.FindStringSubmatch(call)
	if len(matches) < 3 {
		return "", ""
	}
	return matches[1], matches[2]
}

func CheckException(call string, qsotime time.Time, oldresult CLDCheckResult) (CLDCheckResult, bool) { // Result value
	result := oldresult

	// Check CLDMapException here
	er, exists := InExceptionMap(call, qsotime)
	// If exists, return as an DXCC-invalid callsign
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
	}

	// TODO: extract prefix from a callsign

	// NOTREACHED
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
		result1.Adif = AdifInvalid
		result1.Name = NameInvalid
		result1.Prefix = ""
		result1.Cqz = 0
		result1.Cont = ""
		result1.Long = 0.0
		result1.Lat = 0.0
		result1.Deleted = false
		result1.BlockedByWhitelist = false
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

	// If the callsign does not contain slashes
	// branch to another function
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

	// CLDMapException and CLDMapZoneException check
	result2, found2 := CheckException(call, qsotime, result1)
	result3, found3 := CheckZoneException(call, qsotime, result2)
	if found2 || found3 {
		return result3, nil
	}

	// If KL7/JJ1BDX form, also check with JJ1BDX/KL7
	// for CLDMapException and CLDMapZoneException
	if partlength == 2 {
		callswapped := callparts[1] + "/" + callparts[0]

		result31, found31 := CheckException(callswapped, qsotime, result3)
		result32, found32 := CheckZoneException(callswapped, qsotime, result31)
		if found31 || found32 {
			return result32, nil
		}
	}

	// TODO: add split-prefix processing here

	// Remove Distraction Suffixes
	callparts = RemoveDistractionSuffixes(callparts)
	fmt.Printf("truncated callparts: partlength: %d, callparts: %s\n", len(callparts), callparts)

	// TODO: more processing of callsign with slashes

	// NOTREACHED
	return result3, ErrNotReached
}

// Parse a callsign (assuming without slash) and time
// with given callsign and contact/QSO time
// Note well: callsign must be uppercased

func CheckCallsign0(call string, qsotime time.Time) (CLDCheckResult, error) { // Result value
	result1 := InitCLDCheckResult()

	result2, _ := CheckException(call, qsotime, result1)
	result3, _ := CheckZoneException(call, qsotime, result2)

	// TODO: extract prefix from a callsign

	// NOTREACHED
	return result3, nil
}
