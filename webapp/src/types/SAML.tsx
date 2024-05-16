
type SAMLProvider = {
    id: string;
    name: string;
    audience: string;
    issuer: string
    acs: string;
    metadataURL: string;
    allowMissingAttributes: boolean;
  };
