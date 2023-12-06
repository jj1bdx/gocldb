# gocldb:  Club Log DXCC database handling library

## How to use

* Run `gocldb.LoadCtyXml()` to initialize the database
  - Takes a few seconds to startup
* The debug log output is *enabled* by default
  - Use `gocldb.DebugLogger.SetOutput(io.Discard)` to disable debug output
* Use `gocldb.CheckCallsign(call, qsotime)` to search the databse
  - result in `gocldb.CLDCheckResult` format defined in checkcall.go
* See dxcccl command source code for the basic usage details

## Usage example

```go
// Initialize the database
gocldb.LoadCtyXml()
// Disable debug logging
if !(*debugmode) {
  gocldb.DebugLogger.SetOutput(io.Discard)
}
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

