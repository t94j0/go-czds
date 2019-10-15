package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/gammazero/workerpool"
	"github.com/pkg/errors"
)

// ZoneList is a list of zone download links
type ZoneList []string

// NewZoneListFromBody creates a zone download link list from an HTTP body specified in section 5.1 of the API documentation
func NewZoneListFromBody(httpBody io.ReadCloser) (ZoneList, error) {
	body, err := ioutil.ReadAll(httpBody)
	if err != nil {
		return ZoneList{}, err
	}

	rsp := []string{}

	if err := json.Unmarshal(body, &rsp); err != nil {
		return ZoneList{}, err
	}

	return ZoneList(rsp), nil
}

// Zones gets the list of zones without the entire download link.
//
// For example, `https://czds-api.icann.org/czds/downloads/xbox.zone` becomes `xbox`
func (zl ZoneList) Zones() []string {
	cutList := make([]string, 0)
	for _, zone := range zl {
		splitlist := strings.Split(zone, "/")
		zoneExt := splitlist[len(splitlist)-1]
		zone := strings.Split(zoneExt, ".")[0]
		cutList = append(cutList, zone)
	}
	return cutList
}

// DownloadZoneReadersThreaded downloads all zone files and returns a multireader with all zonefiles at the same time
func (c *CZDS) DownloadZoneReadersThreaded(zl ZoneList, dir, fileFormat string, threads int) error {
	wp := workerpool.New(threads)

	for _, z := range zl.Zones() {
		z := z
		wp.Submit(func() {
			if err := c.Zone(z).Save(dir, fileFormat); err != nil {
				log.Println(err)
			}
		})
	}

	wp.StopWait()
	return nil
}

// DownloadZoneReaders applies DownloadZoneReadersThreaded with one thread
func (c *CZDS) DownloadZoneReaders(zl ZoneList, dir, fileFormat string) error {
	return c.DownloadZoneReadersThreaded(zl, dir, fileFormat, 1)
}

func (c *CZDS) makeZoneRequest() (*http.Response, error) {
	req, err := http.NewRequest("GET", c.BaseURL+"/czds/downloads/links", nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", "CRS / 0.1 Authentiction")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// ZoneList returns all TLD zones available to an access token
func (c *CZDS) ZoneList() (ZoneList, error) {
	if c.accessToken == "" {
		return ZoneList{}, ErrMustRefresh
	}

	resp, err := c.makeZoneRequest()
	if err != nil {
		return ZoneList{}, errors.Wrap(err, "Failed to make zone request")
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return ZoneList{}, errors.Wrap(err, "Malformed request")
	case http.StatusInternalServerError:
		return ZoneList{}, errors.Wrap(err, "ICANN internal server error")
	}

	zl, err := NewZoneListFromBody(resp.Body)
	if err != nil {
		return ZoneList{}, err
	}

	return zl, nil
}
