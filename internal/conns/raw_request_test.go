package conns_test

import (
	"io"
	"net/http"
	"strings"
	"testing"

	"terraform-provider-kion/internal/conns"
	"terraform-provider-kion/internal/conns/mocks"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func respJSON(status int, body string) *http.Response {
	return &http.Response{StatusCode: status, Body: io.NopCloser(strings.NewReader(body))}
}

func TestRawGet_ReturnsBodyAndAuth(t *testing.T) {
	t.Parallel()
	doer := mocks.NewMockHTTPDoer(t)
	var got *http.Request
	doer.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
		got = req
		return respJSON(http.StatusOK, `{"data":1}`), nil
	})
	// APIURL is the API root as Configure resolves it, base URL + apipath.
	c := &conns.KionClient{APIURL: "https://kion.example.com/api/", HTTPClient: doer, APIKey: "k"}
	body, err := c.RawGet(t.Context(), "/beta/dashboard/7")
	require.NoError(t, err)
	require.Equal(t, `{"data":1}`, string(body))
	require.Equal(t, http.MethodGet, got.Method)
	require.Equal(t, "https://kion.example.com/api/beta/dashboard/7", got.URL.String())
	require.Equal(t, "Bearer k", got.Header.Get("Authorization"))
}

// A custom or empty `apipath` must reach the raw endpoints too: the helpers
// append to APIURL verbatim rather than re-adding a hardcoded "/api".
func TestRawGet_HonorsResolvedAPIPath(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, apiURL, want string }{
		{"apipath emptied", "https://kion.example.com", "https://kion.example.com/v1/payer/3"},
		{"custom apipath", "https://kion.example.com/custom", "https://kion.example.com/custom/v1/payer/3"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			doer := mocks.NewMockHTTPDoer(t)
			var got *http.Request
			doer.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
				got = req
				return respJSON(http.StatusOK, `{}`), nil
			})
			c := &conns.KionClient{APIURL: tc.apiURL, HTTPClient: doer, APIKey: "k"}
			_, err := c.RawGet(t.Context(), "/v1/payer/3")
			require.NoError(t, err)
			require.Equal(t, tc.want, got.URL.String())
		})
	}
}

func TestRawPost_SendsBody(t *testing.T) {
	t.Parallel()
	doer := mocks.NewMockHTTPDoer(t)
	var sent []byte
	doer.EXPECT().Do(mock.Anything).RunAndReturn(func(req *http.Request) (*http.Response, error) {
		b, rerr := io.ReadAll(req.Body)
		require.NoError(t, rerr)
		sent = b
		return respJSON(http.StatusCreated, `{"record_id":5}`), nil
	})
	c := &conns.KionClient{APIURL: "https://kion.example.com", HTTPClient: doer, AuthToken: "t"}
	body, err := c.RawPost(t.Context(), "/v3/app-role", []byte(`{"name":"x"}`))
	require.NoError(t, err)
	require.Equal(t, `{"record_id":5}`, string(body))
	require.Equal(t, `{"name":"x"}`, string(sent))
}

func TestRawGet_NotFoundIsTyped(t *testing.T) {
	t.Parallel()
	doer := mocks.NewMockHTTPDoer(t)
	doer.EXPECT().Do(mock.Anything).Return(respJSON(http.StatusNotFound, `{"error":"nope"}`), nil)
	c := &conns.KionClient{APIURL: "https://kion.example.com", HTTPClient: doer, APIKey: "k"}
	_, err := c.RawGet(t.Context(), "/v3/app-role/99")
	require.Error(t, err)
	require.True(t, conns.IsRawNotFound(err))
}

func TestRawPatch_ErrorStatus(t *testing.T) {
	t.Parallel()
	doer := mocks.NewMockHTTPDoer(t)
	doer.EXPECT().Do(mock.Anything).Return(respJSON(http.StatusBadRequest, `bad`), nil)
	c := &conns.KionClient{APIURL: "https://kion.example.com", HTTPClient: doer, APIKey: "k"}
	_, err := c.RawPatch(t.Context(), "/v3/webhook/1", []byte(`{}`))
	require.Error(t, err)
	require.False(t, conns.IsRawNotFound(err))
}
