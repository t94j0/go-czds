package czds

import (
	"bytes"
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
)

// DefaultAuthURL is the default ICANN authentication URL
const DefaultAuthURL = "https://account-api.icann.org/api/authenticate"

// DefaultBaseURL is the default ICANN CZDS base URL for getting information about zone files
const DefaultBaseURL = "https://czds-api.icann.org"

// ErrMustRefresh is returned as an error when the RefreshAccessToken function must be run again
var ErrMustRefresh = errors.New("Must refresh access token")

// ErrIncorrectCredentials is returned when the submitted ICANN credentials are incorrect
var ErrIncorrectCredentials = errors.New("ICANN Credentials incorrect")

// ErrManyAuth is returned when one has submitted too many authentication requests
var ErrManyAuth = errors.New("Too many authentications too quickly")

// authData creates a JSON object with the username and password authentication data
func (c *CZDS) authData() ([]byte, error) {
	auth := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{Username: c.username, Password: c.password}

	authData, err := json.Marshal(auth)
	if err != nil {
		return authData, err
	}

	return authData, nil
}

// makeAuthRequest takes a JSON formatted byte string and makes a request to get the ICANN CZDS accessToken
func (c *CZDS) makeAuthRequest(auth []byte) (*http.Response, error) {
	req, err := http.NewRequest("POST", c.AuthURL, bytes.NewBuffer(auth))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "CRS / 0.1 Authentiction")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

// getAccessToken gets the accessToken from an http body
func (c *CZDS) getAccessToken(httpBody io.ReadCloser) (string, error) {
	body, err := ioutil.ReadAll(httpBody)
	if err != nil {
		return "", err
	}

	response := struct {
		AccessToken string `json:"accessToken"`
	}{}

	if err := json.Unmarshal(body, &response); err != nil {
		return "", err
	}

	return response.AccessToken, nil
}

// RefreshAccessToken reauthenticates requests
func (c *CZDS) RefreshAccessToken() error {
	auth, err := c.authData()
	if err != nil {
		return errors.Wrap(err, "Unable to create auth data")
	}

	resp, err := c.makeAuthRequest(auth)
	if err != nil {
		return errors.Wrap(err, "Failed to make request")
	}

	switch resp.StatusCode {
	case http.StatusBadRequest:
		return errors.Wrap(err, "Request malformed")
	case http.StatusUnauthorized:
		return ErrIncorrectCredentials
	case http.StatusTooManyRequests:
		return ErrManyAuth
	case http.StatusInternalServerError:
		return errors.Wrap(err, "ICANN internal server error")
	}

	token, err := c.getAccessToken(resp.Body)
	if err != nil {
		return errors.Wrap(err, "Unable to get access token from body")
	}

	c.accessToken = token

	return nil
}
