# gocldb:  Club Log DXCC database handling library

## How to use

* Run `gocldb.LoadCtyXml()` to initialize the database
  - Takes one or two seconds to startup
  - ~ 200msec on Mac mini 2023 (M2 Pro)
* The debug log output is *enabled* by default
  - Use `gocldb.DebugLogger.SetOutput(io.Discard)` to disable debug output
* Use `gocldb.CheckCallsign(call, qsotime)` to search the databse
  - result in `gocldb.CLDCheckResult` format defined in checkcall.go
    - Use only the public members of `gocldb.CLDCheckResult`
* See ctyxmldump and dxcccl command source code for the basic usage details

## Usage example

```go
// Initialize the database
gocldb.LoadCtyXml()
// Disable debug logging
if !(*debugmode) {
  gocldb.DebugLogger.SetOutput(io.Discard)
}
// Print version string
// gocldb.CLDVersionDateTime has the type time.Time 
fmt.Println(gocldb.CLDVersionDateTime.Format(gocldb.ClublogTimeLayout))
//...
// Look up the database
result, err := gocldb.CheckCallsign(call, qsotime)
```

## cty.xml file

cty.xml is distributed from Club Log with individual explicit
permission for each user.
See [Downloading The Prefixes And Exceptions As XML](https://clublog.freshdesk.com/support/solutions/articles/54902-downloading-the-prefixes-and-exceptions-as-xml)
for the further details to obtain the file.

### File search sequence of cty.xml 

* /usr/local/share/dxcc/cty.xml
* (directory where the executable file resides)/cty.xml

## Tools

* ctyxmldump: Dumping cty.xml loaded data as maps
* dxcccl: search the database with a callsign and optional date/time
* See goadifdxcccl in [goadiftools](https://github.com/jj1bdx/goadiftools)

## LICENSE

MIT

