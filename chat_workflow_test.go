package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRoomNameKeepsAssetUpdatesTogether(t *testing.T) {
	for _, tc := range []struct{ asset, want string }{{"clip-17", "asset-clip-17"}, {"episode-a", "asset-episode-a"}} {
		if got := roomName(tc.asset); got != tc.want {
			t.Fatalf("roomName(%q) = %q, want %q", tc.asset, got, tc.want)
		}
	}
}

func TestCreateRoomUsesPublicChannelType(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/realtime/channel/create" {
			t.Errorf("path = %q", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Errorf("decode request: %v", err)
		}
		if body["type"] != "public" || body["channel"] != "asset-clip-17" {
			t.Errorf("create body = %#v", body)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":true,"data":{"channel":"asset-clip-17"}}`))
	}))
	defer server.Close()
	c := &Client{BaseURL: server.URL, HTTP: server.Client()}
	if name, err := c.CreateRoom("clip-17"); err != nil || name != "asset-clip-17" {
		t.Fatalf("CreateRoom = %q, %v", name, err)
	}
}
