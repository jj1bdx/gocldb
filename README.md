# gocldb:  Club Log DXCC database handling library

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
* dxcldb (TBD): search the database with a callsign and optional date/time
* goadifcldb (TBD): filling in the following entries of ADIF by checking the callsign/datetime pair with the cty.xml database
  - input: call, qso_date/time_on pair or qso_date_off/time_off pair
  - output added: country, cont, cqz, dxcc
  - (Will be moved to [goadiftools](https://github.com/jj1bdx/goadiftools))

## LICENSE

MIT

