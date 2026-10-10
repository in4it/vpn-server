
type OIDCProvider = {
    id: string;
    clientId: string;
    // write-only: sent when creating a provider, never returned by GET /api/oidc
    clientSecret?: string;
    scope: string;
    discoveryURI: string;
    redirectURI: string;
    loginURL: string;
    name: string;
  };
