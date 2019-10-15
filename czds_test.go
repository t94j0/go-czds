package czds_test

import (
	"fmt"
	"io/ioutil"

	czds "github.com/t94j0/go-czds"
)

func ExampleCZDS_DownloadAll() {
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
