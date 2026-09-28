package server

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/artemk1337/avito-mcp/internal/avito"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func TestMCPTools(t *testing.T) {
	requests := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("missing authorization")
		}
		switch r.URL.Path {
		case "/core/v1/items":
			if r.Method != http.MethodGet || r.URL.Query().Get("page") != "2" || r.URL.Query().Get("status") != "active" {
				t.Errorf("wrong list request: %s %s", r.Method, r.URL.String())
			}
			io.WriteString(w, `{"resources":[{"id":123}]}`)
		case "/messenger/v1/accounts/42/chats/chat-1/messages":
			var body struct {
				Type    string `json:"type"`
				Message struct {
					Text string `json:"text"`
				} `json:"message"`
			}
			if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
				t.Error(err)
			}
			if r.Method != http.MethodPost || body.Type != "text" || body.Message.Text != "Hello" {
				t.Errorf("wrong message request: %s %+v", r.Method, body)
			}
			io.WriteString(w, `{"id":"sent"}`)
		default:
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
	}))
	defer srv.Close()
	api, err := avito.NewClient(srv.URL, srv.Client(), avito.Credentials{AccessToken: "test-token"})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	clientTransport, serverTransport := mcp.NewInMemoryTransports()
	serverSession, err := New(api).Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer serverSession.Close()
	clientSession, err := mcp.NewClient(&mcp.Implementation{Name: "test", Version: "1"}, nil).Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer clientSession.Close()

	listed, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "list_items", Arguments: map[string]any{"page": 2, "status": "active"}})
	if err != nil || listed.IsError {
		t.Fatalf("list_items: result=%+v err=%v", listed, err)
	}
	data := listed.StructuredContent.(map[string]any)["data"].(map[string]any)
	if len(data["resources"].([]any)) != 1 {
		t.Fatalf("unexpected list result: %+v", listed.StructuredContent)
	}

	sent, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "send_message", Arguments: map[string]any{"user_id": 42, "chat_id": "chat-1", "text": "Hello"}})
	if err != nil || sent.IsError {
		t.Fatalf("send_message: result=%+v err=%v", sent, err)
	}
	invalid, err := clientSession.CallTool(ctx, &mcp.CallToolParams{Name: "send_message", Arguments: map[string]any{"user_id": 42, "chat_id": "chat-1", "text": ""}})
	if err != nil || !invalid.IsError {
		t.Fatalf("expected tool error: result=%+v err=%v", invalid, err)
	}
	if requests != 2 {
		t.Fatalf("expected 2 API requests, got %d", requests)
	}
}

func TestCallPreservesLargeIDs(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		io.WriteString(w, `{"id":9007199254740993}`)
	}))
	defer srv.Close()
	api, err := avito.NewClient(srv.URL, srv.Client(), avito.Credentials{AccessToken: "test-token"})
	if err != nil {
		t.Fatal(err)
	}
	_, output, err := call(context.Background(), api, http.MethodGet, "/core/v1/items", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := output.Data.(map[string]any)["id"].(json.Number).String(); got != "9007199254740993" {
		t.Fatalf("ID lost precision: %s", got)
	}
}
