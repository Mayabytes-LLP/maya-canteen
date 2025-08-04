---
applyTo: '**/*.go'
---

# WhatsApp Integration Instructions for maya-canteen (2025)

_For AI coding agents working with WhatsApp/whatsmeow in this codebase._

## Architecture & Data Flow

- WhatsApp integration is handled in `internal/handlers/whatsapp_handler.go` using the [whatsmeow](https://pkg.go.dev/go.mau.fi/whatsmeow) Go library.
- WhatsApp client lifecycle and QR login are managed via the WebSocket handler (`internal/handlers/websocket_handler.go`).
- WhatsApp notification API endpoints are registered in `internal/server/routes/whatsapp_routes.go` under `/api/whatsapp`.
- All WhatsApp messages (balance, transactions, CSVs) are sent via the `WhatsAppHandler` abstraction, which wraps whatsmeow's `Client`.
- WhatsApp client instance is injected via a getter function, not a global variable. See `NewWhatsAppHandler` usage.

## Key Workflows

- **Send balance notification to a user:**
	- POST `/api/whatsapp/notify/{id}` with JSON body `{ "message_template": ..., "month": ..., "year": ..., "include_transactions": ... }`.
	- Handler fetches user, balance, and optionally transaction history, then sends WhatsApp message(s).
- **Send notifications to all users:**
	- POST `/api/whatsapp/notify-all` with similar JSON body.
	- Iterates all users with balances, sending messages with a delay to avoid rate limits.
- **WebSocket QR login:**
	- WebSocket endpoint manages QR code login and connection status for WhatsApp client.
	- QR channel must be created _before_ calling `Client.Connect()` (see whatsmeow docs).

## Project-Specific Patterns

- **Phone number validation:** Always use `IsOnWhatsApp` before sending; errors are logged and surfaced to API clients.
- **Message sending:** Use `SendWhatsAppMessage` for text, `SendDocumentMessage` for CSVs. Both wrap whatsmeow's `SendMessage` and `Upload` APIs.
- **Rate limiting:** 300ms delay between bulk messages to avoid WhatsApp bans.
- **Error handling:** All errors are logged with context; API returns structured JSON errors.
- **Dependency injection:** WhatsApp client is _not_ a global; always inject via handler constructors for testability.

## Whatsmeow Usage Notes

- Always check `client.IsLoggedIn()` and `client.IsConnected()` before sending.
- Use `GetQRChannel(ctx)` before `Connect()` for QR login flows.
- Use `IsOnWhatsApp([]string{phone})` to validate recipients.
- For media/CSV, upload with `client.Upload(ctx, data, whatsmeow.MediaDocument)` and send as a document message.
- See [whatsmeow GoDoc](https://pkg.go.dev/go.mau.fi/whatsmeow) for all available methods and error types.

## Examples

- See `internal/handlers/whatsapp_handler.go` for:
	- `SendWhatsAppMessage`, `SendDocumentMessage`, `NotifyUserBalance`, `NotifyAllUsersBalances`
- See `internal/server/routes/whatsapp_routes.go` for route registration.
- See `internal/handlers/websocket_handler.go` for QR login and client management.

## Troubleshooting

- If WhatsApp client is not connected, API returns 500 with a clear error.
- For QR login, ensure QR channel is created _before_ connecting.
- For rate limits or bans, increase delay or check logs for WhatsApp errors.

---
For more, see [whatsmeow documentation](https://pkg.go.dev/go.mau.fi/whatsmeow) and project `README.md`.