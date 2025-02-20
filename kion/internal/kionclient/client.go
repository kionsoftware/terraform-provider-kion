// Package kionclient provides a client for interacting with the Kion application.
package kionclient

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"strings"
)

type RequestError struct {
	StatusCode int
	Err        error
}

func (r RequestError) Error() string {
	return r.Err.Error()
}

func NewRequestError(statusCode int, err error) error {
	return &RequestError{StatusCode: statusCode, Err: err}
}

// Client represents a client to interact with the Kion application.
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient creates a new Client instance.
func NewClient(kionURL, kionAPIKey, kionAPIPath string, skipSSLValidation bool) *Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipSSLValidation}

	client := &Client{
		HTTPClient: &http.Client{
			Transport: customTransport,
		},
		Token: kionAPIKey,
	}

	u, err := url.Parse(kionURL)
	if err != nil {
		log.Fatalf("The URL is not valid: %s, %v", kionURL, err)
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), strings.TrimRight(kionAPIPath, "/"))
	client.HostURL = u.String()

	return client
}

func (client *Client) doRequest(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Authorization", "Bearer "+client.Token)

	res, err := client.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, NewRequestError(0, err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, NewRequestError(res.StatusCode, err)
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return nil, res.StatusCode, NewRequestError(res.StatusCode, fmt.Errorf("url: %s, method: %s, status: %d, body: %s", req.URL.String(), req.Method, res.StatusCode, body))
	}

	return body, res.StatusCode, nil
}

// GET retrieves an element from Kion.
func (client *Client) GET(urlPath string, returnData interface{}) error {
	if returnData != nil {
		v := reflect.ValueOf(returnData)
		if v.Kind() != reflect.Ptr {
			return errors.New("data must be a pointer, not a value")
		}
	}

	req, err := http.NewRequest(http.MethodGet, client.HostURL+urlPath, nil)
	if err != nil {
		return err
	}

	body, statusCode, err := client.doRequest(req)
	if err != nil {
		return err
	}

	if returnData != nil {
		if err := json.Unmarshal(body, returnData); err != nil {
			return NewRequestError(statusCode, fmt.Errorf("could not unmarshal response body: %v", string(body)))
		}
	}

	return nil
}

// POST creates an element in Kion.
func (client *Client) POST(urlPath string, sendData interface{}) (*Creation, error) {
	rb, err := json.Marshal(sendData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, client.HostURL+urlPath, bytes.NewBuffer(rb))
	if err != nil {
		return nil, err
	}

	body, _, err := client.doRequest(req)
	if err != nil {
		return nil, err
	}

	var data Creation
	if err := json.Unmarshal(body, &data); err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %v", string(body))
	}

	return &data, nil
}

// PATCH updates an element in Kion.
func (client *Client) PATCH(urlPath string, sendData interface{}) error {
	return client.doPutPatch(http.MethodPatch, urlPath, sendData)
}

// PUT updates an element in Kion.
func (client *Client) PUT(urlPath string, sendData interface{}) error {
	return client.doPutPatch(http.MethodPut, urlPath, sendData)
}

// doPutPatch is a helper for PUT and PATCH methods.
func (client *Client) doPutPatch(method, urlPath string, sendData interface{}) error {
	rb, err := json.Marshal(sendData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(method, client.HostURL+urlPath, bytes.NewBuffer(rb))
	if err != nil {
		return err
	}

	_, _, err = client.doRequest(req)
	return err
}

// DELETE removes an element from Kion. sendData can be nil.
func (client *Client) DELETE(urlPath string, sendData interface{}) error {
	return client.DeleteWithResponse(urlPath, sendData, nil)
}

// DeleteWithResponse deletes an element from Kion and returns a response.
func (client *Client) DeleteWithResponse(urlPath string, sendData, returnData interface{}) error {
	var req *http.Request
	var err error

	if sendData != nil {
		rb, err := json.Marshal(sendData)
		if err != nil {
			return err
		}
		req, err = http.NewRequest(http.MethodDelete, client.HostURL+urlPath, bytes.NewBuffer(rb))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest(http.MethodDelete, client.HostURL+urlPath, nil)
		if err != nil {
			return err
		}
	}

	body, statusCode, err := client.doRequest(req)
	if err != nil {
		return err
	}

	if returnData != nil {
		if err := json.Unmarshal(body, returnData); err != nil {
			return NewRequestError(statusCode, fmt.Errorf("could not unmarshal response body: %v", string(body)))
		}
	}

	return nil
}

// GETWithParams performs a GET request with query parameters
func (client *Client) GETWithParams(path string, params map[string]string, v interface{}) error {
	req, err := http.NewRequest("GET", client.HostURL+path, nil)
	if err != nil {
		return err
	}

	// Add query parameters
	q := req.URL.Query()
	for key, value := range params {
		q.Add(key, value)
	}
	req.URL.RawQuery = q.Encode()

	body, _, err := client.doRequest(req)
	if err != nil {
		return err
	}

	if v != nil {
		if err := json.Unmarshal(body, v); err != nil {
			return fmt.Errorf("could not unmarshal response body: %v", string(body))
		}
	}

	return nil
}
