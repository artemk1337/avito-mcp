package server

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/artemk1337/avito-mcp/internal/avito"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Output struct {
	Data any `json:"data"`
}

type EmptyInput struct{}

type ListItemsInput struct {
	Page          int    `json:"page,omitempty" jsonschema:"Page number, starting at 1"`
	PerPage       int    `json:"per_page,omitempty" jsonschema:"Items per page, from 1 to 99"`
	Status        string `json:"status,omitempty" jsonschema:"Listing status: active, removed, old, blocked, or rejected"`
	UpdatedAtFrom string `json:"updated_at_from,omitempty" jsonschema:"Update date from, YYYY-MM-DD"`
	Category      int64  `json:"category,omitempty" jsonschema:"Avito category ID"`
}

type GetItemInput struct {
	UserID int64 `json:"user_id" jsonschema:"Avito account ID"`
	ItemID int64 `json:"item_id" jsonschema:"Avito listing ID"`
}

type UpdatePriceInput struct {
	ItemID int64 `json:"item_id" jsonschema:"Avito listing ID"`
	Price  int64 `json:"price" jsonschema:"New price in rubles"`
}

type ListChatsInput struct {
	UserID     int64  `json:"user_id" jsonschema:"Avito account ID"`
	ItemIDs    string `json:"item_ids,omitempty" jsonschema:"Comma-separated listing IDs"`
	UnreadOnly bool   `json:"unread_only,omitempty" jsonschema:"Return only unread chats"`
	ChatTypes  string `json:"chat_types,omitempty" jsonschema:"Comma-separated chat types: u2i, u2u, a2u"`
	Limit      int    `json:"limit,omitempty" jsonschema:"Page size, from 1 to 99"`
	Offset     int    `json:"offset,omitempty" jsonschema:"Zero-based offset"`
}

type GetChatInput struct {
	UserID int64  `json:"user_id" jsonschema:"Avito account ID"`
	ChatID string `json:"chat_id" jsonschema:"Avito chat ID"`
}

type ListMessagesInput struct {
	UserID int64  `json:"user_id" jsonschema:"Avito account ID"`
	ChatID string `json:"chat_id" jsonschema:"Avito chat ID"`
	Limit  int    `json:"limit,omitempty" jsonschema:"Page size, from 1 to 99"`
	Offset int    `json:"offset,omitempty" jsonschema:"Zero-based offset"`
}

type SendMessageInput struct {
	UserID int64  `json:"user_id" jsonschema:"Avito account ID"`
	ChatID string `json:"chat_id" jsonschema:"Avito chat ID"`
	Text   string `json:"text" jsonschema:"Message text, at most 1000 characters"`
}

func New(client *avito.Client) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "avito-mcp", Version: "0.1.0"}, nil)
	mcp.AddTool(s, &mcp.Tool{Name: "get_profile", Description: "Get the authenticated Avito account and its user ID."}, func(ctx context.Context, _ *mcp.CallToolRequest, _ EmptyInput) (*mcp.CallToolResult, Output, error) {
		return call(ctx, client, http.MethodGet, "/core/v1/accounts/self", nil, nil)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "list_items", Description: "List the authenticated account's Avito listings. Avito limits this endpoint to 25 requests per minute; employee-owned listings are excluded."}, func(ctx context.Context, _ *mcp.CallToolRequest, in ListItemsInput) (*mcp.CallToolResult, Output, error) {
		if in.Page < 0 || in.PerPage < 0 || in.PerPage >= 100 || in.Category < 0 {
			return nil, Output{}, errors.New("invalid pagination or category")
		}
		q := url.Values{}
		setInt(q, "page", in.Page)
		setInt(q, "per_page", in.PerPage)
		setString(q, "status", in.Status)
		setString(q, "updatedAtFrom", in.UpdatedAtFrom)
		if in.Category > 0 {
			q.Set("category", strconv.FormatInt(in.Category, 10))
		}
		return call(ctx, client, http.MethodGet, "/core/v1/items", q, nil)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "get_item", Description: "Get one Avito listing by account and listing ID."}, func(ctx context.Context, _ *mcp.CallToolRequest, in GetItemInput) (*mcp.CallToolResult, Output, error) {
		if in.UserID <= 0 || in.ItemID <= 0 {
			return nil, Output{}, errors.New("user_id and item_id must be positive")
		}
		path := fmt.Sprintf("/core/v1/accounts/%d/items/%d/", in.UserID, in.ItemID)
		return call(ctx, client, http.MethodGet, path, nil, nil)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "update_item_price", Description: "Change the price of an Avito listing. This modifies the live listing."}, func(ctx context.Context, _ *mcp.CallToolRequest, in UpdatePriceInput) (*mcp.CallToolResult, Output, error) {
		if in.ItemID <= 0 || in.Price <= 0 {
			return nil, Output{}, errors.New("item_id and price must be positive")
		}
		path := fmt.Sprintf("/core/v1/items/%d/update_price", in.ItemID)
		return call(ctx, client, http.MethodPost, path, nil, map[string]int64{"price": in.Price})
	})
	mcp.AddTool(s, &mcp.Tool{Name: "list_chats", Description: "List Avito Messenger chats for an account."}, func(ctx context.Context, _ *mcp.CallToolRequest, in ListChatsInput) (*mcp.CallToolResult, Output, error) {
		if in.UserID <= 0 || invalidPage(in.Limit, in.Offset) {
			return nil, Output{}, errors.New("invalid user_id or pagination")
		}
		q := url.Values{}
		setString(q, "item_ids", in.ItemIDs)
		setString(q, "chat_types", in.ChatTypes)
		setInt(q, "limit", in.Limit)
		if in.Offset > 0 {
			q.Set("offset", strconv.Itoa(in.Offset))
		}
		if in.UnreadOnly {
			q.Set("unread_only", "true")
		}
		return call(ctx, client, http.MethodGet, fmt.Sprintf("/messenger/v2/accounts/%d/chats", in.UserID), q, nil)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "get_chat", Description: "Get one Avito Messenger chat."}, func(ctx context.Context, _ *mcp.CallToolRequest, in GetChatInput) (*mcp.CallToolResult, Output, error) {
		path, err := chatPath("v2", in.UserID, in.ChatID)
		if err != nil {
			return nil, Output{}, err
		}
		return call(ctx, client, http.MethodGet, path, nil, nil)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "list_messages", Description: "List messages in an Avito chat without marking it read."}, func(ctx context.Context, _ *mcp.CallToolRequest, in ListMessagesInput) (*mcp.CallToolResult, Output, error) {
		path, err := chatPath("v3", in.UserID, in.ChatID)
		if err != nil || invalidPage(in.Limit, in.Offset) {
			return nil, Output{}, errors.New("invalid user_id, chat_id, or pagination")
		}
		q := url.Values{}
		setInt(q, "limit", in.Limit)
		if in.Offset > 0 {
			q.Set("offset", strconv.Itoa(in.Offset))
		}
		return call(ctx, client, http.MethodGet, path+"/messages/", q, nil)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "send_message", Description: "Send a text message in an Avito chat. This contacts the other participant."}, func(ctx context.Context, _ *mcp.CallToolRequest, in SendMessageInput) (*mcp.CallToolResult, Output, error) {
		path, err := chatPath("v1", in.UserID, in.ChatID)
		if err != nil {
			return nil, Output{}, err
		}
		if strings.TrimSpace(in.Text) == "" || len([]rune(in.Text)) > 1000 {
			return nil, Output{}, errors.New("text must contain 1 to 1000 characters")
		}
		body := map[string]any{"type": "text", "message": map[string]string{"text": in.Text}}
		return call(ctx, client, http.MethodPost, path+"/messages", nil, body)
	})
	mcp.AddTool(s, &mcp.Tool{Name: "mark_chat_read", Description: "Mark an Avito chat as read."}, func(ctx context.Context, _ *mcp.CallToolRequest, in GetChatInput) (*mcp.CallToolResult, Output, error) {
		path, err := chatPath("v1", in.UserID, in.ChatID)
		if err != nil {
			return nil, Output{}, err
		}
		return call(ctx, client, http.MethodPost, path+"/read", nil, nil)
	})
	return s
}

func call(ctx context.Context, client *avito.Client, method, path string, query url.Values, body any) (*mcp.CallToolResult, Output, error) {
	data, err := client.Do(ctx, method, path, query, body)
	if err != nil {
		return nil, Output{}, err
	}
	var result any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&result); err != nil {
		return nil, Output{}, err
	}
	return nil, Output{Data: result}, nil
}

func chatPath(version string, userID int64, chatID string) (string, error) {
	if userID <= 0 || chatID == "" || strings.ContainsAny(chatID, "/?#") {
		return "", errors.New("invalid user_id or chat_id")
	}
	return fmt.Sprintf("/messenger/%s/accounts/%d/chats/%s", version, userID, url.PathEscape(chatID)), nil
}

func invalidPage(limit, offset int) bool {
	return limit < 0 || limit >= 100 || offset < 0
}

func setInt(q url.Values, key string, value int) {
	if value > 0 {
		q.Set(key, strconv.Itoa(value))
	}
}

func setString(q url.Values, key, value string) {
	if value != "" {
		q.Set(key, value)
	}
}
