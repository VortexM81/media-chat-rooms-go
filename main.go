package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type ingestRequest struct {
	AssetID string `json:"asset_id"`
}

func main() {
	client, err := NewClient()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/ingest", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var input ingestRequest
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil || input.AssetID == "" {
			http.Error(w, "asset_id is required", http.StatusBadRequest)
			return
		}
		channel, err := client.CreateRoom(input.AssetID)
		if err == nil {
			err = client.Publish(channel, "asset.processing.started", input.AssetID)
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadGateway)
			return
		}
		json.NewEncoder(w).Encode(map[string]string{"channel": channel, "status": "processing.started"})
	})
	log.Println("listening on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
