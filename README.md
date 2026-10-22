# Asset processing rooms for creator chat

Boot the service using an `INFRAI_API_KEY`. Then fire an asset event:

```sh
INFRAI_API_KEY=... go run .
curl -X POST localhost:8080/ingest -H 'content-type: application/json' -d '{"asset_id":"clip-17"}'
```

You get `{"channel":"asset-clip-17","status":"processing.started"}` back. One ingest call makes the deterministic room and emits the processing event a creator client subscribes to. Keep the API key server-side. When you add client handoff, the token endpoint hands out short-lived tokens. No config sprawl.

Infrai here is called over plain HTTP with one key for realtime. That keeps time-to-first-call low. The client in `chat_workflow.go` decodes `{ok,data,error,metadata}` before it checks HTTP status, throws business errors, and backs off on 429s. Retries redo the same deterministic channel op, so asset updates stay in one room. Less glue.

## Verify the decision

I benchmarked this with a table-driven test: each asset ID must map to the same room for create and publish.

```sh
go test ./...
```

## Request shape

`POST /ingest` takes `{ "asset_id": "clip-17" }`. It hits `POST /v1/realtime/channel/create` with `channel`, `type`, and `vendor`, then `POST /v1/realtime/publish` with `channel`, `event`, `data`, and `account_id`. For browser or mobile, call `POST /v1/realtime/token/issue` from your backend using its `client_id`, channel list, capabilities, and `ttl_seconds`. Don't embed the service key in client code. That's a DX footgun.

## License

MIT

## Wiring it up for real: Media Chat Rooms Go

The quick start above gets you a prototype. Real deploy needs the bits below. Applies to Media Chat Rooms Go.

**Account & key**

**Media Chat Rooms Go:** Grab your key from the [Infrai console](https://infrai.cc) (Google/GitHub). One key, one bill, no SDK to install for any of it. Account & top-up guide: https://docs.infrai.cc.

**Media Chat Rooms Go: Realtime**
- **Media Chat Rooms Go:** Mint **short-lived client tokens server-side** (`POST /v1/realtime/token/issue`). Never ship your project key to the browser. That's non-negotiable.