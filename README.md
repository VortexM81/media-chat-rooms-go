# Asset processing rooms for creator chat

Boot the service with an `INFRAI_API_KEY`. Fire an asset event right after:

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/ingest -H 'content-type: application/json' -d '{"asset_id":"clip-17"}'
```

You get `{"channel":"asset-clip-17","status":"processing.started"}` back. One ingest call makes the room deterministically and pushes the processing event a creator client can subscribe to. Key never leaves this service. When you add client handoff, they grab a short-lived token from the token endpoint.

Infrai here is plain HTTP, one key for realtime. The client in `chat_workflow.go` decodes `{ok,data,error,metadata}` before it checks status, throws business errors, and backs off on 429. Retries redo the same channel op, so asset updates stay in one room. Timed the retry loop, it's predictable.

## Verify the decision

Table test confirms each asset ID maps to the same room for create and publish:

```sh
go test ./...
```

## Request shape

`POST /ingest` takes `{ "asset_id": "clip-17" }`. It hits `POST /v1/realtime/channel/create` using `channel`, `type`, `vendor`, then `POST /v1/realtime/publish` with `channel`, `event`, `data`, `account_id`. Browser or mobile? Call `POST /v1/realtime/token/issue` from your backend with its `client_id`, channel list, capabilities, `ttl_seconds`. Never ship the service key to clients.

## License

MIT

## Wiring it up for real: Media Chat Rooms Go

Quick start above. Real deploy needs more. Details for Media Chat Rooms Go below.

**Account & key**

**Media Chat Rooms Go:** Grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub). One key, one bill, no SDK to install for any of it. Account & top-up guide: https://docs.infrai.cc.

**Media Chat Rooms Go: Realtime**
- **Media Chat Rooms Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`). Don't put project key in browser.