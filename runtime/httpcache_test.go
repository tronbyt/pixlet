package runtime

import (
	"bytes"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/tronbyt/pixlet/runtime/testdata"
)

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

func testServer(t *testing.T) string {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(server.Close)
	return server.URL
}

func TestInitHTTP(t *testing.T) {
	c := NewInMemoryCache()
	t.Cleanup(c.Close)
	InitHTTP(c)

	b, err := testdata.FS.ReadFile("httpcache.star")
	require.NoError(t, err)

	url := testServer(t)
	b = bytes.ReplaceAll(b, []byte("https://example.com"), []byte(url))

	app, err := NewApplet(t.Context(), "httpcache.star", b, WithTests(t))
	require.NoError(t, err)
	assert.NotNil(t, app)

	screens, err := app.Run(t.Context())
	require.NoError(t, err)
	assert.NotNil(t, screens)
}

// TestDetermineTTL tests the DetermineTTL function.
func TestDetermineTTL(t *testing.T) {
	type test struct {
		ttl         int
		retryAfter  int
		resHeader   string
		statusCode  int
		method      string
		expectedTTL time.Duration
	}

	tests := map[string]test{
		"test request cache control headers": {
			ttl:         3600,
			resHeader:   "",
			statusCode:  http.StatusOK,
			method:      "GET",
			expectedTTL: 3600 * time.Second,
		},
		"test response cache control headers": {
			ttl:         0,
			resHeader:   "public, max-age=3600, s-maxage=7200, no-transform",
			statusCode:  http.StatusOK,
			method:      "GET",
			expectedTTL: 3600 * time.Second,
		},
		"test too long response cache control headers": {
			ttl:         0,
			resHeader:   "max-age=604800",
			statusCode:  http.StatusOK,
			method:      "GET",
			expectedTTL: 3600 * time.Second,
		},
		"test max-age of zero": {
			ttl:         0,
			resHeader:   "max-age=0",
			statusCode:  http.StatusOK,
			method:      "GET",
			expectedTTL: 5 * time.Second,
		},
		"test both request and response cache control headers": {
			ttl:         3600,
			resHeader:   "public, max-age=60, s-maxage=7200, no-transform",
			statusCode:  http.StatusOK,
			method:      "GET",
			expectedTTL: 3600 * time.Second,
		},
		"test 500 response code": {
			ttl:         3600,
			resHeader:   "",
			statusCode:  http.StatusInternalServerError,
			method:      "GET",
			expectedTTL: 5 * time.Second,
		},
		"test too low ttl": {
			ttl:         3,
			resHeader:   "",
			statusCode:  http.StatusOK,
			method:      "GET",
			expectedTTL: 5 * time.Second,
		},
		"test DELETE request": {
			ttl:         0,
			resHeader:   "",
			statusCode:  http.StatusOK,
			method:      "DELETE",
			expectedTTL: 5 * time.Second,
		},
		"test POST request configured with TTL": {
			ttl:         30,
			resHeader:   "",
			statusCode:  http.StatusOK,
			method:      "POST",
			expectedTTL: 30 * time.Second,
		},
		"test POST request without configured TTL": {
			ttl:         0,
			resHeader:   "",
			statusCode:  http.StatusOK,
			method:      "POST",
			expectedTTL: 5 * time.Second,
		},
		"test 429 request": {
			ttl:         30,
			retryAfter:  60,
			resHeader:   "",
			statusCode:  http.StatusTooManyRequests,
			method:      "GET",
			expectedTTL: 60 * time.Second,
		},
		"test 429 request below minimum": {
			ttl:         30,
			retryAfter:  3,
			resHeader:   "",
			statusCode:  http.StatusTooManyRequests,
			method:      "GET",
			expectedTTL: 5 * time.Second,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			req := &http.Request{
				Header: map[string][]string{
					"X-Tidbyt-Cache-Seconds": {fmt.Sprintf("%d", tc.ttl)},
				},
				Method: tc.method,
			}

			res := &http.Response{
				Header: map[string][]string{
					"Cache-Control": {tc.resHeader},
				},
				StatusCode: tc.statusCode,
			}

			if tc.retryAfter > 0 {
				res.Header.Set("Retry-After", fmt.Sprintf("%d", tc.retryAfter))
			}

			ttl := determineTTL(req, res)
			assert.Equal(t, tc.expectedTTL, ttl)
		})
	}
}

func TestDetermineTTLJitter(t *testing.T) {
	req := &http.Request{
		Header: map[string][]string{
			"X-Tidbyt-Cache-Seconds": {"60"},
		},
		Method: http.MethodGet,
	}

	res := &http.Response{
		StatusCode: http.StatusOK,
	}

	r := rand.New(rand.NewPCG(1, 2))
	ttl := DetermineTTL(req, res, r)
	assert.Equal(t, 63, int(ttl.Seconds()))
}

func TestDetermineTTLNoHeaders(t *testing.T) {
	req := &http.Request{
		Method: http.MethodGet,
	}

	res := &http.Response{
		StatusCode: http.StatusOK,
	}

	ttl := DetermineTTL(req, res, nil)
	assert.LessOrEqual(t, MinRequestTTL, ttl)
}

func TestCacheKey(t *testing.T) {
	url := testServer(t)

	req := httptest.NewRequest(http.MethodGet, url+"/weather?zip=10001", nil)
	req.Header.Set("X-Tidbyt-App", "weather")
	req.Header.Set(TTLHeader, "60")

	key, err := cacheKey(req)
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, HTTPCachePrefix+":weather:"))
	assert.Equal(t, "60", req.Header.Get(TTLHeader))

	reqDifferentTTL := httptest.NewRequest(http.MethodGet, url+"/weather?zip=10001", nil)
	reqDifferentTTL.Header.Set("X-Tidbyt-App", "weather")
	reqDifferentTTL.Header.Set(TTLHeader, "120")
	keyDifferentTTL, err := cacheKey(reqDifferentTTL)
	require.NoError(t, err)
	assert.Equal(t, key, keyDifferentTTL)
}

func TestCacheKeyWithoutAppPrefix(t *testing.T) {
	url := testServer(t)

	req := httptest.NewRequest(http.MethodGet, url+"/noapp", nil)

	key, err := cacheKey(req)
	require.NoError(t, err)
	assert.Len(t, key, 64)
	assert.NotContains(t, key, HTTPCachePrefix+":")
}

func TestParseCacheControl(t *testing.T) {
	const header = "public, max-age=3600, s-maxage=7200, no-transform, private=token"
	for k, v := range parseCacheControl(header) {
		switch k {
		case "public":
			assert.Equal(t, true, v)
		case "max-age":
			assert.Equal(t, 3600, v)
		case "s-maxage":
			assert.Equal(t, 7200, v)
		case "no-transform":
			assert.Equal(t, true, v)
		case "private":
			assert.Equal(t, "token", v)
		}
	}
}

func TestDetermineResponseTTL(t *testing.T) {
	tests := map[string]struct {
		cacheControl string
		want         time.Duration
	}{
		"uses max-age": {
			cacheControl: "public, max-age=120",
			want:         120 * time.Second,
		},
		"caps long max-age": {
			cacheControl: "max-age=999999",
			want:         MaxResponseTTL,
		},
		"returns zero without max-age": {
			cacheControl: "public, no-cache",
			want:         0,
		},
		"returns zero for invalid max-age": {
			cacheControl: "max-age=abc",
			want:         0,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			resp := &http.Response{
				Header: http.Header{
					"Cache-Control": {tc.cacheControl},
				},
			}

			got := determineResponseTTL(resp)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestCacheClientRoundTripCachesResponses(t *testing.T) {
	cache := NewInMemoryCache()
	t.Cleanup(cache.Close)

	transportCalls := 0
	transport := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		transportCalls++
		return &http.Response{
			Status:        "200 OK",
			StatusCode:    http.StatusOK,
			Proto:         "HTTP/1.1",
			ProtoMajor:    1,
			ProtoMinor:    1,
			Header:        http.Header{},
			Body:          io.NopCloser(strings.NewReader(`{"ok":true}`)),
			ContentLength: int64(len(`{"ok":true}`)),
			Request:       req,
		}, nil
	})

	client := &cacheClient{
		cache:     cache,
		transport: transport,
	}

	url := testServer(t)

	req1 := httptest.NewRequest(http.MethodGet, url, nil)
	req1.Header.Set("X-Tidbyt-App", "weather")
	req1.Header.Set(TTLHeader, "60")

	resp1, err := client.RoundTrip(req1)
	require.NoError(t, err)
	assert.Equal(t, "MISS", resp1.Header.Get("tidbyt-cache-status"))
	assert.Equal(t, 1, transportCalls)

	req2 := httptest.NewRequest(http.MethodGet, url, nil)
	req2.Header.Set("X-Tidbyt-App", "weather")
	req2.Header.Set(TTLHeader, "60")

	resp2, err := client.RoundTrip(req2)
	require.NoError(t, err)
	assert.Equal(t, "HIT", resp2.Header.Get("tidbyt-cache-status"))
	assert.Equal(t, 1, transportCalls)
}
