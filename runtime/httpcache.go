package runtime

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"iter"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"net/http/httputil"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/tronbyt/pixlet/runtime/modules/starlarkhttp"
)

const (
	MinRequestTTL      = 5 * time.Second
	MaxResponseTTL     = 1 * time.Hour
	MaxResponseDefault = 20 * 1024 * 1024 // 20MB
	MaxResponseEnv     = "PIXLET_HTTP_MAX_RESPONSE_MB"
	HTTPCachePrefix    = "httpcache"
	TTLHeader          = "X-Tidbyt-Cache-Seconds"
)

// Status codes that are cacheable as defined here:
// https://developer.mozilla.org/en-US/docs/Glossary/Cacheable
var cacheableStatusCodes = []int{200, 201, 202, 203, 204, 206, 300, 301, 404, 405, 410, 414, 501}

type cacheClient struct {
	cache            Cache
	transport        http.RoundTripper
	MaxResponseBytes int64
}

func InitHTTP(cache Cache) {
	cc := &cacheClient{
		cache:            cache,
		transport:        http.DefaultTransport,
		MaxResponseBytes: MaxResponseDefault,
	}

	if rawVal := os.Getenv(MaxResponseEnv); rawVal != "" {
		if parsedVal, err := strconv.ParseInt(rawVal, 10, 64); err == nil {
			cc.MaxResponseBytes = parsedVal << 20
			starlarkhttp.MaxResponseBytes.Store(cc.MaxResponseBytes)
		} else {
			slog.Warn(MaxResponseEnv+" is invalid; using default", "error", err)
		}
	}

	httpClient := &http.Client{
		Transport: cc,
		Timeout:   starlarkhttp.HTTPTimeout * 2,
	}
	starlarkhttp.StarlarkHTTPClient = httpClient
}

// RoundTrip is an approximation of what our internal HTTP proxy does. It should
// behave the same way, and any discrepancy should be considered a bug.
func (c *cacheClient) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	key, err := cacheKey(req)
	if err != nil {
		return nil, fmt.Errorf("generating cache key: %w", err)
	}

	if req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodPost {
		b, exists, err := c.cache.Get(ctx, key)
		if exists && err == nil {
			if res, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(b)), req); err == nil {
				res.Header.Set("tidbyt-cache-status", "HIT")
				return res, nil
			}
		}
	}

	resp, err := c.transport.RoundTrip(req.WithContext(ctx))
	if err == nil && c.MaxResponseBytes > 0 {
		resp.Body = http.MaxBytesReader(nil, resp.Body, c.MaxResponseBytes)
	}

	if err == nil && (req.Method == http.MethodGet || req.Method == http.MethodHead || req.Method == http.MethodPost) {
		ser, err := httputil.DumpResponse(resp, true)
		if err != nil {
			// if httputil.DumpResponse fails, it leaves the response body in an
			// undefined state, so we cannot continue
			if cause := context.Cause(ctx); cause != nil {
				err = cause
			}
			return nil, fmt.Errorf("reading response body: %w", err)
		}

		ttl := DetermineTTL(req, resp, nil)
		_ = c.cache.Set(ctx, key, ser, int64(ttl.Seconds()))
		resp.Header.Set("tidbyt-cache-status", "MISS")
	}

	return resp, err
}

func cacheKey(req *http.Request) (string, error) {
	ttl := req.Header.Get(TTLHeader)
	req.Header.Del(TTLHeader)
	r, err := httputil.DumpRequest(req, true)
	if err != nil {
		return "", fmt.Errorf("serializing request: %w", err)
	}
	if ttl != "" {
		req.Header.Set(TTLHeader, ttl)
	}

	h := sha256.Sum256(r)
	key := hex.EncodeToString(h[:])

	app := req.Header.Get("X-Tidbyt-App")
	if app == "" {
		return key, nil
	}

	return fmt.Sprintf("%s:%s:%s", HTTPCachePrefix, app, key), nil
}

// DetermineTTL determines the TTL for a request based on the request and
// response. We first check request method / response status code to determine
// if we should actually cache the response. Then we check the headers passed in
// from starlark to see if the user configured a TTL. Finally, if the response
// is cachable but the developer didn't configure a TTL, we check the response
// to get a hint at what the TTL should be.
func DetermineTTL(req *http.Request, resp *http.Response, randSource *rand.Rand) time.Duration {
	ttl := determineTTL(req, resp)

	// Jitter the TTL by 10% and double check that it's still greater than the
	// minimum TTL. If it's not, return the minimum TTL. The main thing we want
	// to avoid is a TTL of 0 given it will be cached forever.
	ttl = jitterDuration(ttl, randSource)
	if ttl < MinRequestTTL {
		return MinRequestTTL
	}

	return ttl
}

func determineTTL(req *http.Request, resp *http.Response) time.Duration {
	// If the response is a 429, we want to cache the response for the duration
	// the remote server told us to wait before retrying.
	if resp.StatusCode == http.StatusTooManyRequests {
		retry := MinRequestTTL
		retryAfter := resp.Header.Get("Retry-After")
		if retryAfter != "" {
			if intValue, err := strconv.Atoi(retryAfter); err == nil {
				retry = time.Duration(intValue) * time.Second
			}
		}

		if retry < MinRequestTTL {
			return MinRequestTTL
		}

		return retry
	}

	// Check the status code to determine if the response is cacheable.
	if !slices.Contains(cacheableStatusCodes, resp.StatusCode) {
		return MinRequestTTL
	}

	// Determine the TTL based on the developer's configuration.
	ttl := determineDeveloperTTL(req)

	// We don't want to cache POST requests unless the developer explicitly
	// requests it.
	if ttl == 0 && req.Method != http.MethodGet && req.Method != http.MethodHead {
		return MinRequestTTL
	}

	// If the developer didn't configure a TTL, determine the TTL based on the
	// response.
	if ttl == 0 {
		ttl = determineResponseTTL(resp)
	}

	if ttl < MinRequestTTL {
		return MinRequestTTL
	}

	return ttl
}

func jitterDuration(duration time.Duration, source *rand.Rand) time.Duration {
	if duration <= 0 {
		return duration
	}

	offset := duration / 10
	randMax := 2*offset + 1

	var jitter time.Duration
	if source == nil {
		jitter = rand.N(randMax)
	} else {
		jitter = time.Duration(source.Uint64N(uint64(randMax)))
	}
	return duration + (jitter - offset)
}

func determineResponseTTL(resp *http.Response) time.Duration {
	for k, v := range parseCacheControl(resp.Header.Get("Cache-Control")) {
		if k != "max-age" {
			continue
		}

		intValue, ok := v.(int)
		if !ok {
			continue
		}

		ttl := time.Duration(intValue) * time.Second

		// If we're using a response TTL, we're making the assumption that
		// the remote server is providing a reasonable TTL that a developer
		// didn't configure. In the case of weathermap, the TTL is 1 week,
		// but the developer is requesting a new map every hour. So while the
		// old map _is_ valid for a week, the app only cares about it for
		// one hour.
		return min(ttl, MaxResponseTTL)
	}

	return 0
}

func determineDeveloperTTL(req *http.Request) time.Duration {
	ttlHeader := req.Header.Get("X-Tidbyt-Cache-Seconds")
	if ttlHeader != "" {
		if intValue, err := strconv.Atoi(ttlHeader); err == nil {
			return time.Duration(intValue) * time.Second
		}
	}

	return 0
}

func parseCacheControl(header string) iter.Seq2[string, any] {
	return func(yield func(string, any) bool) {
		for directive := range strings.SplitSeq(header, ",") {
			value := any(true)
			directive = strings.TrimSpace(directive)
			key, strValue, ok := strings.Cut(directive, "=")
			if ok {
				if intValue, err := strconv.Atoi(strValue); err == nil {
					value = intValue
				} else {
					value = strValue
				}
			}

			if !yield(strings.ToLower(key), value) {
				return
			}
		}
	}
}
