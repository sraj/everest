# Zitadel Authentication

## Architecture

```
Browser → Zitadel Login (8082) → OIDC callback → SPA (5173) → API (8080)
                                                       ↓ Bearer id_token
                                              Zitadel Verifier (JWKS validation)
```

- **Frontend**: `oidc-client-ts` handles OIDC code flow (PKCE), stores user in localStorage, silent token renewal
- **Backend**: JWT validation via Zitadel's JWKS (no introspection — public client, no secret)
- **Token**: `user.id_token` sent as Bearer (access_token is opaque, id_token is JWT with `sub`, `email`, `name`)

## Key Files

| File | Purpose |
|---|---|
| `web/src/auth/AuthProvider.tsx` | React context: `login()`, `logout()`, `getAccessToken()` → `user.id_token` |
| `web/src/auth/authConfig.ts` | OIDC client config (issuer, client_id, scopes, localStorage store) |
| `web/src/pages/Home.tsx` | Calls `fetchDocuments(token)` when authenticated, logout button, login redirect |
| `web/src/pages/DocumentEditor.tsx` | Passes token in save/load API calls |
| `web/src/store/documentSlice.ts` | Redux thunks accept `accessToken?` param for Authorization header |
| `internal/infrastructure/zitadel/auth.go` | JWKS-based JWT verifier, Fiber middleware, `Optional()` variant |
| `internal/config/config.go` | `ZITADEL_ISSUER`, `ZITADEL_CLIENT_ID` env vars |
| `cmd/server/main.go` | Initializes verifier, enables middleware when client ID is set |
| `docker-compose.yml` | Zitadel services under `profiles: [zitadel]` |

## Zitadel Console Configuration

After first login to `http://localhost:8082/ui/console` as `zitadel-admin@zitadel.localhost`:

1. **Create a Project**: Instances → your instance → **Projects** → New
2. **Create an Application** in that project:
   - **Type**: Web (Zitadel v4 requires Web for HTTP redirects with code grant; User Agent enforces HTTPS)
   - **Auth Method**: None (PKCE will be used)
3. **Configure OIDC**:
   - **Redirect URIs**: `http://localhost:5173/auth/callback`
   - **Post Logout URIs**: `http://localhost:5173`
   - **Response Types**: Code (`OIDC_RESPONSE_TYPE_CODE`)
   - **Grant Types**: Authorization Code (`OIDC_GRANT_TYPE_AUTHORIZATION_CODE`), Refresh Token
   - **Token Type**: JWT
4. **Copy the Client ID** into `.env`:
   ```bash
   ZITADEL_CLIENT_ID=<your-client-id>
   VITE_ZITADEL_CLIENT_ID=<your-client-id>
   ```

| Setting | Value |
|---|---|
| App Type | User Agent |
| Auth Method | None |
| Redirect URI | `http://localhost:5173/auth/callback` |
| Post Logout URI | `http://localhost:5173` |
| Response Type | Code |
| Grant Types | Authorization Code, Refresh Token |
| Token Type | JWT |

## Disabling Auth

Set `ZITADEL_CLIENT_ID=""` in `.env` — the middleware won't activate, all requests pass through.

## Env Vars

| Variable | Default | Notes |
|---|---|---|
| `ZITADEL_ISSUER` | `http://localhost:8082` | Must match Zitadel's external domain |
| `ZITADEL_CLIENT_ID` | `` | Empty disables auth |
| `VITE_ZITADEL_ISSUER` | `http://localhost:8082` | Frontend OIDC issuer |
| `VITE_ZITADEL_CLIENT_ID` | `` | Frontend client ID |
| `VITE_ZITADEL_REDIRECT_URI` | `http://localhost:5173/auth/callback` | Post-login callback |
| `ZITADEL_MASTERKEY` | `MasterkeyNeedsToHave32Characters` | Zitadel bootstrap |

## Token Validation Flow

1. Frontend sends `Authorization: Bearer <id_token>` on API calls
2. Backend middleware extracts token, validates JWT signature via Zitadel JWKS
3. Validates issuer, algorithm (RS256/384/512), expiry (30s leeway)
4. Sets `c.Locals("user", IntrospectUser{Sub, Email, Name})`
5. Handler extracts `owner_id` from user Sub for document ownership

## Database Columns

- `documents.owner_id` → `VARCHAR(255)` (Zitadel user IDs are not UUIDs)
- `documents.thumbnail_id` → `VARCHAR(255)` (MinIO object keys are UUID strings)
