package main

import (
	"fmt"
	// "fmt" // for debug only
	"github.com/jj1bdx/gocldb"
)

// main program for testing loading cty.xml

func main() {

	gocldb.LoadCtyXml()

	fmt.Println(gocldb.CtyXmlData.Date)

	var sl int

	fmt.Println("=== CLDMapEntity:", len(gocldb.CLDMapEntity))
	sl = 0
	for k, s := range gocldb.CLDMapEntity {
		fmt.Printf("%s: %#v\n", k, s)
		l := len(s)
		if l > sl {
			sl = l
		}
		if l > 1 {
			fmt.Println("==== CLDMapEntity slice length:", l)
		}
	}
	fmt.Println("=== CLDMapEntity max slice length:", sl)

	fmt.Println("=== CLDMapEntityByAdif:", len(gocldb.CLDMapEntityByAdif))
	sl = 0
	for k, s := range gocldb.CLDMapEntityByAdif {
		fmt.Printf("%d (%#x): %#v\n", k, k, s)
	}

	fmt.Println("=== CLDMapException:", len(gocldb.CLDMapException))
	sl = 0
	for k, s := range gocldb.CLDMapException {
		fmt.Printf("%s: %#v\n", k, s)
		l := len(s)
		if l > sl {
			sl = l
		}
		if l > 100 {
			fmt.Println("==== CLDMapException slice length:", l)
		}
	}
	fmt.Println("=== CLDMapException max slice length:", sl)

	fmt.Println("=== CLDMapPrefix:", len(gocldb.CLDMapPrefix))
	sl = 0
	for k, s := range gocldb.CLDMapPrefix {
		fmt.Printf("%s: %#v\n", k, s)
		l := len(s)
		if l > sl {
			sl = l
		}
	}
	fmt.Println("=== CLDMapPrefix max slice length:", sl)

	fmt.Println("=== CLDMapInvalid:", len(gocldb.CLDMapInvalid))
	sl = 0
	for k, s := range gocldb.CLDMapInvalid {
		fmt.Printf("%s: %#v\n", k, s)
		l := len(s)
		if l > sl {
			sl = l
		}
	}
	fmt.Println("=== CLDMapInvalid max slice length:", sl)

	fmt.Println("=== CLDMapZoneException:", len(gocldb.CLDMapZoneException))
	sl = 0
	for k, s := range gocldb.CLDMapZoneException {
		fmt.Printf("%s: %#v\n", k, s)
		l := len(s)
		if l > sl {
			sl = l
		}
	}
	fmt.Println("=== CLDMapZoneException max slice length:", sl)

}
