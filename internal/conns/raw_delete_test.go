package conns_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/conns/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// resp204 is a minimal successful HTTP response with a closable, readable body.
func resp204() *http.Response {
	return &http.Response{StatusCode: http.StatusNoContent, Body: io.NopCloser(strings.NewReader(""))}
}

func TestRawDelete_BuildsRequestAndSendsAPIKey(t *testing.T) {
	t.Parallel()

	doer := mocks.NewMockHTTPDoer(t)
	var got *http.Request
	doer.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
		got = req
		return resp204(), nil
	})

	c := &conns.KionClient{APIURL: "https://kion.example.com/api/", HTTPClient: doer, APIKey: "secret-key"}
	err := c.RawDelete(t.Context(), "/v2/ou/42")

	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, http.MethodDelete, got.Method)
	require.Equal(t, "https://kion.example.com/api/v2/ou/42", got.URL.String())
	require.Equal(t, "application/json", got.Header.Get("Content-Type"))
	require.Equal(t, "Bearer secret-key", got.Header.Get("Authorization"))
}

func TestRawDelete_FallsBackToAuthToken(t *testing.T) {
	t.Parallel()

	doer := mocks.NewMockHTTPDoer(t)
	var got *http.Request
	doer.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
		got = req
		return resp204(), nil
	})

	c := &conns.KionClient{APIURL: "https://kion.example.com", HTTPClient: doer, AuthToken: "bearer-tok"}
	err := c.RawDelete(t.Context(), "/v2/ou/7")

	require.NoError(t, err)
	require.Equal(t, "Bearer bearer-tok", got.Header.Get("Authorization"))
}

func TestRawDelete_TransportErrorIsWrapped(t *testing.T) {
	t.Parallel()

	doer := mocks.NewMockHTTPDoer(t)
	doer.EXPECT().Do(mock.Anything).Return(nil, errors.New("boom"))

	c := &conns.KionClient{APIURL: "https://kion.example.com", HTTPClient: doer, APIKey: "k"}
	err := c.RawDelete(t.Context(), "/v2/ou/1")

	require.Error(t, err)
	require.ErrorContains(t, err, "executing request")
	require.ErrorContains(t, err, "boom")
}

func TestRawDelete_ErrorStatusIsError(t *testing.T) {
	t.Parallel()

	doer := mocks.NewMockHTTPDoer(t)
	doer.EXPECT().Do(mock.Anything).Return(
		&http.Response{StatusCode: http.StatusInternalServerError, Body: io.NopCloser(strings.NewReader(""))}, nil)

	c := &conns.KionClient{APIURL: "https://kion.example.com", HTTPClient: doer, APIKey: "k"}
	err := c.RawDelete(t.Context(), "/v2/ou/1")

	require.Error(t, err)
	require.ErrorContains(t, err, "status 500")
}
