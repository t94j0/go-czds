# Go CZDS API

[![GoDoc](https://godoc.org/github.com/t94j0/go-czds?status.svg)](https://godoc.org/github.com/t94j0/go-czds)

Makes life easy Go developers who work with CZDS. I made this for a long-term research project I'm working on, so it'll likely be supported for a good bit.

## Example

Simple way to get all zone files available.

```go
func main() {
  // Create CZDS object with username and password
  icann := czds.New("example", "password")
  // Get access token
  icann.RefreshAccessToken()

  // Get all available zones reachable by user
  zl, _ := icann.ZoneList()
  dir, _ := ioutil.TempDir("", "example")
  fmt.Println("Output to:", dir)

  // Download all zones into `dir` with the name <tld>.zone
  icann.DownloadZoneReaders(zl, dir, "%s.zone")
}
```
