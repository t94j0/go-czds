package main

type CZDS struct {
	AuthURL     string
	BaseURL     string
	username    string
	password    string
	accessToken string
}

// New creates a new authenticated CZDS object
func New(username, password string) *CZDS {
	return &CZDS{
		AuthURL:  DefaultAuthURL,
		BaseURL:  DefaultBaseURL,
		username: username,
		password: password,
	}
}
