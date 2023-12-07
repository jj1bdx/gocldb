// XML parsing data structure for Club Log cty.xml

// Based on the code generated by a modified code of
// https://github.com/gocomply/xsd2go

// All record attibute integers are represented in uint64
// to prevent limitation of xs:unsignedShort (actually 16bit)

// Time string is represented as TimeString instead of string

package gocldb

import (
	"encoding/xml"
	"io"
	"log"
	"os"
	"path"
	"time"
)

// Time String (of RFC3359 with timezone)
// Example: 1991-03-30T23:59:59+00:00
type TimeString string

// See src/time/format.go
const (
	// Layout string for time.Parse()
	ClublogTimeLayout = "2006-01-02T15:04:05-07:00"
)

// Convert TimeString to time.Time
func ConvertTimeString(ts TimeString) time.Time {
	t, err := time.Parse(ClublogTimeLayout, string(ts))
	if err != nil {
		log.Fatalf("ConvertTimeString() error: %v", err)
	}
	return t
}

// XML nested elements begins here
type Clublog struct {
	XMLName xml.Name `xml:"clublog"`
	Date    string   `xml:"date,attr"`

	Entities          ClublogEntities          `xml:"entities"`
	Exceptions        ClublogExceptions        `xml:"exceptions"`
	Prefixes          ClublogPrefixes          `xml:"prefixes"`
	InvalidOperations ClublogInvalidOperations `xml:"invalid_operations"`
	ZoneExceptions    ClublogZoneExceptions    `xml:"zone_exceptions"`
}

type EntitiesEntity struct {
	XMLName xml.Name `xml:"entity"`

	Adif           uint16     `xml:"adif"`
	Name           string     `xml:"name"`
	Prefix         string     `xml:"prefix"`
	Deleted        bool       `xml:"deleted"`
	Cqz            uint8      `xml:"cqz"`
	Cont           string     `xml:"cont"`
	Long           float64    `xml:"long"`
	Lat            float64    `xml:"lat"`
	Start          TimeString `xml:"start"`
	End            TimeString `xml:"end"`
	Whitelist      bool       `xml:"whitelist"`
	WhitelistStart TimeString `xml:"whitelist_start"`
	WhitelistEnd   TimeString `xml:"whitelist_end"`
}

type ClublogEntities struct {
	XMLName xml.Name `xml:"entities"`

	Entity []EntitiesEntity `xml:",any"`
}

type ExceptionsException struct {
	XMLName xml.Name `xml:"exception"`

	Record uint64     `xml:"record,attr"`
	Call   string     `xml:"call"`
	Entity string     `xml:"entity"`
	Adif   uint16     `xml:"adif"`
	Cqz    uint8      `xml:"cqz"`
	Cont   string     `xml:"cont"`
	Long   float64    `xml:"long"`
	Lat    float64    `xml:"lat"`
	Start  TimeString `xml:"start"`
	End    TimeString `xml:"end"`
}

type ClublogExceptions struct {
	XMLName xml.Name `xml:"exceptions"`

	Exception []ExceptionsException `xml:",any"`
}

type PrefixesPrefix struct {
	XMLName xml.Name `xml:"prefix"`

	Record uint64     `xml:"record,attr"`
	Call   string     `xml:"call"`
	Entity string     `xml:"entity"`
	Adif   uint16     `xml:"adif"`
	Cqz    uint8      `xml:"cqz"`
	Cont   string     `xml:"cont"`
	Long   float64    `xml:"long"`
	Lat    float64    `xml:"lat"`
	Start  TimeString `xml:"start"`
	End    TimeString `xml:"end"`
}

type ClublogPrefixes struct {
	XMLName xml.Name `xml:"prefixes"`

	Prefix []PrefixesPrefix `xml:",any"`
}

type InvalidOperationsInvalid struct {
	XMLName xml.Name `xml:"invalid"`

	Record uint64     `xml:"record,attr"`
	Call   string     `xml:"call"`
	Start  TimeString `xml:"start"`
	End    TimeString `xml:"end"`
}

type ClublogInvalidOperations struct {
	XMLName xml.Name `xml:"invalid_operations"`

	Invalid []InvalidOperationsInvalid `xml:",any"`
}

type ZoneExceptionsZoneException struct {
	XMLName xml.Name `xml:"zone_exception"`

	Record uint64     `xml:"record,attr"`
	Call   string     `xml:"call"`
	Zone   uint8      `xml:"zone"`
	Start  TimeString `xml:"start"`
	End    TimeString `xml:"end"`
}

type ClublogZoneExceptions struct {
	XMLName xml.Name `xml:"zone_exceptions"`

	ZoneException []ZoneExceptionsZoneException `xml:",any"`
}

// XML nested elements ends here

// Global variables for handling raw XML-based structs
var CtyXmlData Clublog
var CtyXmlEntities []EntitiesEntity
var CtyXmlExceptions []ExceptionsException
var CtyXmlPrefixes []PrefixesPrefix
var CtyXmlInvalids []InvalidOperationsInvalid
var CtyXmlZoneExceptions []ZoneExceptionsZoneException

// Prefix (string) is the map key
type CLDEntity struct {
	Adif           uint16
	Name           string
	Deleted        bool
	Cqz            uint8
	Cont           string
	Long           float64
	Lat            float64
	Start          time.Time
	End            time.Time
	Whitelist      bool
	WhitelistStart time.Time
	WhitelistEnd   time.Time
}

// Adif (uint16) is the map key
type CLDEntityByAdif struct {
	Name           string
	Prefix         string
	Deleted        bool
	Cqz            uint8
	Cont           string
	Long           float64
	Lat            float64
	Start          time.Time
	End            time.Time
	Whitelist      bool
	WhitelistStart time.Time
	WhitelistEnd   time.Time
}

// Call (string) is the map key
type CLDException struct {
	Record uint64
	Entity string
	Adif   uint16
	Cqz    uint8
	Cont   string
	Long   float64
	Lat    float64
	Start  time.Time
	End    time.Time
}

// Call (string) is the map key
type CLDPrefix struct {
	Record uint64
	Entity string
	Adif   uint16
	Cqz    uint8
	Cont   string
	Long   float64
	Lat    float64
	Start  time.Time
	End    time.Time
}

// Call (string) is the map key
type CLDInvalid struct {
	Record uint64
	Start  time.Time
	End    time.Time
}

// Call (string) is the map key
type CLDZoneException struct {
	Record uint64
	Zone   uint8
	Start  time.Time
	End    time.Time
}

// Entity by prefix, returning a slice
var CLDMapEntity = make(map[string][]CLDEntity, 500)

// Entity by adif (Entity code)
// Each entity code maps to only one Entity
var CLDMapEntityByAdif = make(map[uint16]CLDEntityByAdif, 500)

// Entity Exception status by callsign, returning a slice
var CLDMapException = make(map[string][]CLDException, 50000)

// Entity by longest-match prefixes, returning a slice
var CLDMapPrefix = make(map[string][]CLDPrefix, 10000)

// DXCC-invalid status by callsign, returning a slice
var CLDMapInvalid = make(map[string][]CLDInvalid, 10000)

// Zone exception by callsign, returning a slice
var CLDMapZoneException = make(map[string][]CLDZoneException, 10000)

// Logger for debug messages in this package
var DebugLogger *log.Logger

// Locate cty.xml and open the file,
// then read all the contents.
// Set CtyXmlData global variable with the database contents.
// Set default logger to stderr.
//
// Search path:
//
//	/usr/local/share/dxcc
//	and the path where the program resides.
func LoadCtyXml() {
	// logger for debugging output
	DebugLogger = log.New(os.Stderr, "gocldb-debug ", log.Ldate|log.Ltime|log.LUTC|log.Lshortfile)

	// Set basedir here
	basename, err := os.Executable()
	if err != nil {
		log.Fatalf("locateCtyXml() basename: %v", err)
	}
	basedir := path.Dir(basename)

	var filename string
	filename = "/usr/local/share/dxcc/cty.xml"
	_, err = os.Stat(filename)
	if !os.IsNotExist(err) {
	} else {
		DebugLogger.Printf("locateCtyXml(): %s does not exist\n", filename)
		filename = basedir + "/cty.xml"
		_, err = os.Stat(filename)
		if !os.IsNotExist(err) {
		} else {
			log.Fatalf("locateCtyXml() unable to find cty.xml: %v",
				err)
		}
	}

	fp, err := os.Open(filename)
	if err != nil {
		log.Fatalf("locateCtyXml() unable to open %s: %v", filename, err)
	}
	buf, err := io.ReadAll(fp)
	if err != nil {
		log.Fatalf("locateCtyXml() unable to io.ReadAll(): %v", err)
	}
	err = xml.Unmarshal(buf, &CtyXmlData)
	if err != nil {
		log.Fatalf("locateCtyXml() unable to xml.Unmarshal(): %v", err)
	}

	CtyXmlEntities = CtyXmlData.Entities.Entity
	CtyXmlExceptions = CtyXmlData.Exceptions.Exception
	CtyXmlPrefixes = CtyXmlData.Prefixes.Prefix
	CtyXmlInvalids = CtyXmlData.InvalidOperations.Invalid
	CtyXmlZoneExceptions = CtyXmlData.ZoneExceptions.ZoneException

	// minimum and maximum time values
	minTime, _ := time.Parse(ClublogTimeLayout, "0001-01-01T00:00:00+00:00")
	maxTime, _ := time.Parse(ClublogTimeLayout, "9999-12-31T23:59:59+00:00")

	for _, s := range CtyXmlEntities {
		var d CLDEntity
		var da CLDEntityByAdif

		adif := s.Adif
		prefix := s.Prefix

		d.Adif = adif
		d.Name = s.Name
		d.Deleted = s.Deleted
		d.Cqz = s.Cqz
		d.Long = s.Long
		d.Lat = s.Lat
		if len(s.Start) > 0 {
			d.Start = ConvertTimeString(s.Start)
		} else {
			d.Start = minTime
		}
		if len(s.End) > 0 {
			d.End = ConvertTimeString(s.End)
		} else {
			d.End = maxTime
		}
		d.Whitelist = s.Whitelist
		if len(s.WhitelistStart) > 0 {
			d.WhitelistStart = ConvertTimeString(s.WhitelistStart)
		} else {
			d.WhitelistStart = minTime
		}
		if len(s.WhitelistEnd) > 0 {
			d.WhitelistEnd = ConvertTimeString(s.WhitelistEnd)
		} else {
			d.WhitelistEnd = maxTime
		}

		CLDMapEntity[prefix] = append(CLDMapEntity[prefix], d)

		da.Name = d.Name
		da.Prefix = prefix
		da.Deleted = d.Deleted
		da.Cqz = d.Cqz
		da.Long = d.Long
		da.Lat = d.Lat
		da.Start = d.Start
		da.End = d.End
		da.Whitelist = d.Whitelist
		da.WhitelistStart = d.WhitelistStart
		da.WhitelistEnd = d.WhitelistEnd

		// Here simple assignment, NOT appending
		CLDMapEntityByAdif[adif] = da
	}

	for _, s := range CtyXmlExceptions {
		var d CLDException

		d.Record = s.Record
		call := s.Call
		d.Entity = s.Entity
		d.Adif = s.Adif
		d.Cqz = s.Cqz
		d.Cont = s.Cont
		d.Long = s.Long
		d.Lat = s.Lat
		if len(s.Start) > 0 {
			d.Start = ConvertTimeString(s.Start)
		} else {
			d.Start = minTime
		}
		if len(s.End) > 0 {
			d.End = ConvertTimeString(s.End)
		} else {
			d.End = maxTime
		}

		CLDMapException[call] = append(CLDMapException[call], d)
	}

	for _, s := range CtyXmlPrefixes {
		var d CLDPrefix

		d.Record = s.Record
		call := s.Call
		d.Entity = s.Entity
		d.Adif = s.Adif
		d.Cqz = s.Cqz
		d.Cont = s.Cont
		d.Long = s.Long
		d.Lat = s.Lat
		if len(s.Start) > 0 {
			d.Start = ConvertTimeString(s.Start)
		} else {
			d.Start = minTime
		}
		if len(s.End) > 0 {
			d.End = ConvertTimeString(s.End)
		} else {
			d.End = maxTime
		}

		CLDMapPrefix[call] = append(CLDMapPrefix[call], d)
	}
	// SPECIAL RULE: locally-added Prefix entries
	// use 0x10000000 and after for the Record entry
	// of locally-added CLDMapPrefix entries
	// SPECIAL RULE: for E5/N CLDPrefix
	CLDMapPrefix["E5/N"] = append(CLDMapPrefix["E5/N"],
		CLDPrefix{
			// User-defined record number
			Record: 0x100000000,
			Entity: "NORTH COOK ISLANDS",
			Adif:   0xbf,
			Cqz:    0x20,
			Cont:   "OC",
			Long:   -157.97,
			Lat:    -9,
			Start:  time.Date(2006, time.July, 1, 0, 0, 0, 0, time.UTC),
			End:    time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC)},
	)

	for _, s := range CtyXmlInvalids {
		var d CLDInvalid

		d.Record = s.Record
		call := s.Call
		if len(s.Start) > 0 {
			d.Start = ConvertTimeString(s.Start)
		} else {
			d.Start = minTime
		}
		if len(s.End) > 0 {
			d.End = ConvertTimeString(s.End)
		} else {
			d.End = maxTime
		}

		CLDMapInvalid[call] = append(CLDMapInvalid[call], d)
	}

	for _, s := range CtyXmlZoneExceptions {
		var d CLDZoneException

		d.Record = s.Record
		call := s.Call
		d.Zone = s.Zone
		if len(s.Start) > 0 {
			d.Start = ConvertTimeString(s.Start)
		} else {
			d.Start = minTime
		}
		if len(s.End) > 0 {
			d.End = ConvertTimeString(s.End)
		} else {
			d.End = maxTime
		}

		CLDMapZoneException[call] = append(CLDMapZoneException[call], d)
	}
}
