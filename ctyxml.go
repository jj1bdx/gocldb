// XML parsing data structure for Club Log cty.xml

// Based on the code generated by a modified code of
// https://github.com/gocomply/xsd2go

// All record attibute integers are represented in uint64
// to prevent limitation of xs:unsignedShort (actually 16bit)

// Time string is represented as TimeString instead of string

// package gocldb
package main

import (
	"encoding/xml"
	"fmt"
	// "fmt" // for debug only
	"io"
	"log"
	"os"
	"path"
)

// Time String (of RFC3359 with timezone)
// Example: 1991-03-30T23:59:59+00:00
type TimeString string

// See src/time/format.go
const (
	// Layout string for time.Parse()
	ClublogTimeLayout = "2006-01-02T15:04:05-07:00"
)

// Element
type Clublog struct {
	XMLName xml.Name `xml:"clublog"`
	Date    string   `xml:"date,attr"`

	Entities          ClublogEntities          `xml:"entities"`
	Exceptions        ClublogExceptions        `xml:"exceptions"`
	Prefixes          ClublogPrefixes          `xml:"prefixes"`
	InvalidOperations ClublogInvalidOperations `xml:"invalid_operations"`
	ZoneExceptions    ClublogZoneExceptions    `xml:"zone_exceptions"`
}

// Element
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

// Element
type ClublogEntities struct {
	XMLName xml.Name `xml:"entities"`

	Entity []EntitiesEntity `xml:",any"`
}

// Element
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

// Element
type ClublogExceptions struct {
	XMLName xml.Name `xml:"exceptions"`

	Exception []ExceptionsException `xml:",any"`
}

// Element
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

// Element
type ClublogPrefixes struct {
	XMLName xml.Name `xml:"prefixes"`

	Prefix []PrefixesPrefix `xml:",any"`
}

// Element
type InvalidOperationsInvalid struct {
	XMLName xml.Name `xml:"invalid"`

	Record uint64     `xml:"record,attr"`
	Call   string     `xml:"call"`
	Start  TimeString `xml:"start"`
	End    TimeString `xml:"end"`
}

// Element
type ClublogInvalidOperations struct {
	XMLName xml.Name `xml:"invalid_operations"`

	Invalid []InvalidOperationsInvalid `xml:",any"`
}

// Element
type ZoneExceptionsZoneException struct {
	XMLName xml.Name `xml:"zone_exception"`

	Record uint64     `xml:"record,attr"`
	Call   string     `xml:"call"`
	Zone   uint8      `xml:"zone"`
	Start  TimeString `xml:"start"`
	End    TimeString `xml:"end"`
}

// Element
type ClublogZoneExceptions struct {
	XMLName xml.Name `xml:"zone_exceptions"`

	ZoneException []ZoneExceptionsZoneException `xml:",any"`
}

var CtyXmlData Clublog
var CtyXmlEntities []EntitiesEntity
var CtyXmlExceptions []ExceptionsException
var CtyXmlPrefixes []PrefixesPrefix
var CtyXmlInvalids []InvalidOperationsInvalid
var CtyXmlZoneExceptions []ZoneExceptionsZoneException

// Locate cty.xml and open the file,
// then read all the contents.
// Set CtyXmlData global variable with the database contents.
//
// Search path:
//
//	/usr/local/share/dxcc
//	and the path where the program resides.
func LoadCtyXml() {
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
		// fmt.Printf("locateCtyXml(): %s does not exist\n", filename)
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

}

// mail program for testing loading cty.xml

func main() {

	LoadCtyXml()

	fmt.Println(CtyXmlData.Date)

	for _, s := range CtyXmlEntities {
		fmt.Println(s)
	}

	for _, s := range CtyXmlExceptions {
		fmt.Println(s)
	}

	for _, s := range CtyXmlPrefixes {
		fmt.Println(s)
	}

	for _, s := range CtyXmlInvalids {
		fmt.Println(s)
	}

	for _, s := range CtyXmlZoneExceptions {
		fmt.Println(s)
	}

}
