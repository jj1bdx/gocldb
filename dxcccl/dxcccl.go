// dxcccl: search callsigns with godxcl library
// usage: dxcccl <callsign> [time]

package main

import (
	"flag"
	"fmt"
	"github.com/jj1bdx/gocldb"
	"io"
	"log"
	"os"
	"strings"
	"time"
)

func main() {

	var err error
	// The variable of flag.Bool is stored AFTER flag.Parse() is executed!
	var debugmode = flag.Bool("d", false, "output debug log if set")

	flag.Usage = func() {
		execname := os.Args[0]
		fmt.Fprintf(flag.CommandLine.Output(),
			"dxcccl: Club Log cty.xml lookup tool\n"+
				"(c) 2023 Kenji Rikitake, JJ1BDX.\n"+
				"\n")
		fmt.Fprintf(flag.CommandLine.Output(),
			"Usage: %s [-d] callsign [time] \n\n", execname)
		fmt.Fprintf(flag.CommandLine.Output(),
			"Acceptable time formats:\n"+
				"    2006-01-02T15:04:05Z (assuming UTC)\n"+
				"    \"2006-01-02 15:04:05\" (assuming UTC, use doublequote)\n"+
				"    2006-01-02 (assuming 0000UTC, date only)\n\n")
		flag.PrintDefaults()
	}

	flag.Parse()

	gocldb.LoadCtyXml()

	// Disable debug logging
	if !(*debugmode) {
		gocldb.DebugLogger.SetOutput(io.Discard)
	}

	args := flag.Args()
	narg := flag.NArg()
	if (narg < 1) || (narg > 2) {
		flag.Usage()
		return
	}

	entry := args[0]
	call := strings.ToUpper(entry)
	var qsotime time.Time
	if narg > 1 {
		datetime := args[1]
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

	fmt.Printf("Callsign:    %s\n", call)
	fmt.Printf("QSO Time:    %s\n", qsotime.Format(time.RFC3339))
	fmt.Printf("Entity Code: %d\n", result.Adif)
	fmt.Printf("Entity Name: %s\n", result.Name)
	fmt.Printf("Prefix:      %s\n", result.Prefix)
	fmt.Printf("CQ Zone:     %d\n", result.Cqz)
	fmt.Printf("Continent:   %s\n", result.Cont)
	fmt.Printf("Latitude:    %.2f\n", result.Long)
	fmt.Printf("Longitude:   %.2f\n", result.Lat)
	fmt.Printf("Deleted:     %t\n", result.Deleted)
	fmt.Printf("Blocked:     %t (by Whitelist)\n", result.BlockedByWhitelist)
	fmt.Printf("\n")

	return

}
