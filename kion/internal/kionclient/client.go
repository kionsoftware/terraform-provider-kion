// Package kionclient provides a client for interacting with the Kion
// application.
package kionclient

import (
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
	requestError := new(RequestError)
	requestError.Err = err
	requestError.StatusCode = statusCode
	return requestError
}

// Client -
type Client struct {
	HostURL    string
	HTTPClient *http.Client
	Token      string
}

// NewClient .
func NewClient(kionURL string, kionAPIKey string, kionAPIPath string, skipSSLValidation bool) *Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipSSLValidation}

	client := Client{
		HTTPClient: &http.Client{
			Transport: customTransport,
		},
	}

	// Append the path to the URL.
	u, err := url.Parse(kionURL)
	if err != nil {
		log.Fatalln("The URL is not valid:", kionURL, err.Error())
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), strings.TrimRight(kionAPIPath, "/"))
	client.HostURL = u.String()

	client.Token = kionAPIKey

	return &client
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

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, res.StatusCode, NewRequestError(res.StatusCode, fmt.Errorf("url: %s, method: %s, status: %d, body: %s", req.URL.String(), req.Method, res.StatusCode, body))
	}

	return body, res.StatusCode, nil
}

// GET - Returns an element from Kion.
func (client *Client) GET(urlPath string, returnData interface{}) error {
	if returnData != nil {
		// Ensure the correct returnData was passed in.
		v := reflect.ValueOf(returnData)
		if v.Kind() != reflect.Ptr {
			return errors.New("data must pass a pointer, not a value")
		}
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", client.HostURL, urlPath), nil)
	if err != nil {
		return err
	}

	body, statusCode, err := client.doRequest(req)
	if err != nil {
		return err
	}

	if returnData != nil {
		err = json.Unmarshal(body, returnData)
		if err != nil {
			return &RequestError{StatusCode: statusCode, Err: fmt.Errorf("could not unmarshal response body: %v", string(body))}
		}
	}

	return nil
}

// POST - creates an element in Kion.
func (client *Client) POST(urlPath string, sendData interface{}) (*Creation, error) {
	//return nil, fmt.Errorf("test error: %v %v %#v", client.HostURL, urlPath, sendData)
	rb, err := json.Marshal(sendData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", client.HostURL, urlPath), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, _, err := client.doRequest(req)
	if err != nil {
		return nil, err
	}

	data := Creation{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		return nil, fmt.Errorf("could not unmarshal response body: %v", string(body))
	}

	// We allow 200 on POST when updating owners so we can't use this logic.
	// if statusCode != http.StatusCreated {
	// 	return &data, fmt.Errorf("received status code: %v | %v", statusCode, string(body))
	// }

	return &data, nil
}

// PATCH - updates an element in Kion.
func (client *Client) PATCH(urlPath string, sendData interface{}) error {
	rb, err := json.Marshal(sendData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s%s", client.HostURL, urlPath), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, _, err = client.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// PUT - updates an element in Kion.
func (client *Client) PUT(urlPath string, sendData interface{}) error {
	rb, err := json.Marshal(sendData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s%s", client.HostURL, urlPath), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, _, err = client.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// DELETE - removes an element from Kion. sendData can be nil.
func (client *Client) DELETE(urlPath string, sendData interface{}) error {
	return client.DeleteWithResponse(urlPath, sendData, nil)
}

func (client *Client) DeleteWithResponse(urlPath string, sendData interface{}, returnData interface{}) error {
	var req *http.Request
	var err error

	if sendData != nil {
		rb, err := json.Marshal(sendData)
		if err != nil {
			return err
		}

		req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s", client.HostURL, urlPath), strings.NewReader(string(rb)))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s", client.HostURL, urlPath), nil)
		if err != nil {
			return err
		}
	}

	body, statusCode, err := client.doRequest(req)
	if err != nil {
		return err
	}

	if returnData != nil {
		err = json.Unmarshal(body, returnData)
		if err != nil {
			return &RequestError{StatusCode: statusCode, Err: fmt.Errorf("could not unmarshal response body: %v", string(body))}
		}
	}

	return nil
}
