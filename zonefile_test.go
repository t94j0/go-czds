package main_test

import (
	"compress/gzip"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"path"
	"strings"
	"testing"

	. "github.com/t94j0/go-czds"
)

const ExampleZone string = "test.com	86400	in	a	127.0.0.1\n"

func makeZoneServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		TargetPath := "/czds/downloads/"

		if !strings.HasPrefix(r.URL.Path, TargetPath) {
			http.NotFound(w, r)
			return
		}

		if r.Header.Get("Authorization") != "Bearer "+TargetAccessToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		requestedZone := strings.TrimPrefix(r.URL.Path, TargetPath)
		switch requestedZone {
		case "badrequest.zone":
			http.Error(w, "malformed request", http.StatusBadRequest)
			return
		case "forbidden.zone":
			w.Header().Set("Content-Type", "text/dns")
			http.Error(w, "", http.StatusForbidden)
			return
		case "toc.zone":
			http.Error(w, "", http.StatusConflict)
			return
		case "notexist.zone":
			http.Error(w, "", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "text/dns")
		gz := gzip.NewWriter(w)
		gz.Write([]byte(ExampleZone))
		gz.Close()
	}))
}

func makeZoneCZDS(t *testing.T) (*CZDS, error) {
	authStub := makeAuthServer(t)
	zoneStub := makeZoneServer(t)

	icann := New(TargetUsername, TargetPassword)
	icann.BaseURL = zoneStub.URL
	icann.AuthURL = authStub.URL

	if err := icann.RefreshAccessToken(); err != nil {
		return nil, err
	}

	return icann, nil
}

func TestZoneFile_Save(t *testing.T) {
	icann, err := makeZoneCZDS(t)
	if err != nil {
		t.Error(err)
	}

	dir, err := ioutil.TempDir("", "example")
	if err != nil {
		t.Error(err)
	}

	zone := icann.Zone("testing")
	if err := zone.Save(dir, "%s"); err != nil {
		t.Error(err)
	}

	data, err := ioutil.ReadFile(path.Join(dir, zone.TLD()))
	if err != nil {
		t.Error(err)
	}
	if string(data) != ExampleZone {
		t.Fail()
	}
}

func TestCZDS_Zone_noaccess(t *testing.T) {
	icann := New(TargetUsername, TargetPassword)
	if _, err := icann.Zone("testing").Reader(); err != ErrMustRefresh {
		t.Fail()
	}
}

func TestCZDS_Zone(t *testing.T) {
	icann, err := makeZoneCZDS(t)
	if err != nil {
		t.Error(err)
	}

	z, err := icann.Zone("testing").Reader()
	if err != nil {
		t.Error(err)
	}

	data, _ := ioutil.ReadAll(z)
	if string(data) != ExampleZone {
		t.Fail()
	}
}

func TestCZDS_Zone_BadRequest(t *testing.T) {
	icann, err := makeZoneCZDS(t)
	if err != nil {
		t.Error(err)
	}

	if _, err := icann.Zone("badrequest").Reader(); err == nil {
		t.Fail()
	}
}

func TestCZDS_Zone_Forbidden(t *testing.T) {
	icann, err := makeZoneCZDS(t)
	if err != nil {
		t.Error(err)
	}

	if _, err := icann.Zone("forbidden").Reader(); err != ErrZoneUnavailable {
		t.Fail()
	}
}

func TestCZDS_Zone_NotExist(t *testing.T) {
	icann, err := makeZoneCZDS(t)
	if err != nil {
		t.Error(err)
	}

	if _, err := icann.Zone("notexist").Reader(); err != ErrZoneNotExist {
		t.Fail()
	}
}

func TestCZDS_Zone_Conflict(t *testing.T) {
	icann, err := makeZoneCZDS(t)
	if err != nil {
		t.Error(err)
	}

	if _, err := icann.Zone("toc").Reader(); err != ErrTOC {
		t.Fail()
	}
}
