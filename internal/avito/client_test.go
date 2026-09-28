package avito

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestClientCredentialsTokenIsReused(t *testing.T) {
	tokenRequests := 0
	apiRequests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			tokenRequests++
			if r.Method != http.MethodPost || r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
				t.Errorf("unexpected token request: %s %s", r.Method, r.Header.Get("Content-Type"))
			}
			if err := r.ParseForm(); err != nil {
				t.Fatal(err)
			}
			if r.Form.Get("grant_type") != "client_credentials" || r.Form.Get("client_id") != "id" || r.Form.Get("client_secret") != "secret" {
				t.Errorf("wrong token form: %v", r.Form)
			}
			io.WriteString(w, `{"access_token":"issued-token","expires_in":3600}`)
		case "/core/v1/items":
			apiRequests++
			if r.Header.Get("Authorization") != "Bearer issued-token" || r.URL.Query().Get("page") != "2" {
				t.Errorf("wrong API request: %s %s", r.Header.Get("Authorization"), r.URL.RawQuery)
			}
			io.WriteString(w, `{"resources":[]}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()
	client, err := NewClient(srv.URL, srv.Client(), Credentials{ClientID: "id", ClientSecret: "secret"})
	if err != nil {
		t.Fatal(err)
	}
	for range 2 {
		data, err := client.Do(context.Background(), http.MethodGet, "/core/v1/items", map[string][]string{"page": {"2"}}, nil)
		if err != nil || !json.Valid(data) {
			t.Fatalf("Do: %s, %v", data, err)
		}
	}
	if tokenRequests != 1 || apiRequests != 2 {
		t.Fatalf("requests: token=%d API=%d", tokenRequests, apiRequests)
	}
}

func TestStaticTokenAndAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/token" {
			t.Error("static token must skip OAuth")
		}
		if r.Header.Get("Authorization") != "Bearer static-token" {
			t.Error("wrong token")
		}
		w.WriteHeader(http.StatusForbidden)
		io.WriteString(w, `{"error":"insufficient scope"}`)
	}))
	defer srv.Close()
	client, err := NewClient(srv.URL, srv.Client(), Credentials{AccessToken: "static-token"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), http.MethodGet, "/core/v1/accounts/self", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "HTTP 403") || !strings.Contains(err.Error(), "insufficient scope") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMissingCredentials(t *testing.T) {
	if _, err := NewClient(BaseURL, nil, Credentials{ClientID: "id"}); err == nil {
		t.Fatal("expected missing credentials error")
	}
}

func TestTokenErrorDoesNotExposeSecret(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"error":"bad secret: private-value"}`)
	}))
	defer srv.Close()
	client, err := NewClient(srv.URL, srv.Client(), Credentials{ClientID: "id", ClientSecret: "private-value"})
	if err != nil {
		t.Fatal(err)
	}
	_, err = client.Do(context.Background(), http.MethodGet, "/core/v1/items", nil, nil)
	if err == nil || !strings.Contains(err.Error(), "HTTP 401") || strings.Contains(err.Error(), "private-value") {
		t.Fatalf("unexpected token error: %v", err)
	}
}
