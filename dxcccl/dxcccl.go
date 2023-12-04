// dxcccl: search callsigns with godxcl library
// usage: dxcccl <callsign> [time]

package main

import (
	"fmt"
	"github.com/jj1bdx/gocldb"
	"log"
	"os"
	"strings"
	"time"
)

func main() {
	argc := len(os.Args)
	Usage := func() {
		execname := os.Args[0]
		fmt.Fprintln(os.Stderr,
			"dxcc: search callsigns with godxcc library")
		fmt.Fprintf(os.Stderr,
			"Usage: %s callsign [time] \n\n", execname)
		fmt.Fprintf(os.Stderr,
			"Time formats:\n"+
				"    2006-01-02T15:04:05Z (assuming UTC)\n"+
				"    \"2006-01-02 15:04:05\" (assuming UTC, use doublequote)\n"+
				"    2006-01-02 (assuming 0000UTC, date only)\n\n")
		fmt.Fprintf(os.Stderr,
			"dxcccl: Club Log cty.xml lookup tool\n"+
				"(c) 2023 Kenji Rikitake, JJ1BDX.\n"+
				"\n")
	}

	gocldb.LoadCtyXml()

	if (argc == 1) || (argc > 3) {
		Usage()
		return
	}

	var err error

	entry := os.Args[1]
	call := strings.ToUpper(entry)
	var qsotime time.Time
	if argc > 2 {
		datetime := os.Args[2]
		// "2006-01-02T15:04:05Z"
		qsotime, err = time.Parse(time.RFC3339, datetime)
		if err != nil {
			// "2006-01-02 15:04:05"
			qsotime, err = time.Parse(time.DateTime, datetime)
			if err != nil {
				// "2006-01-02" (time: 0000UTC)
				qsotime, err = time.Parse(time.DateOnly, datetime)
				if err != nil {
					log.Fatalf("Unable to parse datetime: %v\n", err)
				}
			}
		}
	} else {
		// Use current time in UTC
		qsotime = time.Now().UTC()
	}

	// Look up the database
	result, err := gocldb.CheckCallsign(call, qsotime)
	if err != nil {
		log.Printf("CheckCallsign() error: %v", err)
	}

	fmt.Printf("Callsign:       %s\n", call)
	fmt.Printf("QSO Time:       %s\n", qsotime.Format(time.RFC3339))
	fmt.Printf("Entity Code:    %d\n", result.Adif)
	fmt.Printf("Entity Name:    %s\n", result.Name)
	fmt.Printf("Prefix:         %s\n", result.Prefix)
	fmt.Printf("CQ Zone:        %d\n", result.Cqz)
	fmt.Printf("Continent:      %s\n", result.Cont)
	fmt.Printf("Latitude:       %.2f\n", result.Long)
	fmt.Printf("Longitude:      %.2f\n", result.Lat)
	fmt.Printf("Deleted:        %t\n", result.Deleted)
	fmt.Printf("\n")

	return

}
