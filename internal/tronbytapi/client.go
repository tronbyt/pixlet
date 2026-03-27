package tronbytapi

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"iter"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"time"
)

func NewClient(baseURL, apiToken string) (*Client, error) {
	creds, err := ResolveAPICredentials(baseURL, apiToken)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(creds.baseURL)
	if err != nil {
		return nil, err
	}

	return &Client{
		Client: &http.Client{
			Timeout: 30 * time.Second,
		},
		UserAgent: BuildUserAgent(),
		baseURL:   *u,
		apiToken:  creds.token,
	}, nil
}

type Client struct {
	Client    *http.Client
	UserAgent string
	baseURL   url.URL
	apiToken  string
}

var ErrAPIResponse = errors.New("tronbyt api error")

func (c *Client) makeRequest(ctx context.Context, method, subPath string, body io.Reader, result any) (*http.Response, error) {
	u := c.baseURL
	u.Path = path.Join(u.Path, "v0", subPath)

	req, err := http.NewRequestWithContext(ctx, method, u.String(), body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Authorization", "Bearer "+c.apiToken)

	if c.UserAgent != "" {
		req.Header.Add("User-Agent", BuildUserAgent())
	}

	res, err := c.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		body, _ := io.ReadAll(res.Body)
		return nil, fmt.Errorf("%w %s: %s", ErrAPIResponse, res.Status, body)
	}

	if result != nil {
		if err := json.NewDecoder(res.Body).Decode(result); err != nil {
			return nil, fmt.Errorf("decoding response json: %w", err)
		}
	}

	return res, nil
}

func (c *Client) GetDevices(ctx context.Context) iter.Seq2[*Device, error] {
	return func(yield func(*Device, error) bool) {
		var devices Devices
		if _, err := c.makeRequest(ctx, http.MethodGet, "devices", nil, &devices); err != nil {
			yield(nil, fmt.Errorf("creating request: %w", err))
			return
		}

		for _, device := range devices.Devices {
			if !yield(&device, nil) {
				return
			}
		}
	}
}

func (c *Client) GetInstallations(ctx context.Context, deviceID string) iter.Seq2[*Installation, error] {
	return func(yield func(*Installation, error) bool) {
		u := path.Join("devices", deviceID, "installations")
		var installations Installations

		_, err := c.makeRequest(ctx, http.MethodGet, u, nil, &installations)
		if err != nil {
			yield(nil, fmt.Errorf("creating request: %w", err))
			return
		}

		for _, installation := range installations.Installations {
			if !yield(&installation, nil) {
				return
			}
		}
	}
}

type PushOptions struct {
	InstallationID string
	Background     bool
}

var ErrNoInstallationID = errors.New("background push requires an installation ID")

func (c *Client) Push(ctx context.Context, deviceID string, image io.Reader, opts *PushOptions) error {
	if opts == nil {
		opts = &PushOptions{}
	}

	if opts.Background && opts.InstallationID == "" {
		return ErrNoInstallationID
	}

	var buf strings.Builder

	if f, ok := image.(*os.File); ok {
		if stat, err := f.Stat(); err == nil {
			size := base64.StdEncoding.EncodedLen(int(stat.Size()))
			buf.Grow(size)
		}
	}

	b64enc := base64.NewEncoder(base64.StdEncoding, &buf)
	defer func() { _ = b64enc.Close() }()

	if _, err := io.Copy(b64enc, image); err != nil {
		return fmt.Errorf("encoding image: %w", err)
	}

	if err := b64enc.Close(); err != nil {
		return fmt.Errorf("closing encoder: %w", err)
	}

	payload, err := json.Marshal(
		PushPayload{
			DeviceID:       deviceID,
			Image:          buf.String(),
			InstallationID: opts.InstallationID,
			Background:     opts.Background,
		},
	)
	if err != nil {
		return fmt.Errorf("marshalling request json: %w", err)
	}

	u := path.Join("devices", deviceID, "push")

	_, err = c.makeRequest(ctx, http.MethodPost, u, bytes.NewReader(payload), nil)
	return err
}

func (c *Client) Delete(ctx context.Context, deviceID string, installationID string) error {
	u := path.Join("devices", deviceID, "installations", installationID)
	_, err := c.makeRequest(ctx, http.MethodDelete, u, nil, nil)
	return err
}
