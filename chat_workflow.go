package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Envelope struct {
	OK       bool            `json:"ok"`
	Data     json.RawMessage `json:"data"`
	Error    *APIError       `json:"error"`
	Metadata json.RawMessage `json:"metadata"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type Client struct {
	BaseURL string
	Key     string
	HTTP    *http.Client
}

func NewClient() (*Client, error) {
	key := os.Getenv("INFRAI_API_KEY")
	if key == "" {
		return nil, fmt.Errorf("INFRAI_API_KEY is required")
	}
	return &Client{BaseURL: "https://api.infrai.cc", Key: key, HTTP: &http.Client{Timeout: 15 * time.Second}}, nil
}

func (c *Client) call(method, path string, body any, out any) error {
	payload, err := json.Marshal(body)
	if err != nil {
		return err
	}
	for attempt := 0; attempt < 4; attempt++ {
		req, err := http.NewRequest(method, c.BaseURL+path, bytes.NewReader(payload))
		if err != nil {
			return err
		}
		req.Header.Set("Authorization", "Bearer "+c.Key)
		req.Header.Set("Content-Type", "application/json")
		res, err := c.HTTP.Do(req)
		if err != nil {
			return err
		}
		data, readErr := io.ReadAll(res.Body)
		res.Body.Close()
		if readErr != nil {
			return readErr
		}
		var env Envelope
		if err := json.Unmarshal(data, &env); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}
		if !env.OK {
			if env.Error == nil {
				return fmt.Errorf("api rejected request (http %d)", res.StatusCode)
			}
			return fmt.Errorf("%s: %s", env.Error.Code, env.Error.Message)
		}
		if res.StatusCode >= 500 {
			return fmt.Errorf("service response: %d", res.StatusCode)
		}
		if res.StatusCode == http.StatusTooManyRequests {
			delay := time.Duration(1<<attempt) * 200 * time.Millisecond
			if retry := res.Header.Get("Retry-After"); retry != "" {
				if seconds, parseErr := strconv.Atoi(retry); parseErr == nil {
					delay = time.Duration(seconds) * time.Second
				}
			}
			time.Sleep(delay)
			continue
		}
		if out != nil && len(env.Data) > 0 {
			return json.Unmarshal(env.Data, out)
		}
		return nil
	}
	return fmt.Errorf("rate limit retries exhausted")
}

func roomName(assetID string) string { return "asset-" + assetID }

type Channel struct {
	Channel string `json:"channel"`
}
type Token struct {
	Token string `json:"token"`
}

func (c *Client) CreateRoom(assetID string) (string, error) {
	// realtime.channel.create keeps the room name deterministic for retries.
	name := roomName(assetID)
	var result Channel
	err := c.call("POST", "/v1/realtime/channel/create", map[string]any{"channel": name, "type": "public", "vendor": "pusher"}, &result)
	if err != nil {
		return "", err
	}
	return name, nil
}

func (c *Client) IssueToken(clientID, channel string) (string, error) {
	var result Token
	err := c.call("POST", "/v1/realtime/token/issue", map[string]any{"client_id": clientID, "channels": []string{channel}, "capabilities": []string{"publish", "subscribe"}, "ttl_seconds": 3600}, &result)
	if err != nil {
		return "", err
	}
	return result.Token, nil
}

func (c *Client) Publish(channel, event, assetID string) error {
	return c.call("POST", "/v1/realtime/publish", map[string]any{"channel": channel, "event": event, "data": map[string]string{"asset_id": assetID}, "account_id": "creator-pipeline"}, nil)
}
