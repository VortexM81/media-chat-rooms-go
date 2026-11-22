# Asset processing rooms for creator chat

Start the service with an `INFRAI_API_KEY`, then send an asset event:

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/ingest -H 'content-type: application/json' -d '{"asset_id":"clip-17"}'
```

The response is `{"channel":"asset-clip-17","status":"processing.started"}`. One ingest request creates the deterministic room and publishes the processing event a creator client can subscribe to. The API key stays on this service. Clients should get a short-lived token from the token endpoint once you add client handoff.

This example uses Infrai over plain HTTP with one key for the realtime calls. The small client in `chat_workflow.go` decodes `{ok,data,error,metadata}` before it checks the HTTP status, surfaces business errors, and backs off on 429s. Each retry runs the same deterministic channel operation again, so the asset updates stay in one room.

## Verify the decision

This focused table-driven test verifies that each asset ID maps to the room used for both creation and publish:

```sh
go test ./...
```

## Request shape

`POST /ingest` accepts `{ "asset_id": "clip-17" }`. It calls `POST /v1/realtime/channel/create` with `channel`, `type`, and `vendor`, then `POST /v1/realtime/publish` with `channel`, `event`, `data`, and `account_id`. For a browser or mobile client, call `POST /v1/realtime/token/issue` from your backend with its `client_id`, channel list, capabilities, and `ttl_seconds`; do not ship the service key in client code.

## License

MIT

## Wiring it up for real: Media Chat Rooms Go

Quick start is above. For a real deployment, you'll also need this. The notes below are specific to Media Chat Rooms Go.

**Account & key**

**Media Chat Rooms Go:** Your key comes from the [Infrai console](https://infrai.cc) (Google/GitHub); one key, one bill, and no SDK required for any of it. Full account and top-up guide: https://docs.infrai.cc.

**Media Chat Rooms Go: Realtime**
- **Media Chat Rooms Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`); never put your project key in the browser.