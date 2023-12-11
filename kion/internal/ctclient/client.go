// Package ctclient provides a client for interacting with the Kion
// application.
package ctclient

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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
func NewClient(ctURL string, ctAPIKey string, ctAPIPath string, skipSSLValidation bool) *Client {
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: skipSSLValidation}

	c := Client{
		HTTPClient: &http.Client{
			Transport: customTransport,
		},
	}

	// Append the path to the URL.
	u, err := url.Parse(ctURL)
	if err != nil {
		log.Fatalln("The URL is not valid:", ctURL, err.Error())
	}
	u.Path = path.Join(strings.TrimRight(u.Path, "/"), strings.TrimRight(ctAPIPath, "/"))
	c.HostURL = u.String()

	c.Token = ctAPIKey

	return &c
}

func (c *Client) doRequest(req *http.Request) ([]byte, int, error) {
	req.Header.Set("Authorization", "Bearer "+c.Token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, 0, NewRequestError(0, err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, res.StatusCode, NewRequestError(res.StatusCode, err)
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated {
		return nil, res.StatusCode, NewRequestError(res.StatusCode, fmt.Errorf("url: %s, method: %s, status: %d, body: %s", req.URL.String(), req.Method, res.StatusCode, body))
	}

	return body, res.StatusCode, nil
}

// GET - Returns an element from CT.
func (c *Client) GET(urlPath string, returnData interface{}) error {
	if returnData != nil {
		// Ensure the correct returnData was passed in.
		v := reflect.ValueOf(returnData)
		if v.Kind() != reflect.Ptr {
			return errors.New("data must pass a pointer, not a value")
		}
	}

	req, err := http.NewRequest("GET", fmt.Sprintf("%s%s", c.HostURL, urlPath), nil)
	if err != nil {
		return err
	}

	body, statusCode, err := c.doRequest(req)
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

// POST - creates an element in CT.
func (c *Client) POST(urlPath string, sendData interface{}) (*Creation, error) {
	//return nil, fmt.Errorf("test error: %v %v %#v", c.HostURL, urlPath, sendData)
	rb, err := json.Marshal(sendData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s%s", c.HostURL, urlPath), strings.NewReader(string(rb)))
	if err != nil {
		return nil, err
	}

	body, _, err := c.doRequest(req)
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

// PATCH - updates an element in CT.
func (c *Client) PATCH(urlPath string, sendData interface{}) error {
	rb, err := json.Marshal(sendData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PATCH", fmt.Sprintf("%s%s", c.HostURL, urlPath), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, _, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// PUT - updates an element in CT.
func (c *Client) PUT(urlPath string, sendData interface{}) error {
	rb, err := json.Marshal(sendData)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("PUT", fmt.Sprintf("%s%s", c.HostURL, urlPath), strings.NewReader(string(rb)))
	if err != nil {
		return err
	}

	_, _, err = c.doRequest(req)
	if err != nil {
		return err
	}

	return nil
}

// DELETE - removes an element from CT. sendData can be nil.
func (c *Client) DELETE(urlPath string, sendData interface{}) error {
	return c.DeleteWithResponse(urlPath, sendData, nil)
}

func (c *Client) DeleteWithResponse(urlPath string, sendData interface{}, returnData interface{}) error {
	var req *http.Request
	var err error

	if sendData != nil {
		rb, err := json.Marshal(sendData)
		if err != nil {
			return err
		}

		req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s", c.HostURL, urlPath), strings.NewReader(string(rb)))
		if err != nil {
			return err
		}
	} else {
		req, err = http.NewRequest("DELETE", fmt.Sprintf("%s%s", c.HostURL, urlPath), nil)
		if err != nil {
			return err
		}
	}

	body, statusCode, err := c.doRequest(req)
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
