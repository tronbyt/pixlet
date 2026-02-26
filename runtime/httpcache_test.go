package runtime

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestInitHTTP(t *testing.T) {
	c := NewInMemoryCache()
	t.Cleanup(c.Close)
	InitHTTP(c)

	b, err := os.ReadFile("testdata/httpcache.star")
	assert.NoError(t, err)

	app, err := NewApplet("httpcache.star", b, WithTests(t))
	assert.NoError(t, err)
	assert.NotNil(t, app)

	screens, err := app.Run(t.Context())
	assert.NoError(t, err)
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
