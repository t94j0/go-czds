package czds_test

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/t94j0/array"
	. "github.com/t94j0/go-czds"
)

var zoneList string = `["https://czds-api.icann.org/czds/downloads/xbox.zone",
  "https://czds-api.icann.org/czds/downloads/test.zone"]`

func toReader(data string) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewReader([]byte(data)))
}

func makeZoneListServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/czds/downloads/links" {
			http.NotFound(w, r)
			return
		}

		if r.Header.Get("Authorization") != "Bearer "+TargetAccessToken {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		w.Write([]byte(zoneList))
	}))
}

func TestNewZoneListFromBody(t *testing.T) {
	r := toReader(zoneList)
	if _, err := NewZoneListFromBody(r); err != nil {
		t.Error(err)
	}
}

func TestNewZoneListFromBody_Zones(t *testing.T) {
	r := toReader(zoneList)
	zl, err := NewZoneListFromBody(r)
	if err != nil {
		t.Error(err)
	}

	zones := zl.Zones()

	if len(zones) != 2 && array.In("xbox", zones) && array.In("test", zones) {
		t.Fail()
	}
}

func TestCZDS_ZoneList_noaccess(t *testing.T) {
	zlStub := makeZoneListServer(t)

	icann := New(TargetUsername, TargetPassword)
	icann.BaseURL = zlStub.URL

	if _, err := icann.ZoneList(); err != ErrMustRefresh {
		t.Fail()
	}
}

func TestCZDS_ZoneList(t *testing.T) {
	authStub := makeAuthServer(t)
	zlStub := makeZoneListServer(t)

	icann := New(TargetUsername, TargetPassword)
	icann.AuthURL = authStub.URL
	icann.BaseURL = zlStub.URL

	if err := icann.RefreshAccessToken(); err != nil {
		t.Error(err)
	}

	zl, err := icann.ZoneList()
	if err != nil {
		t.Error(err)
	}
	if !array.In("https://czds-api.icann.org/czds/downloads/xbox.zone", zl) || !array.In("https://czds-api.icann.org/czds/downloads/test.zone", zl) {
		t.Fail()
	}
}

func TestCZDS_ZoneList_Zones(t *testing.T) {
	authStub := makeAuthServer(t)
	zlStub := makeZoneListServer(t)

	icann := New(TargetUsername, TargetPassword)
	icann.AuthURL = authStub.URL
	icann.BaseURL = zlStub.URL

	if err := icann.RefreshAccessToken(); err != nil {
		t.Error(err)
	}

	zl, err := icann.ZoneList()
	if err != nil {
		t.Error(err)
	}
	zs := zl.Zones()
	if !array.In("xbox", zs) || !array.In("test", zs) {
		t.Fail()
	}
}
