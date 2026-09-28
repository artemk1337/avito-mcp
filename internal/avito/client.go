package avito

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

const BaseURL = "https://api.avito.ru"

const maxResponseBytes = 4 << 20

type Credentials struct {
	AccessToken  string
	ClientID     string
	ClientSecret string
}

type Client struct {
	baseURL    string
	httpClient *http.Client
	creds      Credentials
	mu         sync.Mutex
	token      string
	expiresAt  time.Time
}

func NewClient(baseURL string, httpClient *http.Client, creds Credentials) (*Client, error) {
	if creds.AccessToken == "" && (creds.ClientID == "" || creds.ClientSecret == "") {
		return nil, errors.New("set AVITO_ACCESS_TOKEN or both AVITO_CLIENT_ID and AVITO_CLIENT_SECRET")
	}
	parsed, err := url.Parse(baseURL)
	if err != nil || parsed.Scheme == "" || parsed.Host == "" {
		return nil, errors.New("invalid Avito API base URL")
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 20 * time.Second}
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), httpClient: httpClient, creds: creds}, nil
}

func (c *Client) Do(ctx context.Context, method, path string, query url.Values, body any) (json.RawMessage, error) {
	token, err := c.accessToken(ctx)
	if err != nil {
		return nil, err
	}
	var requestBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("encode request: %w", err)
		}
		requestBody = bytes.NewReader(data)
	}
	endpoint := c.baseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, method, endpoint, requestBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return c.send(req)
}

func (c *Client) accessToken(ctx context.Context) (string, error) {
	if c.creds.AccessToken != "" {
		return c.creds.AccessToken, nil
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token != "" && time.Now().Before(c.expiresAt) {
		return c.token, nil
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {c.creds.ClientID},
		"client_secret": {c.creds.ClientSecret},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/token", strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	data, err := c.send(req)
	if err != nil {
		return "", fmt.Errorf("get Avito token: %w", err)
	}
	var response struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int64  `json:"expires_in"`
	}
	if err := json.Unmarshal(data, &response); err != nil || response.AccessToken == "" || response.ExpiresIn <= 0 {
		return "", errors.New("invalid Avito token response")
	}
	c.token = response.AccessToken
	c.expiresAt = time.Now().Add(time.Duration(response.ExpiresIn) * time.Second).Add(-time.Minute)
	return c.token, nil
}

func (c *Client) send(req *http.Request) (json.RawMessage, error) {
	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Avito request: %w", err)
	}
	defer res.Body.Close()
	data, err := io.ReadAll(io.LimitReader(res.Body, maxResponseBytes+1))
	if err != nil {
		return nil, fmt.Errorf("read Avito response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return nil, errors.New("Avito response exceeds 4 MiB")
	}
	if res.StatusCode < 200 || res.StatusCode >= 300 {
		if req.URL.Path == "/token" {
			return nil, fmt.Errorf("Avito token endpoint returned HTTP %d", res.StatusCode)
		}
		return nil, fmt.Errorf("Avito API returned HTTP %d: %s", res.StatusCode, strings.TrimSpace(string(data)))
	}
	if len(data) == 0 {
		return json.RawMessage(`{}`), nil
	}
	if !json.Valid(data) {
		return nil, errors.New("Avito API returned invalid JSON")
	}
	return json.RawMessage(data), nil
}
