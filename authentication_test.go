package main_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	. "github.com/t94j0/go-czds"
)

const TargetUsername = "test_username"
const TargetPassword = "test_password"
const TargetAccessToken = "ABC"

func getAuth(r *http.Request) (string, string, error) {
	req := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", "", err
	}

	if err := json.Unmarshal(body, &req); err != nil {
		return "", "", err
	}

	return req.Username, req.Password, nil
}

func makeAuthServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username, password, err := getAuth(r)
		if err != nil {
			t.Error(err)
		}

		if username != TargetUsername || password != TargetPassword {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		w.Write([]byte("{\"accessToken\":\"" + TargetAccessToken + "\"}"))
	}))
}

func makeTooManyAuthServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "too many auth", http.StatusTooManyRequests)
	}))
}

func TestCZDS_RefreshAccess_success(t *testing.T) {
	apiStub := makeAuthServer(t)
	icann := New(TargetUsername, TargetPassword)
	icann.AuthURL = apiStub.URL
	if err := icann.RefreshAccessToken(); err != nil {
		t.Error(err)
	}
}

func TestCZDS_RefreshAccess_TooMany(t *testing.T) {
	apiStub := makeTooManyAuthServer(t)

	icann := New(TargetUsername, TargetPassword)
	icann.AuthURL = apiStub.URL
	if err := icann.RefreshAccessToken(); err != ErrManyAuth {
		t.Fail()
	}
}

func TestCZDS_RefreshAccess_badcreds(t *testing.T) {
	apiStub := makeAuthServer(t)
	icann := New(TargetUsername, "abc")
	icann.AuthURL = apiStub.URL
	if err := icann.RefreshAccessToken(); err != ErrIncorrectCredentials {
		t.Fail()
	}
}
