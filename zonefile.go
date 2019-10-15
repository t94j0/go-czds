package main

import (
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"

	"github.com/pkg/errors"
)

// ErrZoneUnavailable is returned when the zone exists, but the supplied user account does not have access to it
var ErrZoneUnavailable = errors.New("Zone unavailable to you")

// ErrZoneNotExist is returned when the requested zone file does not exist
var ErrZoneNotExist = errors.New("Zone file does not exist")

// ErrTOC is returned when the user must accept ICANN CZDS terms and conditions before making a request
var ErrTOC = errors.New("User must accept ICANN CZDS terms and conditions")

// ZoneFile is an object which contains the zone file information
type ZoneFile struct {
	err    error
	tld    string
	reader io.Reader
}

// Save the file to disk given a directory
func (zf ZoneFile) Save(dir, fileFormat string) error {
	if zf.err != nil {
		return zf.err
	}

	fileName := fmt.Sprintf(fileFormat, zf.tld)
	outPath := path.Join(dir, fileName)
	outFile, err := os.Create(outPath)
	if err != nil {
		return err
	}

	if _, err := io.Copy(outFile, zf.reader); err != nil {
		return err
	}

	return nil
}

// TLD returns the TLD for the zonefile
func (zf ZoneFile) TLD() string {
	return zf.tld
}

// Reader returns the io.Reader with the zone file
func (zf ZoneFile) Reader() (io.Reader, error) {
	if zf.err != nil {
		return nil, zf.err
	}
	return zf.reader, nil
}

// Zone downloads a zonefile
func (c *CZDS) Zone(tld string) ZoneFile {
	if c.accessToken == "" {
		return ZoneFile{err: ErrMustRefresh}
	}

	url := fmt.Sprintf("%s/czds/downloads/%s.zone", c.BaseURL, tld)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ZoneFile{err: err}
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("User-Agent", "CRS / 0.1 Authentiction")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return ZoneFile{err: err}
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return ZoneFile{err: errors.New("Malformed request")}
	case http.StatusForbidden:
		return ZoneFile{err: ErrZoneUnavailable}
	case http.StatusNotFound:
		return ZoneFile{err: ErrZoneNotExist}
	case http.StatusConflict:
		return ZoneFile{err: ErrTOC}
	}

	bodyReader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return ZoneFile{err: err}
	}

	return ZoneFile{tld: tld, reader: bodyReader}
}
