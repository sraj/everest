import { WebStorageStateStore } from 'oidc-client-ts'

const issuer = import.meta.env.VITE_ZITADEL_ISSUER || 'http://localhost:8082'
const clientId = import.meta.env.VITE_ZITADEL_CLIENT_ID || ''
const redirectUri = import.meta.env.VITE_ZITADEL_REDIRECT_URI || 'http://localhost:5173/auth/callback'
const postLogoutUri = import.meta.env.VITE_ZITADEL_POST_LOGOUT_URI || 'http://localhost:5173'

export const authConfig = {
  authority: issuer,
  client_id: clientId,
  redirect_uri: redirectUri,
  post_logout_redirect_uri: postLogoutUri,
  response_type: 'code',
  scope: 'openid profile email',
  automaticSilentRenew: true,
  loadUserInfo: true,
  userStore: new WebStorageStateStore({ store: window.localStorage }),
}
