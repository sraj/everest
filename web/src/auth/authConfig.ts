import { WebStorageStateStore } from 'oidc-client-ts'

const issuer = import.meta.env.VITE_ZITADEL_ISSUER || 'http://localhost:8082'
const clientId = import.meta.env.VITE_ZITADEL_CLIENT_ID || ''
const redirectUri = import.meta.env.VITE_ZITADEL_REDIRECT_URI || 'http://localhost:5173/auth/callback'
const postLogoutUri = import.meta.env.VITE_ZITADEL_POST_LOGOUT_URI || 'http://localhost:5173'

// Token storage: localStorage survives browser restarts and works across tabs.
// In production without a BFF, this is the pragmatic choice — secure it with:
//   - short token lifetimes (set in Zitadel admin: ~5 min access, ~1 hr refresh)
//   - Content-Security-Policy headers on the API server
//   - automaticSilentRenew to refresh before expiry
// Alternative: sessionStorage clears on tab close, but breaks cross-tab auth.
export const authConfig = {
  authority: issuer,
  client_id: clientId,
  redirect_uri: redirectUri,
  post_logout_redirect_uri: postLogoutUri,
  response_type: 'code',
  scope: 'openid profile email',
  automaticSilentRenew: true,
  loadUserInfo: true,
  userStore: new WebStorageStateStore({
    store: window.localStorage,
  }),
}
