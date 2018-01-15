package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

const (
	// ScaleIODefaultTokenExpiration is the ScaleIO token expiration time.
	ScaleIODefaultTokenExpiration = 8 * time.Hour
	// ScaleIODefaultInactivityTimeout is the ScaleIO token expiration time without
	// activity. We need this for long running task setups.
	ScaleIODefaultInactivityTimeout = 10 * time.Minute
	// ClientDefaultTokenExpiration is the ScaleIO token expiration time divided by 2
	// this accounts for worst case scenarios to ensure we are authenticated.
	ClientDefaultTokenExpiration = ScaleIODefaultTokenExpiration / 2
	// ClientDefaultInactivityTimeout is the ScaleIO token expiration time devided by 2
	// this accounts for worst case scenarios to ensure we are authenticated.
	ClientDefaultInactivityTimeout = ScaleIODefaultInactivityTimeout / 2
)

// SIOClient stores client details for usage without needing to reauth
type SIOClient struct {
	token           string
	client          *http.Client
	address         *url.URL
	verifySSL       bool
	username        string
	password        string
	tokenExpiration time.Time
	lastAccessTime  time.Time
}

// NewSIOClient composes the SIO Client with default values and does a basic auth
func NewSIOClient(gateway string, username string, password string, verifySSL bool) (*SIOClient, error) {
	s := &SIOClient{}
	var c *http.Client
	if !verifySSL {
		c = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		}
	} else {
		c = &http.Client{}
	}
	s.client = c
	u, err := url.Parse(gateway)
	if err != nil {
		return &SIOClient{}, fmt.Errorf("Error while parsing gateway URL: %v", err)
	}
	s.address = u
	s.username = username
	s.password = password
	return s, nil
}

// Authenticate regenerates a token and creates a tokentracker to reauth for us
func (c *SIOClient) Authenticate() error {
	// if this isn't the first attempt, see if the there is time remaining on the token
	if c.token != "" {
		// token could be valid based on time but expired due to inactivity
		// check that it is valid and it isn't timed out due to inactivty
		// if the expiration time is after now and if the last event time is
		// after now, it is still valid
		now := time.Now()
		if c.tokenExpiration.After(now) &&
			c.lastAccessTime.Add(ClientDefaultInactivityTimeout).After(now) {
			return nil
		}
		if err := c.Logout(); err != nil {
			return err
		}
	}
	loginURL := &url.URL{}
	//Make a copy of the base URL
	*loginURL = *c.address
	loginURL.Path = "/api/login"
	req, _ := http.NewRequest("GET", loginURL.String(), nil)
	req.Header.Add("Authorization", "Basic "+base64.StdEncoding.EncodeToString([]byte(c.username+":"+c.password)))
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error while logging in to ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	// Strip out the quotes
	body = bytes.Trim(body, "\"")
	body = append([]byte(":"), body...)
	c.token = base64.StdEncoding.EncodeToString(body)
	c.tokenExpiration = time.Now().Add(ClientDefaultTokenExpiration)
	return nil
}

// Logout invalidates current token and cient session
func (c *SIOClient) Logout() error {
	loginURL := &url.URL{}
	//Make a copy of the base URL
	*loginURL = *c.address
	loginURL.Path = "/api/logout"
	req, _ := http.NewRequest("GET", loginURL.String(), nil)
	req.Header.Add("Authorization", "Basic "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error while logging out of ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	// clear the token for a new one
	c.token = ""
	return nil
}

// GetAPIResponse takes a path and returns the data into the provided object
func (c *SIOClient) GetAPIResponse(path string, v interface{}) error {
	c.updatelastAccessTime() // needed to coordinate client reauth
	fullURL := &url.URL{}
	*fullURL = *c.address
	fullURL.Path = path
	req, _ := http.NewRequest("GET", fullURL.String(), nil)
	req.Header.Add("Authorization", "Basic "+c.token)
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("Error while accessing ScaleIO API: %v", err)
	}
	defer resp.Body.Close()
	// this should never happen but if it does, we will take the error and try
	// to reauth next collection interval
	if resp.StatusCode == http.StatusUnauthorized {
		// auth failed so invalidate the token
		c.Logout()
		return fmt.Errorf("Error while accessing the ScaleIO API: Token Invalid")
	}
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		return fmt.Errorf("Error while parsing data from %s: %v", path, err)
	}
	return nil
}

func (c *SIOClient) updatelastAccessTime() {
	c.lastAccessTime = time.Now()
}
