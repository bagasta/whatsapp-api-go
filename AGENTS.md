# Repository Guidelines

## Project Structure & Module Organization
- `src/` holds the application: `cmd/` (CLI entrypoints `rest` and `mcp`), `main.go`, and `openapi.yaml` for the HTTP API contract. `domains/` defines core entities, `usecase/` orchestrates business logic, `infrastructure/` wires adapters (Fiber HTTP server, WhatsApp client, storage), and `pkg/` provides shared helpers (errors, metrics, utils). Validation rules live in `validations/`; WhatsApp protocol helpers sit in `whatsapp/`; static assets/templates live under `statics/`, `views/`, and `ui/`. Local data persists to `storages/` (SQLite by default).
- Tests (`*_test.go`) sit next to code, notably in `pkg/utils/`, `validations/`, `usecase/`, and `infrastructure/whatsapp/`. Documentation and API references are in `docs/`, `API_REFERENCE.md`, and `REST_API_GUIDE.md`. Container assets are under `docker/` and `docker-compose.yml`; smoke-test scripts live in `scripts/` plus `test_api.sh`.

## Build, Test, and Development Commands
- Work from `src/`.
- Format and vet: `go fmt ./...` then `go vet ./...`.
- Unit tests: `go test ./...` (uses `testify`). Add `-run TestName` to scope locally.
- Run locally: `go run . rest` (REST API on :3000) or `go run . mcp` (MCP server on :8080).
- Build binary: `go build -o ../bin/whatsapp` (or `whatsapp.exe` on Windows), then `../bin/whatsapp rest`.
- Docker: `docker-compose up -d --build` from repo root.
- Smoke tests: with the REST server running, `BASE_URL=http://localhost:3000 PHONE=62xxx ./scripts/rest_autotest.sh` or `./test_api.sh` to exercise QR/session and send flows.

## Coding Style & Naming Conventions
- Go 1.24; always run `go fmt` before committing. Keep packages lower_snake, files lowercase, exported types/functions in PascalCase, and unexported identifiers in camelCase.
- Prefer context-aware, explicit errors wrapped with details; reuse helpers in `pkg/utils/` and error types in `pkg/error`.
- Keep HTTP handlers thin (routing/middleware in `infrastructure/fiber`), pushing logic into `usecase/` and validations into `validations/`.

## Testing Guidelines
- Add table-driven tests in `*_test.go` next to implementations; name cases with `TestSubject_Scenario`. Use `testify` assertions for clarity.
- For changes touching request validation or message sending, cover both success and common failure paths. If touching HTTP handlers, consider a short integration check with `rest_autotest.sh` against a running instance.

## Commit & Pull Request Guidelines
- Follow the existing short, imperative commit style (`fix connected rest`, `rest fix can send message to AI`). Keep commits scoped and descriptive.
- PRs should include: what changed and why, affected endpoints/flags, any config migrations, and screenshots/log snippets for UI or QR/session flows. List commands/tests executed (`go test ./...`, smoke scripts).
- Link relevant issues or user reports when available and note any deployment considerations (e.g., FFmpeg requirement, new env vars in `src/.env.example`).

## Security & Configuration Tips
- Use `src/.env.example` as a base for local config; never commit secrets, tokens, or QR outputs. Keep webhook secrets/API keys in env vars.
- Default storage is SQLite under `storages/`; be mindful of data-at-rest when sharing artifacts. FFmpeg is required for media handling—install it locally or run via Docker if missing.
