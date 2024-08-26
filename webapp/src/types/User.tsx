
type User = {
    id: string;
    login: string;
    password: string;
    role: string;
    oidcID: string;
    samlID: string;
    provisioned: boolean;
    suspended: boolean;
    lastTokenRenewal: string;
    connectionsDisabledOnAuthFailure: boolean;
    lastLogin: string;
  };
